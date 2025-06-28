"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { AuthManager } from "@/utils/auth";
import { Product, ProductResponse, PaginationData, AppliedFilters } from "@/lib/types";
import { LoadingSkeleton } from "@/components/products/LoadingSkeleton";
import { ErrorDisplay } from "@/components/products/ErrorDisplay";
import { ProductFilters } from "@/components/products/ProductFilters";
import { ProductGrid } from "@/components/products/ProductGrid";
import { EmptyState } from "@/components/products/EmptyState";
import { Pagination } from "@/components/products/Pagination";

export default function ProductDashboard() {
  const searchParams = useSearchParams();

  const initialSearchTerm = searchParams.get('search') || '';
  const initialFilters: AppliedFilters = {};
  searchParams.forEach((value, key) => {
    if (key !== 'search' && key !== 'page' && key !== 'pageSize') {
      initialFilters[key as keyof AppliedFilters] = value;
    }
  });
  const initialPage = parseInt(searchParams.get('page') || '1');
  const initialPageSize = parseInt(searchParams.get('pageSize') || '20');

  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [isSearching, setIsSearching] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState(initialSearchTerm);
  const [actualSearchTerm, setActualSearchTerm] = useState(initialSearchTerm);
  const [appliedFilters, setAppliedFilters] = useState<AppliedFilters>(initialFilters);
  const [currentPage, setCurrentPage] = useState(initialPage);
  const [pageSize, setPageSize] = useState(initialPageSize);
  const [pagination, setPagination] = useState<PaginationData>({
    page: initialPage,
    page_size: initialPageSize,
    total: 0,
    total_pages: 0,
    has_next: false,
    has_prev: false,
  });

  useEffect(() => {
    const fetchProducts = async () => {
      if (currentPage === 1) {
        setLoading(true);
      } else {
        setIsSearching(true);
      }

      try {
        const authManager = AuthManager.getInstance();
        let url = `/api/products?page=${currentPage}&page_size=${pageSize}`;

        if (actualSearchTerm) {
          url += `&search=${encodeURIComponent(actualSearchTerm)}`;
        }

        Object.entries(appliedFilters).forEach(([key, value]) => {
          if (value !== undefined) {
            url += `&${key}=${encodeURIComponent(value)}`;
          }
        });

        const response = await authManager.makeAuthenticatedRequest(url);

        if (!response.ok) {
          throw new Error('Failed to fetch products');
        }

        const data: ProductResponse = await response.json();
        setProducts(data.products || []);
        setPagination(data.pagination);
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError("An unknown error occurred");
        }
      } finally {
        setLoading(false);
        setIsSearching(false);
      }
    };

    fetchProducts();
  }, [currentPage, pageSize, actualSearchTerm, appliedFilters]);

  const handlePageChange = (newPage: number) => {
    setCurrentPage(newPage);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handlePageSizeChange = (newPageSize: number) => {
    setCurrentPage(1);
    setPageSize(newPageSize);
  };

  const handleSearch = (term: string) => {
    setActualSearchTerm(term);
    setCurrentPage(1);
    setIsSearching(true);
  };

  const handleFiltersChange = (filters: AppliedFilters) => {
    setAppliedFilters(filters);
    setCurrentPage(1);
  };

  const filteredProducts = products;

  if (loading) {
    return <LoadingSkeleton />;
  }

  if (error) {
    return <ErrorDisplay error={error} />;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      <div className="max-w-7xl mx-auto p-6">
        <div className="mb-8">
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-slate-900 to-slate-700 bg-clip-text text-transparent">
                Products
              </h1>
              <p className="text-muted-foreground mt-1">
                Discover premium keyboards and keycaps from top resellers
              </p>
            </div>
          </div>
        </div>

        <ProductFilters
          onSearch={handleSearch}
          onFiltersChange={handleFiltersChange}
          isSearching={isSearching}
        />

        {filteredProducts.length > 0 ? (
          <ProductGrid products={filteredProducts} />
        ) : (
          <EmptyState searchTerm={actualSearchTerm} />
        )}

        {!loading && products.length > 0 && (
          <Pagination
            pagination={pagination}
            pageSize={pageSize}
            onPageChange={handlePageChange}
            onPageSizeChange={handlePageSizeChange}
          />
        )}
      </div>
    </div>
  );
}