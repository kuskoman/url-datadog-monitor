package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	DefaultMethod          = "GET"
	DefaultInterval        = 60
	DefaultTimeout         = 10
	DefaultDogStatsDHost   = "127.0.0.1"
	DefaultDogStatsDPort   = 8125
)

// Target represents a URL to monitor
type Target struct {
	Name       string            `yaml:"name"`
	URL        string            `yaml:"url"`
	Method     string            `yaml:"method"`
	Headers    map[string]string `yaml:"headers"`
	Labels     map[string]string `yaml:"labels"`
	Interval   int               `yaml:"interval"`
	Timeout    int               `yaml:"timeout"`
	CheckCert  *bool             `yaml:"check_cert"`
	VerifyCert *bool             `yaml:"verify_cert"`
}

// Defaults represents global default settings for all targets
type Defaults struct {
	Method      string            `yaml:"method"`
	Interval    int               `yaml:"interval"`
	Timeout     int               `yaml:"timeout"`
	Headers     map[string]string `yaml:"headers"`
	Labels      map[string]string `yaml:"labels"`
	CheckCert   bool              `yaml:"check_cert"`
	VerifyCert  bool              `yaml:"verify_cert"`
}

// Config represents the structure of config.yaml
type Config struct {
	Defaults Defaults `yaml:"defaults"`
	Targets  []Target `yaml:"targets"`
	Datadog  struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"datadog"`
}

// Load reads the YAML config file and unmarshals it into a Config struct.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config YAML: %w", err)
	}
	
	if cfg.Defaults.Method == "" {
		cfg.Defaults.Method = DefaultMethod
	}
	if cfg.Defaults.Interval <= 0 {
		cfg.Defaults.Interval = DefaultInterval
	}
	if cfg.Defaults.Timeout <= 0 {
		cfg.Defaults.Timeout = DefaultTimeout
	}
	if cfg.Defaults.Headers == nil {
		cfg.Defaults.Headers = make(map[string]string)
	}
	if cfg.Defaults.Labels == nil {
		cfg.Defaults.Labels = make(map[string]string)
	}
	
	defaultCheckCert := true
	defaultVerifyCert := false
	if cfg.Defaults.CheckCert != defaultCheckCert {
		cfg.Defaults.CheckCert = defaultCheckCert
	}
	if cfg.Defaults.VerifyCert != defaultVerifyCert {
		cfg.Defaults.VerifyCert = defaultVerifyCert
	}
	
	for i := range cfg.Targets {
		if cfg.Targets[i].URL == "" {
			return nil, fmt.Errorf("target %d missing required URL field", i)
		}
		
		if cfg.Targets[i].Method == "" {
			cfg.Targets[i].Method = cfg.Defaults.Method
		}
		
		if cfg.Targets[i].Interval <= 0 {
			cfg.Targets[i].Interval = cfg.Defaults.Interval
		}
		
		if cfg.Targets[i].Timeout <= 0 {
			cfg.Targets[i].Timeout = cfg.Defaults.Timeout
		}
		
		if cfg.Targets[i].Name == "" {
			cfg.Targets[i].Name = cfg.Targets[i].URL
		}
		
		if cfg.Targets[i].Headers == nil {
			cfg.Targets[i].Headers = make(map[string]string)
		}
		for k, v := range cfg.Defaults.Headers {
			if _, exists := cfg.Targets[i].Headers[k]; !exists {
				cfg.Targets[i].Headers[k] = v
			}
		}
		
		if cfg.Targets[i].Labels == nil {
			cfg.Targets[i].Labels = make(map[string]string)
		}
		for k, v := range cfg.Defaults.Labels {
			if _, exists := cfg.Targets[i].Labels[k]; !exists {
				cfg.Targets[i].Labels[k] = v
			}
		}
		
		if cfg.Targets[i].CheckCert == nil {
			checkCert := cfg.Defaults.CheckCert
			cfg.Targets[i].CheckCert = &checkCert
		}
		if cfg.Targets[i].VerifyCert == nil {
			verifyCert := cfg.Defaults.VerifyCert
			cfg.Targets[i].VerifyCert = &verifyCert
		}
	}
	
	if cfg.Datadog.Host == "" {
		cfg.Datadog.Host = DefaultDogStatsDHost
	}
	
	if cfg.Datadog.Port == 0 {
		cfg.Datadog.Port = DefaultDogStatsDPort
	}
	
	return &cfg, nil
}