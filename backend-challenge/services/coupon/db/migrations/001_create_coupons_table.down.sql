-- Drop trigger
DROP TRIGGER IF EXISTS update_coupons_updated_at ON coupons;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_coupons_is_valid;
DROP INDEX IF EXISTS idx_coupons_coupon_code;

-- Drop table
DROP TABLE IF EXISTS coupons;