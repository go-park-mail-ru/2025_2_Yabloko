INSERT INTO account (id, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, email, password_hash, created_at, updated_at;