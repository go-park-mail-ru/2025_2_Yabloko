-- migrations/016_insert_store_tags_and_reviews.sql

-- 1. Добавьте теги в таблицу tag (с правильными UUID)
INSERT INTO tag (id, name) VALUES
('550e8400-e29b-41d4-a716-446655440001', 'Электроника'),
('550e8400-e29b-41d4-a716-446655440002', 'Кофе'),
('550e8400-e29b-41d4-a716-446655440003', 'Книги'),
('550e8400-e29b-41d4-a716-446655440004', 'Еда')
ON CONFLICT (id) DO NOTHING;

-- 2. Добавьте связи магазинов с тегами в таблицу store_tag
INSERT INTO store_tag (id, store_id, tag_id) VALUES
(gen_random_uuid(), '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c', '550e8400-e29b-41d4-a716-446655440001'),
(gen_random_uuid(), 'b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2', '550e8400-e29b-41d4-a716-446655440002'),
(gen_random_uuid(), 'c45a7b64-df32-4e84-b2cb-85a3b8e6b0fc', '550e8400-e29b-41d4-a716-446655440003'),
(gen_random_uuid(), 'd0c12a9f-2b2a-4e91-8e0a-13df58d9f8af', '550e8400-e29b-41d4-a716-446655440004')
ON CONFLICT (store_id, tag_id) DO NOTHING;

-- 3. Добавьте тестовые отзывы
INSERT INTO review (id, user_id, store_id, rating, comment, created_at) VALUES
(gen_random_uuid(), (SELECT id FROM account LIMIT 1), '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c', 5, 'Отличный магазин!', NOW()),
(gen_random_uuid(), (SELECT id FROM account LIMIT 1), '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c', 4, 'Хороший ассортимент', NOW())
ON CONFLICT DO NOTHING;