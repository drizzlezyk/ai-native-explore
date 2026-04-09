#!/bin/bash
# Validates that page_size=5 returns at most 5 items
RESPONSE_FILE="$1"

code=$(jq -r '.code' "$RESPONSE_FILE")
if [ "$code" != "" ]; then
    echo "  [FAIL] expected empty code, got: $code" >&2
    exit 1
fi

count=$(jq -r '.data.data | length' "$RESPONSE_FILE")
if [ "$count" -gt 5 ]; then
    echo "  [FAIL] expected at most 5 items, got: $count" >&2
    exit 1
fi

echo "  [CHECK] pagination OK (got $count items, max 5)"
exit 0
