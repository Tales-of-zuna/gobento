import type { NextConfig } from 'next';
import { createMDX } from 'fumadocs-mdx/next';
import { basePath } from './lib/site';

const withMDX = createMDX();

/** Static export for GitHub Pages, served under {@link basePath}. */
const config: NextConfig = {
  output: 'export',
  basePath,
  trailingSlash: true,
  images: { unoptimized: true },
  reactStrictMode: true,
  // three ships modern ESM that React Three Fiber expects to be compiled with the app.
  transpilePackages: ['three'],
};

export default withMDX(config);
