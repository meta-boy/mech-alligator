import { NextRequest, NextResponse } from 'next/server';
import { API_BASE_URL } from '@/utils/constants';

export async function GET(
  req: NextRequest,
  { params }: { params: { id: string } }
) {
  try {
    const authHeader = req.headers.get('authorization');

    if (!authHeader) {
      return NextResponse.json({ message: 'No authentication token' }, { status: 401 });
    }

    const productId = params.id;

    // Build the API URL for specific product
    const apiUrl = new URL(`${API_BASE_URL}/products/${productId}`);

    const response = await fetch(apiUrl.toString(), {
      headers: {
        'Authorization': authHeader,
      },
    });

    if (!response.ok) {
      console.log(response.url);
      console.error('Failed to fetch product:', response.status, response.statusText);
      return NextResponse.json({ message: 'Failed to fetch product' }, { status: response.status });
    }

    const product = await response.json();
    return NextResponse.json(product);
  } catch (error) {
    console.error('Product API error:', error);
    return NextResponse.json({ message: 'Internal server error' }, { status: 500 });
  }
}
