#!/bin/bash

# Configuration
AUTH_URL="http://localhost:8082/api/v0"    # Auth service
PROFILE_URL="http://localhost:8081/api/v0" # Profile service
COOKIES_FILE="/tmp/profile_cookies.txt"
EMAIL="testuser@example.com"
PASSWORD="Password123!"
AVATAR_FILE="./img.jpg"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Initialize cookies file
> "$COOKIES_FILE"

# Function to extract cookie value
get_cookie_value() {
    local cookie_name=$1
    grep "$cookie_name" "$COOKIES_FILE" | awk 'NF>6 {print $7}' 2>/dev/null
}

# Function to make JSON API calls to auth service
make_auth_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local csrf_token=$4

    local headers=("-H" "Content-Type: application/json")
    [ -n "$csrf_token" ] && headers+=("-H" "X-CSRF-Token: $csrf_token")

    local cmd=(curl -s -w "\n%{http_code}" -c "$COOKIES_FILE" -b "$COOKIES_FILE" -X "$method")

    for h in "${headers[@]}"; do
        cmd+=("$h")
    done

    [ -n "$data" ] && cmd+=("-d" "$data")
    cmd+=("$AUTH_URL$endpoint")

    "${cmd[@]}"
}

# Function to make JSON API calls to profile service
make_profile_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local csrf_token=$4

    local headers=("-H" "Content-Type: application/json")
    [ -n "$csrf_token" ] && headers+=("-H" "X-CSRF-Token: $csrf_token")

    local cmd=(curl -s -w "\n%{http_code}" -c "$COOKIES_FILE" -b "$COOKIES_FILE" -X "$method")

    for h in "${headers[@]}"; do
        cmd+=("$h")
    done

    [ -n "$data" ] && cmd+=("-d" "$data")
    cmd+=("$PROFILE_URL$endpoint")

    "${cmd[@]}"
}

# Function to make file upload requests to profile service
make_avatar_upload_request() {
    local endpoint=$1
    local file_path=$2
    local csrf_token=$3

    local cmd=(curl -s -w "\n%{http_code}" -c "$COOKIES_FILE" -b "$COOKIES_FILE" -X "POST")
    
    [ -n "$csrf_token" ] && cmd+=("-H" "X-CSRF-Token: $csrf_token")
    
    cmd+=(-F "avatar=@$file_path")
    cmd+=("$PROFILE_URL$endpoint")

    "${cmd[@]}"
}

# Function to print result
print_result() {
    local op=$1
    local resp=$2
    local code=$(echo "$resp" | tail -1)
    local body=$(echo "$resp" | sed '$d')

    if [[ "$code" -ge 200 && "$code" -lt 300 ]]; then
        echo -e "${GREEN}✓ $op successful (Status: $code)${NC}"
        [ -n "$body" ] && (echo "$body" | jq '.' 2>/dev/null || echo "$body")
    else
        echo -e "${RED}✗ $op failed (Status: $code)${NC}"
        [ -n "$body" ] && echo "$body"
    fi
    echo
}

# Cleanup
cleanup() { 
    rm -f "$COOKIES_FILE" 
}
trap cleanup EXIT

echo -e "${YELLOW}Starting profile & avatar API testing...${NC}"
echo "Auth Service: $AUTH_URL"
echo "Profile Service: $PROFILE_URL"
echo "=================================================="

# 1. CSRF token (from auth service)
echo -e "${YELLOW}1. Getting CSRF token from auth service...${NC}"
CSRF_RESPONSE=$(make_auth_request "GET" "/csrf")
CSRF_TOKEN=$(get_cookie_value "csrf_token")

if [ -z "$CSRF_TOKEN" ]; then
    echo -e "${RED}Failed to get CSRF token${NC}"
    exit 1
fi
echo -e "${GREEN}CSRF token obtained${NC}"
print_result "CSRF Token" "$CSRF_RESPONSE"

# 2. Login (auth service)
echo -e "${YELLOW}2. Logging in via auth service...${NC}"
LOGIN_DATA="{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"
LOGIN_RESPONSE=$(make_auth_request "POST" "/auth/login" "$LOGIN_DATA" "$CSRF_TOKEN")
JWT_TOKEN=$(get_cookie_value "jwt_token")

if [ -n "$JWT_TOKEN" ]; then
    echo -e "${GREEN}JWT token obtained${NC}"
fi
print_result "Login" "$LOGIN_RESPONSE"

