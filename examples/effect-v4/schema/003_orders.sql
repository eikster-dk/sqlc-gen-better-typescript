CREATE TYPE order_status AS ENUM (
    'pending',
    'confirmed',
    'processing',
    'shipped',
    'delivered',
    'cancelled',
    'refunded'
);

CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    customer_id INTEGER NOT NULL REFERENCES customers(id),
    status order_status NOT NULL DEFAULT 'pending',
    total_cents INTEGER NOT NULL DEFAULT 0,
    shipping_address TEXT,
    billing_address TEXT,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX orders_customer_id_idx ON orders(customer_id);
CREATE INDEX orders_status_idx ON orders(status);
CREATE INDEX orders_created_at_idx ON orders(created_at);

-- Full-text search for orders (searching notes, addresses)
ALTER TABLE orders ADD COLUMN search_vector TSVECTOR
    GENERATED ALWAYS AS (
        to_tsvector('english', 
            COALESCE(shipping_address, '') || ' ' || 
            COALESCE(billing_address, '') || ' ' || 
            COALESCE(notes, '')
        )
    ) STORED;

CREATE INDEX orders_search_idx ON orders USING GIN(search_vector);

-- sqlc.embed includes every column, so expose a text projection for reads.
-- The base table retains its native tsvector and GIN index for searching.
CREATE VIEW order_details AS
SELECT id, customer_id, status, total_cents, shipping_address, billing_address,
       notes, created_at, updated_at, search_vector::text AS search_vector
FROM orders;
