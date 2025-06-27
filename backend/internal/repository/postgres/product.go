package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
			p.id, p.name, p.description, p.price, p.currency, p.url, 
			p.config_id, p.in_stock, p.created_at, p.updated_at,
			COALESCE(v.name, '') as vendor,
			COALESCE(
				JSON_AGG(
					DISTINCT pi.url
				) FILTER (WHERE pi.url IS NOT NULL),
				'[]'::json
			) AS image_urls,
			COALESCE(
				JSON_AGG(
					DISTINCT pt.tag
				) FILTER (WHERE pt.tag IS NOT NULL), 
				'[]'::json
			) AS tags
		FROM products p
		LEFT JOIN site_configurations sc ON p.config_id = sc.id
		LEFT JOIN vendors v ON sc.vendor_id = v.id
		LEFT JOIN product_images pi ON p.id = pi.product_id
		LEFT JOIN product_tags pt ON p.id = pt.product_id
		WHERE p.id = $1
		GROUP BY p.id, p.name, p.description, p.price, p.currency, 
				 p.url, p.config_id, p.in_stock, p.created_at, p.updated_at, v.name
	`

	var p product.Product
	var imageURLsJSON, tagsJSON string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Description, &p.Price, &p.Currency, &p.URL,
		&p.ConfigID, &p.InStock, &p.CreatedAt, &p.UpdatedAt,
		&p.Vendor, &imageURLsJSON, &tagsJSON,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Parse JSON arrays
	if err := json.Unmarshal([]byte(imageURLsJSON), &p.ImageURLs); err != nil {
		p.ImageURLs = []string{}
	}
	if err := json.Unmarshal([]byte(tagsJSON), &p.Tags); err != nil {
		p.Tags = []string{}
	}

	return &p, nil
}

func (r *ProductRepository) List(ctx context.Context, req product.ProductListRequest) ([]product.Product, int64, error) {
	// Build WHERE clause and args
	whereClause, args := r.buildWhereClause(req.Filter)

	// Count total items
	countQuery := "SELECT COUNT(DISTINCT p.id) FROM products p LEFT JOIN site_configurations sc ON p.config_id = sc.id LEFT JOIN vendors v ON sc.vendor_id = v.id LEFT JOIN product_images pi ON p.id = pi.product_id LEFT JOIN product_tags pt ON p.id = pt.product_id"
	if whereClause != "" {
		countQuery += " " + whereClause
	}

	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	// Build main query
	query := `SELECT 
		p.id, p.name, p.description, p.price, p.currency, p.url, 
		p.config_id, p.in_stock, p.created_at, p.updated_at,
		COALESCE(v.name, '') as vendor,
		COALESCE(
			JSON_AGG(
				DISTINCT pi.url
			) FILTER (WHERE pi.url IS NOT NULL),
			'[]'::json
		) AS image_urls,
		COALESCE(
			JSON_AGG(
				DISTINCT pt.tag
			) FILTER (WHERE pt.tag IS NOT NULL), 
			'[]'::json
		) AS tags
	FROM products p
	LEFT JOIN site_configurations sc ON p.config_id = sc.id
	LEFT JOIN vendors v ON sc.vendor_id = v.id
	LEFT JOIN product_images pi ON p.id = pi.product_id
	LEFT JOIN product_tags pt ON p.id = pt.product_id`

	if whereClause != "" {
		query += " " + whereClause
	}

	query += ` GROUP BY p.id, p.name, p.description, p.price, p.currency, 
		p.url, p.config_id, p.in_stock, p.created_at, p.updated_at, v.name`

	query += fmt.Sprintf(" ORDER BY p.%s %s", req.Sort.Field, req.Sort.Order)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)

	// Add pagination args
	args = append(args, req.Pagination.PageSize, req.Pagination.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []product.Product
	for rows.Next() {
		var p product.Product
		var imageURLsJSON, tagsJSON string

		err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.Price, &p.Currency, &p.URL,
			&p.ConfigID, &p.InStock, &p.CreatedAt, &p.UpdatedAt,
			&p.Vendor, &imageURLsJSON, &tagsJSON,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan product: %w", err)
		}

		// Parse JSON arrays
		if err := json.Unmarshal([]byte(imageURLsJSON), &p.ImageURLs); err != nil {
			p.ImageURLs = []string{}
		}
		if err := json.Unmarshal([]byte(tagsJSON), &p.Tags); err != nil {
			p.Tags = []string{}
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return products, total, nil
}

func (r *ProductRepository) buildWhereClause(filter product.ProductFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	if filter.Vendor != "" {
		conditions = append(conditions, fmt.Sprintf("v.name ILIKE $%d", argIndex))
		args = append(args, "%"+filter.Vendor+"%")
		argIndex++
	}

	if filter.ConfigID != "" {
		conditions = append(conditions, fmt.Sprintf("p.config_id = $%d", argIndex))
		args = append(args, filter.ConfigID)
		argIndex++
	}

	if filter.Currency != "" {
		conditions = append(conditions, fmt.Sprintf("p.currency = $%d", argIndex))
		args = append(args, filter.Currency)
		argIndex++
	}

	if filter.InStock != nil {
		conditions = append(conditions, fmt.Sprintf("p.in_stock = $%d", argIndex))
		args = append(args, *filter.InStock)
		argIndex++
	}

	if filter.MinPrice != nil {
		conditions = append(conditions, fmt.Sprintf("p.price >= $%d", argIndex))
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil {
		conditions = append(conditions, fmt.Sprintf("p.price <= $%d", argIndex))
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	if filter.CreatedAfter != nil {
		conditions = append(conditions, fmt.Sprintf("p.created_at >= $%d", argIndex))
		args = append(args, *filter.CreatedAfter)
		argIndex++
	}

	if filter.CreatedBefore != nil {
		conditions = append(conditions, fmt.Sprintf("p.created_at <= $%d", argIndex))
		args = append(args, *filter.CreatedBefore)
		argIndex++
	}

	if len(filter.Tags) > 0 {
		tagPlaceholders := make([]string, len(filter.Tags))
		for i, tag := range filter.Tags {
			tagPlaceholders[i] = "$" + strconv.Itoa(argIndex)
			args = append(args, tag)
			argIndex++
		}
		conditions = append(conditions, fmt.Sprintf("pt.tag IN (%s)", strings.Join(tagPlaceholders, ",")))
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	return whereClause, args
}

func (r *ProductRepository) GetDistinctVendors(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT v.name
		FROM vendors v
		INNER JOIN site_configurations sc ON v.id = sc.vendor_id
		INNER JOIN products p ON sc.id = p.config_id
		WHERE v.name IS NOT NULL AND v.name != ''
		ORDER BY v.name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query vendors: %w", err)
	}
	defer rows.Close()

	var vendors []string
	for rows.Next() {
		var vendor string
		if err := rows.Scan(&vendor); err != nil {
			return nil, fmt.Errorf("failed to scan vendor: %w", err)
		}
		vendors = append(vendors, vendor)
	}

	return vendors, rows.Err()
}

func (r *ProductRepository) GetDistinctTags(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT tag
		FROM keyboard_tags
		ORDER BY tag
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}
