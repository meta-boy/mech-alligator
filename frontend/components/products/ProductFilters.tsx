import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Search, Filter } from "lucide-react";
import { useState, KeyboardEvent, useEffect, useRef } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { AuthManager } from "@/utils/auth";
import { FilterOptions, AppliedFilters } from "@/lib/types";
import { FilterPopup } from "./FilterPopup";

interface ProductFiltersProps {
  onSearch: (term: string) => void;
  onFiltersChange: (filters: AppliedFilters) => void;
  isSearching?: boolean;
}

export const ProductFilters = ({ 
  onSearch,
  onFiltersChange,
  isSearching = false
}: ProductFiltersProps) => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const filterButtonRef = useRef<HTMLButtonElement>(null);

  const initialSearchTerm = searchParams.get('search') || '';
  const initialFilters: AppliedFilters = {};
  searchParams.forEach((value, key) => {
    if (key !== 'search') {
      initialFilters[key as keyof AppliedFilters] = value;
    }
  });

  const [inputValue, setInputValue] = useState(initialSearchTerm);
  const [filterOptions, setFilterOptions] = useState<FilterOptions | null>(null);
  const [pendingFilters, setPendingFilters] = useState<AppliedFilters>(initialFilters);
  const [loadingFilters, setLoadingFilters] = useState(true);
  const [showFilterPopup, setShowFilterPopup] = useState(false);

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

  // Close popup when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (filterButtonRef.current && !filterButtonRef.current.contains(event.target as Node)) {
        const filterPopup = document.querySelector('[data-filter-popup]');
        if (filterPopup && !filterPopup.contains(event.target as Node)) {
          setShowFilterPopup(false);
        }
      }
    };

    if (showFilterPopup) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [showFilterPopup]);

  const handleInputChange = (value: string) => {
    setInputValue(value);
  };

  const applyFilters = () => {
    const params = new URLSearchParams();
    if (inputValue) {
      params.set('search', inputValue);
    }
    Object.entries(pendingFilters).forEach(([key, value]) => {
      if (value !== undefined) {
        params.set(key, value);
      }
    });
    router.push(`/?${params.toString()}`);
    onSearch(inputValue);
    onFiltersChange(pendingFilters);
    setShowFilterPopup(false);
  };

  const handleSearch = () => {
    const params = new URLSearchParams();
    if (inputValue) {
      params.set('search', inputValue);
    }
    Object.entries(pendingFilters).forEach(([key, value]) => {
      if (value !== undefined) {
        params.set(key, value);
      }
    });
    router.push(`/?${params.toString()}`);
    onSearch(inputValue);
    onFiltersChange(pendingFilters);
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleSearch();
    }
  };

  const handleFilterChange = (filterType: keyof AppliedFilters, value: string) => {
    setPendingFilters(prevFilters => ({
      ...prevFilters,
      [filterType]: value === 'all' ? undefined : value
    }));
  };

  const clearFilters = () => {
    setInputValue('');
    setPendingFilters({});
    router.push('/');
    onSearch('');
    onFiltersChange({});
    setShowFilterPopup(false);
  };

  const hasActiveFilters = Object.values(pendingFilters).some(value => value !== undefined) || inputValue !== '';
  const activeFilterCount = Object.values(pendingFilters).filter(value => value !== undefined).length;

  if (loadingFilters) {
    return (
      <div className="flex items-center justify-center py-8 mb-6">
        <div className="text-sm text-muted-foreground">Loading filters...</div>
      </div>
    );
  }

  return (
    <div className="mb-6">
      <div className="flex items-center gap-3 max-w-2xl">
        {/* Search Bar */}
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search products..."
            value={inputValue}
            onChange={(e) => handleInputChange(e.target.value)}
            onKeyDown={handleKeyDown}
            className="pl-10 h-11 bg-white border-gray-200 shadow-sm"
          />
        </div>

        {/* Search Button */}
        <Button
          onClick={handleSearch}
          disabled={isSearching}
          className="h-11 px-6"
        >
          {isSearching ? "..." : "Search"}
        </Button>

        {/* Filter Button */}
        <div className="relative">
          <Button
            ref={filterButtonRef}
            variant="outline"
            onClick={() => setShowFilterPopup(!showFilterPopup)}
            className="h-11 px-4 border-gray-200 bg-white hover:bg-gray-50"
          >
            <Filter className="h-4 w-4 mr-2" />
            Filters
            {activeFilterCount > 0 && (
              <Badge variant="secondary" className="ml-2 h-5 w-5 p-0 text-xs">
                {activeFilterCount}
              </Badge>
            )}
          </Button>

          {/* Filter Popup */}
          {showFilterPopup && filterOptions && (
            <div data-filter-popup>
              <FilterPopup
                filterOptions={filterOptions}
                pendingFilters={pendingFilters}
                onFilterChange={handleFilterChange}
                onApplyFilters={applyFilters}
                onClearFilters={clearFilters}
                onClose={() => setShowFilterPopup(false)}
                hasActiveFilters={hasActiveFilters}
              />
            </div>
          )}
        </div>
      </div>

      {/* Active Filters Display */}
      {hasActiveFilters && (
        <div className="mt-3 flex flex-wrap gap-2">
          {inputValue && (
            <Badge variant="outline" className="text-xs bg-blue-50 border-blue-200">
              Search: `{inputValue}`
            </Badge>
          )}
          {Object.entries(pendingFilters).map(([key, value]) => {
            if (!value) return null;
            return (
              <Badge key={key} variant="outline" className="text-xs bg-gray-50 border-gray-200">
                {key.replace('_', ' ')}: {value}
              </Badge>
            );
          })}
        </div>
      )}
    </div>
  );
};