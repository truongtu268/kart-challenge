-- name: CreateOrder :one
INSERT INTO orders (
    coupon_code,
    total_amount,
    discount_amount,
    final_amount,
    status
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetOrder :one
SELECT * FROM orders
WHERE id = $1 LIMIT 1;

-- name: ListOrders :many
SELECT * FROM orders
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2
WHERE id = $1
RETURNING *;

-- name: DeleteOrder :exec
DELETE FROM orders
WHERE id = $1;

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id,
    product_id,
    quantity,
    unit_price,
    total_price
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetOrderItems :many
SELECT * FROM order_items
WHERE order_id = $1
ORDER BY created_at;

-- name: GetOrderItem :one
SELECT * FROM order_items
WHERE id = $1 LIMIT 1;

-- name: UpdateOrderItem :one
UPDATE order_items
SET quantity = $2, unit_price = $3, total_price = $4
WHERE id = $1
RETURNING *;

-- name: DeleteOrderItem :exec
DELETE FROM order_items
WHERE id = $1;

-- name: DeleteOrderItemsByOrderID :exec
DELETE FROM order_items
WHERE order_id = $1;