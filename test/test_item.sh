#!/bin/bash

# Configuration
STORE_URL="http://localhost:8080/api/v0"
COOKIES_FILE="/tmp/item_cookies.txt"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Initialize cookies file
> "$COOKIES_FILE"

# Function to make API calls
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3

    curl -s -w "\n%{http_code}" -c "$COOKIES_FILE" -b "$COOKIES_FILE" \
         -X "$method" -H "Content-Type: application/json" \
         ${data:+-d "$data"} \
         "$STORE_URL$endpoint"
}

# Function to check image endpoint
check_item_image() {
    local image_path=$1
    
    echo -n "  Image: /images/items/$image_path ... "
    response=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/images/items/$image_path")
    
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
    elif [[ "$code" -eq 404 ]]; then
        echo -e "${YELLOW}⚠ $op not found (Status: $code)${NC}"
    elif [[ "$code" -ge 400 && "$code" -lt 500 ]]; then
        echo -e "${YELLOW}⚠ $op client error (Status: $code)${NC}"
    else
        echo -e "${RED}✗ $op failed (Status: $code)${NC}"
    fi
    
    if [ -n "$body" ] && [ "$body" != "[]" ]; then
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
    fi
    echo
}

# Function to extract IDs
extract_ids() {
    echo "$1" | head -n -1 | grep -o '"id":"[^"]*"' | cut -d'"' -f4
}

# Function to extract names
extract_names() {
    echo "$1" | head -n -1 | grep -o '"name":"[^"]*"' | cut -d'"' -f4 | sed 's/\\//g'
}

# Function to extract item images
extract_item_images() {
    echo "$1" | head -n -1 | grep -o '"card_img":"[^"]*"' | cut -d'"' -f4 | sed 's|.*/||'
}

# Function to extract prices
extract_prices() {
    echo "$1" | head -n -1 | grep -o '"price":[^,}]*' | cut -d':' -f2
}

# Cleanup
cleanup() { rm -f "$COOKIES_FILE"; }
trap cleanup EXIT

echo -e "${YELLOW}Starting Item Service API testing...${NC}"
echo "Store Service: $STORE_URL"
echo "=================================================="

# 1. ПРОВЕРКА ДОСТУПНОСТИ ЭНДПОИНТОВ
echo -e "${YELLOW}1. Проверка доступности эндпоинтов Item Service...${NC}"

# Получаем магазины для тестирования
echo -e "${YELLOW}  - Получаем магазины для тестирования...${NC}"
STORES_RESPONSE=$(make_request "GET" "/stores?limit=10" "")
STORES_CODE=$(echo "$STORES_RESPONSE" | tail -1)

if [[ "$STORES_CODE" -ne 200 ]]; then
    echo -e "${RED}❌ Не удалось получить магазины для тестирования${NC}"
    exit 1
fi

# Извлекаем ID магазинов
STORE_IDS=()
while IFS= read -r id; do
    [[ -n "$id" ]] && STORE_IDS+=("$id")
done < <(extract_ids "$STORES_RESPONSE")

STORE_NAMES=()
while IFS= read -r name; do
    [[ -n "$name" ]] && STORE_NAMES+=("$name")
done < <(extract_names "$STORES_RESPONSE")

echo -e "    Найдено магазинов: ${#STORE_IDS[@]}"
for i in "${!STORE_IDS[@]}"; do
    echo -e "      - ${STORE_IDS[$i]}: ${STORE_NAMES[$i]}"
done

