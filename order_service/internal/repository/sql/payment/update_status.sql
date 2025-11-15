UPDATE payment
SET
    status = $2,
    updated_at = NOW()
WHERE
    yookassa_id = $1
    AND status IN ('pending', 'waiting_for_capture')
RETURNING id;