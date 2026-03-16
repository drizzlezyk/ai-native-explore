#!/bin/bash

# Test script for Pod History API endpoint
# GET /v1/cloud/pod/history

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8092}"
ENDPOINT="/web/v1/cloud/pod/history"
XIHE_USERNAME="${XIHE_USERNAME:-testuser}"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print test result
print_result() {
    local test_name="$1"
    local status_code="$2"
    local expected="$3"

    if [ "$status_code" = "$expected" ]; then
        echo -e "${GREEN}✓${NC} $test_name (Status: $status_code)"
    else
        echo -e "${RED}✗${NC} $test_name (Expected: $expected, Got: $status_code)"
    fi
}

# Function to make request with auth bypass (using cookie)
make_request() {
    local query_params="$1"
    local description="$2"
    local expected_status="${3:-200}"

    echo -e "\n${YELLOW}Testing:${NC} $description"
    echo "URL: ${BASE_URL}${ENDPOINT}?${query_params}"

    # Make request with user cookie for auth bypass
    response=$(curl -s -w "\n%{http_code}" \
        -H "Cookie: _U_T_=${XIHE_USERNAME}" \
        "${BASE_URL}${ENDPOINT}?${query_params}")

    # Extract status code (last line)
    status_code=$(echo "$response" | tail -n 1)
    # Extract body (all but last line)
    body=$(echo "$response" | sed '$d')

    print_result "$description" "$status_code" "$expected_status"

    # Print response body if it's JSON
    if echo "$body" | jq . >/dev/null 2>&1; then
        echo "Response:"
        echo "$body" | jq .
    else
        echo "Response: $body"
    fi
}

echo "=================================================="
echo "Testing Pod History API: GET /v1/cloud/pod/history"
echo "=================================================="
echo "Base URL: $BASE_URL"
echo "Using username: $XIHE_USERNAME"
echo "=================================================="

# Test Case 1: Basic request with minimal parameters
make_request "page_num=1&page_size=20" \
    "Basic request without filters" 200

# Test Case 2: With ID filter
make_request "id=test-pod-id&page_num=1&page_size=20" \
    "Filter by pod ID" 200

# Test Case 3: With cards_num filter
make_request "cards_num=1&page_num=1&page_size=10" \
    "Filter by GPU cards number" 200

# Test Case 4: With image filter
make_request "image=ubuntu:20.04&page_num=1&page_size=20" \
    "Filter by image" 200

# Test Case 5: With all filters
make_request "id=test-pod-id&cards_num=2&image=ubuntu:20.04&page_num=1&page_size=10" \
    "Filter with all parameters" 200

# Test Case 6: Large page size
make_request "page_num=1&page_size=100" \
    "Large page size (100)" 200

# Test Case 7: Second page
make_request "page_num=2&page_size=20" \
    "Pagination - second page" 200

# Test Case 8: Invalid page_num (should default to 1)
make_request "page_num=invalid&page_size=20" \
    "Invalid page number (should default to 1)" 200

# Test Case 9: Invalid page_size (should default to 20)
make_request "page_num=1&page_size=invalid" \
    "Invalid page size (should default to 20)" 200

# Test Case 10: No auth (without cookie)
echo -e "\n${YELLOW}Testing:${NC} No authentication (should fail)"
echo "URL: ${BASE_URL}${ENDPOINT}?page_num=1&page_size=20"
response=$(curl -s -w "\n%{http_code}" \
    "${BASE_URL}${ENDPOINT}?page_num=1&page_size=20")
status_code=$(echo "$response" | tail -n 1)
body=$(echo "$response" | sed '$d')
print_result "No authentication" "$status_code" "401"

echo ""
echo "=================================================="
echo "Test Summary"
echo "=================================================="
echo "All tests completed. Review results above."
echo "=================================================="
