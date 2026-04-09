#!/bin/bash
# Cleanup: if POD_ID is set and pod wasn't released (e.g. test failed mid-way), release it
COOKIE_JAR="${COOKIE_JAR:-}"
BASE_URL="${BASE_URL:-http://localhost:8092}"

if [ -z "${POD_ID:-}" ]; then
    echo "  [CLEANUP] no POD_ID, nothing to clean"
    exit 0
fi

echo "  [CLEANUP] releasing pod $POD_ID"
curl --noproxy "*" -s --connect-timeout 5 \
    -X DELETE \
    -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
    "$BASE_URL/web/v1/cloud/pod/$POD_ID" \
    -o /dev/null || true

echo "  [CLEANUP] done"
exit 0
