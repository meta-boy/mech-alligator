import { NextRequest, NextResponse } from 'next/server';
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
    
    // Build the API URL with pagination parameters
    const apiUrl = new URL(`${API_BASE_URL}/products`);
    apiUrl.searchParams.set('page', page);
    apiUrl.searchParams.set('page_size', pageSize);

    const response = await fetch(apiUrl.toString(), {
      headers: {
        'Authorization': authHeader,
      },
    });

    if (!response.ok) {
      return NextResponse.json({ message: 'Failed to fetch products' }, { status: response.status });
    }

    const products = await response.json();
    return NextResponse.json(products);
  } catch (error) {
    console.error('Products API error:', error);
    return NextResponse.json({ message: 'Internal server error' }, { status: 500 });
  }
}
