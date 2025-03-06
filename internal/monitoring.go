package internal

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// DatadogClient represents the Datadog metrics client.
type DatadogClient interface {
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
func CheckTarget(client *http.Client, target Target) (bool, int, time.Duration, error) {
	req, err := http.NewRequest(target.Method, target.URL, nil)
	if err != nil {
		return false, 0, 0, err
	}

	// Add headers
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
	
	// Drain the response body to reuse connections
	_, _ = io.Copy(io.Discard, resp.Body)
	
	status := resp.StatusCode
	// Consider 200-299 as healthy
	if status >= 200 && status < 300 {
		return true, status, duration, nil
	}
	return false, status, duration, nil
}

// MonitorTarget checks a single target and reports its status to Datadog.
func MonitorTarget(client *http.Client, target Target, datadog DatadogClient, logger *slog.Logger) {
	up, status, duration, err := CheckTarget(client, target)
	ms := float64(duration.Milliseconds())
	
	// Prepare tags from target labels and name
	tags := []string{"url:" + target.URL, "name:" + target.Name}
	for k, v := range target.Labels {
		tags = append(tags, k+":"+v)
	}

	// Metrics
	val := 0.0
	if up {
		val = 1.0
	}
	
	if datadog != nil {
		// Send url.up gauge metric (0 for down, 1 for up)
		if err := datadog.Gauge("url.up", val, tags); err != nil {
			logger.Warn("Failed to send url.up metric", 
				slog.String("target", target.Name), 
				slog.String("url", target.URL),
				slog.Any("error", err))
		} else {
			logger.Info("Successfully sent url.up metric", 
				slog.String("target", target.Name), 
				slog.String("url", target.URL),
				slog.Float64("value", val),
				slog.String("tags", tags[0]+","+tags[1]))
		}
		
		// Send response time histogram metric
		if err := datadog.Histogram("url.response_time_ms", ms, tags); err != nil {
			logger.Warn("Failed to send url.response_time_ms metric", 
				slog.String("target", target.Name),
				slog.String("url", target.URL),
				slog.Any("error", err))
		} else {
			logger.Info("Successfully sent url.response_time_ms metric",
				slog.String("target", target.Name),
				slog.String("url", target.URL),
				slog.Float64("value", ms),
				slog.String("tags", tags[0]+","+tags[1]))
		}
	}

	// Logging with all relevant fields
	logAttrs := []any{
		slog.String("target", target.Name),
		slog.String("url", target.URL),
		slog.Float64("response_time_ms", ms),
	}
	
	// Add labels as attributes
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
}

// MonitorTargets starts monitoring all targets with their individual intervals.
func MonitorTargets(cfg *Config, datadog DatadogClient) {
	client := &http.Client{Timeout: 10 * time.Second}
	logger := NewJSONLogger()
	
	logger.Info("Starting target monitoring",
		slog.Int("target_count", len(cfg.Targets)))
	
	// Create a map to track when each target needs to be checked next
	nextChecks := make(map[string]time.Time)
	for _, target := range cfg.Targets {
		nextChecks[target.Name] = time.Now()
	}
	
	// Main monitoring loop
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		
		for _, target := range cfg.Targets {
			nextCheck, ok := nextChecks[target.Name]
			if !ok || now.After(nextCheck) {
				// Time to check this target
				MonitorTarget(client, target, datadog, logger)
				
				// Schedule next check
				interval := time.Duration(target.Interval) * time.Second
				nextChecks[target.Name] = now.Add(interval)
			}
		}
	}
}