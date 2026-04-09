#!/bin/bash
# Validates WebSocket response from /v1/ws/cloud/pod/:id confirms pod is closed
RESPONSE_FILE="$1"

code=$(jq -r '.code' "$RESPONSE_FILE")
if [ "$code" != "" ]; then
    echo "  [FAIL] pod close notification failed, code: $code" >&2
    jq -r '.msg' "$RESPONSE_FILE" >&2
    exit 1
fi

echo "  [CHECK] pod closed OK"
exit 0
