package config

// Reseller represents a store/website that sells products from multiple brands
type Reseller struct {
	ID       string `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`         // StacksKB, Meckeys, etc.
	Country  string `json:"country" db:"country"`   // IN, US, etc.
	Website  string `json:"website" db:"website"`   // Base website URL
	Currency string `json:"currency" db:"currency"` // Default currency for this reseller
	Active   bool   `json:"active" db:"active"`
}

// ResellerConfig represents a specific scraping configuration for a reseller
type ResellerConfig struct {
	ID         string `json:"id" db:"id"`
	ResellerID string `json:"reseller_id" db:"reseller_id"`
	Name       string `json:"name" db:"name"`               // "StacksKB Keyboards", "Meckeys Keycaps"
	URL        string `json:"url" db:"url"`                 // Specific URL to scrape
	SourceType string `json:"source_type" db:"source_type"` // SHOPIFY, WORDPRESS, etc.
	Category   string `json:"category" db:"category"`       // KEYBOARD, KEYCAPS, SWITCHES, etc.
	Active     bool   `json:"active" db:"active"`

	// Scraping options specific to this config
	Options map[string]string `json:"options,omitempty" db:"options"`
}

// Brand represents actual product manufacturers/brands
type Brand struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`       // Wuque Studio, GMK, Cherry, etc.
	Country     string `json:"country" db:"country"` // Country of origin
	Website     string `json:"website,omitempty" db:"website"`
	Description string `json:"description,omitempty" db:"description"`
}

// Job payload for scraping
type ScrapeJobPayload struct {
	ConfigID     string            `json:"config_id"`
	ResellerID   string            `json:"reseller_id"`
	ResellerName string            `json:"reseller_name"`
	URL          string            `json:"url"`
	SourceType   string            `json:"source_type"`
	Category     string            `json:"category"`
	Options      map[string]string `json:"options,omitempty"`
}
