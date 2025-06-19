package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/oklog/ulid/v2"
	"log"
	"math/rand"
	"time"

	"github.com/meta-boy/mech-alligator/internal/database"
	"github.com/meta-boy/mech-alligator/internal/domain/job"
	"github.com/meta-boy/mech-alligator/internal/scraper"
)

type ScrapeJobHandler struct {
	db      *database.DB
	manager *scraper.Manager
	queue   job.Queue
}

func NewScrapeJobHandler(db *database.DB, manager *scraper.Manager, queue job.Queue) *ScrapeJobHandler {
	return &ScrapeJobHandler{
		db:      db,
		manager: manager,
		queue:   queue,
	}
}

func (h *ScrapeJobHandler) Handle(ctx context.Context, j *job.Job) error {
	log.Printf("Processing scrape job %s", j.ID)

	// Parse payload
	var payload job.ScrapeJobPayload
	payloadBytes, err := json.Marshal(j.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	// Create scrape request
	scrapeReq := &scraper.ScrapeRequest{
		ConfigID:    payload.ConfigID,
		VendorID:    payload.VendorID,
		VendorName:  payload.VendorName,
		SiteURL:     payload.SiteURL,
		SiteType:    payload.SiteType,
		Category:    payload.Category,
		Credentials: payload.Credentials,
		Options:     payload.Options,
	}

	start := time.Now()

	// Perform scraping
	var result *scraper.ScrapeResult
	if payload.AllPages {
		// For plugins that support multi-page scraping
		result, err = h.manager.ScrapeByType(ctx, scrapeReq)
	} else {
		result, err = h.manager.ScrapeByType(ctx, scrapeReq)
	}

	if err != nil {
		return fmt.Errorf("scraping failed: %w", err)
	}

	log.Printf("Scraped %d products from %s", len(result.Products), payload.SiteURL)

	// Save products to database
	productsCreated, productsUpdated, imagesProcessed, errors := h.saveProductsToDB(ctx, result.Products, payload)

	// Create job result
	jobResult := job.ScrapeJobResult{
		ProductsCreated: productsCreated,
		ProductsUpdated: productsUpdated,
		ImagesProcessed: imagesProcessed,
		TotalErrors:     len(errors),
		Errors:          errors,
		Duration:        time.Since(start).String(),
		ScrapedAt:       start.Format(time.RFC3339),
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

	log.Printf("Job %s completed: %d created, %d updated, %d images, %d errors",
		j.ID, productsCreated, productsUpdated, imagesProcessed, len(errors))

	return nil
}

func (h *ScrapeJobHandler) GetType() job.JobType {
	return job.JobTypeScrapeProducts
}

func (h *ScrapeJobHandler) saveProductsToDB(ctx context.Context, products []scraper.Product, payload job.ScrapeJobPayload) (int, int, int, []string) {
	var productsCreated, productsUpdated, imagesProcessed int
	var errors []string

	for _, product := range products {
		productID, err := h.saveProduct(ctx, product, payload)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to save product %s: %v", product.Name, err))
			continue
		}

		err = h.enqueueTagJob(ctx, productID)
		if err != nil {
			errors = append(errors, fmt.Sprintf("failed to enqueue tag job for product %s: %v", productID, err))
		}

		// Check if product was created or updated (simplified logic)
		// In a real implementation, you'd track this properly
		productsCreated++

		// Save product images
		imageCount, imageErrors := h.saveProductImages(ctx, productID, product.Images)
		imagesProcessed += imageCount
		errors = append(errors, imageErrors...)

	}

	return productsCreated, productsUpdated, imagesProcessed, errors
}

func (h *ScrapeJobHandler) saveProduct(ctx context.Context, product scraper.Product, payload job.ScrapeJobPayload) (string, error) {
	// Check if product already exists
	query := `
		SELECT id FROM products 
		WHERE name = $1 AND url = $2
		LIMIT 1
	`

	var existingID string
	err := h.db.QueryRowContext(ctx, query, product.Name, product.URL).Scan(&existingID)

	if err != nil && err.Error() != "sql: no rows in result set" {
		return "", fmt.Errorf("failed to check existing product: %w", err)
	}

	if existingID != "" {
		// Update existing product
		updateQuery := `
			UPDATE products 
			SET description = $1, price = $2, currency = $3, in_stock = $4, updated_at = CURRENT_TIMESTAMP
			WHERE id = $5
		`
		_, err = h.db.ExecContext(ctx, updateQuery,
			product.Description, product.Price, product.Currency, product.InStock, existingID)
		return existingID, err
	}

	// Create new product
	insertQuery := `
		INSERT INTO products (id, name, description, price, currency, url, config_id, in_stock, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`

	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano()))).String()[:15]
	_, err = h.db.ExecContext(ctx, insertQuery,
		id, product.Name, product.Description, product.Price,
		product.Currency, product.URL, payload.ConfigID, product.InStock)

	return id, err
}

