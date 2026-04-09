#!/bin/bash
# Validates WebSocket response from /v1/cloud/:cid contains a pod_id and access_url
RESPONSE_FILE="$1"

code=$(jq -r '.code' "$RESPONSE_FILE")
if [ "$code" != "" ]; then
    echo "  [FAIL] pod not ready, code: $code" >&2
    jq -r '.msg' "$RESPONSE_FILE" >&2
    exit 1
fi

pod_id=$(jq -r '.data.id' "$RESPONSE_FILE")
if [ -z "$pod_id" ] || [ "$pod_id" = "null" ]; then
    echo "  [FAIL] no pod id in response" >&2
    exit 1
fi

access_url=$(jq -r '.data.access_url' "$RESPONSE_FILE")
if [ -z "$access_url" ] || [ "$access_url" = "null" ]; then
    echo "  [FAIL] no access_url in response" >&2
    exit 1
fi

echo "  [CHECK] pod ready (id=$pod_id url=$access_url)"
exit 0
