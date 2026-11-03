// Renders the static stills of the mascot from the R3F scene itself, so the
// fallback always matches the 3D version.
//
// Serves scripts/mascot-stills/ with Vite, opens it in headless Chromium at
// 960x960 CSS pixels (device pixel ratio 2), and writes transparent PNG and
// WebP stills at each size in STILL_SIZES to public/mascot/.
//
// Usage: pnpm mascot:stills
import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { chromium } from '@playwright/test';
import react from '@vitejs/plugin-react';
import { createViteServer } from 'vitest/node';

const website = resolve(import.meta.dirname, '..');
const outDir = resolve(website, 'public', 'mascot');
/** Keep in sync with STILL_SIZES in components/mascot/stills.ts. */
const sizes = [480, 960, 64];
const formats = [
  { ext: 'png', type: 'image/png' },
  { ext: 'webp', type: 'image/webp' },
];

const server = await createViteServer({
  configFile: false,
  root: resolve(website, 'scripts', 'mascot-stills'),
  plugins: [react()],
  resolve: { alias: { '@': website } },
  define: { 'process.env.NODE_ENV': JSON.stringify('production') },
  server: { host: '127.0.0.1', port: 0, fs: { allow: [website] } },
  logLevel: 'warn',
});
await server.listen();
const address = server.httpServer?.address();
if (!address || typeof address === 'string') throw new Error('Vite did not report a port');

const browser = await chromium.launch();
try {
  const page = await browser.newPage({
    viewport: { width: 960, height: 960 },
    deviceScaleFactor: 2,
  });
  page.on('pageerror', (error) => console.error('page error:', error));
  await page.goto(`http://127.0.0.1:${address.port}/`);
  await page.waitForFunction(() => window.mascotReady === true, null, { timeout: 60_000 });
  await mkdir(outDir, { recursive: true });
  for (const size of sizes) {
    for (const { ext, type } of formats) {
      const dataUrl = await page.evaluate(([s, t]) => window.exportStill(s, t), [size, type]);
      const file = resolve(outDir, `gobento-mascot-${size}.${ext}`);
      await writeFile(file, Buffer.from(dataUrl.split(',')[1], 'base64'));
      console.log(`wrote ${file}`);
    }
  }
} finally {
  await browser.close();
  await server.close();
}
