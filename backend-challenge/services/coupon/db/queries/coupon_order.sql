-- name: CreateCouponOrder :one
INSERT INTO coupon_order (
    order_id,
    coupon_id
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetCouponOrder :one
SELECT * FROM coupon_order
WHERE id = $1 LIMIT 1;

-- name: GetCouponOrderByOrderID :many
SELECT co.*, c.coupon_code, c.is_valid
FROM coupon_order co
JOIN coupons c ON co.coupon_id = c.id
WHERE co.order_id = $1
ORDER BY co.created_at DESC;

-- name: GetCouponOrderByCouponID :many
SELECT * FROM coupon_order
WHERE coupon_id = $1
ORDER BY created_at DESC;

-- name: GetCouponOrderByOrderAndCoupon :one
SELECT * FROM coupon_order
WHERE order_id = $1 AND coupon_id = $2
LIMIT 1;

-- name: ListCouponOrders :many
SELECT co.*, c.coupon_code, c.is_valid
FROM coupon_order co
JOIN coupons c ON co.coupon_id = c.id
ORDER BY co.created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateCouponOrder :one
UPDATE coupon_order
SET 
    order_id = $2,
    coupon_id = $3
WHERE id = $1
RETURNING *;

-- name: DeleteCouponOrder :exec
DELETE FROM coupon_order
WHERE id = $1;

-- name: DeleteCouponOrderByOrderID :exec
DELETE FROM coupon_order
WHERE order_id = $1;

-- name: DeleteCouponOrderByCouponID :exec
DELETE FROM coupon_order
WHERE coupon_id = $1;

-- name: CountCouponOrders :one
SELECT COUNT(*) FROM coupon_order;

-- name: CountCouponOrdersByOrderID :one
SELECT COUNT(*) FROM coupon_order
WHERE order_id = $1;

-- name: CountCouponOrdersByCouponID :one
SELECT COUNT(*) FROM coupon_order
WHERE coupon_id = $1;

-- name: CheckCouponUsageInOrder :one
SELECT EXISTS(
    SELECT 1 FROM coupon_order
    WHERE order_id = $1 AND coupon_id = $2
) as exists;