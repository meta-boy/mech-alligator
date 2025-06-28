
import { Card, CardContent, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ExternalLink, Package } from "lucide-react";
import { Product } from "@/lib/types";
import { ImageGallery } from "./ImageGallery";

interface ProductCardProps {
  product: Product;
}

export const ProductCard = ({ product }: ProductCardProps) => (
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
          <ExternalLink className="h-3 w-3" />
        </Button>
      </div>
        <div className="absolute bottom-4 left-4">
          <div className="flex items-center gap-2 text-sm text-white drop-shadow-lg bg-black/30 px-2 py-1 rounded-md backdrop-blur-sm w-fit">
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
);