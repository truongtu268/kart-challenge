-- Add category field to products table
ALTER TABLE products ADD COLUMN category VARCHAR(100);

-- Create index for category field
CREATE INDEX idx_products_category ON products(category);