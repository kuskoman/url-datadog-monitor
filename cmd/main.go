package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"url-datadog-exporter/pkg/config"
	"url-datadog-exporter/pkg/exporter"
	"url-datadog-exporter/pkg/monitor"
)

const defaultConfigPath = "config.yaml"

func main() {
	configPath := flag.String("config", defaultConfigPath, "Path to configuration file")
	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", slog.String("signal", sig.String()))
		cancel()
	}()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("Failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	dogstatsd, err := exporter.NewDatadogClient(
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
	
	monitor.Targets(ctx, cfg, dogstatsd)
	
	logger.Info("URL monitor service shutdown complete")
}