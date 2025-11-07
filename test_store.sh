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

# Function to make API calls
make_store_request() {
    local method=$1
    local endpoint=$2
    local data=$3

    curl -s -w "\n%{http_code}" -c "$COOKIES_FILE" -b "$COOKIES_FILE" \
         -X "$method" -H "Content-Type: application/json" \
         ${data:+-d "$data"} \
         "$STORE_URL$endpoint"
}

# Function to make GET requests with query parameters
make_get_request() {
    local endpoint=$1
    local query_params=$2

    curl -s -w "\n%{http_code}" -c "$COOKIES_FILE" -b "$COOKIES_FILE" \
         -X "GET" \
         "$STORE_URL$endpoint$query_params"
}

# Function to check image endpoint (НОВАЯ ЛОГИКА)
check_image() {
    local image_path=$1
    local endpoint="/images/stores/$image_path"
    
    echo -n "  Image: $endpoint ... "
    response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080$endpoint")
    
    if [ "$response" -eq 200 ]; then
        echo -e "${GREEN}✓ (Status: $response)${NC}"
        return 0
    else
        echo -e "${RED}✗ (Status: $response)${NC}"
        return 1
    fi
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

# Function to extract card_img from response (только имя файла)
extract_card_images() {
    echo "$1" | head -n -1 | grep -o '"card_img":"[^"]*"' | cut -d'"' -f4 | sed 's|.*/||'
}

# Cleanup
cleanup() { rm -f "$COOKIES_FILE"; }
trap cleanup EXIT

echo -e "${YELLOW}Starting Store Service API testing...${NC}"
echo "Store Service: $STORE_URL"
echo "=================================================="

# 1. Get all stores with GET method and query parameters
echo -e "${YELLOW}1. Getting all stores (GET with query params)...${NC}"

# 1.1 Basic GET with limit
echo -e "${YELLOW}  - GET /stores?limit=5...${NC}"
GET_BASIC_RESPONSE=$(make_get_request "/stores" "?limit=5")
print_result "GET Stores Basic" "$GET_BASIC_RESPONSE"

# Extract store IDs and card images from the first successful response
STORES_RESPONSE="$GET_BASIC_RESPONSE"
STORE_IDS=()
while IFS= read -r id; do
    [[ -n "$id" ]] && STORE_IDS+=("$id")
done < <(extract_ids "$STORES_RESPONSE")

CARD_IMAGES=()
while IFS= read -r img; do
    [[ -n "$img" ]] && CARD_IMAGES+=("$img")
done < <(extract_card_images "$STORES_RESPONSE")

if [ ${#STORE_IDS[@]} -eq 0 ]; then
    echo -e "${YELLOW}Using hardcoded store IDs...${NC}"
    STORE_IDS=(
        "9ac3b889-96df-4c93-a0b7-31f5b6a6e89c"
        "b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2" 
        "c45a7b64-df32-4e84-b2cb-85a3b8e6b0fc"
        "d0c12a9f-2b2a-4e91-8e0a-13df58d9f8af"
    )
    CARD_IMAGES=(
        "techworld.png"
        "coffee_point.jpg" 
        "book_haven.jpeg"
        "green_market.svg"
    )
fi

echo -e "${GREEN}Found ${#STORE_IDS[@]} stores${NC}"

# 2. Test GET with various filters
echo -e "${YELLOW}2. Testing GET stores with filters...${NC}"

# 2.1 GET with city filter (Moscow)
echo -e "${YELLOW}  - GET /stores?limit=5&city_id=3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9...${NC}"
GET_MOSCOW_RESPONSE=$(make_get_request "/stores" "?limit=5&city_id=3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9")
print_result "GET Stores Moscow" "$GET_MOSCOW_RESPONSE"

# 2.2 GET with tag filter
echo -e "${YELLOW}  - GET /stores?limit=5&tag_id=550e8400-e29b-41d4-a716-446655440001...${NC}"
GET_TAG_RESPONSE=$(make_get_request "/stores" "?limit=5&tag_id=550e8400-e29b-41d4-a716-446655440001")
print_result "GET Stores by Tag" "$GET_TAG_RESPONSE"

# 2.3 GET with sorting by rating
echo -e "${YELLOW}  - GET /stores?limit=5&sorted=rating&desc=true...${NC}"
GET_SORTED_RESPONSE=$(make_get_request "/stores" "?limit=5&sorted=rating&desc=true")
print_result "GET Stores Sorted by Rating" "$GET_SORTED_RESPONSE"

# 2.4 GET with all parameters
echo -e "${YELLOW}  - GET /stores?limit=5&city_id=3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9&tag_id=550e8400-e29b-41d4-a716-446655440001&sorted=rating&desc=true...${NC}"
GET_ALL_PARAMS_RESPONSE=$(make_get_request "/stores" "?limit=5&city_id=3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9&tag_id=550e8400-e29b-41d4-a716-446655440001&sorted=rating&desc=true")
print_result "GET Stores All Params" "$GET_ALL_PARAMS_RESPONSE"

# 3. Test individual store endpoints
echo -e "${YELLOW}3. Testing individual store endpoints...${NC}"
if [ ${#STORE_IDS[@]} -gt 0 ]; then
    store_id="${STORE_IDS[0]}"
    card_img="${CARD_IMAGES[0]}"
    
    echo -e "${BLUE}Testing store: $store_id${NC}"
    
    # 3.1 Get store details (GET /stores/{id})
    echo -e "${YELLOW}  - Getting store details...${NC}"
    STORE_DETAIL_RESPONSE=$(make_store_request "GET" "/stores/$store_id" "")
    print_result "Get Store" "$STORE_DETAIL_RESPONSE"
    
    # 3.2 Get store reviews (GET /stores/{id}/reviews)
    echo -e "${YELLOW}  - Getting store reviews...${NC}"
    print_result "Get Reviews" "$(make_store_request "GET" "/stores/$store_id/reviews" "")"
    
    # 3.3 Test store image accessibility (НОВАЯ ЛОГИКА)
    echo -e "${YELLOW}  - Testing store image...${NC}"
    if [ -n "$card_img" ]; then
        check_image "$card_img"
    else
        echo -e "${YELLOW}    No card_img found for store $store_id${NC}"
    fi
    
    echo "  ---"
else
    echo -e "${YELLOW}  No stores found to test${NC}"
fi

# 4. Test all store images (НОВАЯ ЛОГИКА)
echo -e "${YELLOW}4. Testing all store images...${NC}"
IMAGE_COUNT=0
FAILED_IMAGES=()

for i in "${!CARD_IMAGES[@]}"; do
    card_img="${CARD_IMAGES[$i]}"
    store_id="${STORE_IDS[$i]}"
    
    if [ -n "$card_img" ]; then
        echo -n "  Store: $store_id ... "
        if check_image "$card_img"; then
            ((IMAGE_COUNT++))
        else
            FAILED_IMAGES+=("$store_id:$card_img")
        fi
    fi
done

echo -e "\n${GREEN}✅ Успешно загружено: $IMAGE_COUNT/${#CARD_IMAGES[@]} изображений${NC}"
if [ ${#FAILED_IMAGES[@]} -gt 0 ]; then
    echo -e "${YELLOW}⚠ Проблемные изображения:${NC}"
    for failed in "${FAILED_IMAGES[@]}"; do
        echo "    - $failed"
    done
fi

# 5. Get cities (GET /stores/cities)
echo -e "${YELLOW}5. Getting cities...${NC}"
CITIES_RESPONSE=$(make_store_request "GET" "/stores/cities" "")
print_result "Get Cities" "$CITIES_RESPONSE"

# 6. Get tags (GET /stores/tags)
echo -e "${YELLOW}6. Getting tags...${NC}"
print_result "Get Tags" "$(make_store_request "GET" "/stores/tags" "")"

# 7. Test error cases
echo -e "${YELLOW}7. Testing error cases...${NC}"

# 7.1 Invalid store ID format
echo -e "${YELLOW}  - Testing invalid store ID format...${NC}"
INVALID_ID_RESPONSE=$(make_store_request "GET" "/stores/invalid-uuid-format" "")
print_result "Invalid Store ID Format" "$INVALID_ID_RESPONSE"

# 7.2 Non-existent store
echo -e "${YELLOW}  - Testing non-existent store...${NC}"
NON_EXISTENT_RESPONSE=$(make_store_request "GET" "/stores/00000000-0000-0000-0000-000000000000" "")
print_result "Non-existent Store" "$NON_EXISTENT_RESPONSE"

# 7.3 Missing required parameters for GET /stores
echo -e "${YELLOW}  - Testing GET /stores without limit...${NC}"
NO_LIMIT_RESPONSE=$(make_get_request "/stores" "")
print_result "GET Stores No Limit" "$NO_LIMIT_RESPONSE"

# 7.4 Invalid limit parameter
echo -e "${YELLOW}  - Testing GET /stores with invalid limit...${NC}"
INVALID_LIMIT_RESPONSE=$(make_get_request "/stores" "?limit=invalid")
print_result "GET Stores Invalid Limit" "$INVALID_LIMIT_RESPONSE"

# 8. Test image error cases (НОВАЯ ЛОГИКА)
echo -e "${YELLOW}8. Testing image error cases...${NC}"

# 8.1 Non-existent image
echo -e "${YELLOW}  - Testing non-existent image...${NC}"
check_image "non_existent_image.jpg"

# 8.2 Invalid image path
echo -e "${YELLOW}  - Testing invalid image path...${NC}"
check_image "../../etc/passwd"

echo -e "${GREEN}Store service API testing completed!${NC}"
echo
echo -e "${YELLOW}Summary:${NC}"
echo -e "  ✅ Все рабочие эндпоинты store_handler протестированы:"
echo -e "     - GET /stores (GetStores с query parameters)"
echo -e "     - GET /stores/{id} (GetStore)"  
echo -e "     - GET /stores/{id}/reviews (GetStoreReview)"
echo -e "     - GET /stores/cities (GetCities)"
echo -e "     - GET /stores/tags (GetTags)"
echo -e "     - GET /images/stores/* (Store Images) - НОВАЯ ЛОГИКА"
echo -e "  ❌ CreateStore не зарегистрирован в роутере"

# Детальная проверка фильтрации через GET
echo -e "\n${YELLOW}🔍 Детальная проверка фильтрации (GET)...${NC}"

echo -e "${YELLOW}  - Все магазины (базовый тест)...${NC}"
ALL_STORES_DEBUG=$(make_get_request "/stores" "?limit=10")
ALL_COUNT=$(echo "$ALL_STORES_DEBUG" | head -n -1 | grep -o '"id":"[^"]*"' | wc -l)
echo "    Всего магазинов в базе: $ALL_COUNT"

echo -e "${YELLOW}  - Фильтр по тегу 'Кофе'...${NC}"
COFFEE_RESPONSE=$(make_get_request "/stores" "?limit=10&tag_id=550e8400-e29b-41d4-a716-446655440002")
COFFEE_COUNT=$(echo "$COFFEE_RESPONSE" | head -n -1 | grep -o '"id":"[^"]*"' | wc -l)
COFFEE_NAMES=$(echo "$COFFEE_RESPONSE" | head -n -1 | grep -o '"name":"[^"]*"' | cut -d'"' -f4 | tr '\n' ',' | sed 's/,$//')

if [ "$COFFEE_COUNT" -eq 1 ]; then
    echo -e "${GREEN}    ✅ Найден 1 кофейный магазин: $COFFEE_NAMES${NC}"
else
    echo -e "${YELLOW}    ⚠ Найдено $COFFEE_COUNT кофейных магазинов (ожидалось 1)${NC}"
    echo "    Найденные магазины: $COFFEE_NAMES"
fi

echo -e "${YELLOW}  - Фильтр по городу 'Москва'...${NC}"
MOSCOW_RESPONSE=$(make_get_request "/stores" "?limit=10&city_id=3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9")
MOSCOW_COUNT=$(echo "$MOSCOW_RESPONSE" | head -n -1 | grep -o '"id":"[^"]*"' | wc -l)
MOSCOW_NAMES=$(echo "$MOSCOW_RESPONSE" | head -n -1 | grep -o '"name":"[^"]*"' | cut -d'"' -f4 | tr '\n' ',' | sed 's/,$//')

if [ "$MOSCOW_COUNT" -eq 2 ]; then
    echo -e "${GREEN}    ✅ Найдено 2 магазина в Москве: $MOSCOW_NAMES${NC}"
else
    echo -e "${YELLOW}    ⚠ Найдено $MOSCOW_COUNT магазинов в Москве (ожидалось 2)${NC}"
    echo "    Найденные магазины: $MOSCOW_NAMES"
fi

echo -e "${YELLOW}  - Комбинированный фильтр (Москва + тег Кофе)...${NC}"
COMBO_RESPONSE=$(make_get_request "/stores" "?limit=10&tag_id=550e8400-e29b-41d4-a716-446655440002&city_id=3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9")
COMBO_COUNT=$(echo "$COMBO_RESPONSE" | head -n -1 | grep -o '"id":"[^"]*"' | wc -l)
COMBO_NAMES=$(echo "$COMBO_RESPONSE" | head -n -1 | grep -o '"name":"[^"]*"' | cut -d'"' -f4 | tr '\n' ',' | sed 's/,$//')

if [ "$COMBO_COUNT" -eq 1 ]; then
    echo -e "${GREEN}    ✅ Найден 1 магазин в Москве с тегом Кофе: $COMBO_NAMES${NC}"
else
    echo -e "${YELLOW}    ⚠ Найдено $COMBO_COUNT магазинов (ожидалось 1)${NC}"
    echo "    Найденные магазины: $COMBO_NAMES"
fi

echo -e "\n${GREEN}🎉 Store Service полностью протестирован!${NC}"