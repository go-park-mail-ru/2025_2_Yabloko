INSERT INTO
    item_type (id, item_id, type_id)
VALUES
    -- Наелся лосося → Роллы
    (
        gen_random_uuid(),
        '9e1a4a7f-73a9-4b2d-9332-8aef8d3c11e1',
        'a1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    -- Филадельфия
    (
        gen_random_uuid(),
        '1b0c9c25-3185-4c82-9154-79bb1ef5b53f',
        'a1b2c3d4-e5f6-4000-8000-000000000002'
    ),
    -- Сет "Самурай"
    -- Pizza Heart → Пицца + Напитки
    (
        gen_random_uuid(),
        '3b8dca61-2d4e-4f85-a75b-2ef29f7c13bb',
        'a1b2c3d4-e5f6-4000-8000-000000000003'
    ),
    -- Пепперони
    (
        gen_random_uuid(),
        'f3c7b1ab-41ac-423a-b1cb-4d71b65bdf94',
        'a1b2c3d4-e5f6-4000-8000-000000000007'
    ),
    -- Лимонад
    -- SubJoy → Бургеры + Закуски
    (
        gen_random_uuid(),
        'a4e8bcd0-44b1-40c4-b89a-8df76f36e2fa',
        'a1b2c3d4-e5f6-4000-8000-000000000004'
    ),
    -- Сабвей
    (
        gen_random_uuid(),
        'c5d7e2a9-6b1f-4de3-9b76-0c5e61d4f9b3',
        'a1b2c3d4-e5f6-4000-8000-000000000008'
    ),
    -- Картошка
    -- Все шашлыки → Шашлык + Напитки
    (
        gen_random_uuid(),
        'e8a0d27b-06e3-4bcb-90b9-2eac3b55e6de',
        'a1b2c3d4-e5f6-4000-8000-000000000005'
    ),
    -- Шашлык
    (
        gen_random_uuid(),
        '2f97bcdc-5e45-41cd-92a8-988bcf7eb3b2',
        'a1b2c3d4-e5f6-4000-8000-000000000007'
    ) -- Пиво
    ON CONFLICT (item_id, type_id) DO NOTHING;