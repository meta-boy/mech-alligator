"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { 
  ExternalLink, 
  Package, 
  Search, 
  Grid3X3,
  List,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight
} from "lucide-react";
import { AuthManager } from "@/utils/auth";

interface Product {
  id: string;
  name: string;
  description: string;
  handle: string;
  url: string;
  brand: string;
  reseller: string;
  category: string;
  tags: string[];
  images: string[];
  variant_count: number;
  source_type: string;
  source_id: string;
  reseller_id: string;
  source_metadata: {
    created_at: string;
    published_at: string;
    shopify_handle: string;
    shopify_product_type: string;
    shopify_vendor: string;
    updated_at: string;
  };
}

interface ProductResponse {
  products: Product[];
  pagination: {
    page: number;
    page_size: number;
    total: number;
    total_pages: number;
    has_next: boolean;
    has_prev: boolean;
  };
}

export default function ProductDashboard() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [activeTab, setActiveTab] = useState("all");
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [pagination, setPagination] = useState({
    page: 1,
    page_size: 20,
    total: 0,
    total_pages: 0,
    has_next: false,
    has_prev: false,
  });

  const ImageGallery = ({ images, productName }: { images: string[], productName: string }) => {
    // Generate deterministic random selection based on product name
    const getRandomImages = (imageArray: string[], name: string, count: number = 3) => {
      if (!imageArray || imageArray.length === 0) return [];
      
      // Create a simple hash from the product name for consistent randomness
      let hash = 0;
      for (let i = 0; i < name.length; i++) {
        const char = name.charCodeAt(i);
        hash = ((hash << 5) - hash) + char;
        hash = hash & hash; // Convert to 32-bit integer
      }
      
      // Use the hash to select images
      const selectedImages = [];
      const availableImages = [...imageArray];
      const selectCount = Math.min(count, availableImages.length);
      
      for (let i = 0; i < selectCount; i++) {
        const index = Math.abs(hash + i) % availableImages.length;
        selectedImages.push(availableImages[index]);
        availableImages.splice(index, 1);
      }
      
      return selectedImages;
    };

    const selectedImages = getRandomImages(images, productName, 3);
    const [currentImageIndex, setCurrentImageIndex] = useState(0);
    
    const nextImage = (e: React.MouseEvent) => {
      e.stopPropagation();
      setCurrentImageIndex((prev) => (prev + 1) % selectedImages.length);
    };
    
    const prevImage = (e: React.MouseEvent) => {
      e.stopPropagation();
      setCurrentImageIndex((prev) => (prev - 1 + selectedImages.length) % selectedImages.length);
    };

    if (!selectedImages || selectedImages.length === 0) {
      return (
        <div className="h-48 bg-gradient-to-br from-slate-100 to-slate-200 relative flex items-center justify-center">
          <Package className="h-16 w-16 text-slate-400" />
        </div>
      );
    }

    return (
      <div className="h-48 bg-gradient-to-br from-slate-100 to-slate-200 relative overflow-hidden group">
        <div className="w-full h-full">
          <Image
        src={selectedImages[currentImageIndex]}
        alt={`${productName} - Image ${currentImageIndex + 1}`}
        fill
        className="object-cover transition-all duration-300"
        style={{ objectFit: "cover" }}
        quality={60}
        onError={(e) => {
          const target = e.target as HTMLImageElement;
          target.style.display = 'none';
          target.nextElementSibling?.classList.remove('hidden');
        }}
        sizes="(max-width: 768px) 100vw, 400px"
        priority={currentImageIndex === 0}
          />
        </div>
        
        {/* Fallback for broken images */}
        <div className="absolute inset-0 hidden bg-gradient-to-br from-slate-100 to-slate-200 flex items-center justify-center">
          <Package className="h-16 w-16 text-slate-400" />
        </div>

        {/* Navigation buttons - only show if multiple images */}
        {selectedImages.length > 1 && (
          <>
        <button
          onClick={prevImage}
          className="absolute left-2 top-1/2 -translate-y-1/2 bg-black/50 hover:bg-black/70 text-white p-1 rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-200"
        >
          <ChevronLeft className="h-4 w-4" />
        </button>
        <button
          onClick={nextImage}
          className="absolute right-2 top-1/2 -translate-y-1/2 bg-black/50 hover:bg-black/70 text-white p-1 rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-200"
        >
          <ChevronRight className="h-4 w-4" />
        </button>
        
        {/* Image indicators */}
        <div className="absolute bottom-2 left-1/2 -translate-x-1/2 flex gap-1">
          {selectedImages.map((_, index) => (
            <button
          key={index}
          onClick={(e) => {
            e.stopPropagation();
            setCurrentImageIndex(index);
          }}
          className={`w-2 h-2 rounded-full transition-all duration-200 ${
            index === currentImageIndex 
              ? 'bg-white' 
              : 'bg-white/50 hover:bg-white/70'
          }`}
            />
          ))}
        </div>
          </>
        )}
      </div>
    );
  };

  useEffect(() => {
    const fetchProducts = async () => {
      setLoading(true);
      try {
        const authManager = AuthManager.getInstance();
        const url = `/api/products?page=${currentPage}&page_size=${pageSize}`;
        const response = await authManager.makeAuthenticatedRequest(url);
        
        if (!response.ok) {
          throw new Error('Failed to fetch products');
        }
        
        const data: ProductResponse = await response.json();
        setProducts(data.products);
        setPagination(data.pagination);
      } catch (err) {
        if (err instanceof Error) {
          setError(err.message);
        } else {
          setError("An unknown error occurred");
        }
      } finally {
        setLoading(false);
      }
    };

    fetchProducts();
  }, [currentPage, pageSize]);

  const handlePageChange = (newPage: number) => {
    setCurrentPage(newPage);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handlePageSizeChange = (newPageSize: number) => {
    setCurrentPage(1);
    setPageSize(newPageSize);
  };

  const filteredProducts = products.filter(product => {
    const matchesSearch = product.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         product.reseller.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         product.brand.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesTab = activeTab === "all" || 
                      (activeTab === "published" && product.category === "KEYCAPS") ||
                      (activeTab === "draft" && product.category === "KEYBOARDS");
    
    return matchesSearch && matchesTab;
  });

  const ProductSkeleton = () => (
    <Card className="w-full">
      <CardHeader>
        <Skeleton className="h-6 w-3/4" />
        <Skeleton className="h-4 w-1/2" />
      </CardHeader>
      <CardContent>
        <Skeleton className="h-4 w-full mb-2" />
        <Skeleton className="h-4 w-2/3" />
        <div className="flex gap-2 mt-4">
          <Skeleton className="h-5 w-16" />
          <Skeleton className="h-5 w-20" />
        </div>
      </CardContent>
    </Card>
  );

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 p-6">
        <div className="max-w-7xl mx-auto">
          <div className="mb-8">
            <Skeleton className="h-10 w-64 mb-4" />
            <Skeleton className="h-6 w-96" />
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            {Array.from({ length: 3 }).map((_, i) => (
              <Card key={i}>
                <CardContent className="p-6">
                  <Skeleton className="h-12 w-12 mb-4" />
                  <Skeleton className="h-6 w-24 mb-2" />
                  <Skeleton className="h-8 w-32" />
                </CardContent>
              </Card>
            ))}
          </div>

          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {Array.from({ length: 8 }).map((_, i) => (
              <ProductSkeleton key={i} />
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 flex items-center justify-center p-6">
        <Card className="max-w-md w-full">
          <CardContent className="text-center p-8">
            <Package className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
            <h1 className="text-2xl font-bold text-destructive mb-2">Error</h1>
            <p className="text-muted-foreground mb-6">{error}</p>
            <Button onClick={() => window.location.reload()}>
              Try Again
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      <div className="max-w-7xl mx-auto p-6">
        {/* Header */}
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

        {/* Filters and Search */}
        <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm mb-6">
          <CardContent className="p-6">
            <div className="flex flex-col sm:flex-row gap-4 items-center justify-between">
              <div className="flex items-center gap-4 w-full sm:w-auto">
                <div className="relative flex-1 sm:w-80">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    placeholder="Search products..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-10 border-0 bg-slate-50"
                  />
                </div>
              </div>

              <div className="flex items-center gap-2">
                <Tabs value={activeTab} onValueChange={setActiveTab}>
                  <TabsList className="bg-slate-100">
                    <TabsTrigger value="all" className="text-xs">All Products</TabsTrigger>
                    <TabsTrigger value="published" className="text-xs">Keycaps</TabsTrigger>
                    <TabsTrigger value="draft" className="text-xs">Keyboards</TabsTrigger>
                  </TabsList>
                </Tabs>

                <div className="flex items-center border rounded-md bg-slate-50">
                  <Button
                    variant={viewMode === "grid" ? "default" : "ghost"}
                    size="sm"
                    onClick={() => setViewMode("grid")}
                    className="rounded-r-none"
                  >
                    <Grid3X3 className="h-4 w-4" />
                  </Button>
                  <Button
                    variant={viewMode === "list" ? "default" : "ghost"}
                    size="sm"
                    onClick={() => setViewMode("list")}
                    className="rounded-l-none"
                  >
                    <List className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Products Grid */}
        {viewMode === "grid" ? (
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
            {filteredProducts.map((product) => (
              <Card key={product.id} className="group hover:shadow-lg transition-all duration-300 hover:-translate-y-1 border-0 shadow-sm bg-white/80 backdrop-blur-sm overflow-hidden">
                <div className="relative">
                  <ImageGallery images={product.images} productName={product.name} />
                  <div className="absolute top-4 right-4 flex gap-2">
                    <Badge variant="outline" className="bg-white/90">
                      {product.brand}
                    </Badge>
                    <Button 
                      size="sm"
                      onClick={() => window.open(product.url, '_blank')}
                      className="h-6 px-2 text-xs"
                    >
                      <ExternalLink className="h-3 w-3 mr-1" />
                      View
                    </Button>
                  </div>
                  <div className="absolute bottom-4 left-4 right-4">
                    <div className="flex items-center gap-2 text-sm text-white drop-shadow-lg">
                      <Package className="h-4 w-4" />
                      <span className="capitalize font-medium">{product.reseller}</span>
                    </div>
                  </div>
                </div>
                
                <CardContent className="p-6">
                  <CardTitle className="text-lg leading-tight line-clamp-2 group-hover:text-blue-600 transition-colors mb-2">
                    {product.name}
                  </CardTitle>
                  
                  <div className="flex items-center gap-2 mb-4">
                    <Badge variant="secondary" className="text-xs">
                      {product.category}
                    </Badge>
                    <Badge variant="outline" className="text-xs">
                      {product.variant_count} variant{product.variant_count !== 1 ? 's' : ''}
                    </Badge>
                  </div>
                  
                  {product.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {product.tags.slice(0, 2).map((tag) => (
                        <Badge key={tag} variant="outline" className="text-xs border-slate-200">
                          {tag}
                        </Badge>
                      ))}
                      {product.tags.length > 2 && (
                        <Badge variant="outline" className="text-xs border-slate-200">
                          +{product.tags.length - 2}
                        </Badge>
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>
        ) : (
          // List View
          <div className="space-y-4">
            {filteredProducts.map((product) => (
              <Card key={product.id} className="border-0 shadow-sm bg-white/80 backdrop-blur-sm hover:shadow-md transition-shadow">
                <CardContent className="p-6">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-6 flex-1">
                      <div className="h-16 w-16 rounded-lg overflow-hidden bg-gradient-to-br from-slate-100 to-slate-200 flex-shrink-0">
                        {product.images && product.images.length > 0 ? (
                          <Image
                            src={(() => {
                              // Generate deterministic random selection for list view
                              let hash = 0;
                              for (let i = 0; i < product.name.length; i++) {
                                const char = product.name.charCodeAt(i);
                                hash = ((hash << 5) - hash) + char;
                                hash = hash & hash;
                              }
                              const index = Math.abs(hash) % product.images.length;
                              return product.images[index];
                            })()}
                            alt={product.name}
                            width={64}
                            height={64}
                            className="w-full h-full object-cover"
                            quality={60}
                            onError={(e) => {
                              const target = e.target as HTMLImageElement;
                              target.style.display = 'none';
                              target.nextElementSibling?.classList.remove('hidden');
                            }}
                            sizes="64px"
                            priority
                          />
                        ) : null}
                        <div className={`w-full h-full flex items-center justify-center ${product.images && product.images.length > 0 ? 'hidden' : ''}`}>
                          <Package className="h-8 w-8 text-slate-500" />
                        </div>
                      </div>
                      
                      <div className="flex-1">
                        <div className="flex items-center gap-3 mb-1">
                          <h3 className="font-semibold text-lg">{product.name}</h3>
                          <Badge variant="outline">
                            {product.brand}
                          </Badge>
                        </div>
                        
                        <div className="flex items-center gap-4 text-sm text-muted-foreground mb-2">
                          <span className="capitalize">{product.reseller}</span>
                          <span>•</span>
                          <span>{product.category}</span>
                          <span>•</span>
                          <span>{product.variant_count} variant{product.variant_count !== 1 ? 's' : ''}</span>
                        </div>
                        
                        <div className="flex flex-wrap gap-1">
                          {product.tags.slice(0, 3).map((tag) => (
                            <Badge key={tag} variant="outline" className="text-xs">
                              {tag}
                            </Badge>
                          ))}
                          {product.tags.length > 3 && (
                            <Badge variant="outline" className="text-xs">
                              +{product.tags.length - 3}
                            </Badge>
                          )}
                        </div>
                      </div>
                      
                      <Button 
                        onClick={() => window.open(product.url, '_blank')}
                      >
                        <ExternalLink className="h-4 w-4 mr-2" />
                        View Product
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {/* Pagination */}
        {!loading && products.length > 0 && (
          <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm mt-6">
            <CardContent className="p-6">
              <div className="flex flex-col sm:flex-row items-center justify-between gap-4">
                {/* Page info */}
                <div className="text-sm text-muted-foreground">
                  Showing {((pagination.page - 1) * pageSize) + 1} to{' '}
                  {Math.min(pagination.page * pageSize, pagination.total)} of{' '}
                  {pagination.total} products
                </div>

                {/* Pagination controls */}
                <div className="flex items-center gap-2">
                  {/* Page size selector */}
                  <div className="flex items-center gap-2 mr-4">
                    <span className="text-sm text-muted-foreground">Show:</span>
                    <select
                      value={pageSize}
                      onChange={(e) => handlePageSizeChange(Number(e.target.value))}
                      className="px-2 py-1 text-sm border rounded-md bg-background"
                    >
                      <option value={10}>10</option>
                      <option value={20}>20</option>
                      <option value={50}>50</option>
                      <option value={100}>100</option>
                    </select>
                  </div>

                  {/* Navigation buttons */}
                  <div className="flex items-center gap-1">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(1)}
                      disabled={!pagination.has_prev}
                    >
                      <ChevronsLeft className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(pagination.page - 1)}
                      disabled={!pagination.has_prev}
                    >
                      <ChevronLeft className="h-4 w-4" />
                    </Button>
                    
                    {/* Page numbers */}
                    <div className="flex items-center gap-1 mx-2">
                      {Array.from({ length: Math.min(5, pagination.total_pages) }, (_, i) => {
                        let pageNumber;
                        if (pagination.total_pages <= 5) {
                          pageNumber = i + 1;
                        } else if (pagination.page <= 3) {
                          pageNumber = i + 1;
                        } else if (pagination.page >= pagination.total_pages - 2) {
                          pageNumber = pagination.total_pages - 4 + i;
                        } else {
                          pageNumber = pagination.page - 2 + i;
                        }

                        return (
                          <Button
                            key={pageNumber}
                            variant={pageNumber === pagination.page ? "default" : "outline"}
                            size="sm"
                            onClick={() => handlePageChange(pageNumber)}
                            className="w-8 h-8 p-0"
                          >
                            {pageNumber}
                          </Button>
                        );
                      })}
                    </div>

                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(pagination.page + 1)}
                      disabled={!pagination.has_next}
                    >
                      <ChevronRight className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handlePageChange(pagination.total_pages)}
                      disabled={!pagination.has_next}
                    >
                      <ChevronsRight className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Empty State */}
        {filteredProducts.length === 0 && !loading && (
          <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm">
            <CardContent className="text-center py-16">
              <Package className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-xl font-semibold mb-2">No products found</h3>
              <p className="text-muted-foreground mb-6">
                {searchTerm ? "Try adjusting your search terms" : "No products available at the moment"}
              </p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}