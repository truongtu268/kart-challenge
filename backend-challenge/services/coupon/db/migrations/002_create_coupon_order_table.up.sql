CREATE TABLE IF NOT EXISTS coupon_order (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    coupon_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Foreign key constraint
    CONSTRAINT fk_coupon_order_coupon_id 
        FOREIGN KEY (coupon_id) 
        REFERENCES coupons(id) 
        ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_coupon_order_order_id ON coupon_order(order_id);
CREATE INDEX IF NOT EXISTS idx_coupon_order_coupon_id ON coupon_order(coupon_id);

-- Create unique constraint to prevent duplicate coupon usage for same order
CREATE UNIQUE INDEX IF NOT EXISTS idx_coupon_order_unique 
    ON coupon_order(order_id, coupon_id);

-- Create trigger to automatically update updated_at
CREATE TRIGGER update_coupon_order_updated_at
    BEFORE UPDATE ON coupon_order
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();