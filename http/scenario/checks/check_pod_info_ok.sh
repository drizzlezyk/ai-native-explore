#!/bin/bash
# Validates GET /v1/cloud/pod/:cid returns pod info with access_url
RESPONSE_FILE="$1"

code=$(jq -r '.code' "$RESPONSE_FILE")
if [ "$code" != "" ]; then
    echo "  [FAIL] get pod info failed, code: $code" >&2
    exit 1
fi

access_url=$(jq -r '.data.access_url' "$RESPONSE_FILE")
if [ -z "$access_url" ] || [ "$access_url" = "null" ]; then
    echo "  [FAIL] no access_url in pod info" >&2
    exit 1
fi

echo "  [CHECK] pod info OK (access_url=$access_url)"
exit 0
