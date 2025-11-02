UPDATE account
SET
    name       = $1,
    phone      = $2,
    city_id    = $3,
    address    = $4,
    avatar_url = $5
WHERE id = $6;