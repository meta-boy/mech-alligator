'use client'

import { Link, Heart } from 'lucide-react'
import { Button } from '@/components/ui/button'

interface ProductVariant {
  available: boolean
  url: string
}

interface ProductActionButtonsProps {
  selectedVariant: ProductVariant | null
  isWishlisted: boolean
  setIsWishlisted: (isWishlisted: boolean) => void
  productUrl?: string
}

export default function ProductActionButtons({
  selectedVariant,
  isWishlisted,
  setIsWishlisted,
  productUrl,
}: ProductActionButtonsProps) {
  const handleVisitStore = () => {
    const url = selectedVariant?.url || productUrl
    if (url) {
      window.open(url, '_blank')
    }
  }

  return (
    <div className="flex gap-3">
      <Button 
        size="lg" 
        className="flex-1 h-12"
        disabled={!selectedVariant?.available}
        onClick={handleVisitStore}
      >
        <Link className="w-5 h-5 mr-2" />
        {selectedVariant?.available ? 'Visit the store' : 'Out of Stock'}
      </Button>
      <Button
        size="lg"
        variant="outline"
        onClick={() => setIsWishlisted(!isWishlisted)}
        className="h-12 px-4"
      >
        <Heart className={`w-5 h-5 ${isWishlisted ? 'fill-red-500 text-red-500' : ''}`} />
      </Button>
    </div>
  )
}
