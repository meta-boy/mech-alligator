"use client";

import { useEffect, useState, Suspense } from "react";
import Image from "next/image";
import { useSearchParams } from "next/navigation";
import { Star } from "lucide-react";
import { AuthManager } from "@/utils/auth";
import { Product, ProductResponse, PaginationData, AppliedFilters } from "@/lib/types";
import { LoadingSkeleton } from "@/components/products/LoadingSkeleton";
import { ErrorDisplay } from "@/components/products/ErrorDisplay";
import { ProductFilters } from "@/components/products/ProductFilters";
import { ProductGrid } from "@/components/products/ProductGrid";
import { EmptyState } from "@/components/products/EmptyState";
import { Pagination } from "@/components/products/Pagination";

function ProductDashboardContent() {
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

            <div className="flex items-center gap-4">

              <Image 
                width={64}
                height={64}
                src="/head-logo.png" 

                alt="Mech Alligator Logo" 

                className="w-16 h-16 object-contain"

              />

              <div>

                <h1 className="text-3xl font-bold tracking-tight bg-gradient-to-r from-slate-900 to-slate-700 bg-clip-text text-transparent">

                  Mech Alligator

                </h1>

                <p className="text-muted-foreground mt-1">

                  Discover premium keyboards and keycaps from top resellers in India

                </p>

              </div>

            </div>

            <div className="flex items-center gap-2">

              <a 
                href="https://discord.gg/9Fyd9pYFeX"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1.5 px-3 py-1.5 bg-slate-900 hover:bg-slate-800 text-white rounded-md text-sm font-medium transition-colors duration-200 border border-slate-700 hover:border-slate-600"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M20.317 4.3698a19.7913 19.7913 0 00-4.8851-1.5152.0741.0741 0 00-.0785.0371c-.211.3753-.4447.8648-.6083 1.2495-1.8447-.2762-3.68-.2762-5.4868 0-.1636-.3933-.4058-.8742-.6177-1.2495a.077.077 0 00-.0785-.037 19.7363 19.7363 0 00-4.8852 1.515.0699.0699 0 00-.0321.0277C.5334 9.0458-.319 13.5799.0992 18.0578a.0824.0824 0 00.0312.0561c2.0528 1.5076 4.0413 2.4228 5.9929 3.0294a.0777.0777 0 00.0842-.0276c.4616-.6304.8731-1.2952 1.226-1.9942a.076.076 0 00-.0416-.1057c-.6528-.2476-1.2743-.5495-1.8722-.8923a.077.077 0 01-.0076-.1277c.1258-.0943.2517-.1923.3718-.2914a.0743.0743 0 01.0776-.0105c3.9278 1.7933 8.18 1.7933 12.0614 0a.0739.0739 0 01.0785.0095c.1202.099.246.1981.3728.2924a.077.077 0 01-.0066.1276 12.2986 12.2986 0 01-1.873.8914.0766.0766 0 00-.0407.1067c.3604.698.7719 1.3628 1.225 1.9932a.076.076 0 00.0842.0286c1.961-.6067 3.9495-1.5219 6.0023-3.0294a.077.077 0 00.0313-.0552c.5004-5.177-.8382-9.6739-3.5485-13.6604a.061.061 0 00-.0312-.0286zM8.02 15.3312c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9555-2.4189 2.157-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419-.0001 1.3332-.9555 2.4189-2.1569 2.4189zm7.9748 0c-1.1825 0-2.1569-1.0857-2.1569-2.419 0-1.3332.9554-2.4189 2.1569-2.4189 1.2108 0 2.1757 1.0952 2.1568 2.419 0 1.3332-.9555 2.4189-2.1568 2.4189Z"/>
                </svg>
                Discord
              </a>

              <a 
                href="https://github.com/meta-boy/mech-alligator"
                target="_blank"
                rel="noopener noreferrer"
                className="flex items-center gap-1.5 px-3 py-1.5 bg-slate-900 hover:bg-slate-800 text-white rounded-md text-sm font-medium transition-colors duration-200 border border-slate-700 hover:border-slate-600"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
                  <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/>
                </svg>
                <span>GitHub</span>
                <Star size={12} fill="currentColor" className="text-yellow-400" />
              </a>

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

export default function ProductDashboard() {
  return (
    <Suspense fallback={<LoadingSkeleton />}>
      <ProductDashboardContent />
    </Suspense>
  );
}