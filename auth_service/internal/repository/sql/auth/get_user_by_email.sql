SELECT id, email, password_hash, created_at, updated_at
FROM account
WHERE email = $1;