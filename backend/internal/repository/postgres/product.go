package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/meta-boy/mech-alligator/internal/database"
	"github.com/meta-boy/mech-alligator/internal/domain/product"
)

type ProductRepository struct {
	db *database.DB
}

func NewProductRepository(db *database.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetByID(ctx context.Context, id string) (*product.Product, error) {
	query := `
		SELECT 
			p.id, p.name, p.description, p.handle, p.url, p.brand, p.reseller,
			p.reseller_id, p.category, p.tags, p.images, p.source_type, 
			p.source_id, p.source_metadata
		FROM products p
		WHERE p.id = $1
	`

	var p product.Product
	var tagsArray, imagesArray pq.StringArray
	var sourceMetadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Handle, &p.URL, &p.Brand, &p.Reseller,
		&p.ResellerID, &p.Category, &tagsArray, &imagesArray, &p.SourceType,
		&p.SourceID, &sourceMetadataJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Convert arrays
	p.Tags = []string(tagsArray)
	p.Images = []string(imagesArray)

	// Parse source metadata
	if len(sourceMetadataJSON) > 0 {
		if err := json.Unmarshal(sourceMetadataJSON, &p.SourceMetadata); err != nil {
			p.SourceMetadata = make(map[string]string)
		}
	}

	// Load variants
	variants, err := r.getProductVariants(ctx, p.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product variants: %w", err)
	}
	p.Variants = variants
	p.VariantCount = len(variants)

	return &p, nil
}

func (r *ProductRepository) List(ctx context.Context, req product.ListRequest) ([]product.Product, int64, error) {
	// Build WHERE clause
	whereClause, args := r.buildWhereClause(req)

	// Count total items
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT p.id) 
		FROM products p 
		LEFT JOIN product_variants pv ON p.id = pv.product_id
		%s`, whereClause)

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Build main query
	orderClause := r.buildOrderClause(req.SortBy, req.SortOrder)
	offset := (req.Page - 1) * req.PageSize

	query := fmt.Sprintf(`
		SELECT DISTINCT
			p.id, p.name, p.description, p.handle, p.url, p.brand, p.reseller,
			p.reseller_id, p.category, p.tags, p.images, p.source_type, 
			p.source_id, p.source_metadata,
			(SELECT COUNT(*) FROM product_variants WHERE product_id = p.id) as variant_count
		FROM products p
		LEFT JOIN product_variants pv ON p.id = pv.product_id
		%s
		%s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderClause, len(args)+1, len(args)+2)

	// Add pagination args
	args = append(args, req.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []product.Product
	for rows.Next() {
		var p product.Product
		var tagsArray, imagesArray pq.StringArray
		var sourceMetadataJSON []byte

		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Handle, &p.URL, &p.Brand, &p.Reseller,
			&p.ResellerID, &p.Category, &tagsArray, &imagesArray, &p.SourceType,
			&p.SourceID, &sourceMetadataJSON, &p.VariantCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		// Convert arrays
		p.Tags = []string(tagsArray)
		p.Images = []string(imagesArray)

		// Parse source metadata
		if len(sourceMetadataJSON) > 0 {
			if err := json.Unmarshal(sourceMetadataJSON, &p.SourceMetadata); err != nil {
				p.SourceMetadata = make(map[string]string)
			}
		}

		products = append(products, p)
	}

	return products, total, rows.Err()
}

func (r *ProductRepository) Save(ctx context.Context, p *product.Product) error {
	tx, err := r.db.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if product exists by source
	existingID, err := r.findExistingProduct(ctx, tx, p.SourceType, p.SourceID, p.ResellerID)
	if err != nil {
		return fmt.Errorf("failed to check existing product: %w", err)
	}

	if existingID != "" {
		// Update existing product
		p.ID = existingID
		err = r.updateProduct(ctx, tx, p)
	} else {
		// Insert new product
		err = r.insertProduct(ctx, tx, p)
	}

	if err != nil {
		return err
	}

	// Save variants
	if err := r.saveVariants(ctx, tx, p.ID, p.Variants); err != nil {
		return fmt.Errorf("failed to save variants: %w", err)
	}

	return tx.Commit()
}

