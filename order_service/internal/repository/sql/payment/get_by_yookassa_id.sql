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
    yookassa_id = $1