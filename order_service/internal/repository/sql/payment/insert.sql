INSERT INTO payment (
    id,
    order_id,
    yookassa_id,
    status,
    amount,
    currency,
    description,
    metadata,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    NOW(), NOW()
);