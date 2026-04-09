#!/bin/bash
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
HTTP_DIR=$(cd "$SCRIPT_DIR/.." && pwd)

# Source HTTP helper (provides http_with_cookie function)
source "$HTTP_DIR/http_with_cookie.sh"
# Source WS helper (provides ws_with_cookie function)
if [ -f "$HTTP_DIR/ws_with_cookie.sh" ]; then
    source "$HTTP_DIR/ws_with_cookie.sh"
fi

COOKIE_JAR="$HTTP_DIR/cookies.txt"

SCENARIOS_FILE="${1:-$SCRIPT_DIR/scenarios.json}"
SCENARIO_FILTER="${2:-}"

# Dependency check
for cmd in curl jq; do
    command -v "$cmd" >/dev/null 2>&1 || { echo "ERROR: $cmd is required" >&2; exit 1; }
done

# ---- Globals ----
declare -A SCENARIO_VARS
HTTP_STATUS=""
RESPONSE_BODY=""

BASE_URL=$(jq -r '.base_url' "$SCENARIOS_FILE")

PASS_COUNT=0
FAIL_COUNT=0
declare -a RESULTS=()

# ---- Variable substitution ----
substitute_vars() {
    local text="$1"
    for key in "${!SCENARIO_VARS[@]}"; do
        text="${text//\$$key/${SCENARIO_VARS[$key]}}"
    done
    echo "$text"
}

# ---- HTTP request ----
# Sets globals: RESPONSE_BODY, HTTP_STATUS
do_request() {
    local method="$1"
    local url="$2"
    local body="${3:-}"

    local full_url="${BASE_URL}${url}"

    if [ -n "$body" ]; then
        RESPONSE_BODY=$(http_with_cookie "$method" "$full_url" "$body")
    else
        RESPONSE_BODY=$(http_with_cookie "$method" "$full_url")
    fi

    HTTP_STATUS=$(echo "$RESPONSE_BODY" | grep "HTTP_STATUS:" | tail -1 | cut -d: -f2)
    RESPONSE_BODY=$(echo "$RESPONSE_BODY" | grep -v "HTTP_STATUS:")
}

# ---- WebSocket request ----
# Uses ws_with_cookie (from ws_with_cookie.sh) which sends _U_T_ JWT
# as Sec-Websocket-Protocol header.
# Sets globals: RESPONSE_BODY, HTTP_STATUS
do_ws_request() {
    local url="$1"
    local ws_url="${BASE_URL/http:\/\//ws://}${url}"
    local response_file

    response_file=$(mktemp "/tmp/xihe-scenario-ws-$$-XXXX.json")

    RESPONSE_BODY=$(ws_with_cookie "$ws_url" "$response_file") && HTTP_STATUS="101" || HTTP_STATUS="000"
    [ -f "$response_file" ] && rm -f "$response_file"
}

# ---- Run a single step ----
run_step() {
    local scenario_idx="$1"
    local step_idx="$2"

    local name method url body extract_var extract_path check_script
    name=$(jq -r ".scenarios[$scenario_idx].steps[$step_idx].name" "$SCENARIOS_FILE")
    method=$(jq -r ".scenarios[$scenario_idx].steps[$step_idx].method" "$SCENARIOS_FILE")
    url=$(jq -r ".scenarios[$scenario_idx].steps[$step_idx].url" "$SCENARIOS_FILE")
    body=$(jq -r ".scenarios[$scenario_idx].steps[$step_idx].body // empty" "$SCENARIOS_FILE")
    extract_var=$(jq -r ".scenarios[$scenario_idx].steps[$step_idx].extract.variable // empty" "$SCENARIOS_FILE")
    extract_path=$(jq -r ".scenarios[$scenario_idx].steps[$step_idx].extract.jq_path // empty" "$SCENARIOS_FILE")
    check_script=$(jq -r ".scenarios[$scenario_idx].steps[$step_idx].check_script // empty" "$SCENARIOS_FILE")

    # Substitute variables
    url=$(substitute_vars "$url")
    [ -n "$body" ] && body=$(substitute_vars "$body")

    echo "  [STEP] $name ($method $url)"

    if [ "$method" = "WS" ]; then
        do_ws_request "$url"
        if [ "$HTTP_STATUS" != "101" ] || [ -z "$RESPONSE_BODY" ]; then
            echo "  [FAIL] WebSocket failed (status=$HTTP_STATUS)" >&2
            return 1
        fi
    else
        do_request "$method" "$url" "$body"
        if [[ ! "$HTTP_STATUS" =~ ^2 ]]; then
            echo "  [FAIL] HTTP $HTTP_STATUS: $RESPONSE_BODY" >&2
            return 1
        fi
        # Allow empty response body for HTTP 204 No Content
        if [ -z "$RESPONSE_BODY" ] && [ "$HTTP_STATUS" != "204" ]; then
            echo "  [FAIL] HTTP $HTTP_STATUS: empty response body" >&2
            return 1
        fi
    fi

    # Extract variable
    if [ -n "$extract_var" ] && [ -n "$extract_path" ]; then
        local value
        value=$(echo "$RESPONSE_BODY" | jq -r "$extract_path")
        if [ -z "$value" ] || [ "$value" = "null" ]; then
            echo "  [FAIL] Failed to extract $extract_var via $extract_path" >&2
            return 1
        fi
        SCENARIO_VARS["$extract_var"]="$value"
        export "$extract_var"="$value"
        echo "  [VAR]  $extract_var=$value"
    fi

    # Run check script
    if [ -n "$check_script" ]; then
        local response_file
        response_file=$(mktemp "/tmp/xihe-scenario-response-$$-${scenario_idx}-${step_idx}-XXXX.json")
        echo "$RESPONSE_BODY" > "$response_file"
        # Export all scenario vars for check scripts
        for key in "${!SCENARIO_VARS[@]}"; do
            export "$key"="${SCENARIO_VARS[$key]}"
        done
        if ! bash "$SCRIPT_DIR/$check_script" "$response_file"; then
            echo "  [FAIL] Check script failed: $check_script" >&2
            rm -f "$response_file"
            return 1
        fi
        rm -f "$response_file"
    fi

    echo "  [PASS] $name"
    return 0
}

