-- Migration: 007_create_keyboard_tags_and_product_tags_tables (DOWN)
-- Description: Drop product_tags and keyboard_tags tables

DROP INDEX IF EXISTS idx_product_tags_tag;
DROP INDEX IF EXISTS idx_product_tags_product_id;
DROP TABLE IF EXISTS product_tags;
DROP TABLE IF EXISTS keyboard_tags;

