package service

import (
	"context"
	"fmt"
	"math"

	"github.com/meta-boy/mech-alligator/internal/domain/product"
	"github.com/meta-boy/mech-alligator/internal/repository/postgres"
)

type ProductService struct {
	productRepo *postgres.ProductRepository
}

func NewProductService(productRepo *postgres.ProductRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

func (s *ProductService) GetProduct(ctx context.Context, id string) (*product.Product, error) {
	if id == "" {
		return nil, fmt.Errorf("product id is required")
	}

	return s.productRepo.GetByID(ctx, id)
}

func (s *ProductService) ListProducts(ctx context.Context, req product.ProductListRequest) (*product.ProductListResponse, error) {
	// Validate and set defaults
	req.Pagination.Validate()
	req.Sort.Validate()

	// Get products and total count
	products, total, err := s.productRepo.List(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Calculate pagination metadata
	totalPages := int(math.Ceil(float64(total) / float64(req.Pagination.PageSize)))
	hasNext := req.Pagination.Page < totalPages
	hasPrevious := req.Pagination.Page > 1

	response := &product.ProductListResponse{
		Products: products,
		Pagination: product.PaginationMeta{
			Page:        req.Pagination.Page,
			PageSize:    req.Pagination.PageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrevious: hasPrevious,
		},
		Filter: req.Filter,
		Sort:   req.Sort,
	}

	return response, nil
}

func (s *ProductService) GetFilterOptions(ctx context.Context) (*ProductFilterOptions, error) {
	vendors, err := s.productRepo.GetDistinctVendors(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get vendors: %w", err)
	}

	tags, err := s.productRepo.GetDistinctTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	return &ProductFilterOptions{
		Vendors:    vendors,
		Tags:       tags,
		Currencies: []string{"INR", "USD"},
		SortFields: []string{"name", "price", "created_at", "updated_at"},
		SortOrders: []string{"asc", "desc"},
	}, nil
}

type ProductFilterOptions struct {
	Vendors    []string `json:"vendors"`
	Tags       []string `json:"tags"`
	Currencies []string `json:"currencies"`
	SortFields []string `json:"sort_fields"`
	SortOrders []string `json:"sort_orders"`
}
