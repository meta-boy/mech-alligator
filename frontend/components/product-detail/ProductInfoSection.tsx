import { Star } from 'lucide-react'
import { Badge } from '@/components/ui/badge'

interface ProductInfoSectionProps {
  product: {
    name: string
    brand: string
    category: string
    description: string
  }
}

export default function ProductInfoSection({
  product,
}: ProductInfoSectionProps) {
  const stripHtmlTags = (html: string) => {
    return html.replace(/<[^>]*>/g, '').trim()
  }
  return (
    <div>
      <div className="flex items-center gap-2 mb-2">
        <Badge variant="outline" className="text-xs">
          {product.brand}
        </Badge>
        <Badge variant="secondary" className="text-xs">
          {product.category}
        </Badge>
      </div>
      <h1 className="text-3xl font-bold text-gray-900 mb-4">
        {product.name}
      </h1>
      
      <div className="flex items-center gap-2 mb-4">
        <div className="flex items-center">
          {[...Array(5)].map((_, i) => (
            <Star
              key={i}
              className={`w-4 h-4 ${
                i < 4 ? 'fill-yellow-400 text-yellow-400' : 'text-gray-300'
              }`}
            />
          ))}
        </div>
        <span className="text-sm text-gray-600">4.2 (127 reviews)</span>
      </div>

      <div className="prose prose-sm max-w-none">
        <p className="text-gray-700 leading-relaxed">
          {stripHtmlTags(product.description)}
        </p>
      </div>
    </div>
  )
}
