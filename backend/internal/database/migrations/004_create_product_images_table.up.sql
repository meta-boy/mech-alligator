-- Migration: 004_create_product_images_table
-- Description: Create product images table to store product image URLs

CREATE TABLE IF NOT EXISTS product_images (
    id VARCHAR(15) PRIMARY KEY,
    product_id VARCHAR(15),
    url TEXT,
    uuid VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_product_images_product_id ON product_images (product_id);
CREATE INDEX IF NOT EXISTS idx_product_images_uuid ON product_images (uuid);

-- Create unique constraint for product, uuid, and url combination
CREATE UNIQUE INDEX IF NOT EXISTS idx_product_images_unique ON product_images (product_id, uuid, url);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_product_images_updated_at 
    BEFORE UPDATE ON product_images 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();