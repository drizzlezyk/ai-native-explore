#!/bin/bash
# Validates all items match the expected image
RESPONSE_FILE="$1"
EXPECTED_IMAGE="${EXPECTED_IMAGE:-python3.9-ms2.7.1-cann8.3.RC1}"

code=$(jq -r '.code' "$RESPONSE_FILE")
if [ "$code" != "" ]; then
    echo "  [FAIL] expected empty code, got: $code" >&2
    exit 1
fi

bad_count=$(jq --arg img "$EXPECTED_IMAGE" \
    '[.data.data[] | select(.image != $img)] | length' \
    "$RESPONSE_FILE")

if [ "$bad_count" -gt 0 ]; then
    echo "  [FAIL] $bad_count items have wrong image (expected $EXPECTED_IMAGE)" >&2
    exit 1
fi

total=$(jq -r '.data.data | length' "$RESPONSE_FILE")
echo "  [CHECK] image filter OK (all $total items have image=$EXPECTED_IMAGE)"
exit 0
