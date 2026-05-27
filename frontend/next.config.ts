import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  reactStrictMode: true,
  async rewrites() {
    const backend = process.env.NEXT_PUBLIC_BACKEND_ORIGIN || 'http://localhost:8080';
    return [
      {
        source: '/api/:path*',
        destination: `${backend}/:path*`,
      },
    ];
  },
};

export default nextConfig;
