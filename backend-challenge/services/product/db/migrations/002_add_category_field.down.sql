-- Remove category field from products table
DROP INDEX IF EXISTS idx_products_category;
ALTER TABLE products DROP COLUMN IF EXISTS category;