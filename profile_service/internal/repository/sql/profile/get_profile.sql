SELECT
    id,
    email,
    name,
    phone,
    city_id,
    address,
    addresses_history,
    avatar_url,
    created_at,
    updated_at
FROM account
WHERE id = $1;
