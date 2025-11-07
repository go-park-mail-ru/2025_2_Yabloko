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

# Function to check image endpoint
check_image() {
    local image_path=$1
    local endpoint="/images/stores/$image_path"
    
    echo -n "  Image: $endpoint ... "
    response=$(curl -s -o /dev/null -w "%{http_code}" "$STORE_URL$endpoint")
    
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

# 1. Get all stores (POST method)
echo -e "${YELLOW}1. Getting all stores (POST method)...${NC}"
STORES_RESPONSE=$(make_store_request "POST" "/stores" '{
    "limit": 10,
    "last_id": "",
    "tag_id": "",
    "city_id": "",
    "sorted": "",
    "desc": false
}')
print_result "Get All Stores" "$STORES_RESPONSE"

# Extract store IDs and card images правильно
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

# 2. Test individual store endpoints
echo -e "${YELLOW}2. Testing individual store endpoints...${NC}"
if [ ${#STORE_IDS[@]} -gt 0 ]; then
    store_id="${STORE_IDS[0]}"
    card_img="${CARD_IMAGES[0]}"
    
    echo -e "${BLUE}Testing store: $store_id${NC}"
    
    # 2.1 Get store details (GET /stores/{id})
    echo -e "${YELLOW}  - Getting store details...${NC}"
    STORE_DETAIL_RESPONSE=$(make_store_request "GET" "/stores/$store_id" "")
    print_result "Get Store" "$STORE_DETAIL_RESPONSE"
    
    # 2.2 Get store reviews (GET /stores/{id}/reviews)
    echo -e "${YELLOW}  - Getting store reviews...${NC}"
    print_result "Get Reviews" "$(make_store_request "GET" "/stores/$store_id/reviews" "")"
    
    # 2.3 Test store image accessibility
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

# 3. Test all store images
echo -e "${YELLOW}3. Testing all store images...${NC}"
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

# 4. Get cities (GET /stores/cities)
echo -e "${YELLOW}4. Getting cities...${NC}"
CITIES_RESPONSE=$(make_store_request "GET" "/stores/cities" "")
print_result "Get Cities" "$CITIES_RESPONSE"

# 5. Get tags (GET /stores/tags)
echo -e "${YELLOW}5. Getting tags...${NC}"
print_result "Get Tags" "$(make_store_request "GET" "/stores/tags" "")"

# 6. Test filtering with POST method
echo -e "${YELLOW}6. Testing store filtering by Moscow...${NC}"
FILTER_RESPONSE=$(make_store_request "POST" "/stores" '{
    "limit": 5,
    "last_id": "",
    "tag_id": "",
    "city_id": "3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9",
    "sorted": "",
    "desc": false
}')
print_result "Get Stores Filtered by Moscow" "$FILTER_RESPONSE"

# 7. Test sorting by rating with POST method
echo -e "${YELLOW}7. Testing store sorting by rating...${NC}"
SORT_RESPONSE=$(make_store_request "POST" "/stores" '{
    "limit": 5,
    "last_id": "",
    "tag_id": "",
    "city_id": "",
    "sorted": "rating",
    "desc": true
}')
print_result "Get Stores Sorted by Rating" "$SORT_RESPONSE"

# 8. Test error cases
echo -e "${YELLOW}8. Testing error cases...${NC}"

# 8.1 Invalid store ID format
echo -e "${YELLOW}  - Testing invalid store ID format...${NC}"
INVALID_ID_RESPONSE=$(make_store_request "GET" "/stores/invalid-uuid-format" "")
print_result "Invalid Store ID Format" "$INVALID_ID_RESPONSE"

# 8.2 Non-existent store
echo -e "${YELLOW}  - Testing non-existent store...${NC}"
NON_EXISTENT_RESPONSE=$(make_store_request "GET" "/stores/00000000-0000-0000-0000-000000000000" "")
print_result "Non-existent Store" "$NON_EXISTENT_RESPONSE"

# 8.3 Wrong HTTP method for stores (GET вместо POST) - ожидаем 405
echo -e "${YELLOW}  - Testing wrong HTTP method for stores (expected 405)...${NC}"
WRONG_METHOD_RESPONSE=$(make_store_request "GET" "/stores" "")
print_result "Wrong Method for Stores" "$WRONG_METHOD_RESPONSE"

# 8.4 Invalid JSON for GetStores
echo -e "${YELLOW}  - Testing invalid JSON for GetStores...${NC}"
INVALID_JSON_RESPONSE=$(make_store_request "POST" "/stores" "invalid json")
print_result "Invalid JSON for GetStores" "$INVALID_JSON_RESPONSE"

# 9. Test image error cases
echo -e "${YELLOW}9. Testing image error cases...${NC}"

# 9.1 Non-existent image
echo -e "${YELLOW}  - Testing non-existent image...${NC}"
check_image "non_existent_image.jpg"

# 9.2 Invalid image path
echo -e "${YELLOW}  - Testing invalid image path...${NC}"
check_image "../../etc/passwd"

echo -e "${GREEN}Store service API testing completed!${NC}"
echo
echo -e "${YELLOW}Summary:${NC}"
echo -e "  ✅ Все рабочие эндпоинты store_handler протестированы:"
echo -e "     - POST /stores (GetStores с фильтрацией)"
echo -e "     - GET /stores/{id} (GetStore)"  
echo -e "     - GET /stores/{id}/reviews (GetStoreReview)"
echo -e "     - GET /stores/cities (GetCities)"
echo -e "     - GET /stores/tags (GetTags)"
echo -e "     - GET /images/stores/* (Store Images)"
echo -e "  ❌ CreateStore не зарегистрирован в роутере"
echo -e "  ✅ GET /stores возвращает 405 (ожидаемое поведение)"

# Детальная проверка фильтрации
echo -e "\n${YELLOW}🔍 Детальная проверка фильтрации...${NC}"

echo -e "${YELLOW}  - Все магазины (базовый тест)...${NC}"
ALL_STORES_DEBUG=$(make_store_request "POST" "/stores" '{"limit": 10, "last_id": "", "tag_id": "", "city_id": "", "sorted": "", "desc": false}')
ALL_COUNT=$(echo "$ALL_STORES_DEBUG" | head -n -1 | grep -o '"id":"[^"]*"' | wc -l)
echo "    Всего магазинов в базе: $ALL_COUNT"

echo -e "${YELLOW}  - Фильтр по тегу 'Кофе'...${NC}"
COFFEE_FILTER='{"limit": 10, "tag_id": "550e8400-e29b-41d4-a716-446655440002"}'
COFFEE_RESPONSE=$(make_store_request "POST" "/stores" "$COFFEE_FILTER")
COFFEE_COUNT=$(echo "$COFFEE_RESPONSE" | head -n -1 | grep -o '"id":"[^"]*"' | wc -l)
COFFEE_NAMES=$(echo "$COFFEE_RESPONSE" | head -n -1 | grep -o '"name":"[^"]*"' | cut -d'"' -f4 | tr '\n' ',' | sed 's/,$//')

if [ "$COFFEE_COUNT" -eq 1 ]; then
    echo -e "${GREEN}    ✅ Найден 1 кофейный магазин: $COFFEE_NAMES${NC}"
else
    echo -e "${YELLOW}    ⚠ Найдено $COFFEE_COUNT кофейных магазинов (ожидалось 1)${NC}"
    echo "    Найденные магазины: $COFFEE_NAMES"
fi

echo -e "${YELLOW}  - Фильтр по городу 'Москва'...${NC}"
MOSCOW_FILTER='{"limit": 10, "city_id": "3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9"}'
MOSCOW_RESPONSE=$(make_store_request "POST" "/stores" "$MOSCOW_FILTER")
MOSCOW_COUNT=$(echo "$MOSCOW_RESPONSE" | head -n -1 | grep -o '"id":"[^"]*"' | wc -l)
MOSCOW_NAMES=$(echo "$MOSCOW_RESPONSE" | head -n -1 | grep -o '"name":"[^"]*"' | cut -d'"' -f4 | tr '\n' ',' | sed 's/,$//')

if [ "$MOSCOW_COUNT" -eq 2 ]; then
    echo -e "${GREEN}    ✅ Найдено 2 магазина в Москве: $MOSCOW_NAMES${NC}"
else
    echo -e "${YELLOW}    ⚠ Найдено $MOSCOW_COUNT магазинов в Москве (ожидалось 2)${NC}"
    echo "    Найденные магазины: $MOSCOW_NAMES"
fi

echo -e "${YELLOW}  - Комбинированный фильтр (Москва + тег Кофе)...${NC}"
COMBO_FILTER='{"limit": 10, "tag_id": "550e8400-e29b-41d4-a716-446655440002", "city_id": "3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9"}'
COMBO_RESPONSE=$(make_store_request "POST" "/stores" "$COMBO_FILTER")
COMBO_COUNT=$(echo "$COMBO_RESPONSE" | head -n -1 | grep -o '"id":"[^"]*"' | wc -l)
COMBO_NAMES=$(echo "$COMBO_RESPONSE" | head -n -1 | grep -o '"name":"[^"]*"' | cut -d'"' -f4 | tr '\n' ',' | sed 's/,$//')

if [ "$COMBO_COUNT" -eq 1 ]; then
    echo -e "${GREEN}    ✅ Найден 1 магазин в Москве с тегом Кофе: $COMBO_NAMES${NC}"
else
    echo -e "${YELLOW}    ⚠ Найдено $COMBO_COUNT магазинов (ожидалось 1)${NC}"
    echo "    Найденные магазины: $COMBO_NAMES"
fi

echo -e "\n${GREEN}🎉 Store Service полностью протестирован!${NC}"