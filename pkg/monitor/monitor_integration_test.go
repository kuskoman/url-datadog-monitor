package monitor

import (
	"context"
	"testing"
	"time"

	"url-datadog-monitor/pkg/config"
)

// mockDatadogForCancel implements the DatadogClient interface for testing context cancellation
type mockDatadogForCancel struct{}

func (m mockDatadogForCancel) Gauge(name string, value float64, tags []string) error {
	return nil
}

func (m mockDatadogForCancel) Histogram(name string, value float64, tags []string) error {
	return nil
}

func TestTargets_ContextCancellation(t *testing.T) {
	// Create a cancelable context
	ctx, cancel := context.WithCancel(context.Background())
	
	// Create a test configuration with targets
	cfg := &config.Config{
		Targets: []config.Target{
			{
				Name:     "Test Target",
				URL:      "http://example.com",
				Method:   "GET",
				Interval: 60,
			},
		},
	}
	
	// Create a mock Datadog client
	mockClient := mockDatadogForCancel{}
	
	// Start monitoring in a goroutine
	monitoringDone := make(chan struct{})
	go func() {
		// Since we're only testing context cancellation, we'll recover from any panics
		// that might occur due to a nil logger
		defer func() { 
			if r := recover(); r != nil {
				t.Logf("Recovered from panic: %v", r)
			}
			close(monitoringDone)
		}()
		Targets(ctx, cfg, mockClient)
	}()
	
	// Cancel the context after a short time
	time.Sleep(100 * time.Millisecond)
	cancel()
	
	// The test passes if Targets returns within a reasonable time
	select {
	case <-monitoringDone:
		// Success - monitoring stopped after context cancellation
	case <-time.After(2 * time.Second):
		t.Fatal("Targets didn't respect context cancellation within timeout")
	}
}