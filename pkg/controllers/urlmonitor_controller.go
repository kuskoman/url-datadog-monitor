package controllers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	urlmonitorv1 "github.com/kuskoman/url-datadog-monitor/pkg/api/v1"
	"github.com/kuskoman/url-datadog-monitor/pkg/certcheck"
	"github.com/kuskoman/url-datadog-monitor/pkg/config"
	"github.com/kuskoman/url-datadog-monitor/pkg/exporter"
	"github.com/kuskoman/url-datadog-monitor/pkg/monitor"
)

// URLMonitorReconciler reconciles a URLMonitor object
type URLMonitorReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	MetricsClient exporter.MetricsExporter
	Logger        *slog.Logger

	// Map to track active monitors
	monitors     map[string]context.CancelFunc
	monitorsLock sync.Mutex
}

// NewURLMonitorReconciler creates a new reconciler for URLMonitor resources
func NewURLMonitorReconciler(client client.Client, scheme *runtime.Scheme, metricsClient exporter.MetricsExporter, logger *slog.Logger) *URLMonitorReconciler {
	return &URLMonitorReconciler{
		Client:        client,
		Scheme:        scheme,
		MetricsClient: metricsClient,
		Logger:        logger,
		monitors:      make(map[string]context.CancelFunc),
	}
}

// +kubebuilder:rbac:groups=url-datadog-monitor.kuskoman.github.com,resources=urlmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=url-datadog-monitor.kuskoman.github.com,resources=urlmonitors/status,verbs=get;update;patch

// Reconcile implements the reconciliation loop for URLMonitor resources
func (r *URLMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)
	reqLogger.Info("Reconciling URLMonitor")

	// Fetch the URLMonitor instance
	urlMonitor := &urlmonitorv1.URLMonitor{}
	err := r.Get(ctx, req.NamespacedName, urlMonitor)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request
			r.stopMonitoring(req.String())
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request
		return ctrl.Result{}, err
	}

	// Start or update the monitoring
	r.startOrUpdateMonitoring(ctx, urlMonitor)

	return ctrl.Result{}, nil
}

// startOrUpdateMonitoring starts or updates monitoring for a URLMonitor resource
func (r *URLMonitorReconciler) startOrUpdateMonitoring(ctx context.Context, urlMonitor *urlmonitorv1.URLMonitor) {
	monitorKey := fmt.Sprintf("%s/%s", urlMonitor.Namespace, urlMonitor.Name)

	r.stopMonitoring(monitorKey)

	monitorCtx, cancel := context.WithCancel(context.Background())

	r.monitorsLock.Lock()
	r.monitors[monitorKey] = cancel
	r.monitorsLock.Unlock()

	// Start monitoring in a separate goroutine
	go func() {
		target := config.Target{
			Name:       urlMonitor.Name,
			URL:        urlMonitor.Spec.URL,
			Method:     urlMonitor.Spec.Method,
			Interval:   urlMonitor.Spec.Interval,
			Timeout:    urlMonitor.Spec.Timeout,
			Headers:    urlMonitor.Spec.Headers,
			Labels:     urlMonitor.Spec.Labels,
			CheckCert:  urlMonitor.Spec.CheckCert,
			VerifyCert: urlMonitor.Spec.VerifyCert,
		}

		r.monitorURL(monitorCtx, urlMonitor, target)
	}()
}

// stopMonitoring stops monitoring for a URLMonitor resource
func (r *URLMonitorReconciler) stopMonitoring(key string) {
	r.monitorsLock.Lock()
	defer r.monitorsLock.Unlock()

	if cancel, exists := r.monitors[key]; exists {
		cancel()
		delete(r.monitors, key)
		r.Logger.Info("Stopped monitoring", slog.String("monitor", key))
	}
}

// monitorURL periodically checks a URL and updates the URLMonitor status
func (r *URLMonitorReconciler) monitorURL(ctx context.Context, urlMonitor *urlmonitorv1.URLMonitor, target config.Target) {
	ticker := time.NewTicker(time.Duration(target.Interval) * time.Second)
	defer ticker.Stop()

	r.Logger.Info("Started monitoring",
		slog.String("name", target.Name),
		slog.String("url", target.URL),
		slog.Int("interval", target.Interval))

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			client := &http.Client{
				Timeout: time.Duration(target.Timeout) * time.Second,
			}

			up, status, duration, err := monitor.CheckTarget(client, target)

			tags := []string{"url:" + target.URL, "name:" + target.Name}
			for k, v := range target.Labels {
				tags = append(tags, k+":"+v)
			}

			val := 0.0
			if up {
				val = 1.0
			}

			_ = r.MetricsClient.Gauge(monitor.MetricURLUp, val, tags)
			_ = r.MetricsClient.Histogram(monitor.MetricResponseTime, float64(duration.Milliseconds()), tags)

			statusUpdate := &urlmonitorv1.URLMonitorStatus{
				LastCheckTime: metav1.Now(),
				ResponseTime:  duration.Milliseconds(),
			}

			if err != nil {
				statusUpdate.Status = "Error"
				r.Logger.Warn("Error checking URL",
					slog.String("url", target.URL),
					slog.Any("error", err))
			} else {
				statusUpdate.StatusCode = status
				if up {
					statusUpdate.Status = "Up"
				} else {
					statusUpdate.Status = "Down"
				}
			}

			if monitor.ShouldCheckCertificate(target) {
				certDetails, certErr := certcheck.CheckCertificate(target.URL, *target.VerifyCert)
				if certErr == nil && certDetails != nil {
					daysUntilExpiry := time.Until(certDetails.NotAfter).Hours() / 24

					certVal := 0.0
					if certDetails.IsValid {
						certVal = 1.0
					}

					_ = r.MetricsClient.Gauge(monitor.MetricSSLValid, certVal, tags)
					_ = r.MetricsClient.Gauge(monitor.MetricSSLDaysToExpiry, daysUntilExpiry, tags)

					// Update certificate status
					statusUpdate.Certificate = &urlmonitorv1.CertificateStatus{
						Valid:           certDetails.IsValid,
						Subject:         certDetails.Subject,
						Issuer:          certDetails.Issuer,
						NotAfter:        metav1.NewTime(certDetails.NotAfter),
						DaysUntilExpiry: daysUntilExpiry,
					}
				}
			}

			err = r.updateStatus(ctx, urlMonitor, statusUpdate)
			if err != nil {
				r.Logger.Error("Failed to update URLMonitor status",
					slog.String("name", urlMonitor.Name),
					slog.String("namespace", urlMonitor.Namespace),
					slog.Any("error", err))
			}
		}
	}
}

// updateStatus updates the status of a URLMonitor resource
func (r *URLMonitorReconciler) updateStatus(ctx context.Context, urlMonitor *urlmonitorv1.URLMonitor, status *urlmonitorv1.URLMonitorStatus) error {
	latest := &urlmonitorv1.URLMonitor{}
	err := r.Get(ctx, client.ObjectKey{Namespace: urlMonitor.Namespace, Name: urlMonitor.Name}, latest)
	if err != nil {
		return err
	}

	latest.Status = *status

	return r.Status().Update(ctx, latest)
}

// SetupWithManager sets up the controller with the Manager
func (r *URLMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&urlmonitorv1.URLMonitor{}).
		Complete(r)
}
