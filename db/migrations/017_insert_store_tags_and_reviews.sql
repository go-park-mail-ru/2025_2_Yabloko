INSERT INTO
    tag (id, name)
VALUES
    (
        'b1b2c3d4-e5f6-4000-8000-000000000001',
        'Доставка'
    ),
    (
        'b1b2c3d4-e5f6-4000-8000-000000000002',
        'Есть в зале'
    ),
    ('b1b2c3d4-e5f6-4000-8000-000000000003', 'Острое'),
    (
        'b1b2c3d4-e5f6-4000-8000-000000000004',
        'Вегетарианское'
    ),
    (
        'b1b2c3d4-e5f6-4000-8000-000000000005',
        'Алкоголь'
    ),
    (
        'b1b2c3d4-e5f6-4000-8000-000000000006',
        'Фастфуд'
    ) ON CONFLICT (id) DO NOTHING;

INSERT INTO
    store_tag (id, store_id, tag_id)
VALUES
    -- Наелся лосося: доставка, зал, острое
    (
        gen_random_uuid(),
        'b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2',
        'b1b2c3d4-e5f6-4000-8000-000000000001'
    ),
    (
        gen_random_uuid(),
        'b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2',
        'b1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    (
        gen_random_uuid(),
        'b2f0d6b3-65a2-4c2a-a32f-30a1b73f32e2',
        'b1b2c3d4-e5f6-4000-8000-000000000003'
    ),
    -- Pizza Heart: доставка, зал, вегетарианское
    (
        gen_random_uuid(),
        '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c',
        'b1b2c3d4-e5f6-4000-8000-000000000001'
    ),
    (
        gen_random_uuid(),
        '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c',
        'b1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    (
        gen_random_uuid(),
        '9ac3b889-96df-4c93-a0b7-31f5b6a6e89c',
        'b1b2c3d4-e5f6-4000-8000-000000000004'
    ),
    -- SubJoy: доставка, фастфуд
    (
        gen_random_uuid(),
        'c45a7b64-df32-4e84-b2cb-85a3b8e6b0fc',
        'b1b2c3d4-e5f6-4000-8000-000000000001'
    ),
    (
        gen_random_uuid(),
        'c45a7b64-df32-4e84-b2cb-85a3b8e6b0fc',
        'b1b2c3d4-e5f6-4000-8000-000000000006'
    ),
    -- Все шашлыки: доставка, зал, алкоголь
    (
        gen_random_uuid(),
        'd0c12a9f-2b2a-4e91-8e0a-13df58d9f8af',
        'b1b2c3d4-e5f6-4000-8000-000000000001'
    ),
    (
        gen_random_uuid(),
        'd0c12a9f-2b2a-4e91-8e0a-13df58d9f8af',
        'b1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    (
        gen_random_uuid(),
        'd0c12a9f-2b2a-4e91-8e0a-13df58d9f8af',
        'b1b2c3d4-e5f6-4000-8000-000000000005'
    ) ON CONFLICT (store_id, tag_id) DO NOTHING;
