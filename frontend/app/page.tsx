"use client";

import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Input } from "@/components/ui/input";
import { 
  ExternalLink, 
  Package, 
  Tag, 
  Search, 
  Filter,
  Grid3X3,
  List,
  TrendingUp,
  DollarSign,
  ShoppingBag,
  Eye
} from "lucide-react";

interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  currency: string;
  url: string;
  config_id: string;
  in_stock: boolean;
  created_at: string;
  updated_at: string;
  vendor: string;
  tags: string[];
}

interface ProductResponse {
  products: Product[];
  pagination: {
    page: number;
    page_size: number;
    total_items: number;
    total_pages: number;
    has_next: boolean;
    has_previous: boolean;
  };
  filter: Record<string, any>;
  sort: {
    field: string;
    order: string;
  };
}

// Mock data for demonstration
const mockProducts: Product[] = [
  {
    id: "1",
    name: "Premium Mechanical Keyboard Pro",
    description: "High-quality mechanical keyboard with RGB lighting",
    price: 15999,
    currency: "INR",
    url: "#",
    config_id: "config1",
    in_stock: true,
    created_at: "2024-01-15",
    updated_at: "2024-01-20",
    vendor: "keychron",
    tags: ["mechanical", "rgb", "wireless"]
  },
  {
    id: "2",
    name: "Ultra Compact 60% Keyboard",
    description: "Space-saving compact design for minimalists",
    price: 8999,
    currency: "INR",
    url: "#",
    config_id: "config2",
    in_stock: true,
    created_at: "2024-01-10",
    updated_at: "2024-01-18",
    vendor: "anne pro",
    tags: ["compact", "portable", "gaming"]
  },
  {
    id: "3",
    name: "Artisan Keycap Set - Cherry Profile",
    description: "Premium PBT keycaps with unique artisan designs",
    price: 4599,
    currency: "INR",
    url: "#",
    config_id: "config3",
    in_stock: false,
    created_at: "2024-01-05",
    updated_at: "2024-01-12",
    vendor: "gmk",
    tags: ["keycaps", "artisan", "cherry"]
  },
  {
    id: "4",
    name: "Gaming Mechanical Switch Tester",
    description: "Test different switch types before buying",
    price: 1299,
    currency: "INR",
    url: "#", 
    config_id: "config4",
    in_stock: true,
    created_at: "2024-01-08",
    updated_at: "2024-01-15",
    vendor: "cherry",
    tags: ["switches", "tester", "sample"]
  }
];

