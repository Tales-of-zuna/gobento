import { existsSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { heroSrcSet, STILL_SIZES, stillFileName, stillUrl } from './stills';

describe('stills', () => {
  it('builds URLs under the GitHub Pages base path', () => {
    expect(stillUrl(480, 'webp')).toBe('/gobento/mascot/gobento-mascot-480.webp');
    expect(heroSrcSet('png')).toBe(
      '/gobento/mascot/gobento-mascot-480.png 1x, /gobento/mascot/gobento-mascot-960.png 2x',
    );
  });
});

describe('generated stills', () => {
  it('exist in public/mascot for every size and format', () => {
    for (const size of STILL_SIZES) {
      for (const format of ['webp', 'png'] as const) {
        const file = join(process.cwd(), 'public', 'mascot', stillFileName(size, format));
        expect(existsSync(file), `${file} (run pnpm mascot:stills)`).toBe(true);
      }
    }
  });
});
