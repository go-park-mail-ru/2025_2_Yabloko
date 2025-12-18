INSERT INTO account (id, email, hash, name, city_id, address, avatar_url) 
VALUES (
    '11111111-1111-1111-1111-111111111111'::uuid,
    'empty@example.com',
    '$2a$10$9q4j/k4rB/oWGIAiBr8OXOaW3IyahWI55I.BxKXBzBdO3hsMEYbuS',
    'Пустой Пользователь',
    '3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9'::uuid,
    'ул. Тестовая, 1',
    '/avatars/empty.jpg'
) ON CONFLICT (email) DO NOTHING;

INSERT INTO account (id, email, hash, name, city_id, address, avatar_url) 
VALUES (
    '22222222-2222-2222-2222-222222222222'::uuid,
    'active@example.com',
    '$2a$10$9q4j/k4rB/oWGIAiBr8OXOaW3IyahWI55I.BxKXBzBdO3hsMEYbuS',
    'Активный Покупатель',
    '3b77c3c9-8b6f-4e9f-94f1-7f0a7a4ad5b9'::uuid,
    'ул. Заказов, 10',
    '/avatars/active.jpg'
) ON CONFLICT (email) DO NOTHING;

INSERT INTO "orders" (id, user_id, total_price, status, is_fast, comment, created_at)
VALUES
    ('a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1'::uuid, '22222222-2222-2222-2222-222222222222'::uuid, 1250.00, 'paid'::order_status,      false, NULL, now() - interval '1 day'),
    ('b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2'::uuid, '22222222-2222-2222-2222-222222222222'::uuid, 770.00,  'delivered'::order_status, false, NULL, now() - interval '2 days'),
    ('c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3'::uuid, '22222222-2222-2222-2222-222222222222'::uuid, 340.00,  'on_the_way'::order_status, true,  'Срочный заказ', now() - interval '3 hours'),
    ('d4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4'::uuid, '22222222-2222-2222-2222-222222222222'::uuid, 650.00,  'paid'::order_status,      false, NULL, now() - interval '5 days'),
    ('e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e5e5'::uuid, '22222222-2222-2222-2222-222222222222'::uuid, 1450.00, 'delivered'::order_status, false, NULL, now() - interval '10 days'),
    ('f6f6f6f6-f6f6-f6f6-f6f6-f6f6f6f6f6f6'::uuid, '22222222-2222-2222-2222-222222222222'::uuid, 620.00,  'pending'::order_status,   false, NULL, now() - interval '1 hour'),
    ('a7a7a7a7-a7a7-a7a7-a7a7-a7a7a7a7a7a7'::uuid, '22222222-2222-2222-2222-222222222222'::uuid, 370.00,  'on_the_way'::order_status, true,  'Срочный заказ', now() - interval '2 hours'),
    ('b8b8b8b8-b8b8-b8b8-b8b8-b8b8b8b8b8b8'::uuid, '22222222-2222-2222-2222-222222222222'::uuid, 720.00,  'paid'::order_status,      false, NULL, now() - interval '15 days'),
    ('c9c9c9c9-c9c9-c9c9-c9c9-c9c9c9c9c9c9'::uuid, '22222222-2222-2222-2222-222222222222'::uuid, 710.00,  'delivered'::order_status, false, NULL, now() - interval '20 days'),
    ('d0d0d0d0-d0d0-d0d0-d0d0-d0d0d0d0d0d0'::uuid, '22222222-2222-2222-2222-222222222222'::uuid, 480.00,  'cancelled'::order_status, false, 'Передумал',     now() - interval '7 days')
ON CONFLICT (id) DO NOTHING;


INSERT INTO order_item (id, order_id, store_item_id, price, quantity)
SELECT 
    gen_random_uuid(),
    o.id,
    si.id,
    CASE 
        WHEN o.id = 'a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1'::uuid THEN 1250.00
        WHEN o.id = 'b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2'::uuid THEN 590.00
        WHEN o.id = 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3'::uuid THEN 340.00
        WHEN o.id = 'd4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4'::uuid THEN 650.00
        WHEN o.id = 'e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e5e5'::uuid THEN 1450.00
        WHEN o.id = 'f6f6f6f6-f6f6-f6f6-f6f6-f6f6f6f6f6f6'::uuid THEN 620.00
        WHEN o.id = 'a7a7a7a7-a7a7-a7a7-a7a7-a7a7a7a7a7a7'::uuid THEN 370.00
        WHEN o.id = 'b8b8b8b8-b8b8-b8b8-b8b8-b8b8b8b8b8b8'::uuid THEN 720.00
        WHEN o.id = 'c9c9c9c9-c9c9-c9c9-c9c9-c9c9c9c9c9c9'::uuid THEN 580.00
        WHEN o.id = 'd0d0d0d0-d0d0-d0d0-d0d0-d0d0d0d0d0d0'::uuid THEN 340.00
    END,
    1
FROM 
    (SELECT unnest(ARRAY[
        'a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1',
        'b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2',
        'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3',
        'd4d4d4d4-d4d4-d4d4-d4d4-d4d4d4d4d4d4',
        'e5e5e5e5-e5e5-e5e5-e5e5-e5e5e5e5e5e5',
        'f6f6f6f6-f6f6-f6f6-f6f6-f6f6f6f6f6f6',
        'a7a7a7a7-a7a7-a7a7-a7a7-a7a7a7a7a7a7',
        'b8b8b8b8-b8b8-b8b8-b8b8-b8b8b8b8b8b8',
        'c9c9c9c9-c9c9-c9c9-c9c9-c9c9c9c9c9c9',
        'd0d0d0d0-d0d0-d0d0-d0d0-d0d0d0d0d0d0'
    ])::uuid as id) o
CROSS JOIN LATERAL (
    SELECT id FROM store_item 
    ORDER BY RANDOM() 
    LIMIT 1
) si
ON CONFLICT DO NOTHING;
