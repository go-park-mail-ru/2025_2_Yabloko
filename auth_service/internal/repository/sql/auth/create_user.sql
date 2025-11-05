INSERT INTO account (id, email, hash)
VALUES ($1, $2, $3)
RETURNING id, email, hash, created_at, updated_at;