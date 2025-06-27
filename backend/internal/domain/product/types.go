package product

import "time"

type Product struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       float64   `json:"price" db:"price"`
	Currency    string    `json:"currency" db:"currency"`
	URL         string    `json:"url" db:"url"`
	ConfigID    string    `json:"config_id" db:"config_id"`
	InStock     bool      `json:"in_stock" db:"in_stock"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Joined data
	Vendor    string   `json:"vendor,omitempty" db:"vendor"`
	ImageURLs []string `json:"image_urls" db:"image_urls"`
	Tags      []string `json:"tags" db:"tags"`
}

type ProductFilter struct {
	// Search filters
	Search   string   `json:"search,omitempty"`
	Vendor   string   `json:"vendor,omitempty"`
	ConfigID string   `json:"config_id,omitempty"`
	Currency string   `json:"currency,omitempty"`
	InStock  *bool    `json:"in_stock,omitempty"`
	Tags     []string `json:"tags,omitempty"`

	// Price range filters
	MinPrice *float64 `json:"min_price,omitempty"`
	MaxPrice *float64 `json:"max_price,omitempty"`

	// Date range filters
	CreatedAfter  *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
}

type ProductSort struct {
	Field string `json:"field"` // name, price, created_at, updated_at
	Order string `json:"order"` // asc, desc
}

type ProductPagination struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Offset   int `json:"-"`
}

type ProductListRequest struct {
	Filter     ProductFilter     `json:"filter,omitempty"`
	Sort       ProductSort       `json:"sort,omitempty"`
	Pagination ProductPagination `json:"pagination"`
}

type ProductListResponse struct {
	Products   []Product      `json:"products"`
	Pagination PaginationMeta `json:"pagination"`
	Filter     ProductFilter  `json:"filter,omitempty"`
	Sort       ProductSort    `json:"sort,omitempty"`
}

type PaginationMeta struct {
	Page        int   `json:"page"`
	PageSize    int   `json:"page_size"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

func (p *ProductPagination) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = 20
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
	p.Offset = (p.Page - 1) * p.PageSize
}

func (s *ProductSort) Validate() {
	validFields := map[string]bool{
		"name":       true,
		"price":      true,
		"created_at": true,
		"updated_at": true,
	}

	if !validFields[s.Field] {
		s.Field = "created_at"
	}

	if s.Order != "asc" && s.Order != "desc" {
		s.Order = "desc"
	}
}
