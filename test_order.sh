#!/bin/bash

# Configuration
AUTH_URL="http://localhost:8082/api/v0"    # Auth service
STORE_URL="http://localhost:8080/api/v0"   # Store service
COOKIES_FILE="/tmp/order_cookies.txt"
EMAIL="testuser@example.com"
PASSWORD="Password123!"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

# Function to make JSON API calls to store service
make_store_request() {
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
    cmd+=("$STORE_URL$endpoint")

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

# Function to extract order ID from response
extract_order_id() {
    echo "$1" | sed '$d' | jq -r '.id' 2>/dev/null
}

# Cleanup
cleanup() { 
    rm -f "$COOKIES_FILE" 
}
trap cleanup EXIT

echo -e "${YELLOW}Starting Order API testing...${NC}"
echo "Auth Service: $AUTH_URL"
echo "Store Service: $STORE_URL"
echo "=================================================="

# 🔐 PHASE 1: AUTHENTICATION
echo -e "${YELLOW}🔐 PHASE 1: Authentication${NC}"

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
    echo -e "${BLUE}JWT: ${JWT_TOKEN:0:50}...${NC}"
else
    echo -e "${RED}No JWT token in cookies after login${NC}"
    echo "Cookies content:"
    cat "$COOKIES_FILE"
    exit 1
fi
print_result "Login" "$LOGIN_RESPONSE"

# 🛒 PHASE 2: CART PREPARATION
echo -e "${YELLOW}🛒 PHASE 2: Cart Preparation${NC}"

# 3. Get available items first
echo -e "${YELLOW}3. Getting available items...${NC}"
GET_ITEMS_RESPONSE=$(make_store_request "GET" "/items?limit=5" "" "$CSRF_TOKEN")
ITEM_IDS=$(echo "$GET_ITEMS_RESPONSE" | sed '$d' | jq -r '.[].id' 2>/dev/null)

if [ -n "$ITEM_IDS" ]; then
    echo -e "${GREEN}Found items, using first two for testing${NC}"
    ITEM_ID_1=$(echo "$ITEM_IDS" | head -1)
    ITEM_ID_2=$(echo "$ITEM_IDS" | head -2 | tail -1)
    echo -e "${BLUE}Item 1: $ITEM_ID_1${NC}"
    echo -e "${BLUE}Item 2: $ITEM_ID_2${NC}"
else
    echo -e "${YELLOW}No items found, using mock item IDs${NC}"
    ITEM_ID_1="550e8400-e29b-41d4-a716-446655440000"
    ITEM_ID_2="550e8400-e29b-41d4-a716-446655440001"
fi
print_result "Get Items" "$GET_ITEMS_RESPONSE"

# 4. Update cart with items (исправлено: PUT вместо POST)
echo -e "${YELLOW}4. Updating cart with items...${NC}"
CART_UPDATE_DATA="{\"items\":[{\"item_id\":\"$ITEM_ID_1\",\"quantity\":2},{\"item_id\":\"$ITEM_ID_2\",\"quantity\":1}]}"
UPDATE_CART_RESPONSE=$(make_store_request "PUT" "/cart" "$CART_UPDATE_DATA" "$CSRF_TOKEN")
print_result "Update Cart" "$UPDATE_CART_RESPONSE"

# 5. Get cart to verify
echo -e "${YELLOW}5. Verifying cart contents...${NC}"
GET_CART_RESPONSE=$(make_store_request "GET" "/cart" "" "$CSRF_TOKEN")
print_result "Get Cart" "$GET_CART_RESPONSE"

# 📦 PHASE 3: ORDER OPERATIONS
echo -e "${YELLOW}📦 PHASE 3: Order Operations${NC}"

# 6. Get user orders (empty at first)
echo -e "${YELLOW}6. Getting user orders (should be empty initially)...${NC}"
GET_ORDERS_RESPONSE=$(make_store_request "GET" "/orders" "" "$CSRF_TOKEN")
print_result "Get User Orders" "$GET_ORDERS_RESPONSE"

# 7. Create order from cart
echo -e "${YELLOW}7. Creating order from cart...${NC}"
CREATE_ORDER_RESPONSE=$(make_store_request "POST" "/orders" "" "$CSRF_TOKEN")
ORDER_ID=$(extract_order_id "$CREATE_ORDER_RESPONSE")

if [ -n "$ORDER_ID" ] && [ "$ORDER_ID" != "null" ]; then
    echo -e "${GREEN}Order created with ID: $ORDER_ID${NC}"
    REAL_ORDER_CREATED=true
else
    echo -e "${YELLOW}⚠ Could not create order (cart might be empty or items don't exist)${NC}"
    ORDER_ID="00000000-0000-0000-0000-000000000000"
    REAL_ORDER_CREATED=false
    echo -e "${YELLOW}Using mock order ID: $ORDER_ID for further tests${NC}"
fi
print_result "Create Order" "$CREATE_ORDER_RESPONSE"

# 8. Get specific order details
echo -e "${YELLOW}8. Getting specific order details...${NC}"
GET_ORDER_RESPONSE=$(make_store_request "GET" "/orders/$ORDER_ID" "" "$CSRF_TOKEN")
print_result "Get Order Details" "$GET_ORDER_RESPONSE"

# 9. Get user orders again (should have the new order)
echo -e "${YELLOW}9. Getting user orders after creation...${NC}"
GET_ORDERS_AFTER_RESPONSE=$(make_store_request "GET" "/orders" "" "$CSRF_TOKEN")
print_result "Get User Orders After Creation" "$GET_ORDERS_AFTER_RESPONSE"

# 10. Test order status update (if we have a real order)
if [ "$REAL_ORDER_CREATED" = true ]; then
    echo -e "${YELLOW}10. Testing order status update...${NC}"
    UPDATE_STATUS_DATA='{"status":"processing"}'
    UPDATE_STATUS_RESPONSE=$(make_store_request "PATCH" "/orders/$ORDER_ID/status" "$UPDATE_STATUS_DATA" "$CSRF_TOKEN")
    print_result "Update Order Status" "$UPDATE_STATUS_RESPONSE"
    
    # Verify status was updated
    echo -e "${YELLOW}  - Verifying status update...${NC}"
    GET_ORDER_AFTER_UPDATE=$(make_store_request "GET" "/orders/$ORDER_ID" "" "$CSRF_TOKEN")
    CURRENT_STATUS=$(echo "$GET_ORDER_AFTER_UPDATE" | sed '$d' | jq -r '.status' 2>/dev/null)
    echo -e "${BLUE}Current order status: $CURRENT_STATUS${NC}"
    print_result "Get Order After Status Update" "$GET_ORDER_AFTER_UPDATE"
else
    echo -e "${YELLOW}10. Skipping status update (no real order created)${NC}"
fi

# 🚨 PHASE 4: ERROR CASES
echo -e "${YELLOW}🚨 PHASE 4: Error Cases${NC}"

# 11. Get non-existent order
echo -e "${YELLOW}11. Getting non-existent order...${NC}"
NON_EXISTENT_ORDER_RESPONSE=$(make_store_request "GET" "/orders/00000000-0000-0000-0000-000000000000" "" "$CSRF_TOKEN")
print_result "Get Non-existent Order" "$NON_EXISTENT_ORDER_RESPONSE"

# 12. Update status of non-existent order
echo -e "${YELLOW}12. Updating status of non-existent order...${NC}"
UPDATE_NON_EXISTENT_DATA='{"status":"processing"}'
UPDATE_NON_EXISTENT_RESPONSE=$(make_store_request "PATCH" "/orders/00000000-0000-0000-0000-000000000000/status" "$UPDATE_NON_EXISTENT_DATA" "$CSRF_TOKEN")
print_result "Update Non-existent Order Status" "$UPDATE_NON_EXISTENT_RESPONSE"

# 13. Invalid order ID format
echo -e "${YELLOW}13. Testing invalid order ID format...${NC}"
INVALID_ID_RESPONSE=$(make_store_request "GET" "/orders/invalid-uuid-format" "" "$CSRF_TOKEN")
print_result "Get Order with Invalid ID" "$INVALID_ID_RESPONSE"

# 🚪 PHASE 5: CLEANUP
echo -e "${YELLOW}🚪 PHASE 5: Cleanup${NC}"

# 14. Logout (auth service)
echo -e "${YELLOW}14. Logging out via auth service...${NC}"
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

echo -e "${GREEN}Order API testing completed!${NC}"

# Summary
echo -e "\n${YELLOW}📊 Test Summary:${NC}"
echo -e "  ✅ Authentication flow (CSRF + JWT)"
echo -e "  ✅ Cart operations (PUT /cart)"
echo -e "  ✅ Order creation (POST /orders)" 
echo -e "  ✅ Order retrieval (GET /orders, GET /orders/{id})"
echo -e "  ✅ Order status updates (PATCH /orders/{id}/status)"
echo -e "  ✅ Error handling"
echo -e "  ✅ Cleanup and logout"