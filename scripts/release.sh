#!/usr/bin/env bash
set -e

# Check if a version was provided
if [ -z "$1" ]; then
  echo "Error: No version specified"
  echo "Usage: $0 <version>"
  echo "Example: $0 v1.0.0"
  exit 1
fi

VERSION=$1

# Ensure version starts with 'v'
if [[ ! $VERSION == v* ]]; then
  echo "Error: Version must start with 'v'"
  echo "Example: v1.0.0"
  exit 1
fi

# Version without the 'v' prefix for the Helm chart
CHART_VERSION="${VERSION#v}"

# Check if git is clean
if [ -n "$(git status --porcelain)" ]; then
  echo "Error: Working directory is not clean. Please commit or stash your changes."
  exit 1
fi

# Update Helm chart version and appVersion
CHART_FILE="charts/url-datadog-monitor/Chart.yaml"
if [ -f "$CHART_FILE" ]; then
  echo "Updating Helm chart version to $CHART_VERSION..."

  # Check if yq is installed, otherwise use sed
  if command -v yq &> /dev/null; then
    # Use yq to update the version
    yq e ".version = \"$CHART_VERSION\"" -i "$CHART_FILE"
    yq e ".appVersion = \"$CHART_VERSION\"" -i "$CHART_FILE"
  else
    # Use sed as a fallback
    echo "yq not found, using sed instead..."
    sed -i.bak "s/^version: .*$/version: $CHART_VERSION/" "$CHART_FILE"
    sed -i.bak "s/^appVersion: .*$/appVersion: \"$CHART_VERSION\"/" "$CHART_FILE"
    rm -f "${CHART_FILE}.bak"
  fi

  # Regenerate the Helm README with updated version info
  if command -v helm-docs &> /dev/null; then
    echo "Regenerating Helm chart README..."
    helm-docs -c charts/url-datadog-monitor
  fi

  # Commit the changes
  git add "$CHART_FILE" "charts/url-datadog-monitor/README.md"
  git commit -m "Release $VERSION"
fi

# Create and push git tag
echo "Creating git tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
git push origin master  # Push the commit with updated Chart.yaml
git push origin "$VERSION"

echo "Release process started for $VERSION"
echo "GitHub Actions workflow has been triggered to build binaries and Docker images."
echo "Monitor the progress at: https://github.com/kuskoman/url-datadog-monitor/actions"
