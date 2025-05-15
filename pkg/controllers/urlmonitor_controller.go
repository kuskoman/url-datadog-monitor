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
	"k8s.io/client-go/tools/record"
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
	Scheme              *runtime.Scheme
	MetricsClient       exporter.MetricsExporter
	Logger              *slog.Logger
	KubernetesEventRecorder record.EventRecorder

	// Map to track active monitors
	monitors     map[string]context.CancelFunc
	monitorsLock sync.Mutex
}

// NewURLMonitorReconciler creates a new reconciler for URLMonitor resources
func NewURLMonitorReconciler(client client.Client, scheme *runtime.Scheme, metricsClient exporter.MetricsExporter, logger *slog.Logger, eventRecorder record.EventRecorder) *URLMonitorReconciler {
	return &URLMonitorReconciler{
		Client:                client,
		Scheme:                scheme,
		MetricsClient:         metricsClient,
		Logger:                logger,
		KubernetesEventRecorder: eventRecorder,
		monitors:              make(map[string]context.CancelFunc),
	}
}

// +kubebuilder:rbac:groups=url-datadog-monitor.kuskoman.github.com,resources=urlmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=url-datadog-monitor.kuskoman.github.com,resources=urlmonitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

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
			monitorKey := req.String()
			r.Logger.Info("URLMonitor resource was deleted", slog.String("monitor", monitorKey))
			
			// We can't record a K8s event for a deleted object, but we can log it
			r.stopMonitoring(monitorKey)
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request
		r.KubernetesEventRecorder.Event(urlMonitor, "Warning", "ReconcileError", 
			fmt.Sprintf("Failed to reconcile URLMonitor: %v", err))
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

	// Record an event for the monitoring start
	r.KubernetesEventRecorder.Event(urlMonitor, "Normal", "MonitoringStarted", 
		fmt.Sprintf("Starting URL monitoring for %s with %d second interval", urlMonitor.Spec.URL, urlMonitor.Spec.Interval))

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
		
		// We can't record an event here because we don't have the URLMonitor object
		// Events for stopping will be recorded in the Reconcile method when a resource is being deleted
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
				
				// Record a warning event for the error
				r.KubernetesEventRecorder.Event(urlMonitor, "Warning", "MonitorCheckError", 
					fmt.Sprintf("Error checking URL %s: %v", target.URL, err))
			} else {
				statusUpdate.StatusCode = status
				if up {
					statusUpdate.Status = "Up"
					
					// Record normal event for successful check (but not too frequently)
					if statusUpdate.StatusCode >= 200 && statusUpdate.StatusCode < 300 {
						// Only record events for successful checks occasionally to avoid flooding
						if time.Now().Minute()%10 == 0 { // Record once every ~10 minutes
							r.KubernetesEventRecorder.Event(urlMonitor, "Normal", "URLStatusUp", 
								fmt.Sprintf("URL %s is up with status code %d (response time: %dms)", 
									target.URL, status, duration.Milliseconds()))
						}
					}
				} else {
					statusUpdate.Status = "Down"
					
					// Always record events for down status
					r.KubernetesEventRecorder.Event(urlMonitor, "Warning", "URLStatusDown", 
						fmt.Sprintf("URL %s is down with status code %d", target.URL, status))
				}
			}

			if monitor.ShouldCheckCertificate(target) {
				certDetails, certErr := certcheck.CheckCertificate(target.URL, *target.VerifyCert)
				if certErr == nil && certDetails != nil {
					daysUntilExpiry := time.Until(certDetails.NotAfter).Hours() / 24

					certVal := 0.0
					if certDetails.IsValid {
						certVal = 1.0
					} else {
						// Record an event for invalid certificate
						r.KubernetesEventRecorder.Event(urlMonitor, "Warning", "InvalidCertificate", 
							fmt.Sprintf("SSL certificate for %s is invalid", target.URL))
					}

					// Record event if certificate is expiring soon (less than 14 days)
					if daysUntilExpiry < 14 {
						r.KubernetesEventRecorder.Event(urlMonitor, "Warning", "CertificateExpiringSoon", 
							fmt.Sprintf("SSL certificate for %s expires in %.1f days", target.URL, daysUntilExpiry))
					}

					_ = r.MetricsClient.Gauge(monitor.MetricSSLValid, certVal, tags)
					_ = r.MetricsClient.Gauge(monitor.MetricSSLDaysToExpiry, daysUntilExpiry, tags)

					// Update certificate status
					statusUpdate.Certificate = &urlmonitorv1.CertificateStatus{
						Valid:           certDetails.IsValid,
						Subject:         certDetails.Subject,
						Issuer:          certDetails.Issuer,
						NotAfter:        metav1.NewTime(certDetails.NotAfter),
						DaysUntilExpiry: fmt.Sprintf("%.2f", daysUntilExpiry),
					}
				}
			}

			err = r.updateStatus(ctx, urlMonitor, statusUpdate)
			if err != nil {
				r.Logger.Error("Failed to update URLMonitor status",
					slog.String("name", urlMonitor.Name),
					slog.String("namespace", urlMonitor.Namespace),
					slog.Any("error", err))
				
				// Record event for status update failure
				r.KubernetesEventRecorder.Event(urlMonitor, "Warning", "StatusUpdateFailed", 
					fmt.Sprintf("Failed to update URLMonitor status: %v", err))
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
