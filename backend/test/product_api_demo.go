package main

import (
	"context"
	"fmt"
	"log"
	"net/url"

	"github.com/meta-boy/mech-alligator/internal/config"
	"github.com/meta-boy/mech-alligator/internal/database"
	"github.com/meta-boy/mech-alligator/internal/domain/product"
	"github.com/meta-boy/mech-alligator/internal/repository/postgres"
	"github.com/meta-boy/mech-alligator/internal/service"
)

func main() {
	log.Println("Testing Product API...")

	// Load database configuration
	cfg := config.LoadDatabaseConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid database config: %v", err)
	}

	// Connect to database
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Health(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	log.Println("Database connection established")

	// Create repositories and services
	productRepo := postgres.NewProductRepository(db)
	productService := service.NewProductService(productRepo)

	ctx := context.Background()

	// Test 1: List products with basic pagination
	log.Println("\n=== Test 1: Basic product listing ===")
	req := product.ProductListRequest{
		Pagination: product.ProductPagination{
			Page:     1,
			PageSize: 5,
		},
		Sort: product.ProductSort{
			Field: "created_at",
			Order: "desc",
		},
	}

	response, err := productService.ListProducts(ctx, req)
	if err != nil {
		log.Fatalf("Failed to list products: %v", err)
	}

	log.Printf("Found %d products (page %d of %d)", 
		len(response.Products), response.Pagination.Page, response.Pagination.TotalPages)
	log.Printf("Total items: %d", response.Pagination.TotalItems)

	for i, p := range response.Products {
		log.Printf("  %d. %s - %.2f %s (Stock: %t)", 
			i+1, p.Name, p.Price, p.Currency, p.InStock)
		if p.Vendor != "" {
			log.Printf("     Vendor: %s", p.Vendor)
		}
		if len(p.Tags) > 0 {
			log.Printf("     Tags: %v", p.Tags)
		}
	}

	// Test 2: Filter by vendor
	log.Println("\n=== Test 2: Filter by vendor ===")
	if len(response.Products) > 0 && response.Products[0].Vendor != "" {
		vendorFilter := product.ProductListRequest{
			Filter: product.ProductFilter{
				Vendor: response.Products[0].Vendor,
			},
			Pagination: product.ProductPagination{
				Page:     1,
				PageSize: 3,
			},
		}

		vendorResponse, err := productService.ListProducts(ctx, vendorFilter)
		if err != nil {
			log.Printf("Failed to filter by vendor: %v", err)
		} else {
			log.Printf("Found %d products for vendor '%s'", 
				len(vendorResponse.Products), response.Products[0].Vendor)
		}
	}

	// Test 3: Price range filter
	log.Println("\n=== Test 3: Price range filter ===")
	minPrice := 1000.0
	maxPrice := 5000.0
	priceFilter := product.ProductListRequest{
		Filter: product.ProductFilter{
			MinPrice: &minPrice,
			MaxPrice: &maxPrice,
		},
		Pagination: product.ProductPagination{
			Page:     1,
			PageSize: 5,
		},
	}

	priceResponse, err := productService.ListProducts(ctx, priceFilter)
	if err != nil {
		log.Printf("Failed to filter by price: %v", err)
	} else {
		log.Printf("Found %d products between %.2f and %.2f", 
			len(priceResponse.Products), minPrice, maxPrice)
		for _, p := range priceResponse.Products {
			log.Printf("  - %s: %.2f %s", p.Name, p.Price, p.Currency)
		}
	}

	// Test 4: Search filter
	log.Println("\n=== Test 4: Search filter ===")
	searchFilter := product.ProductListRequest{
		Filter: product.ProductFilter{
			Search: "keyboard",
		},
		Pagination: product.ProductPagination{
			Page:     1,
			PageSize: 3,
		},
	}

	searchResponse, err := productService.ListProducts(ctx, searchFilter)
	if err != nil {
		log.Printf("Failed to search products: %v", err)
	} else {
		log.Printf("Found %d products matching 'keyboard'", len(searchResponse.Products))
		for _, p := range searchResponse.Products {
			log.Printf("  - %s", p.Name)
		}
	}

	// Test 5: Get filter options
	log.Println("\n=== Test 5: Filter options ===")
	options, err := productService.GetFilterOptions(ctx)
	if err != nil {
		log.Printf("Failed to get filter options: %v", err)
	} else {
		log.Printf("Available vendors: %v", options.Vendors)
		log.Printf("Available tags: %v", options.Tags)
		log.Printf("Available currencies: %v", options.Currencies)
	}

	// Test 6: Get single product (if we have any)
	if len(response.Products) > 0 {
		log.Println("\n=== Test 6: Get single product ===")
		productID := response.Products[0].ID
		singleProduct, err := productService.GetProduct(ctx, productID)
		if err != nil {
			log.Printf("Failed to get product: %v", err)
		} else {
			log.Printf("Product details for ID %s:", productID)
			log.Printf("  Name: %s", singleProduct.Name)
			log.Printf("  Price: %.2f %s", singleProduct.Price, singleProduct.Currency)
			log.Printf("  Vendor: %s", singleProduct.Vendor)
			log.Printf("  In Stock: %t", singleProduct.InStock)
			log.Printf("  Tags: %v", singleProduct.Tags)
			log.Printf("  Images: %d", len(singleProduct.ImageURLs))
		}
	}

	log.Println("\n=== Product API Test Complete ===")
}

// Helper function to demonstrate URL building for API calls
func buildProductListURL(baseURL string, filter product.ProductFilter, sort product.ProductSort, pagination product.ProductPagination) string {
	u, _ := url.Parse(baseURL + "/api/products")
	q := u.Query()

	if filter.Search != "" {
		q.Set("search", filter.Search)
	}
	if filter.Vendor != "" {
		q.Set("vendor", filter.Vendor)
	}
	if filter.Currency != "" {
		q.Set("currency", filter.Currency)
	}
	if filter.InStock != nil {
		q.Set("in_stock", fmt.Sprintf("%t", *filter.InStock))
	}
	if filter.MinPrice != nil {
		q.Set("min_price", fmt.Sprintf("%.2f", *filter.MinPrice))
	}
	if filter.MaxPrice != nil {
		q.Set("max_price", fmt.Sprintf("%.2f", *filter.MaxPrice))
	}
	if len(filter.Tags) > 0 {
		q.Set("tags", fmt.Sprintf("%v", filter.Tags))
	}

	q.Set("sort_field", sort.Field)
	q.Set("sort_order", sort.Order)
	q.Set("page", fmt.Sprintf("%d", pagination.Page))
	q.Set("page_size", fmt.Sprintf("%d", pagination.PageSize))

	u.RawQuery = q.Encode()
	return u.String()
}
