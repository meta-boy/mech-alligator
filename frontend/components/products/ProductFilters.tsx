import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Search } from "lucide-react";

interface ProductFiltersProps {
  searchTerm: string;
  onSearchTermChange: (term: string) => void;
  activeTab: string;
  onActiveTabChange: (tab: string) => void;
}

export const ProductFilters = ({ 
  searchTerm, 
  onSearchTermChange, 
  activeTab, 
  onActiveTabChange 
}: ProductFiltersProps) => (
  <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm mb-6">
    <CardContent className="p-6">
      <div className="flex flex-col sm:flex-row gap-4 items-center justify-between">
        <div className="flex items-center gap-4 w-full sm:w-auto">
          <div className="relative flex-1 sm:w-80">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search products..."
              value={searchTerm}
              onChange={(e) => onSearchTermChange(e.target.value)}
              className="pl-10 border-0 bg-slate-50"
            />
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