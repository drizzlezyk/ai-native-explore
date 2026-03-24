#!/bin/bash
# Run all CI checks locally in sequence, generate report, and cleanup

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
REPORT_FILE="$PROJECT_ROOT/CI_FIXES_REPORT.md"
TEMP_DIR="$PROJECT_ROOT/.ci-temp"

# Create temp directory for collecting outputs
mkdir -p "$TEMP_DIR"

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

# Track failures and details
FAILED_CHECKS=()
PASSED_CHECKS=()
declare -A CHECK_ERRORS
declare -A CHECK_FIXES

# Function to run a check and capture errors
run_check() {
    local check_name=$1
    local script_name=$2
    local error_log="$TEMP_DIR/${check_name// /_}_error.log"

    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "▶️  Running: $check_name"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    if bash "$SCRIPT_DIR/$script_name" 2>&1 | tee "$error_log"; then
        PASSED_CHECKS+=("$check_name")
        rm -f "$error_log"  # Clean up log if passed
        echo ""
    else
        FAILED_CHECKS+=("$check_name")

        # Extract error details
        local error_content=$(cat "$error_log" | grep -E "(error|Error|ERROR|failed|FAIL|warning)" | head -20 || echo "Check failed - see detailed logs")
        CHECK_ERRORS["$check_name"]="$error_content"

        # Provide fix suggestions based on check type
        case "$check_name" in
            "golangci-lint")
                CHECK_FIXES["$check_name"]="Run 'gofmt -w .' for formatting issues. Remove unused variables or prefix with '_'. Add proper error handling for unchecked errors."
                ;;
            "typos check")
                CHECK_FIXES["$check_name"]="Run 'typos --write-changes ./' to auto-fix typos. Add false positives to typos.toml if needed."
                ;;
            "gosec security scan")
                CHECK_FIXES["$check_name"]="Fix security issues: use SHA256 instead of MD5, use parameterized SQL queries, validate file paths with filepath.Clean()."
                ;;
            "unit tests")
                CHECK_FIXES["$check_name"]="Fix failing tests: run 'go test -v -run TestName' to debug. Check test coverage with 'go tool cover -html=cover.out'."
                ;;
            "docker build")
                CHECK_FIXES["$check_name"]="Check Dockerfile syntax. Set GITHUB_USER and GITHUB_TOKEN if accessing private repos. Verify go.mod dependencies."
                ;;
        esac

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

# Generate CI Fixes Report
echo "=================================="
echo "📝 Generating CI Fixes Report"
echo "=================================="
echo ""

{
    echo "# CI Fixes Report"
    echo ""
    echo "**Generated:** $(date '+%Y-%m-%d %H:%M:%S')"
    echo "**Project:** $(basename $PROJECT_ROOT)"
    echo ""
    echo "---"
    echo ""

    echo "## Summary"
    echo ""
    echo "- ✅ **Passed:** ${#PASSED_CHECKS[@]} checks"
    echo "- ❌ **Failed:** ${#FAILED_CHECKS[@]} checks"
    echo ""

    if [ ${#PASSED_CHECKS[@]} -gt 0 ]; then
        echo "### Passed Checks"
        echo ""
        for check in "${PASSED_CHECKS[@]}"; do
            echo "- ✅ $check"
        done
        echo ""
    fi

    if [ ${#FAILED_CHECKS[@]} -gt 0 ]; then
        echo "### Failed Checks"
        echo ""
        for check in "${FAILED_CHECKS[@]}"; do
            echo "- ❌ $check"
        done
        echo ""
        echo "---"
        echo ""

        echo "## Error Details and Fixes"
        echo ""

        for check in "${FAILED_CHECKS[@]}"; do
            echo "### ❌ $check"
            echo ""
            echo "**Error Message:**"
            echo '```'
            echo "${CHECK_ERRORS[$check]}"
            echo '```'
            echo ""
            echo "**Recommended Fix:**"
            echo ""
            echo "${CHECK_FIXES[$check]}"
            echo ""
            echo "---"
            echo ""
        done
    else
        echo "🎉 **All CI checks passed!**"
        echo ""
    fi

    echo "## CI Check Details"
    echo ""
    echo "| Check | Description | Script |"
    echo "|-------|-------------|--------|"
    echo "| golangci-lint | Static code analysis | \`run_lint.sh\` |"
    echo "| typos check | Spell checking | \`run_typos.sh\` |"
    echo "| gosec | Security scanning | \`run_gosec.sh\` |"
    echo "| unit tests | Test execution | \`run_tests.sh\` |"
    echo "| docker build | Image build | \`run_docker_build.sh\` |"
    echo ""

} > "$REPORT_FILE"

echo "✅ Report saved to: $REPORT_FILE"
echo ""

# Clean up temporary files
echo "=================================="
echo "🧹 Cleaning up temporary files"
echo "=================================="
echo ""

# Remove temp directory
if [ -d "$TEMP_DIR" ]; then
    rm -rf "$TEMP_DIR"
    echo "✅ Removed: $TEMP_DIR"
fi

# Remove common CI artifacts
CI_ARTIFACTS=(
    "cover.out"
    "coverage.txt"
    "gosec-results.json"
    ".testcoverage.html"
)

for artifact in "${CI_ARTIFACTS[@]}"; do
    if [ -f "$PROJECT_ROOT/$artifact" ]; then
        rm -f "$PROJECT_ROOT/$artifact"
        echo "✅ Removed: $artifact"
    fi
done

echo ""
echo "=================================="
echo "📊 CI Checks Complete"
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
    echo "📄 See detailed error report: $REPORT_FILE"
    echo ""
    exit 1
fi

echo "🎉 All CI checks passed!"
echo "📄 Full report saved to: $REPORT_FILE"
echo ""
