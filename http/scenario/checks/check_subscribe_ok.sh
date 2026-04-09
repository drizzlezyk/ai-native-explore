#!/bin/bash
# Validates POST /v1/cloud/subscribe returns success
RESPONSE_FILE="$1"

code=$(jq -r '.code' "$RESPONSE_FILE")
if [ "$code" != "" ]; then
    echo "  [FAIL] subscribe failed, code: $code" >&2
    jq -r '.msg' "$RESPONSE_FILE" >&2
    exit 1
fi

echo "  [CHECK] subscribe OK"
exit 0
