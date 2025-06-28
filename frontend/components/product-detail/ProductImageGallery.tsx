'use client'

import Image from 'next/image'

interface ProductImageGalleryProps {
  currentImages: string[]
  selectedImageIndex: number
  setSelectedImageIndex: (index: number) => void
  productName: string
}

export default function ProductImageGallery({
  currentImages,
  selectedImageIndex,
  setSelectedImageIndex,
  productName,
}: ProductImageGalleryProps) {
  const mainImage = currentImages[selectedImageIndex] || currentImages[0] || '/placeholder.jpg'

  return (
    <div className="space-y-4">
      <div className="relative aspect-square overflow-hidden rounded-xl border bg-gray-50">
        <Image
          src={mainImage}
          alt={productName}
          fill
          className="object-cover"
          priority
        />
      </div>
      
      {currentImages.length > 1 && (
        <div className="flex gap-2 overflow-x-auto pb-2">
          {currentImages.map((image, index) => (
            <button
              key={index}
              onClick={() => setSelectedImageIndex(index)}
              className={`relative flex-shrink-0 w-20 h-20 rounded-md border-2 overflow-hidden transition-all ${
                selectedImageIndex === index
                  ? 'border-primary ring-2 ring-primary/20'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              <Image
                src={image}
                alt={`${productName} - ${index + 1}`}
                fill
                className="object-cover"
              />
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
