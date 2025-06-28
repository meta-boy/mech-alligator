import { Badge } from '@/components/ui/badge'

interface ProductTagsProps {
  tags: string[]
}

export default function ProductTags({ tags }: ProductTagsProps) {
  if (tags.length === 0) {
    return null
  }

  return (
    <div className="pt-4 border-t">
      <h4 className="text-sm font-medium text-gray-900 mb-2">Tags:</h4>
      <div className="flex flex-wrap gap-2">
        {tags.slice(0, 6).map((tag) => (
          <Badge key={tag} variant="outline" className="text-xs">
            {tag}
          </Badge>
        ))}
      </div>
    </div>
  )
}
