-- +goose Up
INSERT INTO promocode (
    id,
    code,
    relative_discount,
    absolute_discount,
    start_at,
    end_at
) VALUES
    (
        gen_random_uuid(),
        'WELCOME10',
        10.00,      -- 10% скидка
        0,
        now(),
        now() + interval '30 days'
    ),
    (
        gen_random_uuid(),
        'FIXED150',
        0,
        150.00,     -- фиксированная скидка 150
        now(),
        now() + interval '60 days'
    ),
    (
        gen_random_uuid(),
        'BLACK50',
        50.00,      -- 50% скидка
        0,
        now(),
        now() + interval '1 day'
    );

-- +goose Down
DELETE FROM promocode
WHERE code IN ('WELCOME10', 'FIXED150', 'BLACK50');
