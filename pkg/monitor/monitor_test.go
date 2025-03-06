package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	
	"url-datadog-monitor/pkg/config"
)

// mockDatadog is a test implementation of the DatadogClient interface
type mockDatadog struct {
	gaugesCalled     int
	lastGaugeName    string
	lastGaugeValue   float64
	lastGaugeTags    []string
	histogramsCalled int
	lastHistName     string
	lastHistValue    float64
	lastHistTags     []string
}

func (m *mockDatadog) Gauge(name string, value float64, tags []string) error {
	m.gaugesCalled++
	m.lastGaugeName = name
	m.lastGaugeValue = value
	m.lastGaugeTags = tags
	return nil
}

func (m *mockDatadog) Histogram(name string, value float64, tags []string) error {
	m.histogramsCalled++
	m.lastHistName = name
	m.lastHistValue = value
	m.lastHistTags = tags
	return nil
}

func TestCheckTarget_Success(t *testing.T) {
	// Setup a test server that returns 200 OK
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that headers are properly set
		if r.Header.Get("User-Agent") == "TestUserAgent" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("Missing headers"))
		}
	}))
	defer server.Close()

	// Create a test target pointing to our test server
	target := config.Target{
		Name:   "Test Target",
		URL:    server.URL,
		Method: "GET",
		Headers: map[string]string{
			"User-Agent": "TestUserAgent",
		},
		Labels: map[string]string{
			"env": "test",
		},
		Interval: 30,
	}

	client := &http.Client{Timeout: 1 * time.Second}
	
	// Call the function we're testing
	up, status, duration, err := CheckTarget(client, target)
	
	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !up {
		t.Errorf("Expected up to be true for 200 status code")
	}
	if status != http.StatusOK {
		t.Errorf("Expected status 200, got %d", status)
	}
	if duration <= 0 {
		t.Errorf("Expected duration > 0, got %v", duration)
	}
}

func TestCheckTarget_Error(t *testing.T) {
	// Create a target with a URL that will cause an error
	target := config.Target{
		Name:   "Invalid Target",
		URL:    "http://this-does-not-exist.example",
		Method: "GET",
		Interval: 30,
	}

	client := &http.Client{Timeout: 1 * time.Second}
	
	// Call the function we're testing
	up, _, duration, err := CheckTarget(client, target)
	
	// Verify results
	if err == nil {
		t.Fatalf("Expected an error for invalid URL")
	}
	if up {
		t.Errorf("Expected up to be false when there's an error")
	}
	if duration < 0 {
		t.Errorf("Expected non-negative duration, got %v", duration)
	}
}

func TestMonitorTarget(t *testing.T) {
	// Setup a test server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add a small delay to ensure non-zero response time
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Create a test target
	target := config.Target{
		Name:   "Test Monitor",
		URL:    server.URL,
		Method: "GET",
		Labels: map[string]string{
			"env": "test",
		},
		Interval: 30,
	}

	// Create a mock Datadog client
	mock := &mockDatadog{}
	
	// Create a client and logger
	client := &http.Client{Timeout: 1 * time.Second}
	logger := NopLogger()
	
	// Call the function we're testing
	Target(client, target, mock, logger)
	
	// Verify that metrics were sent
	if mock.gaugesCalled != 1 {
		t.Errorf("Expected 1 gauge call, got %d", mock.gaugesCalled)
	}
	if mock.lastGaugeName != "url.up" {
		t.Errorf("Expected gauge name 'url.up', got '%s'", mock.lastGaugeName)
	}
	if mock.lastGaugeValue != 1.0 {
		t.Errorf("Expected gauge value 1.0 for success, got %f", mock.lastGaugeValue)
	}
	
	if mock.histogramsCalled != 1 {
		t.Errorf("Expected 1 histogram call, got %d", mock.histogramsCalled)
	}
	if mock.lastHistName != "url.response_time_ms" {
		t.Errorf("Expected histogram name 'url.response_time_ms', got '%s'", mock.lastHistName)
	}
	if mock.lastHistValue <= 0 {
		t.Errorf("Expected positive response time, got %f", mock.lastHistValue)
	}
	
	// Check that the right tags were sent
	expectedTags := []string{"url:" + server.URL, "name:Test Monitor", "env:test"}
	for _, expected := range expectedTags {
		found := false
		for _, actual := range mock.lastGaugeTags {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find tag '%s' in gauge tags %v", expected, mock.lastGaugeTags)
		}
	}
}