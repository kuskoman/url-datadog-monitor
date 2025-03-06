# URL Datadog Exporter

A Go service that monitors multiple URLs and exports metrics to Datadog.

## Features

- Monitor multiple targets with different configurations
- Support for custom HTTP methods and headers
- Individual check intervals per target
- Custom labels for better metric organization
- Export metrics to Datadog via DogStatsD
- Structured JSON logging using slog

## Configuration

Configuration is done via a YAML file `config.yaml`. You can define global defaults that apply to all targets, and then override them on a per-target basis:

```yaml
defaults:
  method: "GET"
  interval: 60
  timeout: 10
  check_cert: true    # Check SSL certificates for HTTPS URLs
  verify_cert: false  # Don't require valid certificates by default
  headers:
    User-Agent: "Datadog-Monitor"
  labels:
    app: "url-monitor"

targets:
  - name: "Example Site"
    url: "https://example.com"
    labels:
      env: "production"
      service: "website"
    interval: 30
    timeout: 5
  - name: "Httpbin OK" 
    url: "http://httpbin.org/status/200"
    interval: 60
    timeout: 3
    check_cert: false  # Explicitly disable cert check for non-HTTPS
    labels:
      env: "testing"
  - name: "SSL Check Example"
    url: "https://google.com"
    interval: 45
    timeout: 5
    # Using defaults for cert check (enabled)
    # Verifying the certificate chain for this target
    verify_cert: true
    labels:
      env: "production"
      type: "ssl-verification"
datadog:
  host: "127.0.0.1"
  port: 8125
```

The only required field for a target is `url`. All other fields have sensible defaults.

### Configuration Options

**Global Defaults:**
- `method`: HTTP method to use for requests (default: "GET")
- `interval`: Check interval in seconds (default: 60)
- `timeout`: Request timeout in seconds (default: 10)
- `check_cert`: Whether to check SSL certificates for HTTPS URLs (default: true)
- `verify_cert`: Whether to verify certificate validity against system trust store (default: false)
- `headers`: Map of HTTP headers to send with requests
- `labels`: Map of labels to apply to all targets (useful for Datadog tag filtering)

**Target Options:**
- `name`: Name for the target (defaults to URL if not specified)
- `url`: URL to monitor (required)
- `method`: HTTP method (overrides default)
- `interval`: Check interval in seconds (overrides default)
- `timeout`: Request timeout in seconds (overrides default)
- `check_cert`: Whether to check SSL certificate (overrides default)
- `verify_cert`: Whether to verify certificate validity (overrides default)
- `headers`: Map of HTTP headers (merged with default headers)
- `labels`: Map of labels (merged with default labels)

## Metrics

The service exports the following metrics to Datadog:

### URL Health Metrics
- `url_monitor.url.up` - gauge (0 or 1) indicating if the target is up
- `url_monitor.url.response_time_ms` - histogram of response times in milliseconds

### SSL Certificate Metrics (for HTTPS URLs with certificate checking enabled)
- `url_monitor.ssl.valid` - gauge (0 or 1) indicating if the certificate is valid
- `url_monitor.ssl.days_until_expiry` - gauge indicating days until certificate expiry

All metrics include tags:
- `url:https://example.com` - the URL being monitored
- `name:Example Site` - the target name
- Any custom labels defined in the target configuration

### Using Certificate Metrics

The SSL certificate metrics are particularly useful for:

1. **Alerting on expiring certificates**: Create a Datadog alert when `ssl.days_until_expiry` falls below a threshold (e.g., 30 days)
2. **Tracking certificate validity**: Monitor the `ssl.valid` metric to detect certificate issues
3. **Visualizing certificate expiry**: Create dashboards showing certificate expiry timelines for all your services

## Building and Running

```bash
# Build the service
go build -o url-datadog-exporter ./cmd

# Run the service
./url-datadog-exporter
```

Make sure you have the Datadog agent running locally with DogStatsD enabled.

## Development

```bash
# Run tests
go test ./...

# Run with a custom config file
./url-datadog-exporter -config=/path/to/custom-config.yaml
```

## Project Structure

The project is organized into the following packages:

- `cmd/` - Contains the main application entry point
- `pkg/` - Contains the core functionality:
  - `pkg/certcheck/` - SSL certificate checking functionality
  - `pkg/config/` - Configuration loading and processing
  - `pkg/exporter/` - Metrics exporting (Datadog implementation)
  - `pkg/monitor/` - URL monitoring and health checking