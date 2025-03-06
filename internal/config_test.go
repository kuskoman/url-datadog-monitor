package internal

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary YAML file for testing
	content := []byte(`
targets:
  - name: "Example Site"
    url: "https://example.com"
    method: "GET"
    headers:
      User-Agent: "Datadog-Monitor"
    labels:
      env: "production"
      service: "website"
    interval: 30
  - url: "http://test.com"
datadog:
  host: "localhost"
  port: 8125
`)
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	if err != nil {
		t.Fatalf("Could not create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	
	_, err = tmpFile.Write(content)
	if err != nil {
		t.Fatalf("Could not write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify the config fields
	if len(cfg.Targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(cfg.Targets))
	}
	
	// Check first target
	target := cfg.Targets[0]
	if target.Name != "Example Site" {
		t.Errorf("Expected name to be 'Example Site', got '%s'", target.Name)
	}
	if target.URL != "https://example.com" {
		t.Errorf("Expected URL to be 'https://example.com', got '%s'", target.URL)
	}
	if target.Method != "GET" {
		t.Errorf("Expected method to be 'GET', got '%s'", target.Method)
	}
	if target.Interval != 30 {
		t.Errorf("Expected interval to be 30, got %d", target.Interval)
	}
	if target.Headers["User-Agent"] != "Datadog-Monitor" {
		t.Errorf("Expected User-Agent header to be 'Datadog-Monitor', got '%s'", target.Headers["User-Agent"])
	}
	if target.Labels["env"] != "production" {
		t.Errorf("Expected env label to be 'production', got '%s'", target.Labels["env"])
	}
	
	// Check second target (default values should be applied)
	target = cfg.Targets[1]
	if target.URL != "http://test.com" {
		t.Errorf("Expected URL to be 'http://test.com', got '%s'", target.URL)
	}
	if target.Name != "http://test.com" {
		t.Errorf("Expected name to default to URL, got '%s'", target.Name)
	}
	if target.Method != "GET" {
		t.Errorf("Expected method to default to 'GET', got '%s'", target.Method)
	}
	if target.Interval != 60 {
		t.Errorf("Expected interval to default to 60, got %d", target.Interval)
	}
	
	// Check datadog config
	if cfg.Datadog.Host != "localhost" {
		t.Errorf("Expected Datadog host to be 'localhost', got '%s'", cfg.Datadog.Host)
	}
	if cfg.Datadog.Port != 8125 {
		t.Errorf("Expected Datadog port to be 8125, got %d", cfg.Datadog.Port)
	}
}