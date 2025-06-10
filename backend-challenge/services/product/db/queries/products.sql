-- name: GetProduct :one
SELECT id, name, price, category FROM products
WHERE id = $1 LIMIT 1;

-- name: ListProducts :many
SELECT id, name, price, category FROM products
ORDER BY name ASC;