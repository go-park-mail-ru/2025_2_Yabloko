SELECT id, email, name, phone, city_id, address, avatar_url, created_at, updated_at
FROM account
WHERE id = $1;