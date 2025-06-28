package scraper

import "context"

// Core scraping types
type ScrapedProduct struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Handle      string            `json:"handle,omitempty"`
	URL         string            `json:"url"`
	Brand       string            `json:"brand"` // Actual product brand (extracted from product data)
	Category    string            `json:"category"`
	Tags        []string          `json:"tags,omitempty"`
	Images      []string          `json:"images,omitempty"`
	Variants    []ScrapedVariant  `json:"variants"`
	SourceType  string            `json:"source_type"`
	SourceID    string            `json:"source_id"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type ScrapedVariant struct {
	Name      string            `json:"name,omitempty"`
	SKU       string            `json:"sku,omitempty"`
	Price     float64           `json:"price"`
	Currency  string            `json:"currency"`
	Available bool              `json:"available"`
	URL       string            `json:"url,omitempty"`
	Images    []string          `json:"images,omitempty"`
	Options   map[string]string `json:"options,omitempty"`
	SourceID  string            `json:"source_id"`
}

type ScrapeRequest struct {
	URL        string            `json:"url"`
	SourceType string            `json:"source_type"`       // SHOPIFY, WORDPRESS, etc.
	Reseller   string            `json:"reseller"`          // Name of the reseller website
	ResellerID string            `json:"reseller_id"`       // Config ID for the reseller
	Category   string            `json:"category"`          // Category to assign to scraped products
	Options    map[string]string `json:"options,omitempty"` // Plugin-specific options
}

type ScrapeResult struct {
	Products []ScrapedProduct `json:"products"`
	Errors   []string         `json:"errors,omitempty"`
	Stats    ScrapeStats      `json:"stats"`
}

type ScrapeStats struct {
	ProductsFound int    `json:"products_found"`
	VariantsFound int    `json:"variants_found"`
	ErrorCount    int    `json:"error_count"`
	Duration      string `json:"duration"`
	Source        string `json:"source"`
}

// Plugin interface
type Plugin interface {
	Name() string
	SupportedTypes() []string
	Scrape(ctx context.Context, req *ScrapeRequest) (*ScrapeResult, error)
	ValidateRequest(req *ScrapeRequest) error
}
