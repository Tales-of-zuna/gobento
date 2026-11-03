// Renders the static stills of the mascot from the R3F scene itself, so the
// fallback always matches the 3D version.
//
// Serves scripts/mascot-stills/ with Vite, opens it in headless Chromium,
// renders each view at twice the largest still's size, and writes the
// transparent PNG and WebP stills listed in components/mascot/stills.ts to
// public/mascot/.
//
// Usage: pnpm mascot:stills [--design <dir>]
//   --design <dir>  also write large transparent PNGs for the design project:
//                   mascot-tq.png, mascot-front.png, mascot-head.png
import { mkdir, writeFile } from 'node:fs/promises';
import { resolve } from 'node:path';
import { parseArgs } from 'node:util';
import { chromium } from '@playwright/test';
import react from '@vitejs/plugin-react';
import { createViteServer } from 'vitest/node';

const website = resolve(import.meta.dirname, '..');
const outDir = resolve(website, 'public', 'mascot');
const { values: args } = parseArgs({ options: { design: { type: 'string' } } });
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
const { ALL_STILLS } = await server.ssrLoadModule(
  resolve(website, 'components', 'mascot', 'stills.ts'),
);

/** Opens the render page for `view` at `width`x`height` CSS pixels, device pixel ratio 2. */
async function open(browser, view, width, height) {
  const page = await browser.newPage({ viewport: { width, height }, deviceScaleFactor: 2 });
  page.on('pageerror', (error) => console.error('page error:', error));
  await page.goto(`http://127.0.0.1:${address.port}/?view=${view}`);
  await page.waitForFunction(() => window.mascotReady === true, null, { timeout: 60_000 });
  return page;
}

/** Writes the canvas of `page`, scaled to `width`x`height`, to `file`. */
async function save(page, width, height, type, file) {
  const dataUrl = await page.evaluate(
    ([w, h, t]) => window.exportStill(w, h, t),
    [width, height, type],
  );
  await writeFile(file, Buffer.from(dataUrl.split(',')[1], 'base64'));
  console.log(`wrote ${file}`);
}

const browser = await chromium.launch();
try {
  await mkdir(outDir, { recursive: true });
  for (const view of ['hero', 'head']) {
    const stills = ALL_STILLS.filter((still) => still.view === view);
    const largest = stills.reduce((a, b) => (b.width > a.width ? b : a));
    const page = await open(browser, view, largest.width, largest.height);
    for (const still of stills) {
      for (const { ext, type } of formats) {
        await save(page, still.width, still.height, type, resolve(outDir, `${still.name}.${ext}`));
      }
    }
    await page.close();
  }
  if (args.design) {
    const dir = resolve(args.design);
    await mkdir(dir, { recursive: true });
    for (const [view, name, width, height] of [
      ['hero', 'mascot-tq', 1200, 800],
      ['front', 'mascot-front', 1200, 800],
      ['head', 'mascot-head', 512, 512],
    ]) {
      const page = await open(browser, view, width, height);
      await save(page, width, height, 'image/png', resolve(dir, `${name}.png`));
      await page.close();
    }
  }
} finally {
  await browser.close();
  await server.close();
}
