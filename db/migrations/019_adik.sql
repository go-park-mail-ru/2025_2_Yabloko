-- 3. Гарантируем тестового пользователя
INSERT INTO
    account (id, email, hash, name, created_at, updated_at)
VALUES
    (
        '00000000-0000-0000-0000-000000000001',
        'test@example.com',
        '$2a$10$CCCCCCCCCCCCCCCCCCCCC.OOOOOOOOOOOOOOOOOOOOOOOOO',
        'Тест Пользователь',
        NOW(),
        NOW()
    ) ON CONFLICT (id) DO NOTHING;

-- 4. Отзывы
INSERT INTO
    review (
        id,
        user_id,
        store_id,
        rating,
        comment,
        created_at
    )
SELECT
    gen_random_uuid(),
    '00000000-0000-0000-0000-000000000001',
    store_id,
    rating,
    comment,
    NOW()
FROM
    (
        VALUES
            (
                'b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2',
                5,
                'Восхитительные роллы! Свежайший лосось.'
            ),
            (
                'b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2',
                4,
                'Хорошо, но порции маловаты.'
            ),
            (
                '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c',
                5,
                'Лучшая пицца в городе!'
            ),
            (
                '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c',
                4,
                'Хороший ассортимент, но дорого.'
            ),
            (
                'c45a7b64-df32-4e84-b2cb-85a3b8e6b0fc',
                4,
                'Быстро и вкусно. Идеально на обед.'
            ),
            (
                'd0c12a9f-2b2a-4e91-8e0a-13df58d9f8af',
                5,
                'Шашлык тает во рту! Обязательно вернусь.'
            )
    ) AS v(store_id, rating, comment) ON CONFLICT (user_id, store_id) DO NOTHING;