#!/usr/bin/env bash
set -e

# Define the chart directory
CHART_DIR="charts/url-datadog-monitor"

# Check if helm-docs is installed
if ! command -v helm-docs &> /dev/null; then
    echo "helm-docs is not installed. Installing it..."
    go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest
fi

# Check if helm unit test plugin is installed
if ! helm plugin list | grep -q unittest; then
    echo "Helm unittest plugin is not installed. Installing it..."
    helm plugin install https://github.com/quintush/helm-unittest
fi

# Generate the README
echo "Generating Helm chart README..."
helm-docs -c $CHART_DIR

echo "README generated at $CHART_DIR/README.md"

# Run helm unit tests
echo "Running Helm unit tests..."
helm unittest $CHART_DIR

echo "All tests passed successfully!"