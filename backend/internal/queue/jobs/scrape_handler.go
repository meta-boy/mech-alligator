package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/meta-boy/mech-alligator/internal/config"
	"github.com/meta-boy/mech-alligator/internal/database"
	"github.com/meta-boy/mech-alligator/internal/domain/job"
	"github.com/meta-boy/mech-alligator/internal/domain/product"
	"github.com/meta-boy/mech-alligator/internal/repository/postgres"
	"github.com/meta-boy/mech-alligator/internal/scraper"
	"github.com/meta-boy/mech-alligator/internal/scraper/plugins/shopify"
	"github.com/meta-boy/mech-alligator/internal/scraper/plugins/stackskb"
)

type ScrapeJobHandler struct {
	db          *database.DB
	manager     *scraper.Manager
	productRepo *postgres.ProductRepository
}

func NewScrapeJobHandler(db *database.DB, productRepo *postgres.ProductRepository) *ScrapeJobHandler {
	// Initialize scraper manager with plugins
	manager := scraper.NewManager()

	// Register plugins
	if err := manager.RegisterPlugin(shopify.NewShopifyPlugin()); err != nil {
		log.Printf("Warning: Failed to register Shopify plugin: %v", err)
	}

	if err := manager.RegisterPlugin(stackskb.NewStacksKBPlugin()); err != nil {
		log.Printf("Warning: Failed to register StacksKB plugin: %v", err)
	}

	return &ScrapeJobHandler{
		db:          db,
		manager:     manager,
		productRepo: productRepo,
	}
}

func (h *ScrapeJobHandler) GetType() job.JobType {
	return job.JobTypeScrapeProducts
}

func (h *ScrapeJobHandler) Handle(ctx context.Context, j *job.Job) error {
	log.Printf("Processing scrape job %s", j.ID)

	// Parse payload
	var payload config.ScrapeJobPayload
	payloadBytes, err := json.Marshal(j.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Create scrape request
	scrapeReq := &scraper.ScrapeRequest{
		URL:        payload.URL,
		SourceType: payload.SourceType,
		Reseller:   payload.ResellerName,
		ResellerID: payload.ResellerID,
		Category:   payload.Category,
		Options:    payload.Options,
	}

	start := time.Now()
	log.Printf("Starting scrape of %s (%s)", payload.URL, payload.SourceType)

	// Perform scraping using the manager (will auto-select the right plugin)
	result, err := h.manager.ScrapeByType(ctx, scrapeReq)
	if err != nil {
		return fmt.Errorf("scraping failed: %w", err)
	}

	log.Printf("Scraped %d products with %d total variants from %s",
		len(result.Products), result.Stats.VariantsFound, payload.ResellerName)

	// Convert and save products
	saveStats, saveErrors := h.saveProducts(ctx, result.Products, payload)

	// Create job result
	jobResult := ScrapeJobResult{
		ProductsCreated: saveStats.Created,
		ProductsUpdated: saveStats.Updated,
		VariantsTotal:   result.Stats.VariantsFound,
		TotalErrors:     len(result.Errors) + len(saveErrors),
		ScrapeErrors:    result.Errors,
		SaveErrors:      saveErrors,
		Duration:        time.Since(start).String(),
		ScrapedAt:       start.Format(time.RFC3339),
		Source:          payload.ResellerName,
		Category:        payload.Category,
	}

	// Convert result to map for storage
	resultBytes, err := json.Marshal(jobResult)
	if err != nil {
		return fmt.Errorf("failed to marshal job result: %w", err)
	}

	var resultMap map[string]interface{}
	if err := json.Unmarshal(resultBytes, &resultMap); err != nil {
		return fmt.Errorf("failed to unmarshal job result: %w", err)
	}

	j.Result = resultMap

	log.Printf("Job %s completed: %d created, %d updated, %d total errors",
		j.ID, saveStats.Created, saveStats.Updated, jobResult.TotalErrors)

	// print errors
	if len(jobResult.ScrapeErrors) > 0 {
		log.Printf("Scrape errors: %v", jobResult.ScrapeErrors)
	}
	if len(jobResult.SaveErrors) > 0 {
		log.Printf("Save errors: %v", jobResult.SaveErrors)
	}

	return nil
}

func (h *ScrapeJobHandler) saveProducts(ctx context.Context, scrapedProducts []scraper.ScrapedProduct, payload config.ScrapeJobPayload) (*SaveStats, []string) {
	stats := &SaveStats{}
	var errors []string

	for _, sp := range scrapedProducts {
		// Convert scraped product to domain product
		domainProduct := h.convertToProduct(sp, payload)

		// Save to database
		err := h.productRepo.Save(ctx, domainProduct)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to save product '%s': %v", sp.Name, err))
			continue
		}

		// For simplicity, count all as created (repository handles update logic internally)
		stats.Created++

		// Update brand if it's new (optional: maintain brand registry)
		if sp.Brand != "" && sp.Brand != "Unknown" {
			h.updateBrandRegistry(ctx, sp.Brand)
		}
	}

	return stats, errors
}

