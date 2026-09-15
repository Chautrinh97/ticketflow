/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  images: {
    // Event banners are manual URLs in Phase 1 (no File Service yet) —
    // allow any https host rather than an exhausting per-CDN allowlist.
    remotePatterns: [{ protocol: 'https', hostname: '**' }],
  },
}

module.exports = nextConfig
