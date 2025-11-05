SELECT id, email, hash, created_at, updated_at
FROM account
WHERE email = $1;