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

# Check if git is clean
if [ -n "$(git status --porcelain)" ]; then
  echo "Error: Working directory is not clean. Please commit or stash your changes."
  exit 1
fi

# Create and push git tag
echo "Creating git tag $VERSION..."
git tag -a "$VERSION" -m "Release $VERSION"
git push origin "$VERSION"

echo "Release process started for $VERSION"
echo "GitHub Actions workflow has been triggered to build binaries and Docker images."
echo "Monitor the progress at: https://github.com/kuskoman/url-datadog-monitor/actions"