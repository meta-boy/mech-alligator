import type { MetadataRoute } from 'next'

export default function robots(): MetadataRoute.Robots {
  const siteOrigin = (process.env.SITE_ORIGIN || 'https://agg.regator.site').replace(/\/$/, '')

  return {
    rules: [
      {
        userAgent: '*',
        allow: ['/'],
        disallow: ['/api', '/api/*', '/_next', '/_next/*', '/.next', '/.next/*'],
      },
      {
        userAgent: 'Googlebot',
        allow: ['/'],
        crawlDelay: 5,
      },
    ],
    sitemap: `${siteOrigin}/sitemap.xml`,
  }
}