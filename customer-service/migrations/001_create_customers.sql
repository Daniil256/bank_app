-- +goose Up
CREATE TABLE IF NOT EXISTS customers
(
    id          UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL UNIQUE,
    first_name  TEXT,
    last_name   TEXT,
    birth_date  DATE,
    address     TEXT,
    phone       TEXT,
    citizenship TEXT,
    status      TEXT        NOT NULL DEFAULT 'NEW',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_customers_status ON customers (status);
CREATE INDEX IF NOT EXISTS idx_customers_user_id ON customers (user_id);

-- +goose Down
DROP TABLE IF EXISTS customers;