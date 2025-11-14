SELECT
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
FROM
    payment
WHERE
    id = $1