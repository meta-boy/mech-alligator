import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Search, X } from "lucide-react";
import { useState, KeyboardEvent, useEffect } from "react";
import { AuthManager } from "@/utils/auth";
import { FilterOptions, AppliedFilters } from "@/lib/types";

interface ProductFiltersProps {
  searchTerm: string;
  onSearchTermChange: (term: string) => void;
  onSearch: (term: string) => void;
  onFiltersChange: (filters: AppliedFilters) => void;
  isSearching?: boolean;
}

export const ProductFilters = ({ 
  searchTerm, 
  onSearchTermChange,
  onSearch,
  onFiltersChange,
  isSearching = false
}: ProductFiltersProps) => {
  const [inputValue, setInputValue] = useState(searchTerm);
  const [filterOptions, setFilterOptions] = useState<FilterOptions | null>(null);
  const [appliedFilters, setAppliedFilters] = useState<AppliedFilters>({});
  const [loadingFilters, setLoadingFilters] = useState(true);

  // Fetch filter options on component mount
  useEffect(() => {
    const fetchFilterOptions = async () => {
      try {
        const authManager = AuthManager.getInstance();
        const response = await authManager.makeAuthenticatedRequest('/api/products/filter-options');
        
        if (response.ok) {
          const data: FilterOptions = await response.json();
          setFilterOptions(data);
        }
      } catch (error) {
        console.error('Error fetching filter options:', error);
      } finally {
        setLoadingFilters(false);
      }
    };

    fetchFilterOptions();
  }, []);

  const handleInputChange = (value: string) => {
    setInputValue(value);
    onSearchTermChange(value);
  };

  const handleSearch = () => {
    onSearch(inputValue);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  const handleFilterChange = (filterType: keyof AppliedFilters, value: string) => {
    const newFilters = {
      ...appliedFilters,
      [filterType]: value === 'all' ? undefined : value
    };
    setAppliedFilters(newFilters);
    onFiltersChange(newFilters);
  };

  const clearFilters = () => {
    setAppliedFilters({});
    onFiltersChange({});
  };

  const hasActiveFilters = Object.values(appliedFilters).some(value => value !== undefined);

  if (loadingFilters) {
    return (
      <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm mb-6">
        <CardContent className="p-6">
          <div className="flex items-center justify-center py-8">
            <div className="text-sm text-muted-foreground">Loading filters...</div>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm mb-6">
      <CardContent className="p-6">
        <div className="space-y-4">
          {/* Search Bar */}
          <div className="flex items-center gap-4">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Search products..."
                value={inputValue}
                onChange={(e) => handleInputChange(e.target.value)}
                onKeyDown={handleKeyDown}
                className="pl-10 pr-20 border-0 bg-slate-50"
              />
              <Button
                onClick={handleSearch}
                size="sm"
                disabled={isSearching}
                className="absolute right-1 top-1/2 transform -translate-y-1/2 h-8 px-3"
              >
                {isSearching ? "..." : "Search"}
              </Button>
            </div>
          </div>

          {/* Filter Dropdowns */}
          {filterOptions && (
            <div className="flex flex-wrap gap-4 items-center">
              {/* Brand Filter */}
              <div className="min-w-[160px]">
                <Select
                  value={appliedFilters.brand || 'all'}
                  onValueChange={(value) => handleFilterChange('brand', value)}
                >
                  <SelectTrigger className="bg-slate-50 border-0">
                    <SelectValue placeholder="All Brands" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Brands</SelectItem>
                    {filterOptions.brands.map((brand) => (
                      <SelectItem key={brand} value={brand}>
                        {brand}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Reseller Filter */}
              <div className="min-w-[160px]">
                <Select
                  value={appliedFilters.reseller || 'all'}
                  onValueChange={(value) => handleFilterChange('reseller', value)}
                >
                  <SelectTrigger className="bg-slate-50 border-0">
                    <SelectValue placeholder="All Resellers" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Resellers</SelectItem>
                    {filterOptions.resellers.map((reseller) => (
                      <SelectItem key={reseller} value={reseller}>
                        {reseller}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Category Filter */}
              <div className="min-w-[160px]">
                <Select
                  value={appliedFilters.category || 'all'}
                  onValueChange={(value) => handleFilterChange('category', value)}
                >
                  <SelectTrigger className="bg-slate-50 border-0">
                    <SelectValue placeholder="All Categories" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Categories</SelectItem>
                    {filterOptions.categories.map((category) => (
                      <SelectItem key={category} value={category}>
                        {category.replace(/_/g, ' ').toLowerCase().replace(/\b\w/g, l => l.toUpperCase())}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Sort Field Filter */}
              <div className="min-w-[160px]">
                <Select
                  value={appliedFilters.sort_field || 'name'}
                  onValueChange={(value) => handleFilterChange('sort_field', value)}
                >
                  <SelectTrigger className="bg-slate-50 border-0">
                    <SelectValue placeholder="Sort By" />
                  </SelectTrigger>
                  <SelectContent>
                    {filterOptions.sort_fields.map((field) => (
                      <SelectItem key={field} value={field}>
                        {field.charAt(0).toUpperCase() + field.slice(1)}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Sort Order Filter */}
              <div className="min-w-[120px]">
                <Select
                  value={appliedFilters.sort_order || 'asc'}
                  onValueChange={(value) => handleFilterChange('sort_order', value)}
                >
                  <SelectTrigger className="bg-slate-50 border-0">
                    <SelectValue placeholder="Order" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="asc">Ascending</SelectItem>
                    <SelectItem value="desc">Descending</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {/* Clear Filters Button */}
              {hasActiveFilters && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={clearFilters}
                  className="h-10 px-3"
                >
                  <X className="h-4 w-4 mr-1" />
                  Clear Filters
                </Button>
              )}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
};