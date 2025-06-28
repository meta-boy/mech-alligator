import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Search } from "lucide-react";
import { useState, KeyboardEvent } from "react";

interface ProductFiltersProps {
  searchTerm: string;
  onSearchTermChange: (term: string) => void;
  onSearch: (term: string) => void;
  activeTab: string;
  onActiveTabChange: (tab: string) => void;
  isSearching?: boolean;
}

export const ProductFilters = ({ 
  searchTerm, 
  onSearchTermChange,
  onSearch,
  activeTab, 
  onActiveTabChange,
  isSearching = false
}: ProductFiltersProps) => {
  const [inputValue, setInputValue] = useState(searchTerm);

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

  return (
    <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm mb-6">
      <CardContent className="p-6">
        <div className="flex flex-col sm:flex-row gap-4 items-center justify-between">
          <div className="flex items-center gap-4 w-full sm:w-auto">
            <div className="relative flex-1 sm:w-80">
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

          <div className="flex items-center gap-2">
            <Tabs value={activeTab} onValueChange={onActiveTabChange}>
              <TabsList className="bg-slate-100">
                <TabsTrigger value="all" className="text-xs">All Products</TabsTrigger>
                <TabsTrigger value="published" className="text-xs">Keycaps</TabsTrigger>
                <TabsTrigger value="draft" className="text-xs">Keyboards</TabsTrigger>
              </TabsList>
            </Tabs>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};