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

## Docker Images

### Production Images

The project provides several Docker image variants depending on your needs:

#### Scratch-based Images (Minimal)
- `docker/standalone-scratch.Dockerfile` - Standalone mode with minimal scratch-based image
- `docker/operator-scratch.Dockerfile` - Operator mode with minimal scratch-based image

#### Alpine-based Images (Debugging Capability)
- `docker/standalone-alpine.Dockerfile` - Standalone mode with Alpine-based image
- `docker/operator-alpine.Dockerfile` - Operator mode with Alpine-based image

### Development Image
- `docker/dev.Dockerfile` - Development image with hot-reload using Air

### Pre-built Images

Pre-built images are available on GitHub Container Registry:

```bash
# Pull the latest standalone version (scratch-based, recommended for production)
docker pull ghcr.io/kuskoman/url-datadog-monitor:latest-standalone

# Pull the latest operator version (scratch-based, recommended for production)
docker pull ghcr.io/kuskoman/url-datadog-monitor:latest-operator

# Pull a specific version
docker pull ghcr.io/kuskoman/url-datadog-monitor:v1.0.0-standalone-scratch
```

### Building Docker Images Locally

To build any of the Docker images locally:

```bash
# Using Make (recommended)
make docker-build

# Or manually for standalone scratch version
docker build -t url-datadog-monitor:standalone-scratch -f docker/standalone-scratch.Dockerfile .

# For operator scratch version
docker build -t url-datadog-monitor:operator-scratch -f docker/operator-scratch.Dockerfile .

# For alpine versions
docker build -t url-datadog-monitor:standalone-alpine -f docker/standalone-alpine.Dockerfile .
docker build -t url-datadog-monitor:operator-alpine -f docker/operator-alpine.Dockerfile .

# For development with hot-reload
docker build -t url-datadog-monitor:dev -f docker/dev.Dockerfile .
```

#### With Version Information

To include version information in the build:

```bash
docker build \
  --build-arg VERSION=$(git describe --tags --always) \
  --build-arg GIT_COMMIT=$(git rev-parse HEAD) \
  -t url-datadog-monitor:standalone-scratch \
  -f docker/standalone-scratch.Dockerfile .
```

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

## Running Modes

URL Datadog Monitor can be run in two modes:

### 1. Standalone Mode

In standalone mode, the monitor reads a YAML configuration file to determine which URLs to monitor.

```bash
# Build the standalone service
go build -o url-datadog-monitor-standalone ./cmd/standalone

# Run in standalone mode
./url-datadog-monitor-standalone -config=/path/to/config.yaml
```

### 2. Kubernetes Operator Mode

In Kubernetes operator mode, the monitor watches for URLMonitor custom resources and dynamically updates monitoring based on these resources.

```bash
# Build the operator service
go build -o url-datadog-monitor-operator ./cmd/operator

# Run as a Kubernetes operator
./url-datadog-monitor-operator --dogstatsd-host=datadog-agent.monitoring --dogstatsd-port=8125
```

#### Kubernetes Custom Resources

The operator uses a URLMonitor custom resource definition:

```yaml
apiVersion: url-datadog-monitor.kuskoman.github.com/v1
kind: URLMonitor
metadata:
  name: example-com
spec:
  url: https://example.com
  method: GET
  interval: 30
  timeout: 5
  labels:
    env: production
    service: website
  checkCert: true
  verifyCert: true
```

To deploy the CRD:

```bash
kubectl apply -f config/crd/bases/url-datadog-monitor.kuskoman.github.com_urlmonitors.yaml
```

To create a URLMonitor:

```bash
kubectl apply -f config/samples/urlmonitor_v1_samples.yaml
```

## Helm Chart

The project includes a Helm chart to easily deploy URL Datadog Monitor in Kubernetes environments. The chart supports both operator and standalone modes.

### Installing with Helm

```bash
# Add the Helm repository (when published)
# helm repo add url-datadog-monitor https://kuskoman.github.io/url-datadog-monitor
# helm repo update

# For now, install from local chart
# Operator Mode (Default)
helm install url-monitor ./charts/url-datadog-monitor \
  --set datadog.host=datadog-agent.datadog.svc.cluster.local \
  --set datadog.port=8125

# Standalone Mode
helm install url-monitor ./charts/url-datadog-monitor \
  --set mode=standalone \
  --set datadog.host=datadog-agent.datadog.svc.cluster.local \
  --set datadog.port=8125
```

#### Troubleshooting Kubernetes Deployment

If you encounter one of these common errors:

- `failed to get API group resources: unable to retrieve the complete list of server APIs: url-datadog-monitor.kuskoman.github.com/v1: the server could not find the requested resource` - The CRD is not properly registered with the API server.
- `urlmonitors.url-datadog-monitor.kuskoman.github.com is forbidden: User "system:serviceaccount:xxx" cannot list resource "urlmonitors"` - The service account lacks proper permissions.

