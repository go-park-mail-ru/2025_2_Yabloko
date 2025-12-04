INSERT INTO
    category (id, name)
VALUES
    (
        'c1c2c3d4-e5f6-4000-8000-000000000001',
        'Суши'
    ),
    (
        'c1c2c3d4-e5f6-4000-8000-000000000002',
        'Пицца'
    ),
    (
        'c1c2c3d4-e5f6-4000-8000-000000000003',
        'Сендвичи'
    ),
    (
        'c1c2c3d4-e5f6-4000-8000-000000000004',
        'Шашлыки'
    ) ON CONFLICT (id) DO NOTHING;

INSERT INTO
    store_category (id, store_id, category_id)
VALUES
    -- Наелся лосося → Суши
    (
        gen_random_uuid(),
        'b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2',
        'c1c2c3d4-e5f6-4000-8000-000000000001'
    ),
    -- Pizza Heart → Пицца
    (
        gen_random_uuid(),
        '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c',
        'c1c2c3d4-e5f6-4000-8000-000000000002'
    ),
    -- SubJoy → Сендвичи
    (
        gen_random_uuid(),
        'c45a7b64-df32-4e84-b2cb-85a3b8e6b0fc',
        'c1c2c3d4-e5f6-4000-8000-000000000003'
    ),
    -- Все шашлыки → Шашлыки
    (
        gen_random_uuid(),
        'd0c12a9f-2b2a-4e91-8e0a-13df58d9f8af',
        'c1c2c3d4-e5f6-4000-8000-000000000004'
    ) ON CONFLICT (store_id, category_id) DO NOTHING;