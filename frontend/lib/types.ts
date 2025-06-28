export interface Product {
  id: string;
  name: string;
  description: string;
  handle: string;
  url: string;
  brand: string;
  reseller: string;
  category: string;
  tags: string[];
  images: string[];
  variant_count: number;
  source_type: string;
  source_id: string;
  reseller_id: string;
  source_metadata: {
    created_at: string;
    published_at: string;
    shopify_handle: string;
    shopify_product_type: string;
    shopify_vendor: string;
    updated_at: string;
  };
}

export interface PaginationData {
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface ProductResponse {
  products: Product[];
  pagination: PaginationData;
}
