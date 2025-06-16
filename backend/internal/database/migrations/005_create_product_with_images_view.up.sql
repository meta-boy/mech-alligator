-- Migration: 005_create_product_with_images_view
-- Description: Create view that joins products with their images and vendor information

CREATE OR REPLACE VIEW product_with_images AS
SELECT 
    p.id,
    p.name,
    p.description,
    p.price,
    p.currency,
    p.url,
    v.name AS vendor,
    p.in_stock,
    p.config_id,
    COALESCE(
        JSON_AGG(
            CASE 
                WHEN pi.url IS NOT NULL THEN pi.url 
                ELSE NULL 
            END
        ) FILTER (WHERE pi.url IS NOT NULL), 
        '[]'::json
    ) AS image_urls,
    p.created_at,
    p.updated_at
FROM 
    products p
LEFT JOIN 
    product_images pi ON p.id = pi.product_id
LEFT JOIN
    site_configurations sc ON p.config_id = sc.id
LEFT JOIN
    vendors v ON v.id = sc.vendor_id
GROUP BY 
    p.id, p.name, p.description, p.price, p.currency, 
    p.url, v.name, p.in_stock, p.config_id, p.created_at, p.updated_at;