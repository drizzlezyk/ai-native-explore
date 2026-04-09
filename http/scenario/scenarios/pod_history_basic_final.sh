#!/bin/bash
# Final verification for pod_history_basic scenario

if [ -z "${FIRST_POD_ID:-}" ]; then
    echo "  [FAIL] FIRST_POD_ID was not extracted" >&2
    exit 1
fi

if ! [[ "$FIRST_POD_ID" =~ ^[0-9a-f-]{36}$ ]]; then
    echo "  [FAIL] FIRST_POD_ID does not look like a UUID: $FIRST_POD_ID" >&2
    exit 1
fi

echo "  [VERIFY] FIRST_POD_ID=$FIRST_POD_ID is a valid UUID"
exit 0
