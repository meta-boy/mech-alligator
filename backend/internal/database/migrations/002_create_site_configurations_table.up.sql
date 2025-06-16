-- Migration: 002_create_site_configurations_table
-- Description: Create site configurations table for vendor endpoints

CREATE TABLE IF NOT EXISTS site_configurations (
    id VARCHAR(15) PRIMARY KEY,
    vendor_id VARCHAR(15) NOT NULL,
    name VARCHAR(255) NOT NULL,
    endpoint TEXT NOT NULL,
    type VARCHAR(20) CHECK (type IN ('WORDPRESS', 'SHOPIFY')),
    category VARCHAR(30) CHECK (category IN (
        'KEYBOARD', 'COMPONENTS', 'SWITCHES', 'KEYCAPS', 
        'DESKMAT', 'ACCESSORIES_OTHER', 'OTHER'
    )),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (vendor_id) REFERENCES vendors(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_site_configurations_vendor_id ON site_configurations (vendor_id);
CREATE INDEX IF NOT EXISTS idx_site_configurations_category ON site_configurations (category);
CREATE INDEX IF NOT EXISTS idx_site_configurations_active ON site_configurations (active);
CREATE INDEX IF NOT EXISTS idx_site_configurations_type ON site_configurations (type);

-- Create unique constraint for endpoint, category, and vendor combination
CREATE UNIQUE INDEX IF NOT EXISTS idx_site_configurations_unique ON site_configurations (endpoint, category, vendor_id);

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_site_configurations_updated_at 
    BEFORE UPDATE ON site_configurations 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();