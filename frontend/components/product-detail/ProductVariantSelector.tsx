'use client'

import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Label } from '@/components/ui/label'

interface ProductVariant {
  id: string
  name: string
  price: number
  currency: string
}

interface ProductVariantSelectorProps {
  variants: ProductVariant[]
  selectedVariant: ProductVariant | null
  handleVariantChange: (variantId: string) => void
  formatPrice: (price: number, currency: string) => string
}

export default function ProductVariantSelector({
  variants,
  selectedVariant,
  handleVariantChange,
  formatPrice,
}: ProductVariantSelectorProps) {
  if (variants.length <= 1) {
    return null
  }

  return (
    <div className="space-y-3">
      <Label className="text-base font-medium">Choose Variant:</Label>
      <RadioGroup
        value={selectedVariant?.id || ''}
        onValueChange={handleVariantChange}
        className="grid grid-cols-1 gap-3"
      >
        {variants.map((variant) => (
          <div key={variant.id} className="flex items-center space-x-2">
            <RadioGroupItem value={variant.id} id={variant.id} />
            <Label
              htmlFor={variant.id}
              className="flex-1 cursor-pointer flex justify-between items-center p-3 border rounded-lg hover:bg-gray-50"
            >
              <span className="font-medium">{variant.name}</span>
              <span className="font-bold text-primary">
                {formatPrice(variant.price, variant.currency)}
              </span>
            </Label>
          </div>
        ))}
      </RadioGroup>
    </div>
  )
}
