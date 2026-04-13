-- name: GetProduct :one
SELECT id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at
FROM products
WHERE id = $1;

-- name: GetProductBySku :one
SELECT id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at
FROM products
WHERE sku = $1;

-- name: ListProducts :many
SELECT id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at
FROM products
ORDER BY name;

-- name: ListActiveProducts :many
SELECT id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at
FROM products
WHERE is_active = TRUE
ORDER BY name;

-- name: ListProductsByCategory :many
SELECT id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at
FROM products
WHERE category = $1 AND is_active = TRUE
ORDER BY name;

-- name: ListProductsPaginated :many
SELECT id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at
FROM products
WHERE is_active = TRUE
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListProductsCursor :many
SELECT id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at
FROM products
WHERE is_active = TRUE AND id > $1
ORDER BY id
LIMIT $2;

-- name: SearchProducts :many
SELECT id, sku, name, description, price_cents, stock_quantity, category
FROM products
WHERE search_vector @@ plainto_tsquery('english', $1)
  AND is_active = TRUE
ORDER BY name;

-- name: SearchProductsRanked :many
SELECT id, sku, name, description, price_cents, stock_quantity, category,
       ts_rank(search_vector, query) AS rank
FROM products, plainto_tsquery('english', $1) query
WHERE search_vector @@ query
  AND is_active = TRUE
ORDER BY rank DESC
LIMIT $2 OFFSET $3;

-- name: SearchProductsWithHighlight :many
SELECT id, sku, name,
       ts_headline('english', COALESCE(description, ''), plainto_tsquery('english', $1)) AS highlighted_description,
       price_cents, category
FROM products
WHERE search_vector @@ plainto_tsquery('english', $1)
  AND is_active = TRUE;

-- name: SearchProductsWebStyle :many
SELECT id, sku, name, description, price_cents, category
FROM products
WHERE search_vector @@ websearch_to_tsquery('english', $1)
  AND is_active = TRUE
ORDER BY name;

-- name: CreateProduct :one
INSERT INTO products (sku, name, description, price_cents, stock_quantity, is_active, category)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at;

-- name: UpdateProduct :one
UPDATE products
SET sku = $2, name = $3, description = $4, price_cents = $5, stock_quantity = $6, is_active = $7, category = $8, updated_at = NOW()
WHERE id = $1
RETURNING id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at;

-- name: UpdateProductStock :execrows
UPDATE products
SET stock_quantity = stock_quantity + $2, updated_at = NOW()
WHERE id = $1;

-- name: DeactivateProduct :exec
UPDATE products
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;

-- name: GetProductsByIds :many
SELECT id, sku, name, description, price_cents, stock_quantity, is_active, category, created_at, updated_at
FROM products
WHERE id = ANY(sqlc.arg('ids')::int[]);

-- name: CountProductsByCategory :many
SELECT category, COUNT(*) AS product_count
FROM products
WHERE is_active = TRUE
GROUP BY category
ORDER BY product_count DESC;

-- name: GetLowStockProducts :many
SELECT id, sku, name, stock_quantity, category
FROM products
WHERE stock_quantity < $1 AND is_active = TRUE
ORDER BY stock_quantity;
