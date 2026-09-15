CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id),
    provider VARCHAR(50),
    provider_txn_id VARCHAR(128),
    amount NUMERIC(12,2) NOT NULL,
    status VARCHAR(20) NOT NULL,
    raw_payload JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);
