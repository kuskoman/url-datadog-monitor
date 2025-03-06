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

Configuration is done via a YAML file `config.yaml`:

```yaml
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
  - name: "Httpbin OK"
    url: "http://httpbin.org/status/200"
    interval: 60
    labels:
      env: "testing"
datadog:
  host: "127.0.0.1"
  port: 8125
```

The only required field for a target is `url`. All other fields have sensible defaults.

## Metrics

The service exports the following metrics to Datadog:

- `url_monitor.url.up` - gauge (0 or 1) indicating if the target is up
- `url_monitor.url.response_time_ms` - histogram of response times in milliseconds

All metrics include tags:
- `url:https://example.com` - the URL being monitored
- `name:Example Site` - the target name
- Any custom labels defined in the target configuration

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