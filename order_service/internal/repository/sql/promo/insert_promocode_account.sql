INSERT INTO promocode_account (id, user_id, promocode_id)
VALUES ($1, $2::uuid, $3::uuid)