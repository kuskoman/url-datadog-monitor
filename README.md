# URL Datadog Monitor

A Go service that monitors multiple URLs and their SSL certificates, exporting metrics to Datadog.

## Features

- Monitor multiple targets with different configurations
- Support for custom HTTP methods and headers
- Individual check intervals per target
- Custom labels for better metric organization
- Export metrics to Datadog via DogStatsD
- SSL certificate monitoring with expiration tracking
- Certificate chain validation options (verify or just check)
- Configurable certificate verification per target
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

| Metric Name | Type | Description | When Reported |
|-------------|------|-------------|---------------|
| `url_monitor.url.up` | Gauge | 0 or 1 indicating if the target is up (2xx response code) | Every check |
| `url_monitor.url.response_time_ms` | Histogram | Response time in milliseconds | Every successful check |

### SSL Certificate Metrics

These metrics are only reported for HTTPS URLs with certificate checking enabled (`check_cert: true`).

| Metric Name | Type | Description | When Reported |
|-------------|------|-------------|---------------|
| `url_monitor.ssl.valid` | Gauge | 0 or 1 indicating if the certificate is valid | When certificate check is performed |
| `url_monitor.ssl.days_until_expiry` | Gauge | Number of days until certificate expiration | When certificate check is performed |

### Metric Tags

All metrics include the following tags:

| Tag | Example | Description |
|-----|---------|-------------|
| `url` | `url:https://example.com` | The URL being monitored |
| `name` | `name:Example Site` | The target name |
| Custom labels | `env:production`, `service:website` | Any labels defined in the target configuration |

These tags allow you to filter and group metrics in Datadog dashboards and alerts.

## Certificate Monitoring

Certificate monitoring is automatically enabled for HTTPS URLs (unless explicitly disabled with `check_cert: false`). The service performs the following checks:

1. **Certificate Presence**: Verifies the server presents a valid SSL certificate
2. **Hostname Verification**: Checks that the certificate is valid for the requested hostname
3. **Expiration Check**: Verifies that the certificate is not expired and tracks days until expiry
4. **Chain Verification** (optional): When `verify_cert: true` is specified, validates the entire certificate chain against the system's trusted CA store

You can control certificate monitoring behavior with two configuration options:

- `check_cert`: Whether to check the certificate at all (defaults to `true` for HTTPS URLs)
- `verify_cert`: Whether to verify the certificate chain against system CAs (defaults to `false`)

This gives you flexibility to:
- Fully validate certificates (both `check_cert` and `verify_cert` set to `true`)
- Check certificate details but don't require valid chain (`check_cert: true, verify_cert: false`)
- Completely disable certificate checking (`check_cert: false`)

### Using Certificate Metrics

The SSL certificate metrics are particularly useful for:

1. **Alerting on expiring certificates**: Create a Datadog alert when `ssl.days_until_expiry` falls below a threshold (e.g., 30 days)
2. **Tracking certificate validity**: Monitor the `ssl.valid` metric to detect certificate issues
3. **Visualizing certificate expiry**: Create dashboards showing certificate expiry timelines for all your services

## Building and Running

```bash
# Build the service
go build -o url-datadog-monitor ./cmd

# Run the service
./url-datadog-monitor
```

Make sure you have the Datadog agent running locally with DogStatsD enabled.

## Development

```bash
# Run tests
go test ./...

# Run with a custom config file
./url-datadog-monitor -config=/path/to/custom-config.yaml
```

## Project Structure

The project is organized into the following packages:

- `cmd/` - Contains the main application entry point
- `pkg/` - Contains the core functionality:
  - `pkg/certcheck/` - SSL certificate checking functionality
  - `pkg/config/` - Configuration loading and processing
  - `pkg/exporter/` - Metrics exporting (Datadog implementation)
  - `pkg/monitor/` - URL monitoring and health checking