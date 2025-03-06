package exporters

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// DatadogClient implements the DogStatsD client for sending metrics to Datadog.
type DatadogClient struct {
	conn      net.Conn
	addr      string
	namespace string
}

// NewDatadogClient creates a new DogStatsD client for sending metrics.
func NewDatadogClient(host string, port int) (*DatadogClient, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DogStatsD: %w", err)
	}
	
	return &DatadogClient{
		conn:      conn,
		addr:      addr,
		namespace: "url_monitor.",
	}, nil
}

// Close closes the UDP connection.
func (d *DatadogClient) Close() error {
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

// send formats and sends a metrics message to DogStatsD.
func (d *DatadogClient) send(name string, value float64, metricType string, tags []string) error {
	metricName := d.namespace + name
	
	var message strings.Builder
	message.WriteString(metricName)
	message.WriteString(":")
	message.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
	message.WriteString("|")
	message.WriteString(metricType)
	
	if len(tags) > 0 {
		message.WriteString("|#")
		message.WriteString(strings.Join(tags, ","))
	}
	
	_, err := fmt.Fprintf(d.conn, message.String())
	return err
}

// Gauge sends a gauge metric.
func (d *DatadogClient) Gauge(name string, value float64, tags []string) error {
	return d.send(name, value, "g", tags)
}

// Histogram sends a histogram metric.
func (d *DatadogClient) Histogram(name string, value float64, tags []string) error {
	return d.send(name, value, "h", tags)
}

// Count sends a counter metric.
func (d *DatadogClient) Count(name string, value float64, tags []string) error {
	return d.send(name, value, "c", tags)
}