-- Drop trigger
DROP TRIGGER IF EXISTS update_coupon_order_updated_at ON coupon_order;

-- Drop indexes
DROP INDEX IF EXISTS idx_coupon_order_unique;
DROP INDEX IF EXISTS idx_coupon_order_coupon_id;
DROP INDEX IF EXISTS idx_coupon_order_order_id;

-- Drop table
DROP TABLE IF EXISTS coupon_order;