# ---- Run a scenario ----
run_scenario() {
    local scenario_idx="$1"

    local name description final_verify cleanup_script_path
    name=$(jq -r ".scenarios[$scenario_idx].name" "$SCENARIOS_FILE")
    description=$(jq -r ".scenarios[$scenario_idx].description" "$SCENARIOS_FILE")
    final_verify=$(jq -r ".scenarios[$scenario_idx].final_verify_script // empty" "$SCENARIOS_FILE")
    cleanup_script_path=$(jq -r ".scenarios[$scenario_idx].cleanup_script // empty" "$SCENARIOS_FILE")

    echo ""
    echo "===== Scenario: $name ====="
    echo "  $description"

    # Reset variable store
    unset SCENARIO_VARS
    declare -gA SCENARIO_VARS

    local step_count failed_step=""
    step_count=$(jq ".scenarios[$scenario_idx].steps | length" "$SCENARIOS_FILE")

    local scenario_failed=false
    for (( si=0; si<step_count; si++ )); do
        if ! run_step "$scenario_idx" "$si"; then
            step_name=$(jq -r ".scenarios[$scenario_idx].steps[$si].name" "$SCENARIOS_FILE")
            failed_step="$step_name"
            scenario_failed=true
            break
        fi
    done

    # Final verify
    if [ "$scenario_failed" = false ] && [ -n "$final_verify" ] && [ -f "$SCRIPT_DIR/$final_verify" ]; then
        echo "  [VERIFY] Running final verify..."
        for key in "${!SCENARIO_VARS[@]}"; do export "$key"="${SCENARIO_VARS[$key]}"; done
        if ! bash "$SCRIPT_DIR/$final_verify"; then
            echo "  [FAIL] Final verify failed" >&2
            scenario_failed=true
        fi
    fi

    # Cleanup
    if [ -n "$cleanup_script_path" ] && [ -f "$SCRIPT_DIR/$cleanup_script_path" ]; then
        echo "  [CLEANUP] Running cleanup..."
        for key in "${!SCENARIO_VARS[@]}"; do export "$key"="${SCENARIO_VARS[$key]}"; done
        export COOKIE_JAR="$COOKIE_JAR"
        bash "$SCRIPT_DIR/$cleanup_script_path" || true
    fi

    if [ "$scenario_failed" = true ]; then
        echo "===== FAILED: $name (failed at: ${failed_step:-verify}) ====="
        RESULTS+=("FAILED: $name")
        (( FAIL_COUNT++ )) || true
    else
        echo "===== PASSED: $name ====="
        RESULTS+=("PASSED: $name")
        (( PASS_COUNT++ )) || true
    fi
}

# ---- Main ----
main() {
    local scenarios_count
    scenarios_count=$(jq '.scenarios | length' "$SCENARIOS_FILE")

    if [ "$scenarios_count" -eq 0 ]; then
        echo "ERROR: No scenarios defined in $SCENARIOS_FILE" >&2
        exit 1
    fi

    # --list flag
    if [ "$SCENARIO_FILTER" = "--list" ]; then
        echo "Available scenarios:"
        for (( i=0; i<scenarios_count; i++ )); do
            local name desc
            name=$(jq -r ".scenarios[$i].name" "$SCENARIOS_FILE")
            desc=$(jq -r ".scenarios[$i].description" "$SCENARIOS_FILE")
            echo "  [$i] $name - $desc"
        done
        exit 0
    fi

    if [ -n "$SCENARIO_FILTER" ]; then
        # Try by name first
        local idx
        idx=$(jq -r --arg n "$SCENARIO_FILTER" '.scenarios | to_entries[] | select(.value.name == $n) | .key' "$SCENARIOS_FILE" | head -1)
        if [ -z "$idx" ]; then
            # Try by index
            if [[ "$SCENARIO_FILTER" =~ ^[0-9]+$ ]]; then
                idx="$SCENARIO_FILTER"
            else
                echo "ERROR: Scenario '$SCENARIO_FILTER' not found" >&2
                exit 1
            fi
        fi
        run_scenario "$idx"
    else
        # Run all
        for (( i=0; i<scenarios_count; i++ )); do
            run_scenario "$i"
        done
    fi

    echo ""
    echo "===== RESULTS ====="
    for r in "${RESULTS[@]}"; do echo "  $r"; done
    echo "  Total: $PASS_COUNT passed, $FAIL_COUNT failed"

    [ "$FAIL_COUNT" -eq 0 ] || exit 1
}

main