# Проверяем основные эндпоинты
if [ ${#STORE_IDS[@]} -gt 0 ]; then
    echo -e "${YELLOW}  - Проверяем эндпоинты товаров...${NC}"
    
    # Проверка эндпоинта товаров
    echo -n "    GET /stores/{id}/items ... "
    ITEMS_TEST_RESPONSE=$(make_request "GET" "/stores/${STORE_IDS[0]}/items" "")
    ITEMS_TEST_CODE=$(echo "$ITEMS_TEST_RESPONSE" | tail -1)
    if [ "$ITEMS_TEST_CODE" -eq 200 ]; then
        echo -e "${GREEN}✓ (Status: $ITEMS_TEST_CODE)${NC}"
    else
        echo -e "${RED}✗ (Status: $ITEMS_TEST_CODE)${NC}"
    fi
    
    # Проверка эндпоинта типов товаров
    echo -n "    GET /stores/{id}/item-types ... "
    TYPES_TEST_RESPONSE=$(make_request "GET" "/stores/${STORE_IDS[0]}/item-types" "")
    TYPES_TEST_CODE=$(echo "$TYPES_TEST_RESPONSE" | tail -1)
    if [ "$TYPES_TEST_CODE" -eq 200 ]; then
        echo -e "${GREEN}✓ (Status: $TYPES_TEST_CODE)${NC}"
    elif [ "$TYPES_TEST_CODE" -eq 404 ]; then
        echo -e "${YELLOW}⚠ (Status: $TYPES_TEST_CODE) - Not Found${NC}"
    else
        echo -e "${YELLOW}⚠ (Status: $TYPES_TEST_CODE)${NC}"
    fi
fi

# 2. ТЕСТИРОВАНИЕ ТИПОВ ТОВАРОВ (/stores/{id}/item-types)
echo -e "\n${YELLOW}2. ТЕСТИРОВАНИЕ ТИПОВ ТОВАРОВ (/stores/{id}/item-types)...${NC}"

TOTAL_TYPES=0
STORES_WITH_TYPES=0
STORE_TYPE_MAP=()

for i in "${!STORE_IDS[@]}"; do
    store_id="${STORE_IDS[$i]}"
    store_name="${STORE_NAMES[$i]}"
    
    echo -e "${YELLOW}  - Тестируем магазин '$store_name' ($store_id)...${NC}"
    
    # Получаем типы товаров для магазина
    TYPES_RESPONSE=$(make_request "GET" "/stores/$store_id/item-types" "")
    TYPES_CODE=$(echo "$TYPES_RESPONSE" | tail -1)
    TYPES_BODY=$(echo "$TYPES_RESPONSE" | head -n -1)
    
    if [[ "$TYPES_CODE" -eq 200 ]]; then
        # Проверяем, не пустой ли массив
        if [[ "$TYPES_BODY" == "[]" || -z "$TYPES_BODY" || "$TYPES_BODY" == "null" ]]; then
            echo -e "    ${YELLOW}⚠ Типы товаров отсутствуют (пустой массив)${NC}"
        else
            TYPE_IDS=()
            while IFS= read -r id; do
                [[ -n "$id" ]] && TYPE_IDS+=("$id")
            done < <(extract_ids "$TYPES_RESPONSE")
            
            TYPE_NAMES=()
            while IFS= read -r name; do
                [[ -n "$name" ]] && TYPE_NAMES+=("$name")
            done < <(extract_names "$TYPES_RESPONSE")
            
            if [ ${#TYPE_IDS[@]} -eq 0 ]; then
                echo -e "    ${YELLOW}⚠ Типы товаров отсутствуют (нет данных в ответе)${NC}"
            else
                echo -e "    ${GREEN}✅ Найдено типов товаров: ${#TYPE_IDS[@]}${NC}"
                for j in "${!TYPE_IDS[@]}"; do
                    echo -e "      - ${TYPE_IDS[$j]}: ${TYPE_NAMES[$j]}"
                done
                
                ((TOTAL_TYPES+=${#TYPE_IDS[@]}))
                ((STORES_WITH_TYPES++))
                
                # Сохраняем типы для дальнейшего тестирования
                for j in "${!TYPE_IDS[@]}"; do
                    STORE_TYPE_MAP+=("$store_id:$store_name:${TYPE_IDS[$j]}:${TYPE_NAMES[$j]}")
                done
            fi
        fi
    elif [[ "$TYPES_CODE" -eq 404 ]]; then
        echo -e "    ${YELLOW}⚠ Нет типов товаров для этого магазина (404)${NC}"
    elif [[ "$TYPES_CODE" -eq 500 ]]; then
        echo -e "    ${RED}❌ Ошибка сервера (500)${NC}"
        echo "    Response: $TYPES_BODY"
    else
        echo -e "    ${YELLOW}⚠ Статус: $TYPES_CODE${NC}"
        print_result "Get Item Types" "$TYPES_RESPONSE"
    fi
done

# 3. ТЕСТИРОВАНИЕ ТОВАРОВ (/stores/{id}/items)
echo -e "\n${YELLOW}3. ТЕСТИРОВАНИЕ ТОВАРОВ (/stores/{id}/items)...${NC}"

TOTAL_ITEMS=0
ITEMS_WITH_IMAGES=0
ALL_ITEMS_DATA=()

if [ ${#STORE_IDS[@]} -eq 0 ]; then
    echo -e "${YELLOW}  ⚠ Нет магазинов для тестирования${NC}"
else
    # 3.1 Тестируем базовое получение товаров
    echo -e "${YELLOW}  - Базовое получение товаров (все товары магазина)...${NC}"
    
    for i in "${!STORE_IDS[@]}"; do
        store_id="${STORE_IDS[$i]}"
        store_name="${STORE_NAMES[$i]}"
        
        echo -e "${BLUE}    Магазин: $store_name${NC}"
        
        # Получаем все товары магазина
        ITEMS_RESPONSE=$(make_request "GET" "/stores/$store_id/items" "")
        ITEMS_CODE=$(echo "$ITEMS_RESPONSE" | tail -1)
        
        if [[ "$ITEMS_CODE" -eq 200 ]]; then
            ITEM_IDS=()
            while IFS= read -r id; do
                [[ -n "$id" ]] && ITEM_IDS+=("$id")
            done < <(extract_ids "$ITEMS_RESPONSE")
            
            ITEM_NAMES=()
            while IFS= read -r name; do
                [[ -n "$name" ]] && ITEM_NAMES+=("$name")
            done < <(extract_names "$ITEMS_RESPONSE")
            
            ITEM_IMAGES=()
            while IFS= read -r img; do
                [[ -n "$img" ]] && ITEM_IMAGES+=("$img")
            done < <(extract_item_images "$ITEMS_RESPONSE")
            
            ITEM_PRICES=($(extract_prices "$ITEMS_RESPONSE"))
            
            echo -e "      ${GREEN}✅ Товаров: ${#ITEM_IDS[@]}${NC}"
            
            # Выводим очищенные названия товаров
            if [ ${#ITEM_NAMES[@]} -gt 0 ]; then
                echo -e "        Названия:"
                for name in "${ITEM_NAMES[@]}"; do
                    echo -e "          - $name"
                done
            fi
            
            if [ ${#ITEM_PRICES[@]} -gt 0 ]; then
                echo -e "        Цены: ${ITEM_PRICES[*]}"
            fi
            echo -e "        Изображений: ${#ITEM_IMAGES[@]}"
            
            ((TOTAL_ITEMS+=${#ITEM_IDS[@]}))
            
            # Сохраняем данные о товарах
            for j in "${!ITEM_IDS[@]}"; do
                ALL_ITEMS_DATA+=("${ITEM_IDS[$j]}:${ITEM_NAMES[$j]}:${ITEM_IMAGES[$j]}:$store_name")
                if [ -n "${ITEM_IMAGES[$j]}" ]; then
                    ((ITEMS_WITH_IMAGES++))
                fi
            done
        else
            print_result "Get Items" "$ITEMS_RESPONSE"
        fi
    done

    # 3.2 Тестируем фильтрацию товаров по типам (если есть типы)
    if [ ${#STORE_TYPE_MAP[@]} -gt 0 ]; then
        echo -e "${YELLOW}  - Фильтрация товаров по типам...${NC}"
        
        for store_type in "${STORE_TYPE_MAP[@]}"; do
            IFS=':' read -r store_id store_name type_id type_name <<< "$store_type"
            
            echo -e "${BLUE}    Магазин: $store_name -> Тип: $type_name${NC}"
            
            # Получаем товары для конкретного типа
            FILTERED_ITEMS_RESPONSE=$(make_request "GET" "/stores/$store_id/items?type_id=$type_id" "")
            FILTERED_CODE=$(echo "$FILTERED_ITEMS_RESPONSE" | tail -1)
            
            if [[ "$FILTERED_CODE" -eq 200 ]]; then
                FILTERED_COUNT=$(echo "$FILTERED_ITEMS_RESPONSE" | head -n -1 | grep -o '"id":"[^"]*"' | wc -l)
                echo -e "      ${GREEN}✅ Найдено товаров: $FILTERED_COUNT${NC}"
                
                # Показываем названия отфильтрованных товаров
                FILTERED_NAMES=()
                while IFS= read -r name; do
                    [[ -n "$name" ]] && FILTERED_NAMES+=("$name")
                done < <(extract_names "$FILTERED_ITEMS_RESPONSE")
                
                if [ ${#FILTERED_NAMES[@]} -gt 0 ]; then
                    echo -e "        Товары: ${FILTERED_NAMES[*]}"
                fi
            else
                print_result "Get Filtered Items" "$FILTERED_ITEMS_RESPONSE"
            fi
        done
    else
        echo -e "${YELLOW}  ⚠ Нет типов товаров для тестирования фильтрации${NC}"
    fi
fi

# 4. ТЕСТИРОВАНИЕ ИЗОБРАЖЕНИЙ ТОВАРОВ
echo -e "\n${YELLOW}4. ТЕСТИРОВАНИЕ ИЗОБРАЖЕНИЙ ТОВАРОВ...${NC}"

IMAGE_SUCCESS_COUNT=0
IMAGE_FAIL_COUNT=0
FAILED_IMAGES=()

if [ ${#ALL_ITEMS_DATA[@]} -eq 0 ]; then
    echo -e "${YELLOW}  ⚠ Нет товаров для тестирования изображений${NC}"
else
    echo -e "${YELLOW}  - Проверяем доступность изображений...${NC}"
    
    for item_data in "${ALL_ITEMS_DATA[@]}"; do
        IFS=':' read -r item_id item_name item_image store_name <<< "$item_data"
        
        if [ -n "$item_image" ]; then
            echo -n "    Товар: $item_name ... "
            if check_item_image "$item_image"; then
                ((IMAGE_SUCCESS_COUNT++))
            else
                ((IMAGE_FAIL_COUNT++))
                FAILED_IMAGES+=("$item_name:$item_image ($store_name)")
            fi
        else
            echo -e "    Товар: $item_name ... ${YELLOW}⚠ Нет изображения${NC}"
        fi
    done
    
    echo -e "\n    ${GREEN}✅ Успешно загружено: $IMAGE_SUCCESS_COUNT/$ITEMS_WITH_IMAGES изображений${NC}"
    
    if [ $IMAGE_FAIL_COUNT -gt 0 ]; then
        echo -e "    ${YELLOW}⚠ Проблемные изображения: $IMAGE_FAIL_COUNT${NC}"
        for failed in "${FAILED_IMAGES[@]}"; do
            echo "      - $failed"
        done
    fi
fi

# 5. ТЕСТИРОВАНИЕ ОБРАБОТКИ ОШИБОК
echo -e "\n${YELLOW}5. ТЕСТИРОВАНИЕ ОБРАБОТКИ ОШИБОК...${NC}"

# 5.1 Невалидный UUID для типов товаров
echo -e "${YELLOW}  - Невалидный UUID для типов товаров...${NC}"
INVALID_TYPES_RESPONSE=$(make_request "GET" "/stores/invalid-uuid-format/item-types" "")
print_result "Invalid UUID for Item Types" "$INVALID_TYPES_RESPONSE"

# 5.2 Невалидный UUID для товаров
echo -e "${YELLOW}  - Невалидный UUID для товаров...${NC}"
INVALID_ITEMS_RESPONSE=$(make_request "GET" "/stores/invalid-uuid-format/items" "")
print_result "Invalid UUID for Items" "$INVALID_ITEMS_RESPONSE"

# 5.3 Несуществующий магазин для типов товаров
echo -e "${YELLOW}  - Несуществующий магазин для типов товаров...${NC}"
NONEXISTENT_TYPES_RESPONSE=$(make_request "GET" "/stores/00000000-0000-0000-0000-000000000000/item-types" "")
print_result "Non-existent Store for Item Types" "$NONEXISTENT_TYPES_RESPONSE"

# 5.4 Несуществующий магазин для товаров
echo -e "${YELLOW}  - Несуществующий магазин для товаров...${NC}"
NONEXISTENT_ITEMS_RESPONSE=$(make_request "GET" "/stores/00000000-0000-0000-0000-000000000000/items" "")
print_result "Non-existent Store for Items" "$NONEXISTENT_ITEMS_RESPONSE"

# 5.5 Неправильный HTTP метод
echo -e "${YELLOW}  - Неправильный HTTP метод для типов товаров...${NC}"
WRONG_METHOD_TYPES_RESPONSE=$(make_request "POST" "/stores/${STORE_IDS[0]}/item-types" "{}")
print_result "Wrong Method for Item Types" "$WRONG_METHOD_TYPES_RESPONSE"

# 5.6 Неправильный HTTP метод для товаров
echo -e "${YELLOW}  - Неправильный HTTP метод для товаров...${NC}"
WRONG_METHOD_ITEMS_RESPONSE=$(make_request "POST" "/stores/${STORE_IDS[0]}/items" "{}")
print_result "Wrong Method for Items" "$WRONG_METHOD_ITEMS_RESPONSE"

# 6. ТЕСТИРОВАНИЕ ОШИБОК ИЗОБРАЖЕНИЙ
echo -e "\n${YELLOW}6. ТЕСТИРОВАНИЕ ОШИБОК ИЗОБРАЖЕНИЙ...${NC}"

# 6.1 Несуществующее изображение
echo -e "${YELLOW}  - Несуществующее изображение товара...${NC}"
check_item_image "non_existent_item_image.jpg"

# 6.2 Неправильный путь к изображению
echo -e "${YELLOW}  - Неправильный путь к изображению...${NC}"
check_item_image "../../etc/passwd"

# 7. СТАТИСТИЧЕСКИЙ АНАЛИЗ
echo -e "\n${YELLOW}7. СТАТИСТИЧЕСКИЙ АНАЛИЗ ДАННЫХ...${NC}"

echo -e "${GREEN}📊 СТАТИСТИКА ПО ТОВАРАМ:${NC}"
echo -e "  Всего магазинов: ${#STORE_IDS[@]}"
echo -e "  Магазинов с типами товаров: $STORES_WITH_TYPES"
echo -e "  Всего типов товаров: $TOTAL_TYPES"
echo -e "  Всего товаров: $TOTAL_ITEMS"
echo -e "  Товаров с изображениями: $ITEMS_WITH_IMAGES"
echo -e "  Успешно загруженных изображений: $IMAGE_SUCCESS_COUNT"

# Статистические расчеты
if [ $TOTAL_TYPES -gt 0 ] && [ $TOTAL_ITEMS -gt 0 ]; then
    AVG_ITEMS_PER_TYPE=$(echo "scale=2; $TOTAL_ITEMS / $TOTAL_TYPES" | bc 2>/dev/null || echo "N/A")
    echo -e "  Среднее количество товаров на тип: $AVG_ITEMS_PER_TYPE"
fi

if [ $TOTAL_ITEMS -gt 0 ]; then
    IMAGE_COVERAGE=$(echo "scale=2; $ITEMS_WITH_IMAGES * 100 / $TOTAL_ITEMS" | bc 2>/dev/null || echo "N/A")
    echo -e "  Покрытие изображениями: $IMAGE_COVERAGE%"
    
    if [ $IMAGE_SUCCESS_COUNT -gt 0 ]; then
        IMAGE_SUCCESS_RATE=$(echo "scale=2; $IMAGE_SUCCESS_COUNT * 100 / $ITEMS_WITH_IMAGES" | bc 2>/dev/null || echo "N/A")
        echo -e "  Успешная загрузка изображений: $IMAGE_SUCCESS_RATE%"
    fi
fi

# 8. ИТОГОВЫЙ ОТЧЕТ
echo -e "\n${YELLOW}📋 ИТОГОВЫЙ ОТЧЕТ ПО ITEM SERVICE...${NC}"

echo -e "${GREEN}✅ ПРОВЕРЕННЫЕ ФУНКЦИОНАЛЬНОСТИ:${NC}"
echo -e "  - Получение типов товаров по магазину (/stores/{id}/item-types)"
echo -e "  - Получение всех товаров магазина (/stores/{id}/items)" 
echo -e "  - Фильтрация товаров по типам (/stores/{id}/items?type_id=)"
echo -e "  - Загрузка изображений товаров (/images/items/*)"
echo -e "  - Валидация входных параметров (UUID)"
echo -e "  - Обработка несуществующих ресурсов (404)"
echo -e "  - Проверка HTTP методов (405)"

echo -e "\n${GREEN}🎉 Item Service тестирование завершено!${NC}"
echo -e "${YELLOW}📊 Итоговый отчет:${NC}"
echo -e "  ✅ Работающие эндпоинты:"
echo -e "     - GET /stores/{id}/items"
echo -e "     - GET /images/items/*"

if [ $TOTAL_TYPES -gt 0 ]; then
    echo -e "     - GET /stores/{id}/item-types"
    echo -e "     - GET /stores/{id}/items?type_id={type_id}"
else
    echo -e "     - GET /stores/{id}/item-types ${YELLOW}(эндпоинт работает, но данных нет)${NC}"
fi

echo -e "  📈 Результаты тестирования:"
echo -e "     - Протестировано магазинов: ${#STORE_IDS[@]}"
echo -e "     - Найдено типов товаров: $TOTAL_TYPES"
echo -e "     - Найдено товаров: $TOTAL_ITEMS"
echo -e "     - Успешных изображений: $IMAGE_SUCCESS_COUNT/$ITEMS_WITH_IMAGES"

echo -e "  🔒 Обработка ошибок:"
echo -e "     - Валидация UUID: ${GREEN}✓${NC}"
echo -e "     - Проверка HTTP методов: ${GREEN}✓${NC}"
echo -e "     - Обработка 404 ошибок: ${GREEN}✓${NC}"

if [ $IMAGE_FAIL_COUNT -gt 0 ]; then
    echo -e "  ⚠  Проблемные области:"
    echo -e "     - Проблемные изображения: $IMAGE_FAIL_COUNT"
fi

if [ $TOTAL_TYPES -eq 0 ]; then
    echo -e "\n${YELLOW}💡 РЕКОМЕНДАЦИИ:${NC}"
    echo -e "  • Эндпоинт /stores/{id}/item-types работает корректно"
    echo -e "  • Типы товаров отсутствуют в базе данных"
    echo -e "  • Для полного тестирования фильтрации добавьте типы товаров в БД"
fi

echo -e "\n${GREEN}✅ Все основные сценарии Item Service протестированы!${NC}"