package shopify

type Response struct {
	Products []Product `json:"products"`
}

type Product struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Handle      string    `json:"handle"`
	BodyHTML    string    `json:"body_html"`
	PublishedAt string    `json:"published_at"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	Vendor      string    `json:"vendor"`
	ProductType string    `json:"product_type"`
	Tags        []string  `json:"tags"`
	Variants    []Variant `json:"variants"`
	Images      []Image   `json:"images"`
}

type Variant struct {
	ID            int64         `json:"id"`
	Title         string        `json:"title"`
	Option1       string        `json:"option1"`
	Option2       *string       `json:"option2"`
	Option3       *string       `json:"option3"`
	SKU           string        `json:"sku"`
	Available     bool          `json:"available"`
	Price         string        `json:"price"`
	FeaturedImage *VariantImage `json:"featured_image"`
	ProductID     int64         `json:"product_id"`
}

type VariantImage struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Src       string `json:"src"`
	Alt       string `json:"alt"`
}

type Image struct {
	ID        int64  `json:"id"`
	ProductID int64  `json:"product_id"`
	Position  int    `json:"position"`
	Src       string `json:"src"`
	Alt       string `json:"alt"`
}
