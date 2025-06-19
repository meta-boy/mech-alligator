-- Migration: 007_create_keyboard_tags_and_product_tags_tables
-- Description: Create tables for keyboard tags (enum) and product-tag mapping (many-to-many)

CREATE TABLE IF NOT EXISTS keyboard_tags (
    tag VARCHAR(20) PRIMARY KEY CHECK (tag IN (
        'keyboard', 'keycaps', 'switches', 'accessories',
        'linear', 'tactile', 'clicky', 'silent',
        'full_size', 'tkl', 'compact', 'split', 'ergonomic',
        'hot_swap', 'wireless', 'rgb', 'programmable'
    ))
);

CREATE TABLE IF NOT EXISTS product_tags (
    product_id VARCHAR(15) REFERENCES products(id) ON DELETE CASCADE,
    tag VARCHAR(20) REFERENCES keyboard_tags(tag) ON DELETE CASCADE,
    PRIMARY KEY (product_id, tag)
);

CREATE INDEX IF NOT EXISTS idx_product_tags_product_id ON product_tags (product_id);
CREATE INDEX IF NOT EXISTS idx_product_tags_tag ON product_tags (tag);

