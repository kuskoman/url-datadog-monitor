# Kubernetes API Code Generation

This directory contains files used for generating Kubernetes API code for the Custom Resource Definitions (CRDs) used by the operator mode of the URL Datadog Monitor.

## Files

- `boilerplate.go.txt` - A minimal header used in the generated code files.

## Controller Generation Process

The project uses [controller-gen](https://github.com/kubernetes-sigs/controller-tools/tree/master/cmd/controller-gen) to generate Kubernetes Custom Resource Definition (CRD) files and deepcopy methods for the Go types.

### Controller-gen Tool

Controller-gen is a tool from the Kubernetes sigs that generates:

1. DeepCopy methods for Go types (required for Kubernetes custom resources)
2. CRD YAML files based on Go type definitions and comments
3. RBAC manifests based on controller code comments

### How to Generate Code

1. Install controller-gen:

```bash
go install sigs.k8s.io/controller-tools/cmd/controller-gen@latest
```

2. Generate DeepCopy methods for API types:

```bash
controller-gen object:headerFile=hack/boilerplate.go.txt paths="./pkg/api/..."
```

This generates the `zz_generated.deepcopy.go` file in the `pkg/api/v1` directory, which contains the implementation of the DeepCopyObject interface required by Kubernetes.

3. Generate CRD YAML manifests:

```bash
controller-gen crd:trivialVersions=true paths="./pkg/api/..." output:crd:artifacts:config=config/crd/bases
```

This generates the CRD YAML file in the `config/crd/bases` directory, which can be applied to a Kubernetes cluster.

## Marker Comments

The API types use special marker comments to control the CRD generation:

- `// +groupName=url-datadog-monitor.kuskoman.github.com` - Sets the API group name
- `// +kubebuilder:object:root=true` - Marks the type as a root type
- `// +kubebuilder:subresource:status` - Enables the status subresource
- `// +kubebuilder:printcolumn:...` - Adds custom columns to kubectl output
- `// +kubebuilder:validation:...` - Adds validation rules to fields

These markers are used by controller-gen to generate the appropriate CRD YAML with the specified features.
