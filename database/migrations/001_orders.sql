CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
CREATE TABLE IF NOT EXISTS orders (
    id text PRIMARY KEY,
    customer_id bigint NOT NULL,
    amount bigint NOT NULL CHECK (amount > 0),
    recent_orders bigint NOT NULL DEFAULT 0,
    status text NOT NULL CHECK (status IN ('pending','paid','failed','payment_unknown')),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS orders_customer_created_idx ON orders(customer_id, created_at);
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'scalens_monitor') THEN
        CREATE ROLE scalens_monitor LOGIN PASSWORD 'scalens_monitor';
        GRANT pg_monitor TO scalens_monitor;
    END IF;
END
$$;