Follow these steps to resolve either issue:

1. Ensure the CRD is properly installed and has the correct API group:
   ```bash
   kubectl get crd urlmonitors.url-datadog-monitor.kuskoman.github.com
   ```

2. Manually install the CRD first and then install the chart with CRD creation disabled:
   ```bash
   kubectl apply -f charts/url-datadog-monitor/crds/urlmonitors.yaml
   helm install url-monitor ./charts/url-datadog-monitor --set operator.createCRD=false
   ```

3. Check the logs for any initialization issues:
   ```bash
   # Check logs of the main container
   kubectl logs -l app.kubernetes.io/name=url-datadog-monitor

   # Check logs from the CRD installation job
   kubectl logs job/url-monitor-url-datadog-monitor-crd-ready
   ```

4. If you see errors about missing API resources, uninstall the chart, manually apply the CRD, and then reinstall:
   ```bash
   helm uninstall url-monitor
   kubectl apply -f charts/url-datadog-monitor/crds/urlmonitors.yaml
   # Wait a few seconds for the CRD to be fully established
   helm install url-monitor ./charts/url-datadog-monitor --set operator.createCRD=false
   ```

5. Verify the CRD is established and ready:
   ```bash
   kubectl get crd urlmonitors.url-datadog-monitor.kuskoman.github.com -o jsonpath='{.status.conditions[?(@.type=="Established")].status}'
   # Should output: True
   ```

6. If you encounter RBAC permission errors, verify that the ClusterRole references the correct API group:
   ```bash
   kubectl get clusterrole url-monitor-url-datadog-monitor -o yaml | grep -A10 rules
   ```

7. For production deployments, consider setting `operator.leaderElection.enabled=true` only when deploying multiple replicas.

### Helm Chart Testing

The Helm chart includes comprehensive unit tests to ensure correct behavior:

```bash
# Install the Helm unittest plugin if needed
helm plugin install https://github.com/quintush/helm-unittest

# Run the tests
helm unittest ./charts/url-datadog-monitor
```

The GitHub CI pipeline also tests the Helm chart in a real Kubernetes environment using KinD (Kubernetes in Docker) to verify the operator functionality with actual URLMonitor resources.

See the [Helm chart README](./charts/url-datadog-monitor/README.md) for more details on available options and configurations.

## Kubernetes API Generation

This project uses [controller-gen](https://github.com/kubernetes-sigs/controller-tools/tree/master/cmd/controller-gen) to generate boilerplate code for Kubernetes CRDs. The CRD types are defined in `pkg/api/v1/types.go`, and the CRD YAML is generated based on those types.

To regenerate the deepcopy methods and CRD manifests after modifying the API types:

```bash
# Install controller-gen
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest

# Generate deepcopy methods
controller-gen object:headerFile=hack/boilerplate.go.txt paths="./pkg/api/..."

# Generate CRDs
controller-gen crd:trivialVersions=true paths="./pkg/api/..." output:crd:artifacts:config=config/crd/bases
```

## Development

### Running Locally

```bash
# Build both binaries
make build

# Run tests
make test

# Install the CRD
kubectl apply -f config/crd/bases/urlmonitoring.kuskoman.github.com_urlmonitors.yaml

# Run with a custom config file (standalone mode)
./bin/url-datadog-monitor-standalone -config=/path/to/custom-config.yaml

# Run as a Kubernetes operator with custom settings
./bin/url-datadog-monitor-operator --dogstatsd-host=localhost --dogstatsd-port=8125
```

### Creating a Release

To create a new release, use the provided script:

```bash
./scripts/release.sh v1.0.0
```

This will:
1. Update the Helm chart version and appVersion to match the release version
2. Regenerate the Helm chart README with the new version information
3. Commit these changes to the repository
4. Create and push a git tag
5. Trigger the GitHub Actions release workflow, which will:
   - Build binaries for all supported platforms
   - Build Docker images and push them to GitHub Container Registry
   - Create a GitHub release with attached binaries

The script ensures that both the application version and Helm chart version stay in sync, making it easier to track which chart version corresponds to which application release.

## Project Structure

The project is organized into the following packages:

- `cmd/` - Contains the main application entry points:
  - `cmd/standalone/` - Standalone mode implementation with its own main package
  - `cmd/operator/` - Kubernetes operator mode implementation with its own main package
- `pkg/` - Contains the core functionality:
  - `pkg/api/` - Kubernetes API definitions for Custom Resources
  - `pkg/certcheck/` - SSL certificate checking functionality
  - `pkg/config/` - Configuration loading and processing
  - `pkg/controllers/` - Kubernetes controllers for URLMonitor resources
  - `pkg/exporter/` - Metrics exporting (Datadog implementation)
  - `pkg/monitor/` - URL monitoring and health checking
- `config/` - Contains configuration files for Kubernetes:
  - `config/crd/` - Custom Resource Definitions
  - `config/samples/` - Example resources
