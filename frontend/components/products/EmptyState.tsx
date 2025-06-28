import { Card, CardContent } from "@/components/ui/card";
import { Package } from "lucide-react";

interface EmptyStateProps {
  searchTerm: string;
}

export const EmptyState = ({ searchTerm }: EmptyStateProps) => (
  <Card className="border-0 shadow-sm bg-white/80 backdrop-blur-sm">
    <CardContent className="text-center py-16">
      <Package className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
      <h3 className="text-xl font-semibold mb-2">No products found</h3>
      <p className="text-muted-foreground mb-6">
        {searchTerm ? "Try adjusting your search terms" : "No products available at the moment"}
      </p>
    </CardContent>
  </Card>
);