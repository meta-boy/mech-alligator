interface ProductVariant {
  price: number
  currency: string
}

interface ProductPriceDisplayProps {
  selectedVariant: ProductVariant | null
  formatPrice: (price: number, currency: string) => string
}

export default function ProductPriceDisplay({
  selectedVariant,
  formatPrice,
}: ProductPriceDisplayProps) {
  return (
    <div className="bg-gray-50 rounded-lg p-4">
      <div className="flex items-baseline gap-2">
        <span className="text-3xl font-bold text-primary">
          {selectedVariant ? formatPrice(selectedVariant.price, selectedVariant.currency) : 'N/A'}
        </span>
        <span className="text-sm text-gray-600">per unit</span>
      </div>
    </div>
  )
}
