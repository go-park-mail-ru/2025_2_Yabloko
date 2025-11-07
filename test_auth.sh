#!/bin/bash

# Configuration
BASE_URL="http://localhost:8082/api/v0"
COOKIES_FILE="/tmp/test_cookies.txt"
EMAIL="testuser@example.com"
PASSWORD="Password123!"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Initialize cookies file
> "$COOKIES_FILE"

# Function to extract value from cookies
get_cookie_value() {
    local cookie_name=$1
    grep "$cookie_name" "$COOKIES_FILE" | awk -F'\t' '$6 == "'"$cookie_name"'" {print $7}' 2>/dev/null
}

# Function to extract value from JSON response
get_json_value() {
    local json="$1"
    local field="$2"
    echo "$json" | grep -o "\"$field\":\"[^\"]*\"" | head -1 | cut -d'"' -f4
}

# Function to format JSON string (basic indentation)
format_json() {
    local json="$1"
    local indent="$2"
    local result=""
    local level=${indent:-0}
    local spaces=""
    
    # Create indentation spaces
    for ((i=0; i<level; i++)); do
        spaces="$spaces  "
    done
    
    # Check if it looks like JSON (starts with { or [)
    if echo "$json" | grep -q "^[{\[]"; then
        # Very basic JSON formatting
        result=$(echo "$json" | \
            sed 's/},{/},\n'"$spaces"'  {/g' | \
            sed 's/\[{/[\n'"$spaces"'  {/' | \
            sed 's/}\]/\n'"$spaces"']}/' | \
            sed 's/,/,\n'"$spaces"'  /g' | \
            sed 's/{/\{\n'"$spaces"'  /' | \
            sed 's/}/\n'"$spaces"'\}/')
    else
        result="$json"
    fi
    
    echo "$result"
}

# Function to make API calls with proper CSRF handling
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local csrf_token=$4
    
    local curl_cmd=("curl" "-s" "-w" "\n%{http_code}" "-c" "$COOKIES_FILE" "-b" "$COOKIES_FILE" "-X" "$method")
    
    # Add headers
    curl_cmd+=("-H" "Content-Type: application/json")
    
    if [ -n "$csrf_token" ]; then
        curl_cmd+=("-H" "X-CSRF-Token: $csrf_token")
    fi
    
    # Add data if provided
    if [ -n "$data" ]; then
        curl_cmd+=("-d" "$data")
    fi
    
    # Add URL
    curl_cmd+=("$BASE_URL$endpoint")
    
    # Execute command and capture output
    "${curl_cmd[@]}"
}

# Function to print result
print_result() {
    local operation=$1
    local response=$2
    local status_code=$(echo "$response" | tail -1)
    local body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" -ge 200 ] && [ "$status_code" -lt 300 ]; then
        echo -e "${GREEN}✓ $operation successful (Status: $status_code)${NC}"
        if [ -n "$body" ] && [ "$body" != "null" ]; then
            echo -e "${BLUE}Response:${NC}"
            format_json "$body" 1
        fi
    else
        echo -e "${RED}✗ $operation failed (Status: $status_code)${NC}"
        if [ -n "$body" ] && [ "$body" != "null" ]; then
            echo -e "${RED}Error details:${NC}"
            echo "$body"
        fi
    fi
    echo
}

# Function to check dependencies
check_dependencies() {
    local missing_deps=()
    
    if ! command -v "curl" &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    if ! command -v "grep" &> /dev/null; then
        missing_deps+=("grep")
    fi
    
    if ! command -v "awk" &> /dev/null; then
        missing_deps+=("awk")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        echo -e "${RED}Error: Missing dependencies: ${missing_deps[*]}${NC}"
        echo "Please install the required packages:"
        for dep in "${missing_deps[@]}"; do
            echo "  - $dep"
        done
        exit 1
    fi
}

# Cleanup function
cleanup() {
    rm -f "$COOKIES_FILE"
}

# Set trap for cleanup
trap cleanup EXIT

# Display configuration
echo -e "${YELLOW}API Test Configuration:${NC}"
echo -e "  Base URL: ${BLUE}$BASE_URL${NC}"
echo -e "  Email: ${BLUE}$EMAIL${NC}"
echo -e "  Cookies file: ${BLUE}$COOKIES_FILE${NC}"
echo

