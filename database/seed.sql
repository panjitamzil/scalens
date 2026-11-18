\set ON_ERROR_STOP on
TRUNCATE orders;
INSERT INTO orders(id,customer_id,amount,status,created_at)
SELECT md5('seed-v1-' || i), 1 + i % 10000, 1000, 'paid',
       TIMESTAMPTZ '2026-01-01' + (i % 2592000) * INTERVAL '1 second'
FROM generate_series(1, :rows) AS i;
VACUUM ANALYZE orders;
