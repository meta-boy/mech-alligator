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

func NewShopifyPlugin() *ShopifyPlugin {
	return &ShopifyPlugin{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *ShopifyPlugin) GetName() string {
	return "shopify"
}

func (p *ShopifyPlugin) GetVersion() string {
	return "1.0.0"
}

func (p *ShopifyPlugin) GetSupportedTypes() []string {
	return []string{"SHOPIFY"}
}

func (p *ShopifyPlugin) GetRequiredCredentials() []string {
	return []string{} // Public API doesn't need credentials
}

func (p *ShopifyPlugin) GetSupportedOptions() map[string]string {
	return map[string]string{
		"limit":              "Number of products to fetch (default: 250, max: 250)",
		"collection_handle":  "Specific collection handle to scrape (e.g., 'keycaps')",
		"use_storefront_api": "Use Storefront API instead of products.json (requires access token)",
		"page":               "Page number for pagination (default: 1)",
		"include_images":     "Include product images (true/false, default: true)",
		"include_variants":   "Include all variants or just first one (true/false, default: true)",
	}
}

func (p *ShopifyPlugin) Validate(req *scraper.ScrapeRequest) error {
	if req.SiteURL == "" {
		return fmt.Errorf("site_url is required")
	}

	// Clean URL - remove trailing slashes and ensure it's a valid domain
	siteURL := strings.TrimSuffix(req.SiteURL, "/")
	if !strings.HasPrefix(siteURL, "http://") && !strings.HasPrefix(siteURL, "https://") {
		return fmt.Errorf("site_url must include http:// or https://")
	}

	// Check if limit is valid
	if limitStr := req.Options["limit"]; limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err != nil || limit <= 0 || limit > 250 {
			return fmt.Errorf("limit must be a number between 1 and 250")
		}
	}

	return nil
}

func (p *ShopifyPlugin) Scrape(ctx context.Context, req *scraper.ScrapeRequest) (*scraper.ScrapeResult, error) {
	result := &scraper.ScrapeResult{
		Products: []scraper.Product{},
		Errors:   []string{},
	}

	// Build the API URL
	apiURL, err := p.buildAPIURL(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build API URL: %w", err)
	}

	// Fetch products from Shopify
	shopifyProducts, err := p.fetchProducts(ctx, apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// Convert each Shopify product to our standard format
	includeAllVariants := req.Options["include_variants"] != "false"
	includeImages := req.Options["include_images"] != "false"

	for _, shopifyProduct := range shopifyProducts {
		products, errs := p.convertProduct(shopifyProduct, req, includeAllVariants, includeImages)
		result.Products = append(result.Products, products...)
		result.Errors = append(result.Errors, errs...)
	}

	return result, nil
}

func (p *ShopifyPlugin) buildAPIURL(req *scraper.ScrapeRequest) (string, error) {
	baseURL := strings.TrimSuffix(req.SiteURL, "/")

	// Check if we're targeting a specific collection
	if collectionHandle := req.Options["collection_handle"]; collectionHandle != "" {
		baseURL += "/collections/" + collectionHandle
	}

	apiURL := baseURL + "/products.json"

	// Add query parameters
	params := []string{}

	if limit := req.Options["limit"]; limit != "" {
		params = append(params, "limit="+limit)
	} else {
		params = append(params, "limit=250") // Default max
	}

	if page := req.Options["page"]; page != "" {
		params = append(params, "page="+page)
	}

	if len(params) > 0 {
		apiURL += "?" + strings.Join(params, "&")
	}

	return apiURL, nil
}

func (p *ShopifyPlugin) fetchProducts(ctx context.Context, url string) ([]ShopifyProduct, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add user agent to appear more legitimate
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; MechAlligator/1.0)")
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: failed to fetch products from %s", resp.StatusCode, url)
	}

	var shopifyResponse ShopifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&shopifyResponse); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return shopifyResponse.Products, nil
}

func (p *ShopifyPlugin) convertProduct(shopifyProduct ShopifyProduct, req *scraper.ScrapeRequest, includeAllVariants, includeImages bool) ([]scraper.Product, []string) {
	var products []scraper.Product
	var errors []string

	baseURL := strings.TrimSuffix(req.SiteURL, "/")

	// Extract images
	var images []string
	if includeImages {
		for _, img := range shopifyProduct.Images {
			images = append(images, img.Src)
		}
	}

	// Convert tags to string
	tags := strings.Join(shopifyProduct.Tags, ",")

	currency := "INR" // Default, could be improved by storing currency against vendor

	if includeAllVariants && len(shopifyProduct.Variants) > 1 {
		// Create a product for each variant
		for _, variant := range shopifyProduct.Variants {
			product, err := p.convertVariantToProduct(shopifyProduct, variant, baseURL, currency, images, tags, req.ConfigID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to convert variant %d: %v", variant.ID, err))
				continue
			}
			products = append(products, product)
		}
	} else {
		// Create a single product using the first/default variant
		var selectedVariant ShopifyVariant
		if len(shopifyProduct.Variants) > 0 {
			selectedVariant = shopifyProduct.Variants[0]
		}

		product, err := p.convertVariantToProduct(shopifyProduct, selectedVariant, baseURL, currency, images, tags, req.ConfigID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to convert product %d: %v", shopifyProduct.ID, err))
		} else {
			products = append(products, product)
		}
	}

	return products, errors
}