func (r *ProductRepository) insertProduct(ctx context.Context, tx *sql.Tx, p *product.Product) error {
	sourceMetadataJSON, _ := json.Marshal(p.SourceMetadata)

	query := `
		INSERT INTO products (
			name, description, handle, url, brand, reseller, reseller_id,
			category, tags, images, source_type, source_id, source_metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	err := tx.QueryRowContext(ctx, query,
		p.Name, p.Description, p.Handle, p.URL, p.Brand, p.Reseller, p.ResellerID,
		p.Category, pq.Array(p.Tags), pq.Array(p.Images), p.SourceType, p.SourceID, sourceMetadataJSON,
	).Scan(&p.ID)

	return err
}

func (r *ProductRepository) updateProduct(ctx context.Context, tx *sql.Tx, p *product.Product) error {
	sourceMetadataJSON, _ := json.Marshal(p.SourceMetadata)

	query := `
		UPDATE products SET
			name = $2, description = $3, handle = $4, url = $5, brand = $6, 
			reseller = $7, category = $8, tags = $9, images = $10, 
			source_metadata = $11
		WHERE id = $1
	`

	_, err := tx.ExecContext(ctx, query,
		p.ID, p.Name, p.Description, p.Handle, p.URL, p.Brand,
		p.Reseller, p.Category, pq.Array(p.Tags), pq.Array(p.Images), sourceMetadataJSON,
	)

	return err
}

func (r *ProductRepository) saveVariants(ctx context.Context, tx *sql.Tx, productID string, variants []product.Variant) error {
	// Delete existing variants
	_, err := tx.ExecContext(ctx, "DELETE FROM product_variants WHERE product_id = $1", productID)
	if err != nil {
		return fmt.Errorf("failed to delete existing variants: %w", err)
	}

	// Insert new variants
	for i := range variants {
		variant := &variants[i]
		optionsJSON, _ := json.Marshal(variant.Options)

		query := `
			INSERT INTO product_variants (
				product_id, name, sku, price, currency, available, 
				url, images, options, source_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
		`

		err := tx.QueryRowContext(ctx, query,
			productID, variant.Name, variant.SKU, variant.Price,
			variant.Currency, variant.Available, variant.URL, pq.Array(variant.Images),
			optionsJSON, variant.SourceID,
		).Scan(&variant.ID)

		if err != nil {
			return fmt.Errorf("failed to insert variant: %w", err)
		}
	}

	return nil
}

func (r *ProductRepository) getProductVariants(ctx context.Context, productID string) ([]product.Variant, error) {
	query := `
		SELECT id, name, sku, price, currency, available, url, images, options, source_id
		FROM product_variants
		WHERE product_id = $1
		ORDER BY price ASC
	`

	rows, err := r.db.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []product.Variant
	for rows.Next() {
		var v product.Variant
		var imagesArray pq.StringArray
		var optionsJSON []byte

		err := rows.Scan(
			&v.ID, &v.Name, &v.SKU, &v.Price, &v.Currency, &v.Available,
			&v.URL, &imagesArray, &optionsJSON, &v.SourceID,
		)
		if err != nil {
			return nil, err
		}

		v.Images = []string(imagesArray)
		v.ProductID = productID

		// Parse options JSON
		if len(optionsJSON) > 0 {
			if err := json.Unmarshal(optionsJSON, &v.Options); err != nil {
				v.Options = make(map[string]string)
			}
		}

		variants = append(variants, v)
	}

	return variants, rows.Err()
}

func (r *ProductRepository) buildWhereClause(req product.ListRequest) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if req.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+req.Search+"%")
		argIndex++
	}

	if req.Brand != "" {
		conditions = append(conditions, fmt.Sprintf("p.brand ILIKE $%d", argIndex))
		args = append(args, "%"+req.Brand+"%")
		argIndex++
	}

	if req.Reseller != "" {
		conditions = append(conditions, fmt.Sprintf("p.reseller ILIKE $%d", argIndex))
		args = append(args, "%"+req.Reseller+"%")
		argIndex++
	}

	if req.Category != "" {
		conditions = append(conditions, fmt.Sprintf("p.category = $%d", argIndex))
		args = append(args, req.Category)
		argIndex++
	}

	if req.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("pv.price >= $%d", argIndex))
		args = append(args, *req.MinPrice)
		argIndex++
	}

	if req.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("pv.price <= $%d", argIndex))
		args = append(args, *req.MaxPrice)
		argIndex++
	}

	if req.Available != nil {
		conditions = append(conditions, fmt.Sprintf("pv.available = $%d", argIndex))
		args = append(args, *req.Available)
		argIndex++
	}

	if len(req.Tags) > 0 {
		conditions = append(conditions, fmt.Sprintf("p.tags && $%d", argIndex))
		args = append(args, pq.Array(req.Tags))
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

func (r *ProductRepository) buildOrderClause(sortBy, sortOrder string) string {
	// Validate sort fields
	validSortFields := map[string]string{
		"name":     "p.name",
		"brand":    "p.brand",
		"reseller": "p.reseller",
		"price":    "MIN(pv.price)",
	}

	field, ok := validSortFields[sortBy]
	if !ok {
		field = "p.name"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "asc"
	}

	if sortBy == "price" {
		return fmt.Sprintf("GROUP BY p.id, p.name, p.description, p.handle, p.url, p.brand, p.reseller, p.reseller_id, p.category, p.tags, p.images, p.source_type, p.source_id, p.source_metadata ORDER BY %s %s", field, sortOrder)
	}

	return fmt.Sprintf("ORDER BY %s %s", field, sortOrder)
}

func (r *ProductRepository) findExistingProduct(ctx context.Context, tx *sql.Tx, sourceType, sourceID, resellerID string) (string, error) {
	query := `SELECT id FROM products WHERE source_type = $1 AND source_id = $2 AND reseller_id = $3 LIMIT 1`

	var id string
	err := tx.QueryRowContext(ctx, query, sourceType, sourceID, resellerID).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	return id, nil
}

// Filter helper methods
func (r *ProductRepository) GetDistinctBrands(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT brand FROM products WHERE brand IS NOT NULL AND brand != '' ORDER BY brand`
	return r.getDistinctValues(ctx, query)
}

func (r *ProductRepository) GetDistinctResellers(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT reseller FROM products WHERE reseller IS NOT NULL AND reseller != '' ORDER BY reseller`
	return r.getDistinctValues(ctx, query)
}

func (r *ProductRepository) GetDistinctCategories(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT category FROM products WHERE category IS NOT NULL AND category != '' ORDER BY category`
	return r.getDistinctValues(ctx, query)
}

func (r *ProductRepository) getDistinctValues(ctx context.Context, query string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		values = append(values, value)
	}

	return values, rows.Err()
}
