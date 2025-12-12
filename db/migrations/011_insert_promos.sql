INSERT INTO
    promocode (
        id,
        code,
        relative_discount,
        absolute_discount,
        start_at,
        end_at
    )
VALUES
    (
        '11111111-1111-4000-8000-000000000001',
        'WELCOME10',
        10.00,
        0,
        now(),
        now() + interval '30 days'
    ),
    (
        '11111111-1111-4000-8000-000000000002',
        'FIXED150',
        0,
        150.00,
        now(),
        now() + interval '60 days'
    ),
    (
        '11111111-1111-4000-8000-000000000003',
        'BLACK50',
        50.00,
        0,
        now(),
        now() + interval '1 day'
    ) ON CONFLICT (id) DO NOTHING;