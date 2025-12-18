-- db/migrations/XXX_create_users_and_orders.sql (замени XXX на следующий номер)

-- +up

-- Пользователь 1: пустой пользователь (только регистрация)
INSERT INTO account (id, email, hash, name, city_id, address, avatar_url) 
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'empty@example.com',
    '$2a$10$randomhashforemptyuser', -- bcrypt хэш (password: "password")
    'Пустой Пользователь',
    '3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9', -- ID города из ваших данных
    'ул. Тестовая, 1',
    '/avatars/empty.jpg'
) ON CONFLICT (email) DO NOTHING;

-- Пользователь 2: активный пользователь с 10 заказами
INSERT INTO account (id, email, hash, name, city_id, address, avatar_url) 
VALUES (
    '22222222-2222-2222-2222-222222222222',
    'active@example.com',
    '$2a$10$randomhashforactiveuser', -- bcrypt хэш (password: "password")
    'Активный Покупатель',
    '3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9',
    'ул. Заказов, 10',
    '/avatars/active.jpg'
) ON CONFLICT (email) DO NOTHING;

-- 10 заказов для активного пользователя
WITH orders_data AS (
    SELECT 
        unnest(ARRAY[
            'a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1',
            'b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2',
            'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3',
            'd4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4',
            'e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e5e5',
            'f6f6f6f6-f6f6-f6f6-f6f6-f6f6f6f6f6f6',
            'g7g7g7g7-g7g7-g7g7-g7g7-g7g7g7g7g7g7',
            'h8h8h8h8-h8h8-h8h8-h8h8-h8h8h8h8h8h8',
            'i9i9i9i9-i9i9-i9i9-i9i9-i9i9i9i9i9i9',
            'j0j0j0j0-j0j0-j0j0-j0j0-j0j0j0j0j0j0'
        ]) as id,
        unnest(ARRAY[
            1250.00, -- Сет "Самурай"
            590.00,  -- Пицца "Пепперони" + лимонад
            340.00,  -- Сабвей "Турецкий"
            650.00,  -- Шашлык из свинины
            1450.00, -- Сет "Харка"
            620.00,  -- Пицца "Четыре сыра"
            370.00,  -- Сабвей "Итальянский БМТ"
            720.00,  -- Шашлык из баранины
            580.00,  -- Шашлык из курицы + лепешка
            340.00   -- Чизкейк Нью-Йорк + кофе
        ]) as total_price,
        unnest(ARRAY[
            'paid',
            'pending',
            'paid',
            'pending',
            'paid',
            'pending',
            'pending',
            'paid',
            'pending',
            'paid'
        ]) as status,
        unnest(ARRAY[
            false,
            false,
            true,
            false,
            false,
            false,
            true,
            false,
            false,
            false
        ]) as is_fast
)
INSERT INTO "orders" (id, user_id, total_price, status, is_fast, comment, created_at)
SELECT 
    o.id,
    '22222222-2222-2222-2222-222222222222',
    o.total_price,
    o.status::order_status,
    o.is_fast,
    CASE 
        WHEN o.status = 'cancelled' THEN 'Передумал'
        WHEN o.is_fast THEN 'Срочный заказ'
        ELSE NULL 
    END,
    now() - (random() * interval '30 days') -- заказы за последние 30 дней
FROM orders_data o
ON CONFLICT (id) DO NOTHING;