# Main execution
main() {
    echo -e "${YELLOW}Checking dependencies...${NC}"
    check_dependencies
    
    echo -e "${YELLOW}Starting API testing...${NC}"
    echo "=========================================="

    # 1. Get CSRF Token
    echo -e "${YELLOW}1. Getting CSRF token...${NC}"
    CSRF_RESPONSE=$(make_request "GET" "/csrf")
    CSRF_TOKEN=$(get_cookie_value "csrf_token")
    
    # Если токен не в cookies, попробуем из JSON тела
    if [ -z "$CSRF_TOKEN" ]; then
        CSRF_BODY=$(echo "$CSRF_RESPONSE" | head -n -1)
        CSRF_TOKEN=$(get_json_value "$CSRF_BODY" "csrf_token")
    fi

    if [ -n "$CSRF_TOKEN" ]; then
        echo -e "${GREEN}CSRF token obtained successfully${NC}"
        echo -e "${BLUE}CSRF Token: ${CSRF_TOKEN}${NC}"
        print_result "CSRF Token" "$CSRF_RESPONSE"
    else
        echo -e "${RED}Failed to get CSRF token${NC}"
        echo -e "${YELLOW}Debug info:${NC}"
        echo "Cookies file content:"
        cat "$COOKIES_FILE"
        echo "Response body:"
        echo "$CSRF_RESPONSE"
        exit 1
    fi

    # 2. User Registration
    echo -e "${YELLOW}2. Registering user...${NC}"
    REGISTER_DATA="{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"
    REGISTER_RESPONSE=$(make_request "POST" "/auth/signup" "$REGISTER_DATA" "$CSRF_TOKEN")
    print_result "Registration" "$REGISTER_RESPONSE"

    # 3. User Login
    echo -e "${YELLOW}3. Logging in...${NC}"
    LOGIN_DATA="{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"
    LOGIN_RESPONSE=$(make_request "POST" "/auth/login" "$LOGIN_DATA" "$CSRF_TOKEN")
    JWT_TOKEN=$(get_cookie_value "jwt_token")

    if [ -n "$JWT_TOKEN" ]; then
        echo -e "${GREEN}JWT token obtained and stored${NC}"
    else
        echo -e "${YELLOW}No JWT token received in login response${NC}"
        echo -e "${YELLOW}Cookies after login:${NC}"
        cat "$COOKIES_FILE"
    fi
    print_result "Login" "$LOGIN_RESPONSE"

    # 4. Refresh Token (only if login was successful)
    if [ -n "$JWT_TOKEN" ]; then
        echo -e "${YELLOW}4. Refreshing token...${NC}"
        REFRESH_RESPONSE=$(make_request "POST" "/auth/refresh" "" "$CSRF_TOKEN")
        NEW_JWT_TOKEN=$(get_cookie_value "jwt_token")
        
        if [ -n "$NEW_JWT_TOKEN" ]; then
            echo -e "${GREEN}JWT token refreshed successfully${NC}"
        else
            echo -e "${YELLOW}No new JWT token received in refresh response${NC}"
        fi
        print_result "Token Refresh" "$REFRESH_RESPONSE"
    else
        echo -e "${YELLOW}4. Skipping token refresh (no JWT token)${NC}"
    fi

    # 5. Logout (only if we have CSRF token)
    if [ -n "$CSRF_TOKEN" ]; then
        echo -e "${YELLOW}5. Logging out...${NC}"
        LOGOUT_RESPONSE=$(make_request "POST" "/auth/logout" "" "$CSRF_TOKEN")
        print_result "Logout" "$LOGOUT_RESPONSE"
    else
        echo -e "${YELLOW}5. Skipping logout (no CSRF token)${NC}"
    fi

    # Verify tokens are cleared
    FINAL_CSRF=$(get_cookie_value "csrf_token")
    FINAL_JWT=$(get_cookie_value "jwt_token")

    echo -e "${YELLOW}Final verification:${NC}"
    if [ -z "$FINAL_JWT" ]; then
        echo -e "${GREEN}✓ JWT token successfully cleared${NC}"
    else
        echo -e "${RED}✗ JWT token still present${NC}"
    fi

    if [ -n "$FINAL_CSRF" ]; then
        echo -e "${GREEN}✓ CSRF token still present (expected)${NC}"
    else
        echo -e "${YELLOW}⚠ CSRF token cleared${NC}"
    fi

    echo
    echo -e "${GREEN}API testing completed!${NC}"
    
    # Display final cookie file contents for debugging
    echo
    echo -e "${YELLOW}Final cookies file contents:${NC}"
    if [ -s "$COOKIES_FILE" ]; then
        cat "$COOKIES_FILE"
    else
        echo "  (empty)"
    fi
}

# Run main function
main "$@"