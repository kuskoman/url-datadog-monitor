package main

import (
	"flag"
	"log/slog"
	"os"
	"url-datadog-exporter/exporters"
	"url-datadog-exporter/internal"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Create structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

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
	
	// Start the monitoring loop
	internal.MonitorTargets(cfg, dogstatsd)
}