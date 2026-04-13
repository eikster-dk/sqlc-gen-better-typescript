-- name: GetCustomer :one
SELECT id, email, name, phone, created_at, updated_at
FROM customers
WHERE id = $1;
