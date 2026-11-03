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
};

export default withMDX(config);
