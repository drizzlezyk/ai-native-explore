#!/bin/bash
# Run gosec security scanner locally to match the CI environment

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"
GOSEC_REPORT="$PROJECT_ROOT/.ci-temp/gosec-report.json"
GOSEC_BASELINE="$PROJECT_ROOT/.ci-temp/gosec-baseline.json"

echo "🔒 Running gosec security scanner..."

# Check if gosec is installed
if ! command -v gosec &> /dev/null; then
    echo "❌ gosec is not installed"
    echo "Install it with: go install github.com/securego/gosec/v2/cmd/gosec@latest"
    exit 1
fi

# Create temp directory
mkdir -p "$PROJECT_ROOT/.ci-temp"

# Save current report as baseline if it doesn't exist
if [ ! -f "$GOSEC_BASELINE" ]; then
    echo "📝 Creating baseline security scan..."
    gosec -fmt=json -out="$GOSEC_BASELINE" ./... 2>/dev/null || true
fi

# Run gosec and save detailed JSON report
echo "🔍 Scanning for security issues..."
if gosec -fmt=json -out="$GOSEC_REPORT" ./... 2>&1; then
    echo "✅ gosec security scan passed!"

    # Check if there are any issues in the report
    ISSUE_COUNT=$(jq -r '.Issues | length' "$GOSEC_REPORT" 2>/dev/null || echo "0")
    if [ "$ISSUE_COUNT" -gt 0 ]; then
        echo "⚠️  Found $ISSUE_COUNT security issue(s) (non-blocking)"
        echo "📄 Detailed report saved to: $GOSEC_REPORT"
    fi

    exit 0
else
    echo "❌ gosec found security issues!"

    # Count issues
    ISSUE_COUNT=$(jq -r '.Issues | length' "$GOSEC_REPORT" 2>/dev/null || echo "unknown")
    echo "📊 Total issues: $ISSUE_COUNT"
    echo "📄 Detailed report saved to: $GOSEC_REPORT"

    # Show summary
    echo ""
    echo "Issue summary:"
    jq -r '.Issues[] | "  - [\(.severity)] \(.rule_id): \(.file):\(.line)"' "$GOSEC_REPORT" 2>/dev/null | head -10

    if [ "$ISSUE_COUNT" -gt 10 ]; then
        echo "  ... and $((ISSUE_COUNT - 10)) more"
    fi

    echo ""
    echo "💡 To auto-fix security issues, run:"
    echo "   bash .claude/skills/local-ci/scripts/fix_security_issues.sh"

    exit 1
fi
