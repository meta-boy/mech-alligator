import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { X } from "lucide-react";
import { FilterOptions, AppliedFilters } from "@/lib/types";

interface FilterPopupProps {
  filterOptions: FilterOptions;
  pendingFilters: AppliedFilters;
  onFilterChange: (filterType: keyof AppliedFilters, value: string) => void;
  onApplyFilters: () => void;
  onClearFilters: () => void;
  onClose: () => void;
  hasActiveFilters: boolean;
  onSelectOpenChange?: (open: boolean) => void;
}

export const FilterPopup = ({
  filterOptions,
  pendingFilters,
  onFilterChange,
  onApplyFilters,
  onClearFilters,
  onClose,
  hasActiveFilters,
  onSelectOpenChange
}: FilterPopupProps) => {
  const activeFilterCount = Object.values(pendingFilters).filter(value => value !== undefined).length;

  return (
    <div 
      className="absolute top-full right-0 mt-2 z-50 w-80 max-w-sm"
      onClick={(e) => e.stopPropagation()}
    >
      <Card className="border shadow-lg bg-white">
        <CardContent className="p-6">
          <div className="space-y-4">
            {/* Header */}
            <div className="flex items-center justify-between border-b pb-3">
              <div className="flex items-center gap-2">
                <h3 className="font-semibold text-gray-900">Filters</h3>
                {activeFilterCount > 0 && (
                  <Badge variant="secondary" className="text-xs">
                    {activeFilterCount}
                  </Badge>
                )}
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={onClose}
                className="h-6 w-6 p-0"
              >
                <X className="h-4 w-4" />
              </Button>
            </div>

            {/* Active Filters */}
            {hasActiveFilters && (
              <div className="space-y-2">
                <div className="text-xs font-medium text-gray-600 uppercase tracking-wide">
                  Active Filters
                </div>
                <div className="flex flex-wrap gap-2">
                  {Object.entries(pendingFilters).map(([key, value]) => {
                    if (!value) return null;
                    return (
                      <Badge key={key} variant="outline" className="text-xs">
                        {key}: {value}
                        <button
                          onClick={() => onFilterChange(key as keyof AppliedFilters, 'all')}
                          className="ml-1 hover:bg-gray-200 rounded"
                        >
                          <X className="h-3 w-3" />
                        </button>
                      </Badge>
                    );
                  })}
                </div>
              </div>
            )}

            {/* Filter Options */}
            <div className="space-y-4">
              {/* Brand Filter */}
              <div className="space-y-2">
                <label className="text-xs font-medium text-gray-600 uppercase tracking-wide">
                  Brand
                </label>
                <Select
                  value={pendingFilters.brand || 'all'}
                  onValueChange={(value) => onFilterChange('brand', value)}
                  onOpenChange={onSelectOpenChange}
                >
                  <SelectTrigger className="h-9 text-sm">
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
              <div className="space-y-2">
                <label className="text-xs font-medium text-gray-600 uppercase tracking-wide">
                  Reseller
                </label>
                <Select
                  value={pendingFilters.reseller || 'all'}
                  onValueChange={(value) => onFilterChange('reseller', value)}
                  onOpenChange={onSelectOpenChange}
                >
                  <SelectTrigger className="h-9 text-sm">
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
              <div className="space-y-2">
                <label className="text-xs font-medium text-gray-600 uppercase tracking-wide">
                  Category
                </label>
                <Select
                  value={pendingFilters.category || 'all'}
                  onValueChange={(value) => onFilterChange('category', value)}
                  onOpenChange={onSelectOpenChange}
                >
                  <SelectTrigger className="h-9 text-sm">
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

              {/* Sort Section */}
              <div className="space-y-3 border-t pt-3">
                <div className="text-xs font-medium text-gray-600 uppercase tracking-wide">
                  Sort Options
                </div>
                
                <div className="grid grid-cols-2 gap-2">
                  <div className="space-y-1">
                    <label className="text-xs text-gray-500">Sort By</label>
                    <Select
                      value={pendingFilters.sort_field || 'name'}
                      onValueChange={(value) => onFilterChange('sort_field', value)}
                      onOpenChange={onSelectOpenChange}
                    >
                      <SelectTrigger className="h-8 text-xs">
                        <SelectValue />
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

                  <div className="space-y-1">
                    <label className="text-xs text-gray-500">Order</label>
                    <Select
                      value={pendingFilters.sort_order || 'asc'}
                      onValueChange={(value) => onFilterChange('sort_order', value)}
                      onOpenChange={onSelectOpenChange}
                    >
                      <SelectTrigger className="h-8 text-xs">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="asc">Asc</SelectItem>
                        <SelectItem value="desc">Desc</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex gap-2 pt-4 border-t">
              {hasActiveFilters && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onClearFilters}
                  className="flex-1 h-9 text-xs"
                >
                  Clear All
                </Button>
              )}
              <Button
                onClick={onApplyFilters}
                size="sm"
                className="flex-1 h-9 text-xs"
              >
                Apply Filters
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};
