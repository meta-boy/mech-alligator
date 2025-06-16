package shopify

import (
	"net/http"
)

type ShopifyPlugin struct {
	client *http.Client
}

type ShopifyResponse struct {
	Products []ShopifyProduct `json:"products"`
}

type ShopifyProduct struct {
	ID          int64            `json:"id"`
	Title       string           `json:"title"`
	Handle      string           `json:"handle"`
	BodyHTML    string           `json:"body_html"`
	PublishedAt string           `json:"published_at"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
	Vendor      string           `json:"vendor"`
	ProductType string           `json:"product_type"`
	Tags        []string         `json:"tags"`
	Variants    []ShopifyVariant `json:"variants"`
	Images      []ShopifyImage   `json:"images"`
	Options     []ShopifyOption  `json:"options"`
}

type ShopifyVariant struct {
	ID              int64   `json:"id"`
	Title           string  `json:"title"`
	Option1         string  `json:"option1"`
	Option2         *string `json:"option2"`
	Option3         *string `json:"option3"`
	SKU             string  `json:"sku"`
	RequiresShipping bool   `json:"requires_shipping"`
	Taxable         bool    `json:"taxable"`
	FeaturedImage   *ShopifyVariantImage `json:"featured_image"`
	Available       bool    `json:"available"`
	Price           string  `json:"price"`
	Grams           int     `json:"grams"`
	CompareAtPrice  *string `json:"compare_at_price"`
	Position        int     `json:"position"`
	ProductID       int64   `json:"product_id"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type ShopifyVariantImage struct {
	ID         int64   `json:"id"`
	ProductID  int64   `json:"product_id"`
	Position   int     `json:"position"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
	Alt        *string `json:"alt"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Src        string  `json:"src"`
	VariantIDs []int64 `json:"variant_ids"`
}

type ShopifyImage struct {
	ID         int64   `json:"id"`
	CreatedAt  string  `json:"created_at"`
	Position   int     `json:"position"`
	UpdatedAt  string  `json:"updated_at"`
	ProductID  int64   `json:"product_id"`
	VariantIDs []int64 `json:"variant_ids"`
	Src        string  `json:"src"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Alt        *string `json:"alt"`
}

type ShopifyOption struct {
	Name     string   `json:"name"`
	Position int      `json:"position"`
	Values   []string `json:"values"`
}