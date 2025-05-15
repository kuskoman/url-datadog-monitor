package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	urlmonitorv1 "github.com/kuskoman/url-datadog-monitor/pkg/api/v1"
	"github.com/kuskoman/url-datadog-monitor/pkg/controllers"
	"github.com/kuskoman/url-datadog-monitor/pkg/exporter"
	"github.com/kuskoman/url-datadog-monitor/pkg/version"
)

var scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = urlmonitorv1.AddToScheme(scheme)
}

func main() {
	_ = flag.String("metrics-bind-address", ":8080", "The address the metric endpoint binds to")
	probeAddr := flag.String("health-probe-bind-address", ":8081", "The address the probe endpoint binds to")
	enableLeaderElection := flag.Bool("leader-elect", false, "Enable leader election for controller manager")
	dogstatsdHost := flag.String("dogstatsd-host", "127.0.0.1", "Datadog Agent host")
	dogstatsdPort := flag.Int("dogstatsd-port", 8125, "Datadog Agent port")
	flag.Parse()

	setupLog := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	setupLog.Info("Starting URL monitor operator",
		slog.String("version", version.Version),
		slog.String("git_commit", version.GitCommit),
		slog.String("build_date", version.BuildDate))

	dogstatsd, err := exporter.NewDatadogClient(*dogstatsdHost, *dogstatsdPort)
	if err != nil {
		setupLog.Error("Failed to initialize Datadog client", slog.Any("error", err))
		os.Exit(1)
	}
	defer dogstatsd.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		setupLog.Info("Received shutdown signal", slog.String("signal", sig.String()))
		cancel()
	}()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{Development: true})))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		HealthProbeBindAddress: *probeAddr,
		LeaderElection:         *enableLeaderElection,
		LeaderElectionID:       "url-datadog-monitor-operator",
	})
	if err != nil {
		setupLog.Error("Unable to start manager", slog.Any("error", err))
		os.Exit(1)
	}

	eventRecorder := mgr.GetEventRecorderFor("url-datadog-monitor")

	reconciler := controllers.NewURLMonitorReconciler(
		mgr.GetClient(),
		mgr.GetScheme(),
		dogstatsd,
		setupLog,
		eventRecorder,
	)

	if err = reconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error("Unable to create controller", slog.String("controller", "URLMonitor"), slog.Any("error", err))
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error("Unable to set up health check", slog.Any("error", err))
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error("Unable to set up ready check", slog.Any("error", err))
		os.Exit(1)
	}

	setupLog.Info("Starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error("Problem running manager", slog.Any("error", err))
		os.Exit(1)
	}
}
