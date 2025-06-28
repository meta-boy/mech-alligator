package shopify

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/meta-boy/mech-alligator/internal/scraper"
)

type Plugin struct {
	client *http.Client
}

func NewShopifyPlugin() *Plugin {
	return &Plugin{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Plugin) Name() string {
	return "shopify"
}

func (p *Plugin) SupportedTypes() []string {
	return []string{"SHOPIFY"}
}

func (p *Plugin) ValidateRequest(req *scraper.ScrapeRequest) error {
	if req.URL == "" {
		return fmt.Errorf("URL is required")
	}
	if !strings.HasPrefix(req.URL, "http") {
		return fmt.Errorf("URL must include protocol (http/https)")
	}
	return nil
}

func (p *Plugin) Scrape(ctx context.Context, req *scraper.ScrapeRequest) (*scraper.ScrapeResult, error) {
	start := time.Now()

	// Build API URL
	apiURL := p.buildAPIURL(req)

	// Fetch products
	shopifyProducts, err := p.fetchProducts(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// Convert to our format
	var products []scraper.ScrapedProduct
	var errors []string
	variantCount := 0

	for _, sp := range shopifyProducts {
		product, productErrors := p.convertProduct(sp, req)
		if product != nil {
			products = append(products, *product)
			variantCount += len(product.Variants)
		}
		errors = append(errors, productErrors...)
	}

	return &scraper.ScrapeResult{
		Products: products,
		Errors:   errors,
		Stats: scraper.ScrapeStats{
			ProductsFound: len(products),
			VariantsFound: variantCount,
			ErrorCount:    len(errors),
			Duration:      time.Since(start).String(),
			Source:        req.Reseller,
		},
	}, nil
}

func (p *Plugin) buildAPIURL(req *scraper.ScrapeRequest) string {
	baseURL := strings.TrimSuffix(req.URL, "/")

	// Handle collection URLs
	if strings.Contains(baseURL, "/collections/") {
		return baseURL + "/products.json?limit=250"
	}

	// Default to all products
	return baseURL + "/products.json?limit=250"
}

func (p *Plugin) fetchProducts(ctx context.Context, url string) ([]Product, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ProductScraper/1.0)")
	request.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	var shopifyResp Response
	if err := json.NewDecoder(resp.Body).Decode(&shopifyResp); err != nil {
		return nil, err
	}

	return shopifyResp.Products, nil
}

func (p *Plugin) convertProduct(sp Product, req *scraper.ScrapeRequest) (*scraper.ScrapedProduct, []string) {
	var errors []string

	// Extract brand from vendor field or product title
	brand := p.extractBrand(sp.Vendor, sp.Title)

	// Build base URL for product links
	baseURL := strings.TrimSuffix(req.URL, "/")
	if strings.Contains(baseURL, "/collections/") {
		// Extract domain from collection URL
		parts := strings.Split(baseURL, "/collections/")
		if len(parts) > 0 {
			baseURL = parts[0]
		}
	}

	productURL := baseURL + "/products/" + sp.Handle

	// Convert images
	var images []string
	for _, img := range sp.Images {
		images = append(images, img.Src)
	}

	// Convert variants
	var variants []scraper.ScrapedVariant
	for _, sv := range sp.Variants {
		variant, err := p.convertVariant(sv, productURL, baseURL)
		if err != nil {
			errors = append(errors, fmt.Sprintf("variant %d: %s", sv.ID, err.Error()))
			continue
		}
		variants = append(variants, variant)
	}

	// If no variants, skip this product
	if len(variants) == 0 {
		errors = append(errors, fmt.Sprintf("product %d has no valid variants", sp.ID))
		return nil, errors
	}

	product := &scraper.ScrapedProduct{
		Name:        sp.Title,
		Description: sp.BodyHTML,
		Handle:      sp.Handle,
		URL:         productURL,
		Brand:       brand,
		Category:    req.Category,
		Tags:        sp.Tags,
		Images:      images,
		Variants:    variants,
		SourceType:  "SHOPIFY",
		SourceID:    strconv.FormatInt(sp.ID, 10),
		Metadata: map[string]string{
			"shopify_product_type": sp.ProductType,
			"shopify_vendor":       sp.Vendor,
			"shopify_handle":       sp.Handle,
			"published_at":         sp.PublishedAt,
			"created_at":           sp.CreatedAt,
			"updated_at":           sp.UpdatedAt,
		},
	}

	return product, errors
}

func (p *Plugin) convertVariant(sv Variant, productURL, baseURL string) (scraper.ScrapedVariant, error) {
	price, err := strconv.ParseFloat(sv.Price, 64)
	if err != nil {
		return scraper.ScrapedVariant{}, fmt.Errorf("invalid price: %s", sv.Price)
	}

	// Build variant URL
	variantURL := productURL
	if len(sv.Title) > 0 && sv.Title != "Default Title" {
		variantURL += "?variant=" + strconv.FormatInt(sv.ID, 10)
	}

	// Build variant options
	options := make(map[string]string)
	if sv.Option1 != "" {
		options["option1"] = sv.Option1
	}
	if sv.Option2 != nil && *sv.Option2 != "" {
		options["option2"] = *sv.Option2
	}
	if sv.Option3 != nil && *sv.Option3 != "" {
		options["option3"] = *sv.Option3
	}

	// Variant-specific images
	var images []string
	if sv.FeaturedImage != nil {
		images = append(images, sv.FeaturedImage.Src)
	}

	variant := scraper.ScrapedVariant{
		Name:      sv.Title,
		SKU:       sv.SKU,
		Price:     price,
		Currency:  "INR", // Default, should be configurable per reseller
		Available: sv.Available,
		URL:       variantURL,
		Images:    images,
		Options:   options,
		SourceID:  strconv.FormatInt(sv.ID, 10),
	}

	return variant, nil
}

// extractBrand tries to determine the actual brand from Shopify vendor field or product title
func (p *Plugin) extractBrand(vendor, title string) string {
	// If vendor is meaningful, use it
	if vendor != "" && !p.isGenericVendor(vendor) {
		return vendor
	}

	// Try to extract brand from title using common patterns
	title = strings.ToLower(title)

	// Known brand patterns
	brands := []string{
		"wuque studio", "gmk", "cherry", "keychron", "akko", "ducky",
		"leopold", "varmilo", "filco", "topre", "hhkb", "realforce",
		"drop", "novelkeys", "gateron", "kailh", "holy panda", "zealios",
		"artisan", "sa", "oem", "cherry profile", "xda", "dsa",
	}

	for _, brand := range brands {
		if strings.Contains(title, brand) {
			return p.capitalizeWords(brand)
		}
	}

	// Fallback to vendor even if generic
	if vendor != "" {
		return vendor
	}

	return "Unknown"
}

func (p *Plugin) isGenericVendor(vendor string) bool {
	genericVendors := []string{
		"default", "shop", "store", "keyboards", "keycaps", "switches",
		"mechanical", "gaming", "tech", "electronics",
	}

	vendorLower := strings.ToLower(vendor)
	for _, generic := range genericVendors {
		if strings.Contains(vendorLower, generic) {
			return true
		}
	}

	return false
}

func (p *Plugin) capitalizeWords(s string) string {
	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// ShopifyResponse represents the JSON response from Shopify products API