# 3. Get user profile (profile service)
echo -e "${YELLOW}3. Getting user profile from profile service...${NC}"
PROFILE_RESPONSE=$(make_profile_request "GET" "/profiles/me" "" "$CSRF_TOKEN")
USER_ID=$(echo "$PROFILE_RESPONSE" | sed '$d' | jq -r '.id' 2>/dev/null)

if [ -n "$USER_ID" ] && [ "$USER_ID" != "null" ]; then
    echo -e "${GREEN}User ID: $USER_ID${NC}"
else
    echo -e "${YELLOW}Could not extract user ID, using 'me' endpoint${NC}"
    USER_ID="me"
fi
print_result "Get Profile" "$PROFILE_RESPONSE"

# 4. Update profile (profile service)
echo -e "${YELLOW}4. Updating profile via profile service...${NC}"
UPDATE_DATA='{"name":"Test User","phone":"12345678901","address":"Test Address"}'
UPDATE_RESPONSE=$(make_profile_request "PUT" "/profiles/me" "$UPDATE_DATA" "$CSRF_TOKEN")
print_result "Update Profile" "$UPDATE_RESPONSE"

# 5. Upload avatar (profile service)
echo -e "${YELLOW}5. Uploading avatar via profile service...${NC}"
if [ -f "$AVATAR_FILE" ]; then
    echo -e "${GREEN}Using avatar file: $AVATAR_FILE${NC}"
    
    # Get file info
    file_size=$(stat -f%z "$AVATAR_FILE" 2>/dev/null || stat -c%s "$AVATAR_FILE" 2>/dev/null)
    file_type=$(file -b --mime-type "$AVATAR_FILE" 2>/dev/null || echo "unknown")
    echo -e "${BLUE}File info: Size: $((file_size/1024)) KB, Type: $file_type${NC}"
    
    # Try both endpoints
    AVATAR_RESPONSE=$(make_avatar_upload_request "/profiles/me/avatar" "$AVATAR_FILE" "$CSRF_TOKEN")
    if [[ $(echo "$AVATAR_RESPONSE" | tail -1) -ge 400 ]]; then
        echo -e "${YELLOW}Trying direct user ID endpoint...${NC}"
        AVATAR_RESPONSE=$(make_avatar_upload_request "/profiles/$USER_ID/avatar" "$AVATAR_FILE" "$CSRF_TOKEN")
    fi
    
    print_result "Upload Avatar" "$AVATAR_RESPONSE"
else
    echo -e "${YELLOW}⚠ Avatar file not found at $AVATAR_FILE${NC}"
    echo -e "${YELLOW}Current directory: $(pwd)${NC}"
    echo -e "${YELLOW}Files in current directory:${NC}"
    ls -la | head -10
fi

# 6. Get profile after avatar upload to verify (profile service)
echo -e "${YELLOW}6. Getting profile after avatar upload...${NC}"
PROFILE_AFTER_AVATAR=$(make_profile_request "GET" "/profiles/me" "" "$CSRF_TOKEN")
print_result "Get Profile After Avatar" "$PROFILE_AFTER_AVATAR"

# 7. Test direct ID access (profile service)
if [ -n "$USER_ID" ] && [ "$USER_ID" != "me" ]; then
    echo -e "${YELLOW}7. Testing direct user ID access...${NC}"
    DIRECT_PROFILE_RESPONSE=$(make_profile_request "GET" "/profiles/$USER_ID" "" "$CSRF_TOKEN")
    print_result "Get Profile by Direct ID" "$DIRECT_PROFILE_RESPONSE"
fi

# 8. Logout (auth service)
echo -e "${YELLOW}8. Logging out via auth service...${NC}"
LOGOUT_RESPONSE=$(make_auth_request "POST" "/auth/logout" "" "$CSRF_TOKEN")
print_result "Logout" "$LOGOUT_RESPONSE"

# Final verification
echo -e "${YELLOW}Final Verification:${NC}"
FINAL_JWT=$(get_cookie_value "jwt_token")
FINAL_CSRF=$(get_cookie_value "csrf_token")

if [ -z "$FINAL_JWT" ]; then
    echo -e "${GREEN}✓ JWT token cleared${NC}"
else
    echo -e "${RED}✗ JWT token still present${NC}"
fi

if [ -n "$FINAL_CSRF" ]; then
    echo -e "${GREEN}✓ CSRF token present${NC}"
else
    echo -e "${YELLOW}⚠ CSRF token cleared${NC}"
fi

echo -e "${GREEN}Profile & avatar API testing completed!${NC}"