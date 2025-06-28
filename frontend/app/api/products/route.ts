import { NextRequest, NextResponse } from 'next/server';

export const dynamic = 'force-dynamic';
import { API_BASE_URL } from '@/utils/constants';

export async function GET(req: NextRequest) {
  try {
    const authHeader = req.headers.get('authorization');

    if (!authHeader) {
      return NextResponse.json({ message: 'No authentication token' }, { status: 401 });
    }

    // Get query parameters
    const { searchParams } = new URL(req.url);
    const page = searchParams.get('page') || '1';
    const pageSize = searchParams.get('page_size') || '20';
    const search = searchParams.get('search') || '';
    const brand = searchParams.get('brand') || '';
    const reseller = searchParams.get('reseller') || '';
    const category = searchParams.get('category') || '';
    const sortField = searchParams.get('sort_field') || '';
    const sortOrder = searchParams.get('sort_order') || '';
    
    // Build the API URL with all parameters
    const apiUrl = new URL(`${API_BASE_URL}/products`);
    apiUrl.searchParams.set('page', page);
    apiUrl.searchParams.set('page_size', pageSize);
    if (search) {
      apiUrl.searchParams.set('search', search);
    }
    if (brand) {
      apiUrl.searchParams.set('brand', brand);
    }
    if (reseller) {
      apiUrl.searchParams.set('reseller', reseller);
    }
    if (category) {
      apiUrl.searchParams.set('category', category);
    }
    if (sortField) {
      apiUrl.searchParams.set('sort_field', sortField);
    }
    if (sortOrder) {
      apiUrl.searchParams.set('sort_order', sortOrder);
    }

    const response = await fetch(apiUrl.toString(), {
      headers: {
        'Authorization': authHeader,
      },
    });

    if (!response.ok) {
      console.log(response.url);
      console.error('Failed to fetch products:', response.status, response.statusText);
      return NextResponse.json({ message: 'Failed to fetch products' }, { status: response.status });
    }

    const products = await response.json();
    return NextResponse.json(products);
  } catch (error) {
    console.error('Products API error:', error);
    return NextResponse.json({ message: 'Internal server error' }, { status: 500 });
  }
}
