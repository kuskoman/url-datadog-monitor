package exporter

// MetricsExporter is an interface for sending metrics to an external system
type MetricsExporter interface {
	Gauge(name string, value float64, tags []string) error
	Histogram(name string, value float64, tags []string) error
	Count(name string, value float64, tags []string) error
}