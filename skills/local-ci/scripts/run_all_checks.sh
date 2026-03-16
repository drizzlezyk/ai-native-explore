#!/bin/bash
# Run all CI checks locally in sequence

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "=================================="
echo "🚀 Running all CI checks locally"
echo "=================================="
echo ""

# Check prerequisites first
echo "Step 1: Checking prerequisites..."
echo ""

if ! bash "$SCRIPT_DIR/check_prerequisites.sh"; then
    echo ""
    echo "❌ Prerequisites check failed!"
    echo ""
    echo "Please install missing tools first:"
    echo "  bash .claude/local-ci/scripts/install_tools.sh"
    echo ""
    echo "Or install individual tools manually as shown above."
    exit 1
fi

echo ""
echo "Step 2: Running CI checks..."
echo ""

# Track failures
FAILED_CHECKS=()
PASSED_CHECKS=()

# Function to run a check
run_check() {
    local check_name=$1
    local script_name=$2

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "▶️  Running: $check_name"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if bash "$SCRIPT_DIR/$script_name"; then
        PASSED_CHECKS+=("$check_name")
        echo ""
    else
        FAILED_CHECKS+=("$check_name")
        echo "❌ $check_name failed!"
        echo ""
    fi
}

# Run all checks
run_check "golangci-lint" "run_lint.sh"
run_check "typos check" "run_typos.sh"
run_check "gosec security scan" "run_gosec.sh"
run_check "unit tests" "run_tests.sh"
run_check "docker build" "run_docker_build.sh"

# Summary
echo "=================================="
echo "📊 CI Checks Summary"
echo "=================================="
echo ""

if [ ${#PASSED_CHECKS[@]} -gt 0 ]; then
    echo "✅ Passed (${#PASSED_CHECKS[@]}):"
    for check in "${PASSED_CHECKS[@]}"; do
        echo "   - $check"
    done
    echo ""
fi

if [ ${#FAILED_CHECKS[@]} -gt 0 ]; then
    echo "❌ Failed (${#FAILED_CHECKS[@]}):"
    for check in "${FAILED_CHECKS[@]}"; do
        echo "   - $check"
    done
    echo ""
    echo "Run individual checks to see detailed error messages:"
    for check in "${FAILED_CHECKS[@]}"; do
        case "$check" in
            "golangci-lint") echo "   bash scripts/run_lint.sh" ;;
            "typos check") echo "   bash scripts/run_typos.sh" ;;
            "gosec security scan") echo "   bash scripts/run_gosec.sh" ;;
            "unit tests") echo "   bash scripts/run_tests.sh" ;;
            "docker build") echo "   bash scripts/run_docker_build.sh" ;;
        esac
    done
    exit 1
fi

echo "🎉 All CI checks passed!"
