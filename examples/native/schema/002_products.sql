CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    sku TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    price_cents INTEGER NOT NULL,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    category TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX products_sku_idx ON products(sku);
CREATE INDEX products_category_idx ON products(category);
CREATE INDEX products_is_active_idx ON products(is_active);

-- Full-text search vector
ALTER TABLE products ADD COLUMN search_vector TSVECTOR
    GENERATED ALWAYS AS (
        to_tsvector('english', name || ' ' || COALESCE(description, '') || ' ' || COALESCE(category, ''))
    ) STORED;

CREATE INDEX products_search_idx ON products USING GIN(search_vector);