export default function ProductDashboard() {
  const [products, setProducts] = useState<Product[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [activeTab, setActiveTab] = useState("all");

  useEffect(() => {
    // Simulate API call
    const fetchProducts = async () => {
      try {
        // Simulate loading delay
        await new Promise(resolve => setTimeout(resolve, 1000));
        setProducts(mockProducts);
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
  }, []);

  const formatPrice = (price: number, currency: string) => {
    return new Intl.NumberFormat("en-IN", {
      style: "currency",
      currency: currency,
    }).format(price / 100);
  };

  const filteredProducts = products.filter(product => {
    const matchesSearch = product.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         product.vendor.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesTab = activeTab === "all" || 
                      (activeTab === "published" && product.in_stock) ||
                      (activeTab === "draft" && !product.in_stock);
    
    return matchesSearch && matchesTab;
  });

  const totalRevenue = products.reduce((sum, product) => sum + product.price, 0);
  const inStockCount = products.filter(p => p.in_stock).length;

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
                Manage your premium keyboard collection
              </p>
            </div>
            <Button className="bg-blue-600 hover:bg-blue-700">
              <Package className="h-4 w-4 mr-2" />
              Add Product
            </Button>
          </div>

          {/* Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm">
              <CardContent className="p-6">
                <div className="flex items-center gap-4">
                  <div className="h-12 w-12 rounded-full bg-blue-100 flex items-center justify-center">
                    <ShoppingBag className="h-6 w-6 text-blue-600" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Total Products</p>
                    <p className="text-2xl font-bold">{products.length}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
            
            <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm">
              <CardContent className="p-6">
                <div className="flex items-center gap-4">
                  <div className="h-12 w-12 rounded-full bg-green-100 flex items-center justify-center">
                    <TrendingUp className="h-6 w-6 text-green-600" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">In Stock</p>
                    <p className="text-2xl font-bold">{inStockCount}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
            
            <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm">
              <CardContent className="p-6">
                <div className="flex items-center gap-4">
                  <div className="h-12 w-12 rounded-full bg-purple-100 flex items-center justify-center">
                    <DollarSign className="h-6 w-6 text-purple-600" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-muted-foreground">Total Value</p>
                    <p className="text-2xl font-bold">{formatPrice(totalRevenue, "INR")}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
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
                    <TabsTrigger value="all" className="text-xs">All</TabsTrigger>
                    <TabsTrigger value="published" className="text-xs">Published</TabsTrigger>
                    <TabsTrigger value="draft" className="text-xs">Draft</TabsTrigger>
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
                <div className="h-48 bg-gradient-to-br from-slate-100 to-slate-200 relative">
                  <div className="absolute inset-0 bg-gradient-to-br from-blue-500/10 to-purple-500/10" />
                  <div className="absolute top-4 right-4">
                    <Badge 
                      variant={product.in_stock ? "default" : "secondary"} 
                      className={product.in_stock ? "bg-green-500" : ""}
                    >
                      {product.in_stock ? "Published" : "Draft"}
                    </Badge>
                  </div>
                  <div className="absolute bottom-4 left-4 right-4">
                    <div className="flex items-center gap-2 text-sm text-slate-600">
                      <Package className="h-4 w-4" />
                      <span className="capitalize font-medium">{product.vendor}</span>
                    </div>
                  </div>
                </div>
                
                <CardContent className="p-6">
                  <CardTitle className="text-lg leading-tight line-clamp-2 group-hover:text-blue-600 transition-colors mb-2">
                    {product.name}
                  </CardTitle>
                  
                  <div className="flex items-baseline gap-4 mb-4">
                    <div className="text-xl font-bold text-slate-900">
                      {formatPrice(product.price, product.currency)}
                    </div>
                    <div className="flex items-center gap-1 text-sm text-muted-foreground">
                      <Eye className="h-3 w-3" />
                      <span>66</span>
                    </div>
                  </div>
                  
                  {product.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1 mb-4">
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
                  
                  <Button 
                    className="w-full group-hover:shadow-md transition-all duration-300" 
                    onClick={() => window.open(product.url, '_blank')}
                    disabled={!product.in_stock}
                    variant={product.in_stock ? "default" : "secondary"}
                  >
                    <ExternalLink className="h-4 w-4 mr-2" />
                    {product.in_stock ? "View Product" : "Out of Stock"}
                  </Button>
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
                      <div className="h-16 w-16 rounded-lg bg-gradient-to-br from-slate-100 to-slate-200 flex items-center justify-center">
                        <Package className="h-8 w-8 text-slate-500" />
                      </div>
                      
                      <div className="flex-1">
                        <div className="flex items-center gap-3 mb-1">
                          <h3 className="font-semibold text-lg">{product.name}</h3>
                          <Badge 
                            variant={product.in_stock ? "default" : "secondary"}
                            className={product.in_stock ? "bg-green-500" : ""}
                          >
                            {product.in_stock ? "Published" : "Draft"}
                          </Badge>
                        </div>
                        
                        <div className="flex items-center gap-4 text-sm text-muted-foreground mb-2">
                          <span className="capitalize">{product.vendor}</span>
                          <span>•</span>
                          <span>{product.tags.join(", ")}</span>
                        </div>
                        
                        <div className="text-xl font-bold text-slate-900">
                          {formatPrice(product.price, product.currency)}
                        </div>
                      </div>
                      
                      <div className="flex items-center gap-3">
                        <div className="text-center">
                          <div className="text-sm text-muted-foreground">Sales</div>
                          <div className="font-semibold">66</div>
                        </div>
                        <div className="text-center">
                          <div className="text-sm text-muted-foreground">Revenue</div>
                          <div className="font-semibold">{formatPrice(product.price * 66, product.currency)}</div>
                        </div>
                      </div>
                      
                      <Button 
                        onClick={() => window.open(product.url, '_blank')}
                        disabled={!product.in_stock}
                        variant={product.in_stock ? "default" : "secondary"}
                      >
                        <ExternalLink className="h-4 w-4 mr-2" />
                        View
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        )}

        {/* Empty State */}
        {filteredProducts.length === 0 && !loading && (
          <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm">
            <CardContent className="text-center py-16">
              <Package className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-xl font-semibold mb-2">No products found</h3>
              <p className="text-muted-foreground mb-6">
                {searchTerm ? "Try adjusting your search terms" : "Start by adding your first product"}
              </p>
              <Button>
                <Package className="h-4 w-4 mr-2" />
                Add Product
              </Button>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}