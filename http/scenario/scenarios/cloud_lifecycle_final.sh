#!/bin/bash
# Final verify: confirm POD_ID was extracted and looks like a UUID
if [ -z "${POD_ID:-}" ]; then
    echo "  [FAIL] POD_ID was not extracted" >&2
    exit 1
fi

if ! [[ "$POD_ID" =~ ^[0-9a-f-]{36}$ ]]; then
    echo "  [FAIL] POD_ID does not look like a UUID: $POD_ID" >&2
    exit 1
fi

echo "  [VERIFY] lifecycle complete (POD_ID=$POD_ID)"
exit 0
