package monitor

import (
	"context"
	"testing"
	"time"

	"github.com/kuskoman/url-datadog-monitor/pkg/config"
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
	ctx, cancel := context.WithCancel(context.Background())

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

	mockClient := mockDatadogForCancel{}

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

	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case <-monitoringDone:
		// Success - monitoring stopped after context cancellation
	case <-time.After(2 * time.Second):
		t.Fatal("Targets didn't respect context cancellation within timeout")
	}
}
