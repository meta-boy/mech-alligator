package main

import (
	"context"
	"fmt"
	"log"

	"github.com/meta-boy/mech-alligator/internal/scraper"
	"github.com/meta-boy/mech-alligator/internal/scraper/plugins/shopify"
)

func main() {
	// Create a new Shopify plugin instance
	plugin := shopify.NewShopifyPlugin()

	// Define the scrape request
	// The SiteURL should be the base domain, and collection_handle specifies the collection.
	// The plugin will construct the final URL like: SiteURL/collections/collection_handle/products.json
	req := &scraper.ScrapeRequest{
		SiteURL:  "https://www.keebsmod.com",
		ConfigID: "test-shopify-scrape", // A unique ID for this scrape configuration
		Options: map[string]string{
			"collection_handle": "keycaps",
			// "limit": "10", // You can uncomment and set a limit for testing, max 250 per page
			// "include_images": "true", // Default is true
			// "include_variants": "true", // Default is true
		},
		// Credentials are not required for the public Shopify products.json API
	}

	log.Printf("Attempting to scrape with plugin: %s (Version: %s)", plugin.GetName(), plugin.GetVersion())
	log.Printf("Site URL: %s", req.SiteURL)
	if ch, ok := req.Options["collection_handle"]; ok {
		log.Printf("Collection Handle: %s", ch)
	}

	// Validate the request (optional, but good practice)
	err := plugin.Validate(req)
	if err != nil {
		log.Fatalf("Scrape request validation failed: %v", err)
	}
	log.Println("Scrape request validated successfully.")

	// Create a context
	ctx := context.Background()

	// Scrape all pages for the given request
	log.Println("Starting scrape process (all pages)...")
	result, err := plugin.ScrapeAllPages(ctx, req)
	if err != nil {
		log.Fatalf("Failed to scrape all pages: %v", err)
	}

	log.Printf("Scraping completed. Found %d products.", len(result.Products))

	if len(result.Errors) > 0 {
		log.Println("Encountered errors during scraping:")
		for i, errMsg := range result.Errors {
			log.Printf("Error %d: %s", i+1, errMsg)
		}
	}

	log.Println("\n--- Scraped Products ---")
	if len(result.Products) == 0 {
		log.Println("No products found.")
	} else {
		for i, product := range result.Products {
			fmt.Printf("\nProduct %d:\n", i+1)
			fmt.Printf("  ID:          %s\n", product.ID)
			fmt.Printf("  Name:        %s\n", product.Name)
			fmt.Printf("  Price:       %.2f %s\n", product.Price, product.Currency)
			fmt.Printf("  URL:         %s\n", product.URL)
			fmt.Printf("  In Stock:    %t\n", product.InStock)
			if len(product.Images) > 0 {
				fmt.Printf("  Images:      %s (and %d more)\n", product.Images[0], len(product.Images)-1)
			} else {
				fmt.Printf("  Images:      None\n")
			}
			fmt.Printf("  Metadata:\n")
			for key, val := range product.Metadata {
				fmt.Printf("    %s: %s\n", key, val)
			}
			// To keep output concise, don't print full description
			// fmt.Printf("  Description: %s\n", product.Description)
		}
	}

	log.Println("\n--- End of Scraped Products ---")
	log.Println("Shopify plugin test finished.")
}
