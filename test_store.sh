#!/bin/bash

# Configuration
STORE_URL="http://localhost:8080/api/v0"
COOKIES_FILE="/tmp/store_cookies.txt"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Initialize cookies file
> "$COOKIES_FILE"

# Function to make JSON API calls
make_store_request() {
    local method=$1
    local endpoint=$2
    local data=$3

    curl -s -w "\n%{http_code}" -c "$COOKIES_FILE" -b "$COOKIES_FILE" \
         -X "$method" -H "Content-Type: application/json" \
         ${data:+-d "$data"} \
         "$STORE_URL$endpoint"
}

# Function to print result
print_result() {
    local op=$1
    local resp=$2
    local code=$(echo "$resp" | tail -1)
    local body=$(echo "$resp" | head -n -1)

    echo -n "Response: "
    if [[ "$code" -ge 200 && "$code" -lt 300 ]]; then
        echo -e "${GREEN}✓ $op successful (Status: $code)${NC}"
    else
        echo -e "${RED}✗ $op failed (Status: $code)${NC}"
    fi
    
    if [ -n "$body" ]; then
        echo "$body"
    fi
    echo
}

# Function to extract IDs
extract_ids() {
    echo "$1" | head -n -1 | grep -o '"id":"[^"]*"' | cut -d'"' -f4
}

# Cleanup
cleanup() { rm -f "$COOKIES_FILE"; }
trap cleanup EXIT

echo -e "${YELLOW}Starting Store Service API testing...${NC}"
echo "Store Service: $STORE_URL"
echo "=================================================="

# 1. Get all stores
echo -e "${YELLOW}1. Getting all stores...${NC}"
STORES_RESPONSE=$(make_store_request "POST" "/stores" '{
    "limit": 10,
    "last_id": "",
    "tag_id": "",
    "city_id": "",
    "sorted": "",
    "desc": false
}')
print_result "Get All Stores" "$STORES_RESPONSE"

# Extract store IDs
STORE_IDS=($(extract_ids "$STORES_RESPONSE"))
if [ ${#STORE_IDS[@]} -eq 0 ]; then
    echo -e "${YELLOW}Using hardcoded store IDs...${NC}"
    STORE_IDS=(
        "9ac3b889-96df-4c93-a0b7-31f5b6a6e89c"
        "b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2" 
        "c45a7b64-df32-4e84-b2cb-85a3b8e6b0fc"
        "d0c12a9f-2b2a-4e91-8e0a-13df58d9f8af"
    )
fi

echo -e "${GREEN}Found ${#STORE_IDS[@]} stores${NC}"

# 2. Test individual stores (только 1 для скорости)
echo -e "${YELLOW}2. Testing individual store endpoints...${NC}"
for store_id in "${STORE_IDS[@]:0:1}"; do
    echo -e "${BLUE}Testing store: $store_id${NC}"
    
    echo -e "${YELLOW}  - Getting store details...${NC}"
    print_result "Get Store" "$(make_store_request "GET" "/stores/$store_id" "")"
    
    echo -e "${YELLOW}  - Getting store reviews...${NC}"
    print_result "Get Reviews" "$(make_store_request "GET" "/stores/$store_id/reviews" "")"
    
    echo -e "${YELLOW}  - Getting store items...${NC}"
    print_result "Get Items" "$(make_store_request "GET" "/stores/$store_id/items" "")"
    
    echo -e "${YELLOW}  - Getting item types...${NC}"
    print_result "Get Item Types" "$(make_store_request "GET" "/stores/$store_id/item-types" "")"
    
    echo "  ---"
done

# 3. Get cities
echo -e "${YELLow}3. Getting cities...${NC}"
CITIES_RESPONSE=$(make_store_request "GET" "/stores/cities" "")
print_result "Get Cities" "$CITIES_RESPONSE"

# 4. Get tags
echo -e "${YELLOW}4. Getting tags...${NC}"
print_result "Get Tags" "$(make_store_request "GET" "/stores/tags" "")"

# 5. Test filtering with proper variable substitution
echo -e "${YELLOW}5. Testing store filtering by Moscow...${NC}"
FILTER_RESPONSE=$(make_store_request "POST" "/stores" '{
    "limit": 5,
    "last_id": "",
    "tag_id": "",
    "city_id": "3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9",
    "sorted": "",
    "desc": false
}')
print_result "Get Stores Filtered by Moscow" "$FILTER_RESPONSE"

# 6. Test sorting
echo -e "${YELLOW}6. Testing store sorting by rating...${NC}"
SORT_RESPONSE=$(make_store_request "POST" "/stores" '{
    "limit": 5,
    "last_id": "",
    "tag_id": "",
    "city_id": "",
    "sorted": "rating",
    "desc": true
}')
print_result "Get Stores Sorted by Rating" "$SORT_RESPONSE"

echo -e "${GREEN}Store service API testing completed!${NC}"