func (h *ScrapeJobHandler) saveProductImages(ctx context.Context, productID string, images []string) (int, []string) {
	var imageCount int
	var errors []string

	if len(images) == 0 {
		return 0, nil
	}

	// Generate all image UUIDs for this product
	imageUUIDs := make([]string, 0, len(images))
	imageURLMap := make(map[string]string) // uuid -> url
	for _, imageURL := range images {
		if imageURL == "" {
			continue
		}
		imageID := h.generateImageID(productID, imageURL)
		imageUUIDs = append(imageUUIDs, imageID)
		imageURLMap[imageID] = imageURL
	}

	if len(imageUUIDs) == 0 {
		return 0, nil
	}

	// Fetch all existing image IDs in one query
	query := `SELECT uuid, id FROM product_images WHERE uuid = ANY($1)`
	rows, err := h.db.QueryContext(ctx, query, imageUUIDs)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to fetch existing images: %v", err))
		return 0, errors
	}
	defer rows.Close()

	existingImages := make(map[string]string) // uuid -> id
	for rows.Next() {
		var uuid, id string
		if err := rows.Scan(&uuid, &id); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to scan image row: %v", err))
			continue
		}
		existingImages[uuid] = id
	}

	for _, imageUUID := range imageUUIDs {
		imageURL := imageURLMap[imageUUID]
		if existingID, ok := existingImages[imageUUID]; ok {
			// Update existing image
			updateQuery := `
				UPDATE product_images 
				SET product_id = $1, url = $2, updated_at = CURRENT_TIMESTAMP
				WHERE id = $3
			`
			_, err = h.db.ExecContext(ctx, updateQuery, productID, imageURL, existingID)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to update image: %v", err))
				continue
			}
			imageCount++
		} else {
			// Insert new image
			insertQuery := `
				INSERT INTO product_images (uuid, product_id, url, created_at, updated_at)
				VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
			`
			_, err = h.db.ExecContext(ctx, insertQuery, imageUUID, productID, imageURL)
			if err != nil {
				errors = append(errors, fmt.Sprintf("Failed to insert image: %v", err))
				continue
			}
			imageCount++
		}
	}

	return imageCount, errors
}

func (h *ScrapeJobHandler) enqueueTagJob(ctx context.Context, productID string) error {
	payload := map[string]interface{}{
		"product_id": productID,
	}

	j := &job.Job{
		ID:          ulid.MustNew(ulid.Timestamp(time.Now()), rand.New(rand.NewSource(time.Now().UnixNano()))).String(),
		Type:        job.JobTypeTagProduct,
		Priority:    job.PriorityNormal,
		Payload:     payload,
		MaxAttempts: 3,
		ScheduledAt: time.Now(),
	}

	return h.queue.Enqueue(ctx, j)
}

func (h *ScrapeJobHandler) generateImageID(productID, imageURL string) string {
	// Simple ID generation - in production, you might want to use UUID
	return fmt.Sprintf("%s-%d", productID, len(imageURL))
}
