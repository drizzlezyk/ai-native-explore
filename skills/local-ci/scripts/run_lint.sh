#!/bin/bash
# Run golangci-lint locally to match the CI environment

set -e

echo "🔍 Running golangci-lint..."

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo "❌ golangci-lint is not installed"
    echo "Install it with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$(go env GOPATH)/bin"
    exit 1
fi

# Check if .golangci.yml exists
if [ ! -f .golangci.yml ]; then
    echo "❌ .golangci.yml not found in project root"
    exit 1
fi

# Run golangci-lint with the same arguments as CI
golangci-lint run -v --config=.golangci.yml --max-same-issues=0

echo "✅ golangci-lint passed!"