-- Товары в заказах (по 1-2 товара на заказ)
INSERT INTO order_item (id, order_id, store_item_id, price, quantity)
VALUES
    -- Заказ 1: Сет "Самурай" (1 товар)
    (gen_random_uuid(), 'a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%Самурай%' LIMIT 1),
        1250.00, 1),
    
    -- Заказ 2: Пицца "Пепперони" + лимонад (2 товара)
    (gen_random_uuid(), 'b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%Пепперони%' LIMIT 1),
        590.00, 1),
    (gen_random_uuid(), 'b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%Лимонад%' LIMIT 1),
        180.00, 1),
    
    -- Заказ 3: Сабвей "Турецкий" (1 товар)
    (gen_random_uuid(), 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%Турецкий%' LIMIT 1),
        340.00, 1),
    
    -- Заказ 4: Шашлык из свинины (1 товар)
    (gen_random_uuid(), 'd4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%свинины%' LIMIT 1),
        650.00, 1),
    
    -- Заказ 5: Сет "Харка" (1 товар)
    (gen_random_uuid(), 'e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e5e5',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%Харка%' LIMIT 1),
        1450.00, 1),
    
    -- Заказ 6: Пицца "Четыре сыра" (1 товар)
    (gen_random_uuid(), 'f6f6f6f6-f6f6-f6f6-f6f6-f6f6f6f6f6f6',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%четыре сыра%' LIMIT 1),
        620.00, 1),
    
    -- Заказ 7: Сабвей "Итальянский БМТ" (1 товар)
    (gen_random_uuid(), 'g7g7g7g7-g7g7-g7g7-g7g7-g7g7g7g7g7g7',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%Итальянский БМТ%' LIMIT 1),
        370.00, 1),
    
    -- Заказ 8: Шашлык из баранины (1 товар)
    (gen_random_uuid(), 'h8h8h8h8-h8h8-h8h8-h8h8-h8h8h8h8h8h8',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%баранины%' LIMIT 1),
        720.00, 1),
    
    -- Заказ 9: Шашлык из курицы + лепешка (2 товара)
    (gen_random_uuid(), 'i9i9i9i9-i9i9-i9i9-i9i9-i9i9i9i9i9i9',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%курицы%' LIMIT 1),
        580.00, 1),
    (gen_random_uuid(), 'i9i9i9i9-i9i9-i9i9-i9i9-i9i9i9i9i9i9',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%лепешка%' LIMIT 1),
        130.00, 1),
    
    -- Заказ 10: Чизкейк Нью-Йорк + кофе (2 товара)
    (gen_random_uuid(), 'j0j0j0j0-j0j0-j0j0-j0j0-j0j0j0j0j0j0',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%Чизкейк%' LIMIT 1),
        340.00, 1),
    (gen_random_uuid(), 'j0j0j0j0-j0j0-j0j0-j0j0-j0j0j0j0j0j0',
        (SELECT si.id FROM store_item si 
         JOIN item i ON si.item_id = i.id 
         WHERE i.name LIKE '%Американо%' LIMIT 1),
        140.00, 1)
ON CONFLICT DO NOTHING;

-- +down

-- Удалить заказы и пользователей (в обратном порядке)
DELETE FROM order_item WHERE order_id IN (
    'a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1',
    'b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2',
    'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3',
    'd4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4',
    'e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e5e5',
    'f6f6f6f6-f6f6-f6f6-f6f6-f6f6f6f6f6f6',
    'g7g7g7g7-g7g7-g7g7-g7g7-g7g7g7g7g7g7',
    'h8h8h8h8-h8h8-h8h8-h8h8-h8h8h8h8h8h8',
    'i9i9i9i9-i9i9-i9i9-i9i9-i9i9i9i9i9i9',
    'j0j0j0j0-j0j0-j0j0-j0j0-j0j0j0j0j0j0'
);

DELETE FROM "orders" WHERE id IN (
    'a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1',
    'b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2',
    'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3',
    'd4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4',
    'e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e5e5',
    'f6f6f6f6-f6f6-f6f6-f6f6-f6f6f6f6f6f6',
    'g7g7g7g7-g7g7-g7g7-g7g7-g7g7g7g7g7g7',
    'h8h8h8h8-h8h8-h8h8-h8h8-h8h8h8h8h8h8',
    'i9i9i9i9-i9i9-i9i9-i9i9-i9i9i9i9i9i9',
    'j0j0j0j0-j0j0-j0j0-j0j0-j0j0j0j0j0j0'
);

DELETE FROM account WHERE email IN ('empty@example.com', 'active@example.com');