CREATE TABLE IF NOT EXISTS orders (
    order_uid TEXT PRIMARY KEY,
    track_number TEXT,
    entry TEXT,
    locale TEXT,
    internal_signature TEXT,
    customer_id TEXT,
    delivery_service TEXT,
    shardkey TEXT,
    sm_id INT,
    date_created TIMESTAMP WITH TIME ZONE,
    oof_shard TEXT,
    payload JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);


CREATE INDEX IF NOT EXISTS idx_orders_date_created ON orders (date_created);


CREATE TABLE IF NOT EXISTS bad_messages (
    id SERIAL PRIMARY KEY,
    raw_message TEXT,
    error TEXT,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
