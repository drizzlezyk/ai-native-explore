#!/bin/bash
# Run unit tests with coverage check locally to match the CI environment

set -e

echo "🧪 Running unit tests with coverage..."

# Check if tests directory exists
if [ ! -d tests ]; then
    echo "❌ tests directory not found"
    exit 1
fi

# Check if test script exists
if [ ! -f tests/test_coverage_get.sh ]; then
    echo "❌ tests/test_coverage_get.sh not found"
    exit 1
fi

# Run the test coverage script
cd tests
bash -x test_coverage_get.sh
cd ..

# Check if cover.out was generated
if [ ! -f cover.out ]; then
    echo "❌ cover.out not generated"
    exit 1
fi

echo "📊 Checking test coverage..."

# Check if go-test-coverage is installed
if command -v go-test-coverage &> /dev/null; then
    # Use go-test-coverage if available
    go-test-coverage --config=.github/workflows/.testcoverage.yml
else
    # Fallback to basic coverage report
    echo "⚠️  go-test-coverage not installed, showing basic coverage report"
    echo "Install it with: go install github.com/vladopajic/go-test-coverage/v2@latest"

    # Calculate total coverage
    total_coverage=$(go tool cover -func=cover.out | grep total | awk '{print $3}' | sed 's/%//')
    echo "Total coverage: ${total_coverage}%"

    # Check against minimum threshold (0.2% from .testcoverage.yml)
    min_coverage=0.2
    if (( $(echo "$total_coverage < $min_coverage" | bc -l) )); then
        echo "❌ Coverage ${total_coverage}% is below minimum ${min_coverage}%"
        exit 1
    fi
fi

echo "✅ Unit tests and coverage check passed!"
