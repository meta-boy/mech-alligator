// ProductAPI Client Example
// This file demonstrates how to use the Product API from a frontend application

class ProductAPI {
    constructor(baseURL = 'http://localhost:8080/api') {
        this.baseURL = baseURL;
    }

    // Helper method to build query parameters
    buildQueryParams(params) {
        const searchParams = new URLSearchParams();
        
        Object.keys(params).forEach(key => {
            const value = params[key];
            if (value !== null && value !== undefined && value !== '') {
                if (Array.isArray(value)) {
                    searchParams.set(key, value.join(','));
                } else {
                    searchParams.set(key, value.toString());
                }
            }
        });
        
        return searchParams.toString();
    }

    // Fetch products with filters, sorting, and pagination
    async getProducts(options = {}) {
        const {
            // Filters
            search,
            vendor,
            configId,
            currency,
            inStock,
            tags,
            minPrice,
            maxPrice,
            createdAfter,
            createdBefore,
            
            // Sorting
            sortField = 'created_at',
            sortOrder = 'desc',
            
            // Pagination
            page = 1,
            pageSize = 20
        } = options;

        const params = this.buildQueryParams({
            search,
            vendor,
            config_id: configId,
            currency,
            in_stock: inStock,
            tags,
            min_price: minPrice,
            max_price: maxPrice,
            created_after: createdAfter,
            created_before: createdBefore,
            sort_field: sortField,
            sort_order: sortOrder,
            page,
            page_size: pageSize
        });

        const url = `${this.baseURL}/products?${params}`;
        const response = await fetch(url);
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return await response.json();
    }

