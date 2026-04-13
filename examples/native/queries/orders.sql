-- name: GetOrder :one
SELECT id, customer_id, status, total_cents, shipping_address, billing_address, notes, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: GetOrderWithCustomer :one
SELECT 
    o.id, o.status, o.total_cents, o.shipping_address, o.billing_address, o.notes, o.created_at, o.updated_at,
    c.id AS customer_id, c.email AS customer_email, c.name AS customer_name, c.phone AS customer_phone
FROM orders o
JOIN customers c ON o.customer_id = c.id
WHERE o.id = $1;

-- name: ListOrders :many
SELECT id, customer_id, status, total_cents, shipping_address, billing_address, notes, created_at, updated_at
FROM orders
ORDER BY created_at DESC;

-- name: ListOrdersByCustomer :many
SELECT id, customer_id, status, total_cents, shipping_address, billing_address, notes, created_at, updated_at
FROM orders
WHERE customer_id = $1
ORDER BY created_at DESC;

-- name: ListOrdersByStatus :many
SELECT id, customer_id, status, total_cents, shipping_address, billing_address, notes, created_at, updated_at
FROM orders
WHERE status = $1
ORDER BY created_at DESC;

-- name: ListOrdersPaginated :many
SELECT id, customer_id, status, total_cents, shipping_address, billing_address, notes, created_at, updated_at
FROM orders
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListOrdersByDateRange :many
-- Example: duplicate column params get auto-suffixed (createdAt, createdAt_2)
SELECT id, customer_id, status, total_cents, shipping_address, billing_address, notes, created_at, updated_at
FROM orders
WHERE created_at >= $1 AND created_at < $2
ORDER BY created_at DESC;

-- name: ListOrdersByDateRangeNamed :many
-- Example: using sqlc.arg() for explicit param names (preferred approach)
SELECT id, customer_id, status, total_cents, shipping_address, billing_address, notes, created_at, updated_at
FROM orders
WHERE created_at >= sqlc.arg('start_date') AND created_at < sqlc.arg('end_date')
ORDER BY created_at DESC;

-- name: ListRecentOrdersWithCustomer :many
SELECT 
    o.id, o.status, o.total_cents, o.created_at,
    c.name AS customer_name, c.email AS customer_email
FROM orders o
JOIN customers c ON o.customer_id = c.id
ORDER BY o.created_at DESC
LIMIT $1;

-- name: SearchOrders :many
SELECT id, customer_id, status, total_cents, shipping_address, notes, created_at
FROM orders
WHERE search_vector @@ websearch_to_tsquery('english', $1)
ORDER BY created_at DESC;

-- name: CreateOrder :one
INSERT INTO orders (customer_id, status, shipping_address, billing_address, notes)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, customer_id, status, total_cents, shipping_address, billing_address, notes, created_at, updated_at;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateOrderAddresses :exec
UPDATE orders
SET shipping_address = $2, billing_address = $3, updated_at = NOW()
WHERE id = $1;

-- name: UpdateOrderTotal :exec
UPDATE orders
SET total_cents = (
    SELECT COALESCE(SUM((unit_price_cents * quantity) - discount_cents), 0)
    FROM order_lines
    WHERE order_id = $1
), updated_at = NOW()
WHERE id = $1;

-- name: DeleteOrder :exec
DELETE FROM orders
WHERE id = $1;

-- name: CountOrdersByStatus :many
SELECT status, COUNT(*) AS order_count
FROM orders
GROUP BY status
ORDER BY order_count DESC;

-- name: GetCustomerOrderStats :one
SELECT 
    COUNT(*) AS total_orders,
    COALESCE(SUM(total_cents), 0) AS total_spent,
    COALESCE(AVG(total_cents), 0) AS avg_order_value
FROM orders
WHERE customer_id = $1;

-- name: GetOrdersWithLineCount :many
SELECT 
    o.id, o.customer_id, o.status, o.total_cents, o.created_at,
    COUNT(ol.id) AS line_count
FROM orders o
LEFT JOIN order_lines ol ON o.id = ol.order_id
GROUP BY o.id
ORDER BY o.created_at DESC
LIMIT $1 OFFSET $2;

-- ============================================
-- Order Lines Queries
-- ============================================

-- name: GetOrderLine :one
SELECT id, order_id, product_id, quantity, unit_price_cents, discount_cents, created_at
FROM order_lines
WHERE id = $1;

-- name: ListOrderLines :many
SELECT id, order_id, product_id, quantity, unit_price_cents, discount_cents, created_at
FROM order_lines
WHERE order_id = $1
ORDER BY created_at;

-- name: ListOrderLinesWithProduct :many
SELECT 
    ol.id, ol.order_id, ol.quantity, ol.unit_price_cents, ol.discount_cents, ol.created_at,
    p.id AS product_id, p.sku AS product_sku, p.name AS product_name
FROM order_lines ol
JOIN products p ON ol.product_id = p.id
WHERE ol.order_id = $1
ORDER BY ol.created_at;

-- name: GetFullOrderDetails :many
SELECT 
    o.id AS order_id, o.status, o.total_cents AS order_total, o.shipping_address, o.created_at AS order_date,
    c.id AS customer_id, c.name AS customer_name, c.email AS customer_email,
    ol.id AS line_id, ol.quantity, ol.unit_price_cents, ol.discount_cents,
    p.id AS product_id, p.sku, p.name AS product_name
FROM orders o
JOIN customers c ON o.customer_id = c.id
JOIN order_lines ol ON o.id = ol.order_id
JOIN products p ON ol.product_id = p.id
WHERE o.id = $1
ORDER BY ol.created_at;

-- name: CreateOrderLine :one
INSERT INTO order_lines (order_id, product_id, quantity, unit_price_cents, discount_cents)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, order_id, product_id, quantity, unit_price_cents, discount_cents, created_at;

-- name: UpdateOrderLineQuantity :exec
UPDATE order_lines
SET quantity = $2
WHERE id = $1;

-- name: DeleteOrderLine :exec
DELETE FROM order_lines
WHERE id = $1;

-- name: DeleteOrderLines :execrows
DELETE FROM order_lines
WHERE order_id = $1;

-- name: GetOrderLineTotal :one
SELECT 
    COALESCE(SUM((unit_price_cents * quantity) - discount_cents), 0) AS total
FROM order_lines
WHERE order_id = $1;

-- name: GetProductSalesStats :one
SELECT 
    p.id, p.sku, p.name,
    COALESCE(SUM(ol.quantity), 0) AS total_sold,
    COALESCE(SUM(ol.quantity * ol.unit_price_cents), 0) AS total_revenue
FROM products p
LEFT JOIN order_lines ol ON p.id = ol.product_id
WHERE p.id = $1
GROUP BY p.id;

-- name: GetTopSellingProducts :many
SELECT 
    p.id, p.sku, p.name, p.category,
    SUM(ol.quantity) AS total_sold,
    SUM(ol.quantity * ol.unit_price_cents) AS total_revenue
FROM products p
JOIN order_lines ol ON p.id = ol.product_id
GROUP BY p.id
ORDER BY total_sold DESC
LIMIT $1;

-- name: GetOrdersByProductIds :many
SELECT DISTINCT o.id, o.customer_id, o.status, o.total_cents, o.created_at
FROM orders o
JOIN order_lines ol ON o.id = ol.order_id
WHERE ol.product_id = ANY(sqlc.arg('product_ids')::int[])
ORDER BY o.created_at DESC;
