package internal

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Target represents a URL to monitor
type Target struct {
	Name     string            `yaml:"name"`
	URL      string            `yaml:"url"`
	Method   string            `yaml:"method"`
	Headers  map[string]string `yaml:"headers"`
	Labels   map[string]string `yaml:"labels"`
	Interval int               `yaml:"interval"`
	Timeout  int               `yaml:"timeout"`  // timeout in seconds
}

// Config represents the structure of config.yaml
type Config struct {
	Targets []Target `yaml:"targets"`
	Datadog struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"datadog"`
}

// LoadConfig reads the YAML config file and unmarshals it into a Config struct.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config YAML: %w", err)
	}
	
	// Set defaults for any unspecified fields
	for i := range cfg.Targets {
		// URL is the only required field
		if cfg.Targets[i].URL == "" {
			return nil, fmt.Errorf("target %d missing required URL field", i)
		}
		
		// Set defaults for optional fields
		if cfg.Targets[i].Method == "" {
			cfg.Targets[i].Method = "GET"
		}
		
		if cfg.Targets[i].Interval <= 0 {
			cfg.Targets[i].Interval = 60 // Default to 60 seconds
		}
		
		if cfg.Targets[i].Timeout <= 0 {
			cfg.Targets[i].Timeout = 10 // Default to 10 seconds timeout
		}
		
		// Set a default name if not provided
		if cfg.Targets[i].Name == "" {
			cfg.Targets[i].Name = cfg.Targets[i].URL
		}
	}
	
	// Default Datadog settings if not provided
	if cfg.Datadog.Host == "" {
		cfg.Datadog.Host = "127.0.0.1"
	}
	
	if cfg.Datadog.Port == 0 {
		cfg.Datadog.Port = 8125
	}
	
	return &cfg, nil
}