    // Get a single product by ID
    async getProduct(id) {
        if (!id) {
            throw new Error('Product ID is required');
        }

        const url = `${this.baseURL}/products/?id=${encodeURIComponent(id)}`;
        const response = await fetch(url);
        
        if (!response.ok) {
            if (response.status === 404) {
                return null;
            }
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return await response.json();
    }

    // Get available filter options
    async getFilterOptions() {
        const url = `${this.baseURL}/products/filter-options`;
        const response = await fetch(url);
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return await response.json();
    }
}

// Usage Examples
async function examples() {
    const api = new ProductAPI();

    try {
        // Example 1: Basic product listing
        console.log('=== Basic Product Listing ===');
        const basicList = await api.getProducts({
            page: 1,
            pageSize: 10
        });
        console.log(`Found ${basicList.pagination.total_items} products`);
        console.log(`Page ${basicList.pagination.page} of ${basicList.pagination.total_pages}`);
        
        // Example 2: Search and filter
        console.log('\n=== Search and Filter ===');
        const searchResults = await api.getProducts({
            search: 'keyboard',
            minPrice: 1000,
            maxPrice: 5000,
            inStock: true,
            currency: 'INR',
            sortField: 'price',
            sortOrder: 'asc'
        });
        console.log(`Found ${searchResults.products.length} keyboards between ₹1000-₹5000`);
        
        // Example 3: Filter by tags
        console.log('\n=== Filter by Tags ===');
        const taggedProducts = await api.getProducts({
            tags: ['mechanical', 'rgb'],
            pageSize: 5
        });
        console.log(`Found ${taggedProducts.products.length} mechanical RGB products`);
        
        // Example 4: Get single product
        if (basicList.products.length > 0) {
            console.log('\n=== Single Product ===');
            const productId = basicList.products[0].id;
            const product = await api.getProduct(productId);
            console.log(`Product: ${product.name} - ₹${product.price}`);
            console.log(`Vendor: ${product.vendor}`);
            console.log(`Tags: ${product.tags.join(', ')}`);
        }
        
        // Example 5: Get filter options
        console.log('\n=== Filter Options ===');
        const filterOptions = await api.getFilterOptions();
        console.log('Available vendors:', filterOptions.vendors.slice(0, 5));
        console.log('Available tags:', filterOptions.tags);
        
    } catch (error) {
        console.error('API Error:', error.message);
    }
}

// React Hook Example
function useProducts(filters = {}, dependencies = []) {
    const [products, setProducts] = React.useState([]);
    const [pagination, setPagination] = React.useState(null);
    const [loading, setLoading] = React.useState(false);
    const [error, setError] = React.useState(null);
    
    const api = React.useMemo(() => new ProductAPI(), []);
    
    React.useEffect(() => {
        let isCancelled = false;
        
        const fetchProducts = async () => {
            setLoading(true);
            setError(null);
            
            try {
                const result = await api.getProducts(filters);
                if (!isCancelled) {
                    setProducts(result.products);
                    setPagination(result.pagination);
                }
            } catch (err) {
                if (!isCancelled) {
                    setError(err.message);
                }
            } finally {
                if (!isCancelled) {
                    setLoading(false);
                }
            }
        };
        
        fetchProducts();
        
        return () => {
            isCancelled = true;
        };
    }, dependencies);
    
    return { products, pagination, loading, error };
}

// React Component Example
function ProductList() {
    const [filters, setFilters] = React.useState({
        search: '',
        vendor: '',
        minPrice: '',
        maxPrice: '',
        inStock: null,
        sortField: 'created_at',
        sortOrder: 'desc',
        page: 1,
        pageSize: 20
    });
    
    const { products, pagination, loading, error } = useProducts(
        filters, 
        [JSON.stringify(filters)]
    );
    
    const handleFilterChange = (key, value) => {
        setFilters(prev => ({
            ...prev,
            [key]: value,
            page: 1 // Reset to first page when filtering
        }));
    };
    
    const handlePageChange = (page) => {
        setFilters(prev => ({ ...prev, page }));
    };
    
    if (loading) return <div>Loading products...</div>;
    if (error) return <div>Error: {error}</div>;
    
    return (
        <div className="product-list">
            {/* Filter Controls */}
            <div className="filters">
                <input
                    type="text"
                    placeholder="Search products..."
                    value={filters.search}
                    onChange={(e) => handleFilterChange('search', e.target.value)}
                />
                
                <select
                    value={filters.sortField}
                    onChange={(e) => handleFilterChange('sortField', e.target.value)}
                >
                    <option value="created_at">Date Added</option>
                    <option value="price">Price</option>
                    <option value="name">Name</option>
                </select>
                
                <select
                    value={filters.sortOrder}
                    onChange={(e) => handleFilterChange('sortOrder', e.target.value)}
                >
                    <option value="desc">Descending</option>
                    <option value="asc">Ascending</option>
                </select>
                
                <input
                    type="number"
                    placeholder="Min Price"
                    value={filters.minPrice}
                    onChange={(e) => handleFilterChange('minPrice', e.target.value)}
                />
                
                <input
                    type="number"
                    placeholder="Max Price"
                    value={filters.maxPrice}
                    onChange={(e) => handleFilterChange('maxPrice', e.target.value)}
                />
                
                <label>
                    <input
                        type="checkbox"
                        checked={filters.inStock === true}
                        onChange={(e) => handleFilterChange('inStock', e.target.checked ? true : null)}
                    />
                    In Stock Only
                </label>
            </div>
            
            {/* Product Grid */}
            <div className="products">
                {products.map(product => (
                    <div key={product.id} className="product-card">
                        <h3>{product.name}</h3>
                        <p>₹{product.price} {product.currency}</p>
                        <p>Vendor: {product.vendor}</p>
                        <p>In Stock: {product.in_stock ? 'Yes' : 'No'}</p>
                        {product.tags.length > 0 && (
                            <div className="tags">
                                {product.tags.map(tag => (
                                    <span key={tag} className="tag">{tag}</span>
                                ))}
                            </div>
                        )}
                    </div>
                ))}
            </div>
            
            {/* Pagination */}
            {pagination && (
                <div className="pagination">
                    <button
                        disabled={!pagination.has_previous}
                        onClick={() => handlePageChange(pagination.page - 1)}
                    >
                        Previous
                    </button>
                    
                    <span>
                        Page {pagination.page} of {pagination.total_pages}
                        ({pagination.total_items} total items)
                    </span>
                    
                    <button
                        disabled={!pagination.has_next}
                        onClick={() => handlePageChange(pagination.page + 1)}
                    >
                        Next
                    </button>
                </div>
            )}
        </div>
    );
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { ProductAPI, useProducts };
}
