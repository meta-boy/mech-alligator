-- Migration: 001_create_vendors_table
-- Description: Create vendors table to store vendor information

CREATE TABLE IF NOT EXISTS vendors (
    id VARCHAR(15) PRIMARY KEY,
    name VARCHAR(255),
    home_url TEXT,
    country VARCHAR(10) DEFAULT 'IN' CHECK (country IN ('IN', 'OTHER')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_vendors_country ON vendors (country);
CREATE INDEX IF NOT EXISTS idx_vendors_name ON vendors (name);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_vendors_updated_at 
    BEFORE UPDATE ON vendors 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();