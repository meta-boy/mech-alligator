package main

import 
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
_ "github.com/lib/pq"
C
)
	// Connect to database
	db, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/mech_alligator?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Test query - get first product with images
	query := `
query := `
SELECT 
p.id, p.name,
COALESCE(
JSON_AGG(
DISTINCT pi.url
) FILTER (WHERE pi.url IS NOT NULL),
'[]'::json
) AS image_urls
FROM products p
LEFT JOIN product_images pi ON p.id = pi.product_id
GROUP BY p.id, p.name
LIMIT 1
	var id, name, imageURLsJSON string
	err = db.QueryRowContext(context.Background(), query).Scan(&id, &name, &imageURLsJSON)
	if err != nil {
		log.Printf("Query error: %v", err)
		return
	}
	fmt.Printf("Product ID: %s\n", id)
	fmt.Printf("Product Name: %s\n", name)
	fmt.Printf("Raw image URLs JSON: %s\n", imageURLsJSON)
	// Parse JSON
	var imageURLs []string
	if err := json.Unmarshal([]byte(imageURLsJSON), &imageURLs); err != nil {
		fmt.Printf("JSON parse error: %v\n", err)
	} else {
		fmt.Printf("Parsed image URLs: %v\n", imageURLs)
	}
	// Check if any images exist
	var imageCount int
	countQuery := "SELECT COUNT(*) FROM product_images WHERE product_id = $1"
	err = db.QueryRowContext(context.Background(), countQuery, id).Scan(&imageCount)
	if err != nil {
		log.Printf("Count query error: %v", err)
	} else {
		fmt.Printf("Number of images for product %s: %d\n", id, imageCount)
	}
}
}
