#!/bin/bash
# Validates generic pod history success response shape
RESPONSE_FILE="$1"

code=$(jq -r '.code' "$RESPONSE_FILE")
if [ "$code" != "" ]; then
    echo "  [FAIL] expected empty code, got: $code" >&2
    exit 1
fi

data_type=$(jq -r '.data.data | type' "$RESPONSE_FILE")
if [ "$data_type" != "array" ]; then
    echo "  [FAIL] .data.data should be array, got: $data_type" >&2
    exit 1
fi

total=$(jq -r '.data.total' "$RESPONSE_FILE")
if ! [[ "$total" =~ ^[0-9]+$ ]]; then
    echo "  [FAIL] .data.total should be a non-negative integer, got: $total" >&2
    exit 1
fi

echo "  [CHECK] pod history shape OK (total=$total)"
exit 0
