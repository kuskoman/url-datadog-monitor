package exporters

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"
)

// mockUDPServer simulates a DogStatsD server for testing
type mockUDPServer struct {
	conn      net.PacketConn
	addr      string
	received  chan string
	listening bool
}

func newMockUDPServer(t *testing.T) *mockUDPServer {
	addr := "127.0.0.1:0" // Use port 0 to get a free port
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		t.Fatalf("Failed to create mock UDP server: %v", err)
	}
	
	server := &mockUDPServer{
		conn:     conn,
		addr:     conn.LocalAddr().String(),
		received: make(chan string, 10),
	}
	
	// Start receiving packets
	go server.listen()
	
	// Give the server a moment to start listening
	time.Sleep(100 * time.Millisecond)
	
	return server
}

func (s *mockUDPServer) listen() {
	s.listening = true
	buffer := make([]byte, 1024)
	
	for s.listening {
		n, _, err := s.conn.ReadFrom(buffer)
		if err != nil {
			if !s.listening {
				// Server was closed, just return
				return
			}
			continue
		}
		
		s.received <- string(buffer[:n])
	}
}

func (s *mockUDPServer) close() {
	s.listening = false
	s.conn.Close()
	close(s.received)
}

func TestDatadogClient_SendMetrics(t *testing.T) {
	// Create a mock UDP server
	server := newMockUDPServer(t)
	defer server.close()
	
	// Parse the server address to get host and port
	addrParts := strings.Split(server.addr, ":")
	if len(addrParts) != 2 {
		t.Fatalf("Invalid address format: %s", server.addr)
	}
	
	host := addrParts[0]
	// Parse port string to int
	port, err := strconv.Atoi(addrParts[1])
	if err != nil {
		t.Fatalf("Failed to parse port number: %v", err)
	}
	
	// Create the Datadog client pointing to our mock server
	client, err := NewDatadogClient(host, port)
	if err != nil {
		t.Fatalf("Failed to create Datadog client: %v", err)
	}
	defer client.Close()
	
	// Test gauge metric
	err = client.Gauge("test.gauge", 42.5, []string{"tag1:value1", "tag2:value2"})
	if err != nil {
		t.Errorf("Error sending gauge metric: %v", err)
	}
	
	// Check received message
	select {
	case msg := <-server.received:
		expected := "url_monitor.test.gauge:42.5|g|#tag1:value1,tag2:value2"
		if msg != expected {
			t.Errorf("Expected gauge message '%s', got '%s'", expected, msg)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timed out waiting for gauge message")
	}
	
	// Test histogram metric
	err = client.Histogram("test.histogram", 123, []string{"tag1:value1"})
	if err != nil {
		t.Errorf("Error sending histogram metric: %v", err)
	}
	
	// Check received message
	select {
	case msg := <-server.received:
		expected := "url_monitor.test.histogram:123|h|#tag1:value1"
		if msg != expected {
			t.Errorf("Expected histogram message '%s', got '%s'", expected, msg)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timed out waiting for histogram message")
	}
}