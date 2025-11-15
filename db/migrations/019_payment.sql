CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TYPE payment_status AS ENUM (
    'pending',
    'waiting_for_capture',
    'succeeded',
    'canceled'
);

CREATE TABLE IF NOT EXISTS payment (
    id uuid PRIMARY KEY,
    order_id uuid NOT NULL REFERENCES "orders" (id) ON DELETE CASCADE ON UPDATE CASCADE,
    yookassa_id varchar(50) NOT NULL UNIQUE,
    status payment_status NOT NULL DEFAULT 'pending',
    amount numeric(12, 2) NOT NULL CHECK (amount > 0),
    currency varchar(3) NOT NULL DEFAULT 'RUB',
    description varchar(128),
    metadata jsonb,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp
);

CREATE TRIGGER trg_update_payment_updated_at
BEFORE UPDATE ON payment
FOR EACH ROW
EXECUTE FUNCTION update_updated_at();

CREATE INDEX idx_payment_order_id ON payment (order_id);
CREATE INDEX idx_payment_yookassa_id ON payment (yookassa_id);
CREATE INDEX idx_payment_status ON payment (status);