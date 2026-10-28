-- name: GetProduct :one
SELECT * FROM products
WHERE id = ?;

-- name: ListProducts :many
SELECT * FROM products
ORDER BY name;

-- name: CreateProduct :exec
INSERT INTO products (name, price)
VALUES (?, ?);
