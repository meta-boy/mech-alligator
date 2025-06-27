package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/meta-boy/mech-alligator/internal/domain/product"
	"github.com/meta-boy/mech-alligator/internal/service"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	// Extract product ID from URL path
	productID := r.URL.Query().Get("id")
	if productID == "" {
		http.Error(w, "product id is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	p, err := h.productService.GetProduct(ctx, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if p == nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	req, err := h.parseListRequest(r)
	if err != nil {
		http.Error(w, "Invalid request parameters: "+err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	response, err := h.productService.ListProducts(ctx, *req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ProductHandler) GetFilterOptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	options, err := h.productService.GetFilterOptions(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(options)
}

func (h *ProductHandler) parseListRequest(r *http.Request) (*product.ProductListRequest, error) {
	req := &product.ProductListRequest{
		Filter: product.ProductFilter{},
		Sort: product.ProductSort{
			Field: "created_at",
			Order: "desc",
		},
		Pagination: product.ProductPagination{
			Page:     1,
			PageSize: 20,
		},
	}

	// Parse filter parameters
	if search := r.URL.Query().Get("search"); search != "" {
		req.Filter.Search = search
	}

	if vendor := r.URL.Query().Get("vendor"); vendor != "" {
		req.Filter.Vendor = vendor
	}

	if configID := r.URL.Query().Get("config_id"); configID != "" {
		req.Filter.ConfigID = configID
	}

	if currency := r.URL.Query().Get("currency"); currency != "" {
		req.Filter.Currency = currency
	}

	if inStockStr := r.URL.Query().Get("in_stock"); inStockStr != "" {
		inStock, err := strconv.ParseBool(inStockStr)
		if err == nil {
			req.Filter.InStock = &inStock
		}
	}

	if tagsStr := r.URL.Query().Get("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		req.Filter.Tags = tags
	}

	if minPriceStr := r.URL.Query().Get("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			req.Filter.MinPrice = &minPrice
		}
	}

	if maxPriceStr := r.URL.Query().Get("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			req.Filter.MaxPrice = &maxPrice
		}
	}

	if createdAfterStr := r.URL.Query().Get("created_after"); createdAfterStr != "" {
		if createdAfter, err := time.Parse(time.RFC3339, createdAfterStr); err == nil {
			req.Filter.CreatedAfter = &createdAfter
		}
	}

	if createdBeforeStr := r.URL.Query().Get("created_before"); createdBeforeStr != "" {
		if createdBefore, err := time.Parse(time.RFC3339, createdBeforeStr); err == nil {
			req.Filter.CreatedBefore = &createdBefore
		}
	}

	// Parse sort parameters
	if sortField := r.URL.Query().Get("sort_field"); sortField != "" {
		req.Sort.Field = sortField
	}

	if sortOrder := r.URL.Query().Get("sort_order"); sortOrder != "" {
		req.Sort.Order = sortOrder
	}

	// Parse pagination parameters
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Pagination.Page = page
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			req.Pagination.PageSize = pageSize
		}
	}

	return req, nil
}
