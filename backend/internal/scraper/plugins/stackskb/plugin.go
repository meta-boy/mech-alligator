package stackskb

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/meta-boy/mech-alligator/internal/scraper"
)

type Plugin struct {
	client *http.Client
}

func NewStacksKBPlugin() *Plugin {
	return &Plugin{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *Plugin) Name() string {
	return "stackskb"
}

func (p *Plugin) SupportedTypes() []string {
	return []string{"STACKS", "STACKSKB"}
}

func (p *Plugin) ValidateRequest(req *scraper.ScrapeRequest) error {
	if req.URL == "" {
		return fmt.Errorf("URL is required")
	}
	if !strings.Contains(req.URL, "stackskb.com") {
		return fmt.Errorf("URL must be from stackskb.com")
	}
	return nil
}

func (p *Plugin) Scrape(ctx context.Context, req *scraper.ScrapeRequest) (*scraper.ScrapeResult, error) {
	start := time.Now()

	// Determine if this is a product listing or product detail page
	if p.isProductDetailPage(req.URL) {
		return p.scrapeProductDetail(ctx, req, start)
	} else {
		return p.scrapeProductListing(ctx, req, start)
	}
}

func (p *Plugin) isProductDetailPage(url string) bool {
	// Listing pages contain /product-category/
	// Everything else is a product detail page
	return !strings.Contains(url, "/product-category/")
}

func (p *Plugin) scrapeProductListing(ctx context.Context, req *scraper.ScrapeRequest, start time.Time) (*scraper.ScrapeResult, error) {
	baseURL := p.getBaseCategoryURL(req.URL)
	allProducts := []scraper.ScrapedProduct{}
	var allErrors []string

	// Fetch first page to determine total pages
	firstPageHTML, err := p.fetchPage(ctx, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch first page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(firstPageHTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract products from first page
	firstPageProducts, pageErrors := p.extractProductsFromListing(doc, req)
	allProducts = append(allProducts, firstPageProducts...)
	allErrors = append(allErrors, pageErrors...)

	// Determine total pages
	totalPages := p.extractTotalPages(doc)

	// If there are more pages, scrape them
	if totalPages > 1 {
		for page := 2; page <= totalPages; page++ {
			// Add delay to be respectful
			time.Sleep(2 * time.Second)

			pageURL := fmt.Sprintf("%s/page/%d/", strings.TrimSuffix(baseURL, "/"), page)
			pageHTML, err := p.fetchPage(ctx, pageURL)
			if err != nil {
				allErrors = append(allErrors, fmt.Sprintf("page %d: %v", page, err))
				continue
			}

			pageDoc, err := goquery.NewDocumentFromReader(strings.NewReader(pageHTML))
			if err != nil {
				allErrors = append(allErrors, fmt.Sprintf("page %d: failed to parse HTML: %v", page, err))
				continue
			}

			pageProducts, pagePageErrors := p.extractProductsFromListing(pageDoc, req)
			allProducts = append(allProducts, pageProducts...)
			allErrors = append(allErrors, pagePageErrors...)
		}
	}

	return &scraper.ScrapeResult{
		Products: allProducts,
		Errors:   allErrors,
		Stats: scraper.ScrapeStats{
			ProductsFound: len(allProducts),
			VariantsFound: p.countVariants(allProducts),
			ErrorCount:    len(allErrors),
			Duration:      time.Since(start).String(),
			Source:        req.Reseller,
		},
	}, nil
}

func (p *Plugin) scrapeProductDetail(ctx context.Context, req *scraper.ScrapeRequest, start time.Time) (*scraper.ScrapeResult, error) {
	html, err := p.fetchPage(ctx, req.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch product page: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	product, err := p.extractProductDetail(doc, req)
	if err != nil {
		return nil, fmt.Errorf("failed to extract product: %w", err)
	}

	return &scraper.ScrapeResult{
		Products: []scraper.ScrapedProduct{*product},
		Errors:   []string{},
		Stats: scraper.ScrapeStats{
			ProductsFound: 1,
			VariantsFound: len(product.Variants),
			ErrorCount:    0,
			Duration:      time.Since(start).String(),
			Source:        req.Reseller,
		},
	}, nil
}

func (p *Plugin) fetchPage(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ProductScraper/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	html, err := doc.Html()
	if err != nil {
		return "", err
	}

	return html, nil
}

func (p *Plugin) getBaseCategoryURL(urlStr string) string {
	// Remove /page/X/ from URL if present
	re := regexp.MustCompile(`/page/\d+/?$`)
	baseURL := re.ReplaceAllString(urlStr, "/")
	return strings.TrimSuffix(baseURL, "/") + "/"
}

func (p *Plugin) extractTotalPages(doc *goquery.Document) int {
	totalPages := 1

	doc.Find("nav.woocommerce-pagination ul.page-numbers a.page-numbers").Each(func(i int, s *goquery.Selection) {
		pageText := strings.TrimSpace(s.Text())
		if pageNum, err := strconv.Atoi(pageText); err == nil {
			if pageNum > totalPages {
				totalPages = pageNum
			}
		}
	})

	return totalPages
}

func (p *Plugin) extractProductsFromListing(doc *goquery.Document, req *scraper.ScrapeRequest) ([]scraper.ScrapedProduct, []string) {
	var products []scraper.ScrapedProduct
	var errors []string

	doc.Find("li.product").Each(func(i int, s *goquery.Selection) {
		product, err := p.extractProductFromListingItem(s, req)
		if err != nil {
			errors = append(errors, fmt.Sprintf("product %d: %v", i, err))
			return
		}
		if product != nil {
			products = append(products, *product)
		}
	})

	return products, errors
}

func (p *Plugin) extractProductFromListingItem(s *goquery.Selection, req *scraper.ScrapeRequest) (*scraper.ScrapedProduct, error) {
	// Extract product name and URL
	titleLink := s.Find("h2.woocommerce-loop-product__title a")
	if titleLink.Length() == 0 {
		return nil, fmt.Errorf("no title link found")
	}

	name := strings.TrimSpace(titleLink.Text())
	productURL, exists := titleLink.Attr("href")
	if !exists {
		return nil, fmt.Errorf("no product URL found")
	}

	// Extract image
	var images []string
	imageLink := s.Find("a.woocommerce-loop-image-link img")
	if imageLink.Length() > 0 {
		if src, exists := imageLink.Attr("data-src"); exists {
			images = append(images, src)
		} else if src, exists := imageLink.Attr("src"); exists {
			images = append(images, src)
		}
	}

	// Extract price
	priceElement := s.Find("span.price")
	price, currency := p.extractPrice(priceElement)

	// Generate handle
	handle := p.generateHandle(name)

	// Extract categories from classes
	categories := p.extractCategoriesFromClasses(s)

	// Determine brand
	brand := p.determineBrand(name, categories)

	// Generate tags
	tags := p.generateTags(name, categories)

	// Create basic variant
	variants := []scraper.ScrapedVariant{
		{
			Name:      "Default",
			Price:     price,
			Currency:  currency,
			Available: true,
			URL:       productURL,
			Images:    images,
			Options:   make(map[string]string),
			SourceID:  "default",
		},
	}

	// Generate source ID from URL or name
	sourceID := p.generateSourceID(productURL, name)

	product := &scraper.ScrapedProduct{
		Name:        name,
		Description: "", // Listing pages don't have full descriptions
		Handle:      handle,
		URL:         productURL,
		Brand:       brand,
		Category:    req.Category,
		Tags:        tags,
		Images:      images,
		Variants:    variants,
		SourceType:  "STACKS",
		SourceID:    sourceID,
		Metadata: map[string]string{
			"listing_page": "true",
			"categories":   strings.Join(categories, ","),
		},
	}

	return product, nil
}

func (p *Plugin) extractProductDetail(doc *goquery.Document, req *scraper.ScrapeRequest) (*scraper.ScrapedProduct, error) {
	// Extract product title
	title := strings.TrimSpace(doc.Find("h1.product_title").Text())
	if title == "" {
		return nil, fmt.Errorf("no product title found")
	}

	// Extract description
	description := p.extractDescription(doc)

	// Extract images
	images := p.extractImages(doc)

	// Extract price
	priceElement := doc.Find("p.price")
	basePrice, currency := p.extractPrice(priceElement)

	// Extract brand
	brand := p.extractBrandFromAttributes(doc, title)

	// Extract categories and tags
	categories, tags := p.extractCategoriesAndTags(doc, title)

	// Extract variants
	variants := p.extractVariants(doc, basePrice, currency, images, req.URL)

	// Generate handle and source ID
	handle := p.generateHandle(title)
	sourceID := p.extractSourceID(doc, req.URL, title)

	product := &scraper.ScrapedProduct{
		Name:        title,
		Description: description,
		Handle:      handle,
		URL:         req.URL,
		Brand:       brand,
		Category:    req.Category,
		Tags:        tags,
		Images:      images,
		Variants:    variants,
		SourceType:  "STACKS",
		SourceID:    sourceID,
		Metadata: map[string]string{
			"categories":  strings.Join(categories, ","),
			"detail_page": "true",
		},
	}

	return product, nil
}

func (p *Plugin) extractPrice(priceElement *goquery.Selection) (float64, string) {
	currency := "INR" // Default currency

	// Look for sale price first (ins element)
	insElement := priceElement.Find("ins span.woocommerce-Price-amount")
	if insElement.Length() > 0 {
		return p.parsePrice(insElement.Text()), currency
	}

	// Regular price
	priceAmounts := priceElement.Find("span.woocommerce-Price-amount")
	if priceAmounts.Length() > 0 {
		// Take first price if multiple (min price in range)
		return p.parsePrice(priceAmounts.First().Text()), currency
	}

	return 0.0, currency
}

func (p *Plugin) parsePrice(priceText string) float64 {
	// Remove currency symbols and non-numeric characters
	re := regexp.MustCompile(`[^\d,.]`)
	cleanPrice := re.ReplaceAllString(priceText, "")
	cleanPrice = strings.ReplaceAll(cleanPrice, ",", "")

	if price, err := strconv.ParseFloat(cleanPrice, 64); err == nil {
		return price
	}
	return 0.0
}

func (p *Plugin) extractDescription(doc *goquery.Document) string {
	var descriptions []string

	// Short description
	shortDesc := doc.Find("div.woocommerce-product-details__short-description")
	if shortDesc.Length() > 0 {
		if html, err := shortDesc.Html(); err == nil {
			descriptions = append(descriptions, html)
		}
	}

	// Full description from tab
	descTab := doc.Find("div#tab-description")
	if descTab.Length() > 0 {
		// Remove heading if present
		descTab.Find("h2").Remove()
		if html, err := descTab.Html(); err == nil {
			descriptions = append(descriptions, html)
		}
	}

	return strings.Join(descriptions, "")
}

func (p *Plugin) extractImages(doc *goquery.Document) []string {
	var images []string
	seenImages := make(map[string]bool)

	doc.Find("div.woocommerce-product-gallery__wrapper div.woocommerce-product-gallery__image").Each(func(i int, s *goquery.Selection) {
		// Try to get full size image URL
		imgLink := s.Find("a")
		img := s.Find("img")

		var imageURL string
		if imgLink.Length() > 0 {
			if href, exists := imgLink.Attr("href"); exists {
				imageURL = href
			}
		}

		if imageURL == "" && img.Length() > 0 {
			if src, exists := img.Attr("data-src"); exists {
				imageURL = src
			} else if src, exists := img.Attr("src"); exists {
				imageURL = src
			}
		}

		if imageURL != "" && !seenImages[imageURL] {
			images = append(images, imageURL)
			seenImages[imageURL] = true
		}
	})

	return images
}

func (p *Plugin) extractBrandFromAttributes(doc *goquery.Document, title string) string {
	// Check product attributes table
	foundBrand := ""
	doc.Find("table.woocommerce-product-attributes tr").Each(func(i int, s *goquery.Selection) {
		label := strings.ToLower(strings.TrimSpace(s.Find("th.woocommerce-product-attributes-item__label").Text()))
		if strings.Contains(label, "manufacturer") || strings.Contains(label, "brand") {
			value := strings.TrimSpace(s.Find("td.woocommerce-product-attributes-item__value").Text())
			if value != "" {
				foundBrand = value
			}
		}
	})

	if foundBrand != "" {
		return foundBrand
	}

	// Fallback to brand detection from title
	return p.determineBrand(title, []string{})
}

func (p *Plugin) extractCategoriesAndTags(doc *goquery.Document, title string) ([]string, []string) {
	var categories []string

	// Extract from breadcrumbs
	doc.Find("nav.kadence-breadcrumbs div.kadence-breadcrumb-container a").Each(func(i int, s *goquery.Selection) {
		catName := strings.TrimSpace(s.Text())
		if catName != "Home" && catName != "Store" && catName != "" {
			categories = append(categories, strings.ToUpper(catName))
		}
	})

	// Extract from product meta
	doc.Find("div.product_meta span.posted_in a").Each(func(i int, s *goquery.Selection) {
		catName := strings.TrimSpace(s.Text())
		catUpper := strings.ToUpper(catName)

		// Check if already in categories
		found := false
		for _, existing := range categories {
			if existing == catUpper {
				found = true
				break
			}
		}
		if !found {
			categories = append(categories, catUpper)
		}
	})

	// Generate tags
	tags := p.generateTags(title, categories)

	return categories, tags
}

func (p *Plugin) extractVariants(doc *goquery.Document, basePrice float64, currency string, baseImages []string, productURL string) []scraper.ScrapedVariant {
	var variants []scraper.ScrapedVariant

	// Look for variations form
	variationsForm := doc.Find("form.variations_form")
	if variationsForm.Length() == 0 {
		// No variations, create default variant
		return []scraper.ScrapedVariant{
			{
				Name:      "Default",
				Price:     basePrice,
				Currency:  currency,
				Available: true,
				URL:       productURL,
				Images:    baseImages,
				Options:   make(map[string]string),
				SourceID:  "default",
			},
		}
	}

	// Try to extract variations data from data attribute
	if variationsData, exists := variationsForm.Attr("data-product_variations"); exists {
		// Unescape HTML entities
		variationsData = html.UnescapeString(variationsData)

		var variationsJSON []map[string]interface{}
		if err := json.Unmarshal([]byte(variationsData), &variationsJSON); err == nil {
			for _, varData := range variationsJSON {
				variant := p.createVariantFromJSON(varData, currency, baseImages, productURL)
				variants = append(variants, variant)
			}
		}
	}

	// If no variants extracted from JSON, try to extract from form elements
	if len(variants) == 0 {
		variants = p.extractVariantsFromForm(doc, basePrice, currency, baseImages, productURL)
	}

	// If still no variants, create default
	if len(variants) == 0 {
		variants = []scraper.ScrapedVariant{
			{
				Name:      "Default",
				Price:     basePrice,
				Currency:  currency,
				Available: true,
				URL:       productURL,
				Images:    baseImages,
				Options:   make(map[string]string),
				SourceID:  "default",
			},
		}
	}

	return variants
}

func (p *Plugin) createVariantFromJSON(varData map[string]interface{}, currency string, baseImages []string, productURL string) scraper.ScrapedVariant {
	variant := scraper.ScrapedVariant{
		Currency: currency,
		Images:   make([]string, len(baseImages)),
		URL:      productURL,
		Options:  make(map[string]string),
	}
	copy(variant.Images, baseImages)

	// Extract name from attributes
	if attributes, ok := varData["attributes"].(map[string]interface{}); ok {
		var nameParts []string
		for key, value := range attributes {
			// Clean up attribute key (remove attribute_ and pa_ prefixes)
			cleanKey := strings.ReplaceAll(key, "attribute_", "")
			cleanKey = strings.ReplaceAll(cleanKey, "pa_", "")

			if valueStr, ok := value.(string); ok && valueStr != "" {
				variant.Options[cleanKey] = valueStr
				nameParts = append(nameParts, valueStr)
			}
		}
		if len(nameParts) > 0 {
			variant.Name = strings.Join(nameParts, " ")
		} else {
			variant.Name = "Default"
		}
	} else {
		variant.Name = "Default"
	}

	// Extract availability
	if isInStock, ok := varData["is_in_stock"].(bool); ok {
		variant.Available = isInStock
	} else {
		variant.Available = true
	}

	// Extract price - handle both float64 and string types
	if displayPrice, ok := varData["display_price"].(float64); ok {
		variant.Price = displayPrice
	} else if displayPriceStr, ok := varData["display_price"].(string); ok {
		if price, err := strconv.ParseFloat(displayPriceStr, 64); err == nil {
			variant.Price = price
		} else {
			variant.Price = 0.0
		}
	} else {
		variant.Price = 0.0
	}

	// Extract source ID - handle both number and string types
	if variationID, ok := varData["variation_id"].(float64); ok {
		variant.SourceID = fmt.Sprintf("%.0f", variationID)
	} else if variationIDStr, ok := varData["variation_id"].(string); ok {
		variant.SourceID = variationIDStr
	} else {
		variant.SourceID = "unknown"
	}

	// Extract variant-specific image
	if imageData, ok := varData["image"].(map[string]interface{}); ok {
		if fullSrc, ok := imageData["full_src"].(string); ok && fullSrc != "" {
			variant.Images = []string{fullSrc}
		}
	}

	// Build variant URL with query parameter
	if variant.SourceID != "unknown" && variant.SourceID != "" && productURL != "" {
		variant.URL = fmt.Sprintf("%s?variant=%s", productURL, variant.SourceID)
	}

	return variant
}

func (p *Plugin) extractVariantsFromForm(doc *goquery.Document, basePrice float64, currency string, baseImages []string, productURL string) []scraper.ScrapedVariant {
	var variants []scraper.ScrapedVariant
	variationOptions := make(map[string][]map[string]string)

	// Find variation options from select elements in variations table
	doc.Find("table.variations tr").Each(func(i int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find("label").Text())
		selectElem := s.Find("select")

		if selectElem.Length() > 0 && label != "" {
			var options []map[string]string

			selectElem.Find("option").Each(func(j int, opt *goquery.Selection) {
				value, exists := opt.Attr("value")
				if !exists || value == "" {
					return
				}

				text := strings.TrimSpace(opt.Text())
				if text == "" || text == "Choose an option" {
					return
				}

				options = append(options, map[string]string{
					"value": value,
					"text":  text,
				})
			})

			if len(options) > 0 {
				variationOptions[label] = options
			}
		}
	})

	// Create all possible combinations of variants
	if len(variationOptions) > 0 {
		variants = p.generateVariantCombinations(variationOptions, basePrice, currency, baseImages, productURL)
	}

	return variants
}

func (p *Plugin) generateVariantCombinations(variationOptions map[string][]map[string]string, basePrice float64, currency string, baseImages []string, productURL string) []scraper.ScrapedVariant {
	var variants []scraper.ScrapedVariant

	// Convert map to slice for easier iteration
	var attributes []string
	var optionLists [][]map[string]string

	for attrName, options := range variationOptions {
		attributes = append(attributes, attrName)
		optionLists = append(optionLists, options)
	}

	if len(attributes) == 0 {
		return variants
	}

	// Generate all combinations
	combinations := p.generateCombinations(optionLists, 0, []map[string]string{})

	for _, combination := range combinations {
		var nameParts []string
		options := make(map[string]string)
		sourceIDParts := []string{}

		for i, option := range combination {
			attrName := attributes[i]
			nameParts = append(nameParts, option["text"])
			options[strings.ToLower(attrName)] = option["value"]
			sourceIDParts = append(sourceIDParts, option["value"])
		}

		variant := scraper.ScrapedVariant{
			Name:      strings.Join(nameParts, " - "),
			Price:     basePrice,
			Currency:  currency,
			Available: true,
			URL:       productURL,
			Images:    make([]string, len(baseImages)),
			Options:   options,
			SourceID:  strings.Join(sourceIDParts, "-"),
		}
		copy(variant.Images, baseImages)

		variants = append(variants, variant)
	}

	return variants
}

func (p *Plugin) generateCombinations(optionLists [][]map[string]string, index int, current []map[string]string) [][]map[string]string {
	if index == len(optionLists) {
		// Make a copy of current combination
		combination := make([]map[string]string, len(current))
		copy(combination, current)
		return [][]map[string]string{combination}
	}

	var results [][]map[string]string
	for _, option := range optionLists[index] {
		newCurrent := append(current, option)
		results = append(results, p.generateCombinations(optionLists, index+1, newCurrent)...)
	}

	return results
}

func (p *Plugin) generateHandle(productName string) string {
	if productName == "" {
		return "unknown-product"
	}

	// Convert to lowercase
	handle := strings.ToLower(productName)

	// Replace common separators and special characters with hyphens
	handle = regexp.MustCompile(`[^\p{L}\p{N}]+`).ReplaceAllString(handle, "-")

	// Remove leading/trailing hyphens
	handle = strings.Trim(handle, "-")

	// Collapse multiple consecutive hyphens
	handle = regexp.MustCompile(`-+`).ReplaceAllString(handle, "-")

	// Ensure it's not empty
	if handle == "" {
		handle = "unnamed-product"
	}

	return handle
}

func (p *Plugin) ensureMaxLength(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// For very long strings, use a combination of truncated name + hash
	if len(s) > maxLen {
		// Take first part of the string + hash of full string
		hashLen := 8                      // 8 character hash
		prefixLen := maxLen - hashLen - 1 // -1 for separator

		if prefixLen < 1 {
			// If maxLen is too small, just return hash
			return fmt.Sprintf("%x", md5.Sum([]byte(s)))[:maxLen]
		}

		// Truncate at word boundary if possible
		prefix := s[:prefixLen]
		if lastHyphen := strings.LastIndex(prefix, "-"); lastHyphen > prefixLen/2 {
			prefix = prefix[:lastHyphen]
		}
		prefix = strings.TrimSuffix(prefix, "-")

		// Create hash of full string
		hash := fmt.Sprintf("%x", md5.Sum([]byte(s)))[:hashLen]

		return prefix + "-" + hash
	}

	return s
}

// Alternative: Simple truncation approach
func (p *Plugin) generateSourceIDSimple(productURL, productName string) string {
	var baseID string

	// Try URL first
	if productURL != "" {
		if urlID := p.extractIDFromURL(productURL); urlID != "" {
			baseID = urlID
		}
	}

	// Fallback to name
	if baseID == "" {
		baseID = p.generateHandle(productName)
	}

	// Hard truncation to 45 characters
	if len(baseID) > 45 {
		baseID = baseID[:45]
		baseID = strings.TrimSuffix(baseID, "-")
	}

	// Ensure not empty
	if baseID == "" {
		baseID = "product"
	}

	return baseID
}

func (p *Plugin) extractCategoriesFromClasses(s *goquery.Selection) []string {
	var categories []string
	if classes, exists := s.Attr("class"); exists {
		for _, class := range strings.Fields(classes) {
			if strings.HasPrefix(class, "product_cat-") {
				category := strings.ReplaceAll(strings.ReplaceAll(class, "product_cat-", ""), "-", " ")
				categories = append(categories, strings.Title(category))
			}
		}
	}
	return categories
}

func (p *Plugin) determineBrand(title string, categories []string) string {
	titleLower := strings.ToLower(title)
	knownBrands := []string{
		"epbt", "gmk", "sa", "cherry", "gateron", "kailh",
		"akko", "keychron", "drop", "wuque studio",
	}

	for _, brand := range knownBrands {
		if strings.Contains(titleLower, brand) {
			return strings.Title(brand)
		}
	}

	return "StacksKB"
}

func (p *Plugin) generateTags(title string, categories []string) []string {
	var tags []string
	seenTags := make(map[string]bool)

	// Add category-based tags
	for _, cat := range categories {
		catLower := strings.ToLower(cat)
		if strings.Contains(catLower, " ") {
			catLower = strings.ReplaceAll(catLower, " ", "-")
		}
		if !seenTags[catLower] {
			tags = append(tags, catLower)
			seenTags[catLower] = true
		}
	}

	// Add title-based tags
	re := regexp.MustCompile(`\b\w+\b`)
	titleWords := re.FindAllString(strings.ToLower(title), -1)
	excludeWords := map[string]bool{
		"the": true, "and": true, "for": true, "with": true, "from": true,
	}

	for _, word := range titleWords {
		if len(word) > 2 && !excludeWords[word] && !seenTags[word] {
			tags = append(tags, word)
			seenTags[word] = true
		}
	}

	return tags
}

func (p *Plugin) generateSourceID(productURL, productName string) string {
	var baseID string

	// Try to extract from URL first
	if productURL != "" {
		if urlID := p.extractIDFromURL(productURL); urlID != "" {
			baseID = urlID
		}
	}

	// Fallback to generating from product name
	if baseID == "" {
		baseID = p.generateHandle(productName)
	}

	// Ensure we never exceed 50 characters
	return p.ensureMaxLength(baseID, 45) // Leave 5 chars buffer
}

func (p *Plugin) extractIDFromURL(productURL string) string {
	u, err := url.Parse(productURL)
	if err != nil {
		return ""
	}

	path := strings.Trim(u.Path, "/")
	if path == "" {
		return ""
	}

	pathParts := strings.Split(path, "/")

	// Handle different URL patterns:
	// /products/product-handle
	// /collections/collection-name/products/product-handle
	// /product/product-handle
	for i, part := range pathParts {
		if (part == "products" || part == "product") && i+1 < len(pathParts) {
			return pathParts[i+1]
		}
	}

	// If no products path found, use the last segment
	if len(pathParts) > 0 {
		lastPart := pathParts[len(pathParts)-1]
		// Skip common file extensions and query parameters
		if !strings.Contains(lastPart, ".") && lastPart != "" {
			return lastPart
		}
	}

	return ""
}

func (p *Plugin) extractSourceID(doc *goquery.Document, productURL, title string) string {
	// Try to extract SKU from product meta
	sku := strings.TrimSpace(doc.Find("div.product_meta span.sku_wrapper span.sku").Text())
	if sku != "" {
		return sku
	}

	// Fallback to URL-based ID
	return p.generateSourceID(productURL, title)
}

func (p *Plugin) countVariants(products []scraper.ScrapedProduct) int {
	count := 0
	for _, product := range products {
		count += len(product.Variants)
	}
	return count
}