func (p *ShopifyPlugin) convertVariantToProduct(shopifyProduct ShopifyProduct, variant ShopifyVariant, baseURL, currency string, images []string, tags, configID string) (scraper.Product, error) {
	// Parse price
	var price float64
	if variant.Price != "" {
		priceFloat, err := strconv.ParseFloat(variant.Price, 64)
		if err != nil {
			price = 0
		} else {
			price = priceFloat
		}
	}

	// Build product URL
	productURL := baseURL + "/products/" + shopifyProduct.Handle
	if len(shopifyProduct.Variants) > 1 {
		productURL += "?variant=" + strconv.FormatInt(variant.ID, 10)
	}

	// Create unique ID for variant
	productID := strconv.FormatInt(shopifyProduct.ID, 10)
	if len(shopifyProduct.Variants) > 1 {
		productID += "-" + strconv.FormatInt(variant.ID, 10)
	}

	// Build product name
	productName := shopifyProduct.Title
	if len(shopifyProduct.Variants) > 1 && variant.Title != "Default Title" {
		productName += " - " + variant.Title
	}

	// Add variant-specific image if available
	variantImages := make([]string, len(images))
	copy(variantImages, images)
	if variant.FeaturedImage != nil {
		// Prepend variant image to the beginning
		variantImages = append([]string{variant.FeaturedImage.Src}, variantImages...)
	}

	// Build metadata
	metadata := map[string]string{
		"shopify_product_id": strconv.FormatInt(shopifyProduct.ID, 10),
		"shopify_variant_id": strconv.FormatInt(variant.ID, 10),
		"handle":             shopifyProduct.Handle,
		"vendor":             shopifyProduct.Vendor,
		"product_type":       shopifyProduct.ProductType,
		"tags":               tags,
		"created_at":         shopifyProduct.CreatedAt,
		"updated_at":         shopifyProduct.UpdatedAt,
		"variant_sku":        variant.SKU,
		"variant_position":   strconv.Itoa(variant.Position),
		"variant_grams":      strconv.Itoa(variant.Grams),
		"requires_shipping":  strconv.FormatBool(variant.RequiresShipping),
		"taxable":            strconv.FormatBool(variant.Taxable),
	}

	if variant.CompareAtPrice != nil {
		metadata["compare_at_price"] = *variant.CompareAtPrice
	}

	if variant.Option1 != "" {
		metadata["option1"] = variant.Option1
	}
	if variant.Option2 != nil && *variant.Option2 != "" {
		metadata["option2"] = *variant.Option2
	}
	if variant.Option3 != nil && *variant.Option3 != "" {
		metadata["option3"] = *variant.Option3
	}

	return scraper.Product{
		ID:          productID,
		Name:        productName,
		Description: shopifyProduct.BodyHTML,
		Price:       price,
		Currency:    currency,
		URL:         productURL,
		InStock:     variant.Available,
		Images:      variantImages,
		Metadata:    metadata,
	}, nil
}

func (p *ShopifyPlugin) ScrapeAllPages(ctx context.Context, req *scraper.ScrapeRequest) (*scraper.ScrapeResult, error) {
	allProducts := []scraper.Product{}
	allErrors := []string{}
	page := 1
	limit := 250

	if limitStr := req.Options["limit"]; limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	for {
		// Clone request and set page
		pageReq := *req
		if pageReq.Options == nil {
			pageReq.Options = make(map[string]string)
		}
		pageReq.Options["page"] = strconv.Itoa(page)
		pageReq.Options["limit"] = strconv.Itoa(limit)

		result, err := p.Scrape(ctx, &pageReq)
		if err != nil {
			return nil, fmt.Errorf("failed to scrape page %d: %w", page, err)
		}

		if len(result.Products) == 0 {
			break // No more products
		}

		allProducts = append(allProducts, result.Products...)
		allErrors = append(allErrors, result.Errors...)

		// If we got fewer products than the limit, we've reached the end
		if len(result.Products) < limit {
			break
		}

		page++
	}

	return &scraper.ScrapeResult{
		Products: allProducts,
		Errors:   allErrors,
	}, nil
}
