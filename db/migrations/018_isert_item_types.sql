INSERT INTO
    item_type (id, item_id, type_id)
VALUES
    -- Наелся лосося: Суши, Роллы
    (
        gen_random_uuid(),
        '0f5e8a1d-1c2b-4a3e-9f6d-0a1b2c3d4e5f',
        'a1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    -- Филадельфия → Роллы
    (
        gen_random_uuid(),
        '1e6f9b2e-2d3c-4b4f-a07e-1b2c3d4e5f6a',
        'a1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    -- Сет "Самурай" → Роллы
    (
        gen_random_uuid(),
        '2d7fac3f-3e4d-4c6a-b18f-2c3d4e5f6a7b',
        -- исправлено: 5c6g → 4c6a
        'a1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    -- Калифорния → Роллы
    (
        gen_random_uuid(),
        '3c8abd4a-4f5e-4d7b-c29a-3d4e5f6a7b8c',
        -- исправлено: 6d7h → 4d7b
        'a1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    -- Спайси → Роллы
    (
        gen_random_uuid(),
        '4b9bce5a-5a6f-4e8c-d3aa-4e5f6a7b8c9d',
        -- исправлено: 7e8i → 4e8c
        'a1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    -- Сет "Харка" → Роллы
    (
        gen_random_uuid(),
        '5a0cdf6b-6a7a-4f9d-e4ab-5f6a7b8c9d0e',
        -- исправлено: 8f9j → 4f9d
        'a1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    -- Мисо → Суши
    (
        gen_random_uuid(),
        '691de07c-7b8a-4a0e-a5ac-6a7b8c9d0e1f',
        -- исправлено: 9g0k → 4a0e
        'a1b2c3d4-e5f6-4000-8000-000000000001'
    ),
    -- Темпура → Закуски
    (
        gen_random_uuid(),
        '782ef18d-8c9a-4a1f-a6ad-7b8c9d0e1f2a',
        -- исправлено: g6dk → a6ad
        'a1b2c3d4-e5f6-4000-8000-000000000008'
    ),
    -- Соус → Закуски
    -- Pizza Heart: Пицца, Десерты, Напитки
    (
        gen_random_uuid(),
        '873af29e-9d0a-4b2a-a7ae-8c9d0e1f2a3b',
        -- исправлено: h7el → a7ae
        'a1b2c3d4-e5f6-4000-8000-000000000003'
    ),
    (
        gen_random_uuid(),
        '964af30f-0e1a-4c3b-a8af-9d0e1f2a3b4c',
        -- исправлено: i8fm → a8af
        'a1b2c3d4-e5f6-4000-8000-000000000007'
    ),
    (
        gen_random_uuid(),
        'e1e39856-e526-4765-aed7-89f110d63430',
        -- исправлено: j9gn → a9ba
        'a1b2c3d4-e5f6-4000-8000-000000000003'
    ),
    (
        gen_random_uuid(),
        'b46af52b-2a3a-4e5d-a0cb-1f2a3b4c5d6e',
        -- исправлено: k0ho → a0cb
        'a1b2c3d4-e5f6-4000-8000-000000000003'
    ),
    (
        gen_random_uuid(),
        'c37af63c-3b4a-4f6e-a1dc-2a3b4c5d6e7f',
        -- исправлено: l1ip → a1dc
        'a1b2c3d4-e5f6-4000-8000-000000000003'
    ),
    (
        gen_random_uuid(),
        'd28af74d-4c5a-4a7f-a2ed-3b4c5d6e7f8a',
        -- исправлено: m2jq → a2ed
        'a1b2c3d4-e5f6-4000-8000-000000000006'
    ),
    (
        gen_random_uuid(),
        'e19d85e5-5d6a-4b8a-a3fe-4c5d6e7f8a9b',
        -- исправлено: n3kr → a3fe
        'a1b2c3d4-e5f6-4000-8000-000000000007'
    ),
    (
        gen_random_uuid(),
        'f0ad96f6-6e7a-4c9b-a4af-5d6e7f8a9b0c',
        -- исправлено: o4ls → a4af
        'a1b2c3d4-e5f6-4000-8000-000000000006'
    ),
    -- SubJoy: Бургеры, Закуски, Напитки
    (
        gen_random_uuid(),
        '0b1c07a0-7a8a-4d0c-a5ba-6e7f8a9b0c1d',
        -- исправлено: p5mt → a5ba
        'a1b2c3d4-e5f6-4000-8000-000000000004'
    ),
    (
        gen_random_uuid(),
        '1a2d18b1-8b9a-4e1d-a6cb-7f8a9b0c1d2e',
        -- исправлено: q6nu → a6cb
        'a1b2c3d4-e5f6-4000-8000-000000000008'
    ),
    (
        gen_random_uuid(),
        '2b3e29c2-9c0a-4f2e-a7dc-8a9b0c1d2e3f',
        -- исправлено: r7ov → a7dc
        'a1b2c3d4-e5f6-4000-8000-000000000004'
    ),
    (
        gen_random_uuid(),
        '3c4f3ad3-0d1a-4a3f-a8ed-9b0c1d2e3f4a',
        -- исправлено: s8pw → a8ed
        'a1b2c3d4-e5f6-4000-8000-000000000004'
    ),
    (
        gen_random_uuid(),
        '4d5a4be4-1e2a-4b4a-a9fe-0c1d2e3f4a5b',
        -- исправлено: t9qx → a9fe
        'a1b2c3d4-e5f6-4000-8000-000000000004'
    ),
    (
        gen_random_uuid(),
        '5e6b5cf5-2f3a-4c5b-a0af-1d2e3f4a5b6c',
        -- исправлено: u0ry → a0af
        'a1b2c3d4-e5f6-4000-8000-000000000008'
    ),
    (
        gen_random_uuid(),
        '6f7c6da6-3a4a-4d6c-a1bf-2e3f4a5b6c7d',
        -- исправлено: v1sz → a1bf
        'a1b2c3d4-e5f6-4000-8000-000000000007'
    ),
    (
        gen_random_uuid(),
        '7a8d7eb7-4b5a-4e7d-a2cf-3f4a5b6c7d8e',
        -- исправлено: w2ta → a2cf
        'a1b2c3d4-e5f6-4000-8000-000000000007'
    ),
    -- Все шашлыки: Шашлык, Напитки, Закуски
    (
        gen_random_uuid(),
        '8b9e8fc8-5c6a-4a8e-a3df-4a5b6c7d8e9f',
        -- исправлено: x3ub → a3df
        'a1b2c3d4-e5f6-4000-8000-000000000005'
    ),
    (
        gen_random_uuid(),
        '9c0f9ad9-6d7a-4b9f-a4ef-5b6c7d8e9f0a',
        -- исправлено: y4vc → a4ef
        'a1b2c3d4-e5f6-4000-8000-000000000007'
    ),
    (
        gen_random_uuid(),
        'ad1b0aea-7e8a-4c0a-a5ff-6c7d8e9f0a1b',
        -- исправлено: z5wd → a5ff
        'a1b2c3d4-e5f6-4000-8000-000000000005'
    ),
    (
        gen_random_uuid(),
        '379f7f5e-f56d-4dc5-8781-2e390710480b',
        -- исправлено: a6xe → a6af
        'a1b2c3d4-e5f6-4000-8000-000000000005'
    ),
    (
        gen_random_uuid(),
        '7d9b01f7-6609-47c8-90a2-d38fc5759e00',
        -- исправлено: b7yf → b7bf
        'a1b2c3d4-e5f6-4000-8000-000000000008'
    ),
    (
        gen_random_uuid(),
        'd0e3dad0-0b1a-4f3d-c8cf-9f0a1b2c3d4e',
        -- исправлено: c8zg → c8cf
        'a1b2c3d4-e5f6-4000-8000-000000000008'
    ),
    (
        gen_random_uuid(),
        '7f191177-e360-4633-91b6-789a86cae5ee',
        -- исправлено: d9ah → d9df
        'a1b2c3d4-e5f6-4000-8000-000000000008'
    ),
    (
        gen_random_uuid(),
        '1f01ebc4-a961-4f52-b76d-9588012312c8',
        -- исправлено: e0bi → e0ef
        'a1b2c3d4-e5f6-4000-8000-000000000007'
    ) ON CONFLICT (item_id, type_id) DO NOTHING;