func (h *ScrapeJobHandler) convertToProduct(sp scraper.ScrapedProduct, payload config.ScrapeJobPayload) *product.Product {
	// Convert scraped product to domain product
	domainProduct := &product.Product{
		Name:           sp.Name,
		Description:    sp.Description,
		Handle:         sp.Handle,
		URL:            sp.URL,
		Brand:          sp.Brand,
		Reseller:       payload.ResellerName,
		ResellerID:     payload.ResellerID,
		Category:       payload.Category,
		Tags:           sp.Tags,
		Images:         sp.Images,
		SourceType:     sp.SourceType,
		SourceID:       sp.SourceID,
		SourceMetadata: sp.Metadata,
	}

	// Convert variants
	for _, sv := range sp.Variants {
		variant := product.Variant{
			Name:      sv.Name,
			SKU:       sv.SKU,
			Price:     sv.Price,
			Currency:  sv.Currency,
			Available: sv.Available,
			URL:       sv.URL,
			Images:    sv.Images,
			Options:   sv.Options,
			SourceID:  sv.SourceID,
		}
		domainProduct.Variants = append(domainProduct.Variants, variant)
	}

	domainProduct.VariantCount = len(domainProduct.Variants)

	return domainProduct
}

func (h *ScrapeJobHandler) updateBrandRegistry(ctx context.Context, brandName string) {
	// Optional: Keep a registry of brands for analytics/filtering
	// This is a simple insert-ignore operation using UUID generation
	query := `
		INSERT INTO brands (id, name) 
		VALUES (gen_random_uuid(), $1) 
		ON CONFLICT (name) DO NOTHING
	`

	_, err := h.db.ExecContext(ctx, query, brandName)
	if err != nil {
		log.Printf("Warning: Failed to update brand registry for '%s': %v", brandName, err)
	}
}

func (h *ScrapeJobHandler) generateBrandID(name string) string {
	// Simple ID generation from name
	// In production, you might want more sophisticated ID generation
	return fmt.Sprintf("brand_%s",
		strings.ToLower(
			strings.ReplaceAll(
				strings.ReplaceAll(name, " ", "_"),
				".", "")))
}

// Job result structure
type ScrapeJobResult struct {
	ProductsCreated int      `json:"products_created"`
	ProductsUpdated int      `json:"products_updated"`
	VariantsTotal   int      `json:"variants_total"`
	TotalErrors     int      `json:"total_errors"`
	ScrapeErrors    []string `json:"scrape_errors,omitempty"`
	SaveErrors      []string `json:"save_errors,omitempty"`
	Duration        string   `json:"duration"`
	ScrapedAt       string   `json:"scraped_at"`
	Source          string   `json:"source"`
	Category        string   `json:"category"`
}

type SaveStats struct {
	Created int
	Updated int
	Errors  int
}

type ScrapeAllSitesHandler struct{}

func NewScrapeAllSitesHandler() *ScrapeAllSitesHandler {
	return &ScrapeAllSitesHandler{}
}

func (h *ScrapeAllSitesHandler) GetType() job.JobType {
	return job.JobTypeScrapeAllSites
}

func (h *ScrapeAllSitesHandler) Handle(ctx context.Context, j *job.Job) error {
	// This job type is typically just a coordinator that creates individual scrape jobs
	// The actual work is done by ScrapeJobHandler for each individual site
	log.Printf("ScrapeAllSites job %s completed - individual scrape jobs were already created", j.ID)
	return nil
}
