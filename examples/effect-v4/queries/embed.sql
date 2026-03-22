-- Examples demonstrating sqlc.embed macro for nested structures
-- These queries showcase how to group related table columns into nested objects

-- name: GetOrderWithCustomerEmbed :one
-- Example: Get a single order with customer details using sqlc.embed
-- This groups all orders columns under 'order' and all customers columns under 'customer'
SELECT sqlc.embed(orders), sqlc.embed(customers)
FROM orders
JOIN customers ON orders.customer_id = customers.id
WHERE orders.id = $1;

-- name: ListOrdersWithCustomerEmbed :many
-- Example: List orders with customer details using sqlc.embed
-- Returns an array of orders, each with nested order and customer objects
SELECT sqlc.embed(orders), sqlc.embed(customers)
FROM orders
JOIN customers ON orders.customer_id = customers.id
ORDER BY orders.created_at DESC;
