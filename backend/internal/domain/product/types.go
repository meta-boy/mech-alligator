package product

type Product struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Handle       string    `json:"handle" db:"handle"` // URL-friendly identifier
	URL          string    `json:"url" db:"url"`
	Brand        string    `json:"brand" db:"brand"`       // Actual product brand/vendor (Wuque Studio, GMK, etc.)
	Reseller     string    `json:"reseller" db:"reseller"` // Where it was scraped from (StacksKB, Meckeys, etc.)
	Category     string    `json:"category" db:"category"` // KEYBOARD, KEYCAPS, SWITCHES, etc.
	Tags         []string  `json:"tags" db:"tags"`
	Images       []string  `json:"images" db:"images"`
	Variants     []Variant `json:"variants,omitempty"`
	VariantCount int       `json:"variant_count" db:"variant_count"`

	// Source tracking
	SourceType     string            `json:"source_type" db:"source_type"` // SHOPIFY, WORDPRESS, etc.
	SourceID       string            `json:"source_id" db:"source_id"`     // Original ID from source
	ResellerID     string            `json:"reseller_id" db:"reseller_id"` // Config/reseller ID
	SourceMetadata map[string]string `json:"source_metadata,omitempty" db:"source_metadata"`
}

type Variant struct {
	ID        string   `json:"id" db:"id"`
	ProductID string   `json:"product_id" db:"product_id"`
	Name      string   `json:"name" db:"name"`
	SKU       string   `json:"sku,omitempty" db:"sku"`
	Price     float64  `json:"price" db:"price"`
	Currency  string   `json:"currency" db:"currency"`
	Available bool     `json:"available" db:"available"`
	URL       string   `json:"url,omitempty" db:"url"`       // Variant-specific URL if different
	Images    []string `json:"images,omitempty" db:"images"` // Variant-specific images

	// Variant options (color, size, etc.)
	Options map[string]string `json:"options,omitempty" db:"options"`

	// Source tracking
	SourceID string `json:"source_id" db:"source_id"` // Original variant ID from source
}

// Request/Response types for API
type ListRequest struct {
	Search    string   `json:"search,omitempty"`
	Brand     string   `json:"brand,omitempty"`    // Filter by product brand
	Reseller  string   `json:"reseller,omitempty"` // Filter by reseller
	Category  string   `json:"category,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	MinPrice  *float64 `json:"min_price,omitempty"`
	MaxPrice  *float64 `json:"max_price,omitempty"`
	Available *bool    `json:"available,omitempty"` // Filter by availability

	// Pagination
	Page     int `json:"page"`
	PageSize int `json:"page_size"`

	// Sorting
	SortBy    string `json:"sort_by"`    // name, price, brand, reseller
	SortOrder string `json:"sort_order"` // asc, desc
}

type ProductListResponse struct {
	Products   []Product      `json:"products"`
	Pagination PaginationMeta `json:"pagination"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// Filter helpers
func (r *ListRequest) Validate() {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.PageSize < 1 || r.PageSize > 100 {
		r.PageSize = 20
	}
	if r.SortBy == "" {
		r.SortBy = "name"
	}
	if r.SortOrder != "asc" && r.SortOrder != "desc" {
		r.SortOrder = "asc"
	}
}
