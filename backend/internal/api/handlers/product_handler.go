package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/meta-boy/mech-alligator/internal/domain/product"
	"github.com/meta-boy/mech-alligator/internal/service"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// GET /api/products
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	req := h.parseListRequest(r)
	req.Validate()

	ctx := r.Context()
	response, err := h.productService.ListProducts(ctx, *req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GET /api/products/{id}
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	productID := strings.TrimPrefix(r.URL.Path, "/api/products/")
	if productID == "" {
		http.Error(w, "product id required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	product, err := h.productService.GetProduct(ctx, productID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if product == nil {
		http.Error(w, "product not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// GET /api/products/filters
func (h *ProductHandler) GetFilterOptions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filters, err := h.productService.GetFilterOptions(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(filters)
}

func (h *ProductHandler) parseListRequest(r *http.Request) *product.ListRequest {
	req := &product.ListRequest{
		Page:      1,
		PageSize:  20,
		SortBy:    "name",
		SortOrder: "asc",
	}

	// Parse query parameters
	query := r.URL.Query()

	if search := query.Get("search"); search != "" {
		req.Search = search
	}
	if brand := query.Get("brand"); brand != "" {
		req.Brand = brand
	}
	if reseller := query.Get("reseller"); reseller != "" {
		req.Reseller = reseller
	}
	if category := query.Get("category"); category != "" {
		req.Category = category
	}
	if tags := query.Get("tags"); tags != "" {
		req.Tags = strings.Split(tags, ",")
	}

	if minPrice := query.Get("min_price"); minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			req.MinPrice = &price
		}
	}
	if maxPrice := query.Get("max_price"); maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			req.MaxPrice = &price
		}
	}

	if available := query.Get("available"); available != "" {
		if avail, err := strconv.ParseBool(available); err == nil {
			req.Available = &avail
		}
	}

	if page := query.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			req.Page = p
		}
	}
	if pageSize := query.Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 {
			req.PageSize = ps
		}
	}

	if sortBy := query.Get("sort_by"); sortBy != "" {
		req.SortBy = sortBy
	}
	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		req.SortOrder = sortOrder
	}

	return req
}
