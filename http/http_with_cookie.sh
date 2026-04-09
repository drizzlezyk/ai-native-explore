#!/bin/bash

# Cookie 文件路径
COOKIE_JAR="$(dirname "$0")/cookies.txt"

# curl 参数
CURL_OPTS=(
    --noproxy "*"
    -s
    --connect-timeout 10
    --max-time 30
    -L
)

# 执行 HTTP 请求并自动刷新 Cookie
# 用法: http_with_cookie [METHOD] [URL] [BODY]
#   METHOD  - HTTP 方法 (GET, POST, DELETE等)
#   URL     - 请求的 URL
#   BODY    - 可选的请求体
# 返回: 0 成功, 1 失败
# 输出: 响应内容 或 错误信息
http_with_cookie() {
    local method="${1:-GET}"
    local url="${2:-http://localhost:8092/web/v1/cloud/pod/history}"
    local body="${3:-}"

    local response
    local status
    local header_file

    header_file=$(mktemp "/tmp/xihe-http-headers-$$-XXXX.txt")

    if [ -n "$body" ]; then
        response=$(curl "${CURL_OPTS[@]}" \
            -X "$method" \
            -b "$COOKIE_JAR" \
            -c "$COOKIE_JAR" \
            -D "$header_file" \
            -H "Content-Type: application/json" \
            -d "$body" \
            -w "\nHTTP_STATUS:%{http_code}" \
            "$url")
    else
        response=$(curl "${CURL_OPTS[@]}" \
            -X "$method" \
            -b "$COOKIE_JAR" \
            -c "$COOKIE_JAR" \
            -D "$header_file" \
            -w "\nHTTP_STATUS:%{http_code}" \
            "$url")
    fi

    status=$?

    if [ $status -ne 0 ]; then
        echo "HTTP_STATUS:000"
        return 1
    fi

    http_status=$(echo "$response" | grep "HTTP_STATUS:" | tail -1 | cut -d: -f2)
    response_body=$(echo "$response" | grep -v "HTTP_STATUS:")

    # Check HTTP status
    if [[ ! "$http_status" =~ ^2 ]]; then
        echo "HTTP_STATUS:$http_status"
        [ -n "$response_body" ] && echo "$response_body"
        rm -f "$header_file"
        return 1
    fi

    if [ -s "$header_file" ]; then
        sync_cookie_jar_from_headers "$header_file"
    fi

    # Empty response is valid for HTTP 204 No Content
    if [ -z "$response_body" ] && [ "$http_status" = "204" ]; then
        echo "HTTP_STATUS:$http_status"
        rm -f "$header_file"
        return 0
    fi

    # Empty response response for non-204 status is error
    if [ -z "$response_body" ]; then
        echo "HTTP_STATUS:$http_status"
        rm -f "$header_file"
        return 1
    fi

    echo "$response_body"
    echo "HTTP_STATUS:$http_status"

    rm -f "$header_file"

    return 0
}

sync_cookie_jar_from_headers() {
    local header_file="$1"
    local tmp_cookie
    declare -A updates=()
    declare -A updated_names=()

    while IFS= read -r line; do
        local cookie pair name value
        cookie=${line#Set-Cookie: }
        pair=${cookie%%;*}
        name=${pair%%=*}
        value=${pair#*=}
        [ -n "$name" ] || continue
        updates["$name"]="$value"
        updated_names["$name"]=1
    done < <(grep -i '^Set-Cookie:' "$header_file" || true)

    if [ ${#updates[@]} -eq 0 ]; then
        return 0
    fi

    tmp_cookie=$(mktemp "/tmp/xihe-cookie-jar-$$-XXXX.txt")
    : > "$tmp_cookie"

    while IFS=$'\t' read -r domain flag path secure expire name value; do
        [ -z "$domain" ] && continue
        [[ "$domain" =~ ^# ]] && continue
        [ -z "$name" ] && continue
        if [ -n "${updates[$name]+x}" ]; then
            value="${updates[$name]}"
            unset 'updated_names[$name]'
        fi
        printf '%s\t%s\t%s\t%s\t%s\t%s\t%s\n' "$domain" "$flag" "$path" "$secure" "$expire" "$name" "$value" >> "$tmp_cookie"
    done < "$COOKIE_JAR"

    for name in "${!updated_names[@]}"; do
        printf 'localhost\tFALSE\t/\tFALSE\t18934560000\t%s\t%s\n' "$name" "${updates[$name]}" >> "$tmp_cookie"
    done

    mv "$tmp_cookie" "$COOKIE_JAR"
}

# 如果直接执行脚本，则调用函数
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    http_with_cookie "$@"
fi
