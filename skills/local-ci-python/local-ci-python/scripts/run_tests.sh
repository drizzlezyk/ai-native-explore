#!/bin/bash
# Run unit tests with coverage validation
# Checks: 10% baseline coverage, 80% incremental coverage for changed code

set -euo pipefail

echo "🧪 Running unit tests with coverage validation..."
echo ""

# Configuration
BASELINE_COVERAGE=10    # Minimum overall coverage (%)
INCREMENTAL_COVERAGE=80 # Minimum coverage for changed code (%)

# Determine Python command
if command -v python3 &> /dev/null; then
    PYTHON_CMD="python3"
else
    PYTHON_CMD="python"
fi

# Check if Python is installed
if ! command -v $PYTHON_CMD &> /dev/null; then
    echo "❌ Python is not installed"
    exit 1
fi

# Check if pytest is installed for the selected Python interpreter
if ! $PYTHON_CMD -m pytest --version > /dev/null 2>&1; then
    echo "❌ pytest is not installed"
    echo ""
    echo "Install with:"
    echo "  pip install pytest pytest-cov"
    echo ""
    echo "Or run:"
    echo "  bash .claude/skills/local-ci-python/scripts/install_tools.sh"
    exit 1
fi

# Check if pytest-cov is installed
if ! $PYTHON_CMD -c "import pytest_cov" 2>/dev/null; then
    echo "❌ pytest-cov is not installed"
    echo ""
    echo "Install with:"
    echo "  pip install pytest-cov"
    echo ""
    echo "Or run:"
    echo "  bash .claude/skills/local-ci-python/scripts/install_tools.sh"
    exit 1
fi

# Run tests with coverage
echo "Running tests..."
# Remove stale coverage artifacts to avoid reading old successful results
rm -f coverage.xml .coverage

if ! $PYTHON_CMD -m pytest -p pytest_cov --cov=. --cov-report=term --cov-report=xml --cov-report=html -v 2>&1 | tee test_output.txt; then
    echo ""
    echo "❌ Tests failed!"
    echo ""
    echo "To debug:"
    echo "  - Check test_output.txt for details"
    echo "  - Run specific test: pytest -v tests/test_name.py::test_function"
    rm -f test_output.txt
    exit 1
fi

echo ""

# Ensure test run actually executed tests
if grep -E -q "collected 0 items|no tests ran" test_output.txt; then
    echo "❌ No tests were collected."
    echo ""
    echo "Add tests under tests/ or files matching test_*.py"
    rm -f test_output.txt
    exit 1
fi

# Check if coverage file was generated
if [ ! -f coverage.xml ]; then
    echo "❌ coverage.xml not generated"
    exit 1
fi

# Calculate overall coverage from XML file
echo "📊 Analyzing coverage..."

# Extract line-rate from coverage.xml (line-rate is a decimal like 0.12 for 12%)
total_coverage=$($PYTHON_CMD -c "
import xml.etree.ElementTree as ET
try:
    tree = ET.parse('coverage.xml')
    root = tree.getroot()
    line_rate = float(root.attrib['line-rate'])
    print(f'{line_rate * 100:.2f}')
except Exception as e:
    print('0')
    exit(1)
")

if [ -z "$total_coverage" ] || [ "$total_coverage" = "0" ]; then
    echo "❌ Failed to calculate coverage"
    exit 1
fi

echo "Overall coverage: ${total_coverage}%"

# Check baseline coverage without external bc dependency
if $PYTHON_CMD -c "import sys; sys.exit(0 if float('$total_coverage') < float('$BASELINE_COVERAGE') else 1)"; then
    echo "❌ Coverage ${total_coverage}% is below baseline ${BASELINE_COVERAGE}%"
    echo ""
    echo "To improve coverage:"
    echo "  1. View coverage report: open htmlcov/index.html"
    echo "  2. Or run: coverage report --show-missing"
    echo "  3. Add tests for uncovered functions"
    exit 1
fi

echo "✅ Baseline coverage check passed (${total_coverage}% >= ${BASELINE_COVERAGE}%)"
echo ""

# Check incremental coverage (for changed files)
if git rev-parse --git-dir > /dev/null 2>&1; then
    echo "📈 Checking incremental coverage for changed files..."

    # Get list of changed Python files (excluding test files)
    changed_files=$(git diff --name-only --diff-filter=ACM HEAD | grep '\.py$' | grep -v '^test_' | grep -v '/test_' || true)

    if [ -z "$changed_files" ]; then
        echo "ℹ️  No changed Python files found (excluding tests)"
        echo "✅ Incremental coverage check skipped"
    else
        echo "Changed files:"
        echo "$changed_files" | sed 's/^/  - /'
        echo ""

        # Extract coverage for changed files using coverage.xml
        failed_files=()

        while IFS= read -r file; do
            if [ -f "$file" ]; then
                # Get coverage for this file from coverage.xml
                file_coverage=$($PYTHON_CMD -c "
import xml.etree.ElementTree as ET
import sys

try:
    tree = ET.parse('coverage.xml')
    root = tree.getroot()

    # Find the file in the coverage report
    for package in root.findall('.//package'):
        for class_elem in package.findall('.//class'):
            filename = class_elem.attrib.get('filename', '')
            # Normalize paths for comparison
            if filename.replace('./', '') == '$file'.replace('./', ''):
                line_rate = float(class_elem.attrib.get('line-rate', 0))
                print(f'{line_rate * 100:.2f}')
                sys.exit(0)

    # File not in coverage report
    print('0')
except Exception as e:
    print('0')
    sys.exit(1)
")

                if [ -z "$file_coverage" ] || [ "$file_coverage" = "0" ] || [ "$file_coverage" = "0.00" ]; then
                    echo "  ⚠️  $file: no coverage data"
                    failed_files+=("$file (no coverage)")
                elif $PYTHON_CMD -c "import sys; sys.exit(0 if float('$file_coverage') < float('$INCREMENTAL_COVERAGE') else 1)"; then
                    echo "  ❌ $file: ${file_coverage}% (< ${INCREMENTAL_COVERAGE}%)"
                    failed_files+=("$file (${file_coverage}%)")
                else
                    echo "  ✅ $file: ${file_coverage}%"
                fi
            fi
        done <<< "$changed_files"

        echo ""

        if [ ${#failed_files[@]} -gt 0 ]; then
            echo "❌ Incremental coverage check failed for ${#failed_files[@]} file(s):"
            for file in "${failed_files[@]}"; do
                echo "  - $file"
            done
            echo ""
            echo "To improve incremental coverage:"
            echo "  1. View coverage: open htmlcov/index.html"
            echo "  2. Focus on changed files listed above"
            echo "  3. Add tests to reach ${INCREMENTAL_COVERAGE}% coverage"
            exit 1
        fi

        echo "✅ Incremental coverage check passed (all changed files >= ${INCREMENTAL_COVERAGE}%)"
    fi
else
    echo "ℹ️  Not a git repository - skipping incremental coverage check"
fi

echo ""

# Generate coverage summary
echo "📄 Coverage summary:"
if command -v coverage &> /dev/null; then
    coverage report | tail -20
else
    echo "Install 'coverage' tool for detailed reports: pip install coverage"
fi

echo ""
echo "✅ All coverage checks passed!"
echo ""
echo "View detailed coverage:"
echo "  open htmlcov/index.html"
echo "  or: coverage report --show-missing"

# Cleanup
rm -f test_output.txt
