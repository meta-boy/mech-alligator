-- Resellers table (websites that sell products)
CREATE TABLE resellers (
                           id VARCHAR(15) PRIMARY KEY,
                           name VARCHAR(255) NOT NULL,
                           country VARCHAR(3) DEFAULT 'IN',
                           website TEXT,
                           currency VARCHAR(3) DEFAULT 'INR',
                           active BOOLEAN DEFAULT true
);

-- Reseller configurations (specific scraping endpoints)
CREATE TABLE reseller_configs (
                                  id VARCHAR(15) PRIMARY KEY,
                                  reseller_id VARCHAR(15) NOT NULL REFERENCES resellers(id) ON DELETE CASCADE,
                                  name VARCHAR(255) NOT NULL,
                                  url TEXT NOT NULL,
                                  source_type VARCHAR(20) NOT NULL, -- SHOPIFY, WORDPRESS, etc.
                                  category VARCHAR(30) NOT NULL,    -- KEYBOARD, KEYCAPS, SWITCHES, etc.
                                  active BOOLEAN DEFAULT true,
                                  options JSONB DEFAULT '{}'
);

-- Brands table (actual product manufacturers)
CREATE TABLE brands (
                        id VARCHAR(15) PRIMARY KEY,
                        name VARCHAR(255) NOT NULL UNIQUE,
                        country VARCHAR(3),
                        website TEXT,
                        description TEXT
);

-- Products table (simplified)
CREATE TABLE products (
                          id VARCHAR(15) PRIMARY KEY,
                          name TEXT NOT NULL,
                          description TEXT,
                          handle VARCHAR(255), -- URL-friendly identifier
                          url TEXT NOT NULL,
                          brand VARCHAR(255),   -- Brand name (denormalized for performance)
                          reseller VARCHAR(255), -- Reseller name (denormalized)
                          reseller_id VARCHAR(15) REFERENCES resellers(id),
                          category VARCHAR(30),
                          tags TEXT[], -- PostgreSQL array of tags
                          images TEXT[], -- Array of image URLs

    -- Source tracking
                          source_type VARCHAR(20), -- SHOPIFY, WORDPRESS, etc.
                          source_id VARCHAR(50),   -- Original ID from source system
                          source_metadata JSONB DEFAULT '{}'
);

-- Product variants table
CREATE TABLE product_variants (
                                  id VARCHAR(15) PRIMARY KEY,
                                  product_id VARCHAR(15) NOT NULL REFERENCES products(id) ON DELETE CASCADE,
                                  name VARCHAR(255),
                                  sku VARCHAR(100),
                                  price DECIMAL(10,2) NOT NULL,
                                  currency VARCHAR(3) DEFAULT 'INR',
                                  available BOOLEAN DEFAULT false,
                                  url TEXT, -- Variant-specific URL if different
                                  images TEXT[], -- Variant-specific images
                                  options JSONB DEFAULT '{}', -- Color, size, etc. as key-value pairs
                                  source_id VARCHAR(50) -- Original variant ID from source
);

-- Jobs table (simplified)
CREATE TABLE jobs (
                      id VARCHAR(255) PRIMARY KEY,
                      type VARCHAR(50) NOT NULL,
                      status VARCHAR(20) DEFAULT 'pending',
                      payload JSONB NOT NULL DEFAULT '{}',
                      result JSONB DEFAULT '{}',
                      error_message TEXT,
                      attempts INTEGER DEFAULT 0,
                      max_attempts INTEGER DEFAULT 3,
                      scheduled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_products_brand ON products (brand);
CREATE INDEX idx_products_reseller ON products (reseller);
CREATE INDEX idx_products_category ON products (category);
CREATE INDEX idx_products_source ON products (source_type, source_id);
CREATE INDEX idx_product_variants_product_id ON product_variants (product_id);
CREATE INDEX idx_product_variants_price ON product_variants (price);
CREATE INDEX idx_product_variants_available ON product_variants (available);
CREATE INDEX idx_jobs_status_scheduled ON jobs (status, scheduled_at);


CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Step 1: Drop all foreign key constraints first
ALTER TABLE reseller_configs DROP CONSTRAINT IF EXISTS reseller_configs_reseller_id_fkey;
ALTER TABLE products DROP CONSTRAINT IF EXISTS products_reseller_id_fkey;
ALTER TABLE product_variants DROP CONSTRAINT IF EXISTS product_variants_product_id_fkey;

-- Step 2: Modify resellers table (parent table first)
ALTER TABLE resellers
    ALTER COLUMN id TYPE UUID USING gen_random_uuid(),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- Step 3: Modify reseller_configs table
ALTER TABLE reseller_configs
    ALTER COLUMN id TYPE UUID USING gen_random_uuid(),
    ALTER COLUMN id SET DEFAULT gen_random_uuid(),
    ALTER COLUMN reseller_id TYPE UUID USING gen_random_uuid();

-- Step 4: Modify brands table (independent)
ALTER TABLE brands
    ALTER COLUMN id TYPE UUID USING gen_random_uuid(),
    ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- Step 5: Modify products table
ALTER TABLE products
    ALTER COLUMN id TYPE UUID USING gen_random_uuid(),
    ALTER COLUMN id SET DEFAULT gen_random_uuid(),
    ALTER COLUMN reseller_id TYPE UUID USING gen_random_uuid();

-- Step 6: Modify product_variants table
ALTER TABLE product_variants
    ALTER COLUMN id TYPE UUID USING gen_random_uuid(),
    ALTER COLUMN id SET DEFAULT gen_random_uuid(),
    ALTER COLUMN product_id TYPE UUID USING gen_random_uuid();

-- Step 7: Recreate foreign key constraints
ALTER TABLE reseller_configs
    ADD CONSTRAINT reseller_configs_reseller_id_fkey
        FOREIGN KEY (reseller_id) REFERENCES resellers(id) ON DELETE CASCADE;

ALTER TABLE products
    ADD CONSTRAINT products_reseller_id_fkey
        FOREIGN KEY (reseller_id) REFERENCES resellers(id) ON DELETE SET NULL;

ALTER TABLE product_variants
    ADD CONSTRAINT product_variants_product_id_fkey
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;