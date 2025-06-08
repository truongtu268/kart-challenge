-- name: CreateCoupon :one
INSERT INTO coupons (
    coupon_code,
    is_valid
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetCoupon :one
SELECT * FROM coupons
WHERE id = $1 LIMIT 1;

-- name: GetCouponByCode :one
SELECT * FROM coupons
WHERE coupon_code = $1 LIMIT 1;

-- name: ListCoupons :many
SELECT * FROM coupons
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListValidCoupons :many
SELECT * FROM coupons
WHERE is_valid = true
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateCoupon :one
UPDATE coupons
SET 
    coupon_code = $2,
    is_valid = $3
WHERE id = $1
RETURNING *;

-- name: UpdateCouponValidity :one
UPDATE coupons
SET is_valid = $2
WHERE id = $1
RETURNING *;

-- name: DeleteCoupon :exec
DELETE FROM coupons
WHERE id = $1;

-- name: CountCoupons :one
SELECT COUNT(*) FROM coupons;

-- name: CountValidCoupons :one
SELECT COUNT(*) FROM coupons
WHERE is_valid = true;

-- name: SearchCouponsByCode :many
SELECT * FROM coupons
WHERE coupon_code ILIKE '%' || $1 || '%'
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;