#!/bin/bash
# WebSocket request helper — sends _U_T_ JWT as Sec-Websocket-Protocol header.
# Streams response body to stdout.
#
# Usage: source http/ws_with_cookie.sh && ws_with_cookie "ws://host/path"
# Returns: response body on stdout, status message on stderr.

ws_with_cookie() {
    local url="${1:-}"

    if [ -z "$url" ]; then
        echo "Usage: ws_with_cookie <ws_url>" >&2
        return 1
    fi

    local cookie_jar="${COOKIE_JAR:-$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/cookies.txt}"

    local ws_token
    ws_token=$(awk '!/^#/ && NF==7 && $6=="_U_T_" {print $7}' "$cookie_jar")
    if [ -z "$ws_token" ]; then
        echo "ws_with_cookie: _U_T_ token not found in $cookie_jar" >&2
        return 1
    fi

    curl --noproxy "*" \
        --connect-timeout 10 \
        --max-time "${WS_TIMEOUT:-120}" \
        -N \
        -s \
        -H "Cookie: $(awk '!/^#/ && NF==7 {printf "%s%s=%s", sep, $6, $7; sep="; "} END {print ""}' "$cookie_jar")" \
        -H "Sec-Websocket-Protocol: $ws_token" \
        "$url" 2>/dev/null
}
