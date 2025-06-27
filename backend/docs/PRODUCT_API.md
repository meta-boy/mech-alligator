# Product API Documentation

This document describes the Product API endpoints for fetching products with pagination and filtering capabilities.

## Base URL
```
http://localhost:8080/api
```

## Endpoints

### 1. List Products (GET /api/products)

Fetch a paginated list of products with optional filtering and sorting.

#### Query Parameters

##### Filtering
- `search` (string): Search in product name and description
- `vendor` (string): Filter by vendor name (partial match)
- `config_id` (string): Filter by specific site configuration ID
- `currency` (string): Filter by currency (INR, USD)
- `in_stock` (boolean): Filter by stock status (true/false)
- `tags` (string): Comma-separated list of tags to filter by
- `min_price` (float): Minimum price filter
- `max_price` (float): Maximum price filter
- `created_after` (datetime): Filter products created after this date (RFC3339 format)
- `created_before` (datetime): Filter products created before this date (RFC3339 format)

##### Sorting
- `sort_field` (string): Field to sort by (name, price, created_at, updated_at)
- `sort_order` (string): Sort order (asc, desc)

##### Pagination
- `page` (int): Page number (default: 1)
- `page_size` (int): Items per page (default: 20, max: 100)

#### Example Requests

**Basic listing:**
```bash
GET /api/products?page=1&page_size=10
```

**Search with filters:**
```bash
GET /api/products?search=keyboard&vendor=keychron&min_price=1000&max_price=5000&in_stock=true
```

**With sorting:**
```bash
GET /api/products?sort_field=price&sort_order=asc&page=1&page_size=20
```

**Filter by tags:**
```bash
GET /api/products?tags=keyboard,mechanical&currency=USD
```

**Date range filtering:**
```bash
GET /api/products?created_after=2024-01-01T00:00:00Z&created_before=2024-12-31T23:59:59Z
```

#### Response Format

```json
{
  "products": [
    {
      "id": "abc123",
      "name": "Product Name",
      "description": "Product description",
      "price": 2500.00,
      "currency": "INR",
      "url": "https://example.com/product",
      "config_id": "config123",
      "in_stock": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z",
      "vendor": "Vendor Name",
      "image_urls": [
        "https://example.com/image1.jpg",
        "https://example.com/image2.jpg"
      ],
      "tags": ["keyboard", "mechanical", "rgb"]
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 150,
    "total_pages": 8,
    "has_next": true,
    "has_previous": false
  },
  "filter": {
    "search": "keyboard",
    "vendor": "keychron",
    "min_price": 1000,
    "max_price": 5000,
    "in_stock": true
  },
  "sort": {
    "field": "price",
    "order": "asc"
  }
}
```

### 2. Get Single Product (GET /api/products/)

Fetch details of a specific product by ID.

#### Query Parameters
- `id` (string, required): Product ID

#### Example Request
```bash
GET /api/products/?id=abc123
```

#### Response Format
```json
{
  "id": "abc123",
  "name": "Product Name",
  "description": "Detailed product description",
  "price": 2500.00,
  "currency": "INR",
  "url": "https://example.com/product",
  "config_id": "config123",
  "in_stock": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "vendor": "Vendor Name",
  "image_urls": [
    "https://example.com/image1.jpg",
    "https://example.com/image2.jpg"
  ],
  "tags": ["keyboard", "mechanical", "rgb"]
}
```

### 3. Get Filter Options (GET /api/products/filter-options)

Get available filter options for building dynamic filter UIs.

#### Example Request
```bash
GET /api/products/filter-options
```

#### Response Format
```json
{
  "vendors": ["Keychron", "Das Keyboard", "Corsair"],
  "tags": ["keyboard", "keycaps", "switches", "accessories"],
  "currencies": ["INR", "USD"],
  "sort_fields": ["name", "price", "created_at", "updated_at"],
  "sort_orders": ["asc", "desc"]
}
```

## Error Responses

All endpoints return appropriate HTTP status codes:

- `200 OK`: Success
- `400 Bad Request`: Invalid parameters
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

Error response format:
```json
{
  "error": "Error message description"
}
```

## Data Types

### Product Object
- `id`: String - Unique product identifier
- `name`: String - Product name
- `description`: String - Product description (may contain HTML)
- `price`: Float - Product price
- `currency`: String - Currency code (INR, USD)
- `url`: String - Product URL on vendor site
- `config_id`: String - Site configuration ID
- `in_stock`: Boolean - Stock availability
- `created_at`: DateTime - Creation timestamp (RFC3339)
- `updated_at`: DateTime - Last update timestamp (RFC3339)
- `vendor`: String - Vendor name (from joined data)
- `image_urls`: Array of strings - Product image URLs
- `tags`: Array of strings - Product tags

### Pagination Object
- `page`: Integer - Current page number
- `page_size`: Integer - Items per page
- `total_items`: Integer - Total number of items
- `total_pages`: Integer - Total number of pages
- `has_next`: Boolean - Whether there's a next page
- `has_previous`: Boolean - Whether there's a previous page

## Performance Considerations

1. **Pagination**: Use reasonable page sizes (recommended: 10-50 items)
2. **Filtering**: Combine filters to reduce result sets
3. **Sorting**: Database indexes exist for commonly sorted fields
4. **Search**: Text search is case-insensitive and searches both name and description

## Examples

### Frontend JavaScript Usage

```javascript
// Fetch products with filters
async function fetchProducts(filters = {}) {
  const params = new URLSearchParams({
    page: filters.page || 1,
    page_size: filters.pageSize || 20,
    sort_field: filters.sortField || 'created_at',
    sort_order: filters.sortOrder || 'desc',
    ...filters
  });

  const response = await fetch(`/api/products?${params}`);
  return response.json();
}

// Usage examples
const keyboards = await fetchProducts({
  search: 'keyboard',
  tags: 'mechanical,rgb',
  min_price: 1000,
  max_price: 10000,
  in_stock: true
});

const expensiveProducts = await fetchProducts({
  sort_field: 'price',
  sort_order: 'desc',
  min_price: 5000
});
```

### cURL Examples

```bash
# Get first page of products
curl "http://localhost:8080/api/products?page=1&page_size=10"

# Search for keyboards under 5000 INR
curl "http://localhost:8080/api/products?search=keyboard&max_price=5000&currency=INR"

# Get products from specific vendor
curl "http://localhost:8080/api/products?vendor=keychron&in_stock=true"

# Get filter options
curl "http://localhost:8080/api/products/filter-options"

# Get specific product
curl "http://localhost:8080/api/products/?id=abc123"
```
