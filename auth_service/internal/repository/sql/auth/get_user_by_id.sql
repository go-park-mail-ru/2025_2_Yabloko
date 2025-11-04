SELECT id, email, hash, created_at, updated_at
FROM account
WHERE id = $1;