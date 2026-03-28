#!/bin/bash
# Run bandit security scanner to detect vulnerabilities

set -e

# Parse arguments
JSON_OUTPUT=""
QUIET=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --json-output)
            JSON_OUTPUT="$2"
            shift 2
            ;;
        --quiet)
            QUIET=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

if [ "$QUIET" = false ]; then
    echo "🔒 Running security scan with bandit..."
    echo ""
fi

# Check if bandit is installed
if ! command -v bandit &> /dev/null; then
    echo "❌ bandit is not installed"
    echo ""
    echo "Install with:"
    echo "  pip install bandit"
    echo ""
    echo "Or run:"
    echo "  bash .claude/skills/local-ci-python/scripts/install_tools.sh"
    exit 1
fi

# Check for custom bandit configuration
CONFIG_ARG=""
if [ -f .bandit ]; then
    [ "$QUIET" = false ] && echo "Using custom configuration: .bandit"
    CONFIG_ARG="-c .bandit"
elif [ -f .bandit.yaml ]; then
    [ "$QUIET" = false ] && echo "Using custom configuration: .bandit.yaml"
    CONFIG_ARG="-c .bandit.yaml"
else
    [ "$QUIET" = false ] && echo "Using default bandit configuration"
fi

[ "$QUIET" = false ] && echo ""

# Determine output format
if [ -n "$JSON_OUTPUT" ]; then
    # JSON output mode
    [ "$QUIET" = false ] && echo "Scanning for security vulnerabilities (JSON output)..."

    # Create parent directory if needed
    JSON_DIR=$(dirname "$JSON_OUTPUT")
    mkdir -p "$JSON_DIR"

    if bandit $CONFIG_ARG -r . -f json -o "$JSON_OUTPUT" 2>&1 | grep -v "Test results written to"; then
        [ "$QUIET" = false ] && echo ""
        [ "$QUIET" = false ] && echo "✅ No security issues found!"
        [ "$QUIET" = false ] && echo "JSON report: $JSON_OUTPUT"
        exit 0
    else
        EXIT_CODE=$?

        # Ensure JSON file exists even on error
        if [ ! -f "$JSON_OUTPUT" ]; then
            echo '{"errors": [], "results": [], "metrics": {}}' > "$JSON_OUTPUT"
        fi

        [ "$QUIET" = false ] && echo ""
        [ "$QUIET" = false ] && echo "❌ Security issues detected!"
        [ "$QUIET" = false ] && echo "JSON report: $JSON_OUTPUT"

        # Determine Python command
        if command -v python3 &> /dev/null; then
            PYTHON_CMD="python3"
        else
            PYTHON_CMD="python"
        fi

        # Parse issue count from JSON
        ISSUE_COUNT=$($PYTHON_CMD -c "
import json
try:
    with open('$JSON_OUTPUT', 'r') as f:
        data = json.load(f)
        print(len(data.get('results', [])))
except:
    print('0')
" 2>/dev/null || echo "0")
        [ "$QUIET" = false ] && echo "Found $ISSUE_COUNT issue(s)"

        exit $EXIT_CODE
    fi
else
    # Text output mode (original behavior)
    [ "$QUIET" = false ] && echo "Scanning for security vulnerabilities..."

    if bandit $CONFIG_ARG -r . -ll 2>&1 | tee bandit_output.txt; then
        echo ""
        echo "✅ No security issues found!"
        rm -f bandit_output.txt
        exit 0
    else
        EXIT_CODE=$?
        echo ""
        echo "❌ Security issues detected!"
        echo ""
        echo "Common fixes:"
        echo ""
        echo "B105/B106 - Hardcoded passwords:"
        echo "  Use environment variables: password = os.getenv('DB_PASSWORD')"
        echo ""
        echo "B104 - Binding to 0.0.0.0:"
        echo "  Consider binding to localhost for development: app.run(host='127.0.0.1')"
        echo ""
        echo "B608 - SQL injection:"
        echo "  Use parameterized queries: cursor.execute('SELECT * FROM users WHERE id = %s', (user_id,))"
        echo ""
        echo "B602/B605 - Shell injection:"
        echo "  Avoid shell=True: subprocess.run(['ls', '-la']) instead of subprocess.run('ls -la', shell=True)"
        echo ""
        echo "B303/B324 - Weak cryptography:"
        echo "  Use strong algorithms: hashlib.sha256() instead of hashlib.md5()"
        echo ""
        echo "B301-B324 - Unsafe deserialization:"
        echo "  Avoid pickle for untrusted data. Use JSON or other safe formats."
        echo ""
        echo "For detailed fixes, see:"
        echo "  .claude/skills/local-ci-python/references/security-best-practices.md"
        echo ""
        echo "To exclude specific issues (use sparingly):"
        echo "  Add # nosec comment: password = 'temp'  # nosec B105"
        echo "  Or configure .bandit to exclude rules globally"
        echo ""

        rm -f bandit_output.txt
        exit $EXIT_CODE
    fi
fi
