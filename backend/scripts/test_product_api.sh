#!/bin/bash

# Product API Test Script
# Make sure the API server is running on localhost:8080

BASE_URL="http://localhost:8080/api"

echo "=== Product API Test Script ==="
echo "Base URL: $BASE_URL"
echo

# Test 1: Health check
echo "1. Health check..."
curl -s "$BASE_URL/../health" | head -1
echo -e "\n"

# Test 2: Get filter options
echo "2. Get filter options..."
curl -s "$BASE_URL/products/filter-options" | jq '.' 2>/dev/null || curl -s "$BASE_URL/products/filter-options"
echo -e "\n"

# Test 3: List products (basic)
echo "3. List products (first 5)..."
curl -s "$BASE_URL/products?page=1&page_size=5" | jq '.products[0:2] | .[].name' 2>/dev/null || curl -s "$BASE_URL/products?page=1&page_size=5"
echo -e "\n"

# Test 4: Search products
echo "4. Search for 'keyboard'..."
curl -s "$BASE_URL/products?search=keyboard&page_size=3" | jq '.products[].name' 2>/dev/null || curl -s "$BASE_URL/products?search=keyboard&page_size=3"
echo -e "\n"

# Test 5: Filter by price range
echo "5. Filter by price range (1000-5000)..."
curl -s "$BASE_URL/products?min_price=1000&max_price=5000&page_size=3" | jq '.products[] | {name, price, currency}' 2>/dev/null || curl -s "$BASE_URL/products?min_price=1000&max_price=5000&page_size=3"
echo -e "\n"

# Test 6: Filter by stock
echo "6. Filter by in-stock products..."
curl -s "$BASE_URL/products?in_stock=true&page_size=3" | jq '.products[] | {name, in_stock}' 2>/dev/null || curl -s "$BASE_URL/products?in_stock=true&page_size=3"
echo -e "\n"

# Test 7: Sort by price ascending
echo "7. Sort by price (ascending)..."
curl -s "$BASE_URL/products?sort_field=price&sort_order=asc&page_size=3" | jq '.products[] | {name, price}' 2>/dev/null || curl -s "$BASE_URL/products?sort_field=price&sort_order=asc&page_size=3"
echo -e "\n"

# Test 8: Get specific product (if available)
echo "8. Get first product details..."
PRODUCT_ID=$(curl -s "$BASE_URL/products?page_size=1" | jq -r '.products[0].id' 2>/dev/null)
if [ "$PRODUCT_ID" != "null" ] && [ "$PRODUCT_ID" != "" ]; then
    echo "Product ID: $PRODUCT_ID"
    curl -s "$BASE_URL/products/?id=$PRODUCT_ID" | jq '{name, price, vendor, tags}' 2>/dev/null || curl -s "$BASE_URL/products/?id=$PRODUCT_ID"
else
    echo "No products found to get details"
fi
echo -e "\n"

echo "=== Test Complete ==="
