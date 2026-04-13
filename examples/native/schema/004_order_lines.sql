CREATE TABLE order_lines (
    id SERIAL PRIMARY KEY,
    order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL REFERENCES products(id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price_cents INTEGER NOT NULL,
    discount_cents INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX order_lines_order_id_idx ON order_lines(order_id);
CREATE INDEX order_lines_product_id_idx ON order_lines(product_id);

-- Unique constraint to prevent duplicate products in same order
CREATE UNIQUE INDEX order_lines_order_product_idx ON order_lines(order_id, product_id);
