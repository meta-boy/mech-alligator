package service

import (
	"context"
	"fmt"
	"github.com/meta-boy/mech-alligator/internal/scraper"
	"math"

	"github.com/meta-boy/mech-alligator/internal/domain/product"
	"github.com/meta-boy/mech-alligator/internal/repository/postgres"
)

type ProductService struct {
	productRepo *postgres.ProductRepository
}

func NewProductService(productRepo *postgres.ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*product.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("product id is required")
	}
	return s.productRepo.GetByID(ctx, id)
}

func (s *ProductService) ListProducts(ctx context.Context, req product.ListRequest) (*product.ProductListResponse, error) {
	// Validate request
	req.Validate()

	// Get products and total count
	products, total, err := s.productRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Calculate pagination
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	hasNext := req.Page < totalPages
	hasPrev := req.Page > 1

	return &product.ProductListResponse{
		Products: products,
		Pagination: product.PaginationMeta{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      total,
			TotalPages: totalPages,
			HasNext:    hasNext,
			HasPrev:    hasPrev,
		},
	}, nil
}

func (s *ProductService) GetFilterOptions(ctx context.Context) (*FilterOptions, error) {
	brands, err := s.productRepo.GetDistinctBrands(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get brands: %w", err)
	}

	resellers, err := s.productRepo.GetDistinctResellers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get resellers: %w", err)
	}

	categories, err := s.productRepo.GetDistinctCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return &FilterOptions{
		Brands:     brands,
		Resellers:  resellers,
		Categories: categories,
		SortFields: []string{"name", "price", "brand", "reseller"},
		SortOrders: []string{"asc", "desc"},
	}, nil
}

type FilterOptions struct {
	Brands     []string `json:"brands"`
	Resellers  []string `json:"resellers"`
	Categories []string `json:"categories"`
	SortFields []string `json:"sort_fields"`
	SortOrders []string `json:"sort_orders"`
}

// SaveScrapedProducts saves products from scraping
func (s *ProductService) SaveScrapedProducts(ctx context.Context, scrapedProducts []scraper.ScrapedProduct, resellerID, resellerName string) (*SaveResult, error) {
	var created, updated, errors int

	for _, sp := range scrapedProducts {
		err := s.saveScrapedProduct(ctx, sp, resellerID, resellerName)
		if err != nil {
			errors++
			continue
		}

		// For simplicity, counting all as created
		// In real implementation, you'd check if product exists
		created++
	}

	return &SaveResult{
		Created: created,
		Updated: updated,
		Errors:  errors,
	}, nil
}

func (s *ProductService) saveScrapedProduct(ctx context.Context, sp scraper.ScrapedProduct, resellerID, resellerName string) error {
	// Convert scraped product to our domain model
	p := &product.Product{
		Name:           sp.Name,
		Description:    sp.Description,
		Handle:         sp.Handle,
		URL:            sp.URL,
		Brand:          sp.Brand,
		Reseller:       resellerName,
		ResellerID:     resellerID,
		Category:       sp.Category,
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
		p.Variants = append(p.Variants, variant)
	}

	p.VariantCount = len(p.Variants)

	return s.productRepo.Save(ctx, p)
}

type SaveResult struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
	Errors  int `json:"errors"`
}
