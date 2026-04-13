-- name: GetCustomer :one
SELECT id, email, name, phone, created_at, updated_at
FROM customers
WHERE id = $1;

-- name: GetCustomerByEmail :one
SELECT id, email, name, phone, created_at, updated_at
FROM customers
WHERE email = $1;

-- name: ListCustomers :many
SELECT id, email, name, phone, created_at, updated_at
FROM customers
ORDER BY created_at DESC;

-- name: ListCustomersPaginated :many
SELECT id, email, name, phone, created_at, updated_at
FROM customers
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: SearchCustomersByName :many
SELECT id, email, name, phone, created_at, updated_at
FROM customers
WHERE name ILIKE '%' || $1 || '%'
ORDER BY name;

-- name: CreateCustomer :one
INSERT INTO customers (email, name, phone)
VALUES ($1, $2, $3)
RETURNING id, email, name, phone, created_at, updated_at;

-- name: UpdateCustomer :one
UPDATE customers
SET email = $2, name = $3, phone = $4, updated_at = NOW()
WHERE id = $1
RETURNING id, email, name, phone, created_at, updated_at;

-- name: UpdateCustomerEmail :exec
UPDATE customers
SET email = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteCustomer :exec
DELETE FROM customers
WHERE id = $1;

-- name: CountCustomers :one
SELECT COUNT(*) AS total
FROM customers;

-- name: GetCustomersByIds :many
SELECT id, email, name, phone, created_at, updated_at
FROM customers
WHERE id = ANY(sqlc.arg('ids')::int[]);
