# url-datadog-monitor

![Version: 0.0.2](https://img.shields.io/badge/Version-0.0.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.2](https://img.shields.io/badge/AppVersion-0.0.2-informational?style=flat-square)

A Helm chart for deploying URL Datadog Monitor on Kubernetes

**Homepage:** <https://github.com/kuskoman/url-datadog-monitor>

## Installation

### Add Helm Repository

```bash
helm repo add url-datadog-monitor https://kuskoman.github.io/url-datadog-monitor
helm repo update
```

### Operator Mode (Default)

Deploy as an operator to monitor URLs defined by URLMonitor custom resources:

```bash
helm install url-monitor url-datadog-monitor/url-datadog-monitor \
  --set datadog.host=datadog-agent.datadog.svc.cluster.local \
  --set datadog.port=8125
```

After installation, create URLMonitor resources:

```yaml
apiVersion: urlmonitoring.kuskoman.github.com/v1
kind: URLMonitor
metadata:
  name: example-com
spec:
  url: https://example.com
  method: GET
  interval: 60
  timeout: 10
  checkCert: true
  verifyCert: false
  labels:
    env: production
    service: website
```

### Standalone Mode

Deploy in standalone mode with a predefined list of target URLs:

```bash
helm install url-monitor url-datadog-monitor/url-datadog-monitor \
  --set mode=standalone \
  --set datadog.host=datadog-agent.datadog.svc.cluster.local \
  --set datadog.port=8125
```

## Configuration Modes

The chart supports two operational modes:

- **Operator Mode** (default): Runs as a Kubernetes operator, watching URLMonitor custom resources to dynamically manage URL monitoring
- **Standalone Mode**: Runs with a static configuration defined in values.yaml

## Requirements

- Kubernetes: >= 1.21.0-0
- Helm: >= 3.3.0
- A running Datadog Agent with DogStatsD enabled

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| crd.annotations | object | `{}` |  |
| datadog.host | string | `"datadog-agent.datadog.svc.cluster.local"` |  |
| datadog.port | int | `8125` |  |
| fullnameOverride | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/kuskoman/url-datadog-monitor"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| mode | string | `"operator"` |  |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| operator.createCRD | bool | `true` |  |
| operator.installSamples | bool | `true` |  |
| operator.leaderElection.enabled | bool | `false` |  |
| operator.rbac.create | bool | `true` |  |
| podAnnotations | object | `{}` |  |
| podSecurityContext | object | `{}` |  |
| probes.liveness.enabled | bool | `true` |  |
| probes.liveness.failureThreshold | int | `3` |  |
| probes.liveness.initialDelaySeconds | int | `10` |  |
| probes.liveness.periodSeconds | int | `30` |  |
| probes.liveness.timeoutSeconds | int | `5` |  |
| probes.readiness.enabled | bool | `true` |  |
| probes.readiness.failureThreshold | int | `2` |  |
| probes.readiness.initialDelaySeconds | int | `5` |  |
| probes.readiness.periodSeconds | int | `10` |  |
| probes.readiness.timeoutSeconds | int | `5` |  |
| replicaCount | int | `1` |  |
| resources.limits.cpu | string | `"100m"` |  |
| resources.limits.memory | string | `"128Mi"` |  |
| resources.requests.cpu | string | `"10m"` |  |
| resources.requests.memory | string | `"64Mi"` |  |
| sampleURLMonitors[0].name | string | `"example-com"` |  |
| sampleURLMonitors[0].spec.checkCert | bool | `true` |  |
| sampleURLMonitors[0].spec.interval | int | `60` |  |
| sampleURLMonitors[0].spec.labels.env | string | `"production"` |  |
| sampleURLMonitors[0].spec.labels.service | string | `"website"` |  |
| sampleURLMonitors[0].spec.method | string | `"GET"` |  |
| sampleURLMonitors[0].spec.timeout | int | `10` |  |
| sampleURLMonitors[0].spec.url | string | `"https://example.com"` |  |
| sampleURLMonitors[0].spec.verifyCert | bool | `false` |  |
| sampleURLMonitors[1].name | string | `"google-com"` |  |
| sampleURLMonitors[1].spec.checkCert | bool | `true` |  |
| sampleURLMonitors[1].spec.interval | int | `30` |  |
| sampleURLMonitors[1].spec.labels.env | string | `"production"` |  |
| sampleURLMonitors[1].spec.labels.service | string | `"search"` |  |
| sampleURLMonitors[1].spec.method | string | `"GET"` |  |
| sampleURLMonitors[1].spec.timeout | int | `5` |  |
| sampleURLMonitors[1].spec.url | string | `"https://google.com"` |  |
| sampleURLMonitors[1].spec.verifyCert | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| securityContext.runAsNonRoot | bool | `true` |  |
| securityContext.runAsUser | int | `65534` |  |
| service.port | int | `8080` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| standalone.config.datadog.host | string | `"${DATADOG_HOST}"` |  |
| standalone.config.datadog.port | string | `"${DATADOG_PORT}"` |  |
| standalone.config.defaults.check_cert | bool | `true` |  |
| standalone.config.defaults.headers.User-Agent | string | `"Datadog-Monitor"` |  |
| standalone.config.defaults.interval | int | `60` |  |
| standalone.config.defaults.labels.app | string | `"url-monitor"` |  |
| standalone.config.defaults.method | string | `"GET"` |  |
| standalone.config.defaults.timeout | int | `10` |  |
| standalone.config.defaults.verify_cert | bool | `false` |  |
| standalone.config.targets[0].interval | int | `30` |  |
| standalone.config.targets[0].labels.env | string | `"production"` |  |
| standalone.config.targets[0].labels.service | string | `"website"` |  |
| standalone.config.targets[0].name | string | `"Example Site"` |  |
| standalone.config.targets[0].timeout | int | `5` |  |
| standalone.config.targets[0].url | string | `"https://example.com"` |  |
| standalone.config.targets[1].interval | int | `60` |  |
| standalone.config.targets[1].name | string | `"Google"` |  |
| standalone.config.targets[1].timeout | int | `5` |  |
| standalone.config.targets[1].url | string | `"https://google.com"` |  |
| standalone.config.targets[1].verify_cert | bool | `true` |  |
| tolerations | list | `[]` |  |

### Notable Configuration Options

#### General Settings
- `mode`: Choose between "operator" (default) or "standalone" mode
- `datadog.host`: Hostname of the Datadog agent
- `datadog.port`: Port for DogStatsD on the Datadog agent

#### Operator Mode Settings
- `operator.createCRD`: Whether to create the URLMonitor CRD (set to false if installed separately)
- `operator.installSamples`: Deploy sample URLMonitor resources
- `operator.rbac.create`: Create RBAC resources for the operator
- `operator.leaderElection`: Configuration for leader election in multi-replica deployments

#### Standalone Mode Settings
- `standalone.config`: Configuration for the standalone mode, with targets to monitor

#### Container Settings
- `image.tag`: Specify a particular image tag (defaults to appropriate tag for selected mode)
- `resources`: Configure resource requests and limits
- `securityContext`: Customize security settings

