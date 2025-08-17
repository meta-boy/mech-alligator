import { API_BASE_URL } from '@/utils/constants';

type SitemapEntry = {
  url: string;
  lastModified?: string;
};

/**
 * Server-side sitemap generator for Next.js App Router.
 * - Authenticates to the backend using API_USERNAME/API_PASSWORD
 * - Fetches products (paginated) from backend and adds each product page
 * - Falls back to only the root URL if authentication or fetch fails
 */
export default async function sitemap(): Promise<SitemapEntry[]> {
  const siteOrigin = (process.env.SITE_ORIGIN || 'https://agg.regator.site').replace(/\/$/, '');

  const entries: SitemapEntry[] = [
    { url: siteOrigin + '/' },
  ];

  // Require backend credentials to include dynamic product pages.
  const username = process.env.API_USERNAME;
  const password = process.env.API_PASSWORD;
  if (!username || !password) {
    return entries;
  }

  try {
    // Login to backend to obtain a token
    const loginRes = await fetch(`${API_BASE_URL}/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, password }),
      cache: 'no-store',
    });

    if (!loginRes.ok) return entries;

    const loginData = await loginRes.json();
    const token = loginData?.token;
    if (!token) return entries;

    // Fetch products with pagination. Use a large page_size but handle multiple pages.
    const pageSize = 1000;
    let page = 1;
    const allProducts: any[] = [];

    while (true) {
      const productsRes = await fetch(`${API_BASE_URL}/products?page=${page}&page_size=${pageSize}`, {
        headers: { Authorization: `Bearer ${token}` },
        cache: 'no-store',
      });

      if (!productsRes.ok) break;

      const data = await productsRes.json();
      const products = data?.products || [];
      allProducts.push(...products);

      const pagination = data?.pagination;
      if (!pagination || !pagination.has_next) break;
      page += 1;
    }

    for (const p of allProducts) {
      const id = p?.id;
      if (!id) continue;
      const last = p?.source_metadata?.updated_at || p?.source_metadata?.published_at;
      entries.push({
        url: `${siteOrigin}/${encodeURIComponent(String(id))}`,
        lastModified: last ? new Date(last).toISOString() : undefined,
      });
    }
  } catch (err) {
    // on any failure return at least the root entry
    return entries;
  }

  return entries;
}
