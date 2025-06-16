-- Migration: 003_create_products_table
-- Description: Create products table to store product information

CREATE TABLE IF NOT EXISTS products (
    id VARCHAR(15) PRIMARY KEY,
    name TEXT,
    description TEXT,
    price DECIMAL(10,2),
    currency VARCHAR(3) CHECK (currency IN ('INR', 'USD')) DEFAULT 'INR',
    url TEXT,
    config_id VARCHAR(15),
    in_stock BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (config_id) REFERENCES site_configurations(id) ON DELETE SET NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_products_config_id ON products (config_id);
CREATE INDEX IF NOT EXISTS idx_products_currency ON products (currency);
CREATE INDEX IF NOT EXISTS idx_products_in_stock ON products (in_stock);
CREATE INDEX IF NOT EXISTS idx_products_price ON products (price);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products (created_at);

-- Create unique constraint for name and url combination
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_name_url ON products (name, url);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();