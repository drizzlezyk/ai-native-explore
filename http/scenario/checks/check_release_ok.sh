#!/bin/bash
# Validates DELETE /v1/cloud/pod/:id returns success
RESPONSE_FILE="$1"

code=$(jq -r '.code' "$RESPONSE_FILE")
if [ "$code" != "" ]; then
    echo "  [FAIL] release pod failed, code: $code" >&2
    exit 1
fi

echo "  [CHECK] pod release OK"
exit 0
