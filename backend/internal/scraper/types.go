package scraper

import (
	"context"
	"time"
)

type Product struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Price       float64           `json:"price"`
	Currency    string            `json:"currency"`
	URL         string            `json:"url"`
	InStock     bool              `json:"in_stock"`
	Images      []string          `json:"images"`
	Metadata    map[string]string `json:"metadata,omitempty"` // For plugin-specific data
}

type ScrapeRequest struct {
	ConfigID    string            `json:"config_id"`
	VendorID    string            `json:"vendor_id"`
	VendorName  string            `json:"vendor_name"`
	SiteURL     string            `json:"site_url"`
	SiteType    string            `json:"site_type"`    // SHOPIFY, WORDPRESS, CUSTOM
	Category    string            `json:"category"`     // KEYBOARD, COMPONENTS, etc.
	Credentials map[string]string `json:"credentials"`  // API keys, tokens, etc.
	Options     map[string]string `json:"options"`      // Plugin-specific options
}

type ScrapeResult struct {
	Products []Product `json:"products"`
	Errors   []string  `json:"errors,omitempty"`
	Metadata struct {
		ScrapedAt    time.Time `json:"scraped_at"`
		TotalFound   int       `json:"total_found"`
		TotalErrors  int       `json:"total_errors"`
		Duration     string    `json:"duration"`
		PluginName   string    `json:"plugin_name"`
		PluginVersion string   `json:"plugin_version"`
	} `json:"metadata"`
}

type Plugin interface {
	// GetName returns the plugin name (e.g., "shopify", "wordpress")
	GetName() string
	
	// GetVersion returns the plugin version
	GetVersion() string
	
	// GetSupportedTypes returns the site types this plugin supports
	GetSupportedTypes() []string
	
	// Validate checks if the scrape request is valid for this plugin
	Validate(req *ScrapeRequest) error
	
	// Scrape performs the actual scraping operation
	Scrape(ctx context.Context, req *ScrapeRequest) (*ScrapeResult, error)
	
	// GetRequiredCredentials returns the required credential keys
	GetRequiredCredentials() []string
	
	// GetSupportedOptions returns the supported option keys with descriptions
	GetSupportedOptions() map[string]string
}