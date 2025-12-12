INSERT INTO account (id, email, hash, addresses_history)
VALUES ($1, $2, $3, '[]'::jsonb);
