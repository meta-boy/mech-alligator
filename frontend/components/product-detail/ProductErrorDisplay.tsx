import { Card } from '@/components/ui/card'

interface ProductErrorDisplayProps {
  error: string | null
}

export default function ProductErrorDisplay({ error }: ProductErrorDisplayProps) {
  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      <Card className="p-8 text-center">
        <h2 className="text-2xl font-bold text-red-600 mb-4">Error Loading Product</h2>
        <p className="text-gray-600">{error || 'An unknown error occurred'}</p>
      </Card>
    </div>
  )
}
