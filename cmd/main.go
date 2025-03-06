package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"url-datadog-exporter/exporters"
	"url-datadog-exporter/internal"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Create structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", slog.String("signal", sig.String()))
		cancel()
	}()

	// Load configuration from YAML file
	cfg, err := internal.LoadConfig(*configPath)
	if err != nil {
		logger.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	// Initialize Datadog client
	dogstatsd, err := exporters.NewDatadogClient(
		cfg.Datadog.Host, 
		cfg.Datadog.Port,
	)
	if err != nil {
		logger.Error("Failed to initialize Datadog client", slog.Any("error", err))
		os.Exit(1)
	}
	defer dogstatsd.Close()

	logger.Info("Starting URL monitor service", 
		slog.Int("target_count", len(cfg.Targets)))
	
	// Start the monitoring loop with context for graceful shutdown
	internal.MonitorTargets(ctx, cfg, dogstatsd)
	
	logger.Info("URL monitor service shutdown complete")
}