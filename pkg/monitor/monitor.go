package monitor

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
	
	"github.com/kuskoman/url-datadog-monitor/pkg/certcheck"
	"github.com/kuskoman/url-datadog-monitor/pkg/config"
)

const (
	MetricURLUp             = "url.up"
	MetricResponseTime      = "url.response_time_ms"
	MetricSSLValid          = "ssl.valid"
	MetricSSLDaysToExpiry   = "ssl.days_until_expiry"
	HealthyStatusMin        = 200
	HealthyStatusMax        = 300
	TickInterval            = 1 * time.Second
)

// ShouldCheckCertificate determines if a certificate should be checked for a target
func ShouldCheckCertificate(target config.Target) bool {
	return strings.HasPrefix(strings.ToLower(target.URL), certcheck.SchemeHTTPS) && 
		target.CheckCert != nil && *target.CheckCert
}

// MetricsClient represents the interface for sending metrics
type MetricsClient interface {
	Gauge(name string, value float64, tags []string) error
	Histogram(name string, value float64, tags []string) error
}

// NewJSONLogger creates a new JSON structured logger.
func NewJSONLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

// NopLogger creates a no-op logger for testing.
func NopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// CheckTarget performs an HTTP request to the target and returns true if the response status is 2xx (OK).
func CheckTarget(client *http.Client, target config.Target) (bool, int, time.Duration, error) {
	req, err := http.NewRequest(target.Method, target.URL, nil)
	if err != nil {
		return false, 0, 0, err
	}

	for key, value := range target.Headers {
		req.Header.Set(key, value)
	}
	
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)
	
	if err != nil {
		return false, 0, duration, err
	}
	defer resp.Body.Close()
	
	_, _ = io.Copy(io.Discard, resp.Body)
	
	status := resp.StatusCode
	if status >= HealthyStatusMin && status < HealthyStatusMax {
		return true, status, duration, nil
	}
	return false, status, duration, nil
}

// Target checks a single target and reports its status to the metrics client.
func Target(client *http.Client, target config.Target, metrics MetricsClient, logger *slog.Logger) {
	up, status, duration, err := CheckTarget(client, target)
	ms := float64(duration.Milliseconds())
	
	tags := []string{"url:" + target.URL, "name:" + target.Name}
	for k, v := range target.Labels {
		tags = append(tags, k+":"+v)
	}

	val := 0.0
	if up {
		val = 1.0
	}
	
	if metrics != nil {
		if err := metrics.Gauge(MetricURLUp, val, tags); err != nil {
			logger.Warn("Failed to send url.up metric", 
				slog.String("target", target.Name), 
				slog.String("url", target.URL),
				slog.Any("error", err))
		} else {
			logger.Info("Successfully sent url.up metric", 
				slog.String("target", target.Name), 
				slog.String("url", target.URL),
				slog.Float64("value", val))
		}
		
		if err := metrics.Histogram(MetricResponseTime, ms, tags); err != nil {
			logger.Warn("Failed to send url.response_time_ms metric", 
				slog.String("target", target.Name),
				slog.String("url", target.URL),
				slog.Any("error", err))
		} else {
			logger.Info("Successfully sent url.response_time_ms metric",
				slog.String("target", target.Name),
				slog.String("url", target.URL),
				slog.Float64("value", ms))
		}
	}

	logAttrs := []any{
		slog.String("target", target.Name),
		slog.String("url", target.URL),
		slog.Float64("response_time_ms", ms),
	}
	
	for k, v := range target.Labels {
		logAttrs = append(logAttrs, slog.String("label_"+k, v))
	}
	
	if err != nil {
		logAttrs = append(logAttrs, slog.Any("error", err))
		logger.Error("Target check failed", logAttrs...)
	} else {
		logAttrs = append(logAttrs, slog.Int("status", status))
		if !up {
			logger.Warn("Target is unhealthy", logAttrs...)
		} else {
			logger.Info("Target is healthy", logAttrs...)
		}
	}
	
	if ShouldCheckCertificate(target) {
		certDetails, certErr := certcheck.CheckCertificate(target.URL, *target.VerifyCert)
		
		if certErr != nil && certDetails == nil {
			logger.Error("Failed to check certificate",
				slog.String("target", target.Name),
				slog.String("url", target.URL),
				slog.Any("error", certErr))
		} else if certDetails != nil {
			certcheck.LogCertificateInfo(logger, target.URL, certDetails)
			
			daysUntilExpiry := time.Until(certDetails.NotAfter).Hours() / 24
			
			if metrics != nil {
				certVal := 0.0
				if certDetails.IsValid {
					certVal = 1.0
				}
				
				if err := metrics.Gauge(MetricSSLValid, certVal, tags); err != nil {
					logger.Warn("Failed to send ssl.valid metric", 
						slog.String("target", target.Name),
						slog.String("url", target.URL),
						slog.Any("error", err))
				}
				
				if err := metrics.Gauge(MetricSSLDaysToExpiry, daysUntilExpiry, tags); err != nil {
					logger.Warn("Failed to send ssl.days_until_expiry metric", 
						slog.String("target", target.Name),
						slog.String("url", target.URL),
						slog.Any("error", err))
				}
			}
		}
	}
}

// Targets starts monitoring all targets with their individual intervals.
// The function will run until the context is canceled.
func Targets(ctx context.Context, cfg *config.Config, metrics MetricsClient) {
	logger := NewJSONLogger()
	
	logger.Info("Starting target monitoring",
		slog.Int("target_count", len(cfg.Targets)))
	
	nextChecks := make(map[string]time.Time)
	for _, target := range cfg.Targets {
		nextChecks[target.Name] = time.Now()
	}
	
	ticker := time.NewTicker(TickInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping target monitoring due to context cancellation")
			return
			
		case now := <-ticker.C:
			for _, target := range cfg.Targets {
				nextCheck, ok := nextChecks[target.Name]
				if !ok || now.After(nextCheck) {
					client := &http.Client{
						Timeout: time.Duration(target.Timeout) * time.Second,
					}
					
					Target(client, target, metrics, logger)
					
					interval := time.Duration(target.Interval) * time.Second
					nextChecks[target.Name] = now.Add(interval)
				}
			}
		}
	}
}