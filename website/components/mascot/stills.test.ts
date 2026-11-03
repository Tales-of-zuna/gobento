import { existsSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { ALL_STILLS, headStill, heroSrcSet, HERO_STILLS, stillFileName, stillUrl } from './stills';

describe('stills', () => {
  it('renders the hero landscape at 1x and 2x', () => {
    expect(HERO_STILLS.map(({ width, height }) => [width, height])).toEqual([
      [720, 480],
      [1440, 960],
    ]);
  });

  it('builds URLs under the GitHub Pages base path', () => {
    expect(stillUrl(headStill(64), 'png')).toBe('/gobento/mascot/gobento-head-64.png');
    expect(heroSrcSet('webp')).toBe(
      '/gobento/mascot/gobento-mascot-720x480.webp 1x, /gobento/mascot/gobento-mascot-1440x960.webp 2x',
    );
  });
});

describe('generated stills', () => {
  it.each(ALL_STILLS.map((still) => [still.name, still] as const))(
    '%s exists in public/mascot as WebP and PNG',
    (_, still) => {
      for (const format of ['webp', 'png'] as const) {
        const file = join(process.cwd(), 'public', 'mascot', stillFileName(still, format));
        expect(existsSync(file), `${file} (run pnpm mascot:stills)`).toBe(true);
      }
    },
  );
});
