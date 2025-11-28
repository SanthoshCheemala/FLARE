import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  reactCompiler: true,
  distDir: process.env.DIST_DIR || ".next",
  env: {
    NEXT_PUBLIC_APP_MODE: process.env.NEXT_PUBLIC_APP_MODE,
    API_URL: process.env.API_URL,
  },
  async rewrites() {
    const apiUrl = process.env.API_URL || "http://localhost:8080";
    console.log(`[Next.js] Proxying /api to ${apiUrl}`);
    return [
      {
        source: "/api/:path*",
        destination: `${apiUrl}/:path*`,
      },
    ];
  },
};

export default nextConfig;
