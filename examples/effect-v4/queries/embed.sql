-- Examples demonstrating sqlc.embed macro for nested structures
-- These queries showcase how to group related table columns into nested objects

-- name: GetOrderWithCustomerEmbed :one
-- Example: Get a single order with customer details using sqlc.embed
-- The order_details view casts search_vector to text before it reaches the driver.
-- Columns are grouped under 'orderDetail' and 'customer'.
SELECT sqlc.embed(order_details), sqlc.embed(customers)
FROM order_details
JOIN customers ON order_details.customer_id = customers.id
WHERE order_details.id = $1;

-- name: ListOrdersWithCustomerEmbed :many
-- Example: List orders with customer details using sqlc.embed
-- Returns an array of orders, each with nested order and customer objects
SELECT sqlc.embed(order_details), sqlc.embed(customers)
FROM order_details
JOIN customers ON order_details.customer_id = customers.id
ORDER BY order_details.created_at DESC;
