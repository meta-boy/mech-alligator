'use client'

import { useState, useEffect } from 'react'
import { formatPrice } from '@/lib/formatters'
import ProductLoadingSkeleton from '@/components/product-detail/ProductLoadingSkeleton'
import ProductErrorDisplay from '@/components/product-detail/ProductErrorDisplay'
import ProductImageGallery from '@/components/product-detail/ProductImageGallery'
import ProductInfoSection from '@/components/product-detail/ProductInfoSection'
import ProductVariantSelector from '@/components/product-detail/ProductVariantSelector'
import ProductPriceDisplay from '@/components/product-detail/ProductPriceDisplay'
import ProductActionButtons from '@/components/product-detail/ProductActionButtons'
import ProductTags from '@/components/product-detail/ProductTags'
import { AuthManager } from '@/utils/auth'

interface ProductVariant {
  id: string
  product_id: string
  name: string
  price: number
  currency: string
  available: boolean
  url: string
  images: string[]
  options: {
    option1: string
  }
  source_id: string
}

interface Product {
  id: string
  name: string
  description: string
  handle: string
  url: string
  brand: string
  reseller: string
  category: string
  tags: string[]
  images: string[]
  variants: ProductVariant[]
  variant_count: number
  source_type: string
  source_id: string
  reseller_id: string
}

interface ProductDetailPageProps {
  params: { product: string }
}

export default function ProductDetailPage({ params }: ProductDetailPageProps) {
  const productId = params.product
  const [product, setProduct] = useState<Product | null>(null)
  const [selectedVariant, setSelectedVariant] = useState<ProductVariant | null>(null)
  const [selectedImageIndex, setSelectedImageIndex] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isWishlisted, setIsWishlisted] = useState(false)

  useEffect(() => {
    const fetchProduct = async () => {
      try {
        const authManager = AuthManager.getInstance();
        const url = `/api/products/${productId}`;
        const response = await authManager.makeAuthenticatedRequest(url);
        
        if (!response.ok) {
          throw new Error('Failed to fetch product')
        }
        
        const productData: Product = await response.json()
        setProduct(productData)
        setSelectedVariant(productData.variants[0] || null)
        setLoading(false)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred')
        setLoading(false)
      }
    }

    fetchProduct()
  }, [productId])

  const handleVariantChange = (variantId: string) => {
    const variant = product?.variants.find(v => v.id === variantId)
    if (variant) {
      setSelectedVariant(variant)
      // Update main image to variant's first image if available
      if (variant.images.length > 0) {
        const variantImageIndex = product?.images.findIndex(img => img === variant.images[0])
        if (variantImageIndex !== -1) {
          setSelectedImageIndex(variantImageIndex || 0)
        }
      }
    }
  }

  

  if (loading) {
    return <ProductLoadingSkeleton />
  }

  if (error || !product) {
    return <ProductErrorDisplay error={error || 'Product not found'} />
  }

  const currentImages = product.images || []

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
        {/* Product Images */}
        <ProductImageGallery
          currentImages={currentImages}
          selectedImageIndex={selectedImageIndex}
          setSelectedImageIndex={setSelectedImageIndex}
          productName={product.name}
        />

        {/* Product Info */}
        <div className="space-y-6">
          <ProductInfoSection
            product={product}
          />

          {/* Variant Selection */}
          <ProductVariantSelector
            variants={product.variants}
            selectedVariant={selectedVariant}
            handleVariantChange={handleVariantChange}
            formatPrice={formatPrice}
          />

          {/* Price */}
          <ProductPriceDisplay
            selectedVariant={selectedVariant}
            formatPrice={formatPrice}
          />

          {/* Action Buttons */}
          <ProductActionButtons
            selectedVariant={selectedVariant}
            isWishlisted={isWishlisted}
            setIsWishlisted={setIsWishlisted}
            productUrl={product.url}
          />

          {/* Tags */}
          <ProductTags tags={product.tags} />
        </div>
      </div>
    </div>
  )
}
