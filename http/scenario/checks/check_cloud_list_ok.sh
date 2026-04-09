#!/bin/bash
# Validates GET /v1/cloud returns a list with at least one cloud config
RESPONSE_FILE="$1"

code=$(jq -r '.code' "$RESPONSE_FILE")
if [ "$code" != "" ]; then
    echo "  [FAIL] expected empty code, got: $code" >&2
    exit 1
fi

count=$(jq -r '.data | length' "$RESPONSE_FILE")
if [ "$count" -eq 0 ]; then
    echo "  [FAIL] cloud list is empty" >&2
    exit 1
fi

echo "  [CHECK] cloud list OK ($count configs found)"
exit 0
