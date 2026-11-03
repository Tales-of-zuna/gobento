/**
 * Pre-rendered stills of the mascot in its static pose, generated from the
 * R3F scene by `pnpm mascot:stills` into `public/mascot/`. The generator
 * reads this module, so these lists are the single source of truth.
 */
import { basePath } from '@/lib/site';

export type StillFormat = 'webp' | 'png';

/** A rendered still: which camera view, its pixel size, and its file stem. */
export interface Still {
  readonly view: 'hero' | 'head';
  readonly width: number;
  readonly height: number;
  readonly name: string;
}

/** CSS size of the landscape hero still (its 1x render). */
export const HERO_SIZE = { width: 720, height: 480 } as const;

function heroStill(scale: number): Still {
  const width = HERO_SIZE.width * scale;
  const height = HERO_SIZE.height * scale;
  return { view: 'hero', width, height, name: `gobento-mascot-${width}x${height}` };
}

/** The hero at 1x and 2x. */
export const HERO_STILLS = [heroStill(1), heroStill(2)] as const;

/** The round head alone, for the favicon, avatars, and social images. */
export const HEAD_STILLS = [64, 128, 256].map((size): Still => ({
  view: 'head',
  width: size,
  height: size,
  name: `gobento-head-${size}`,
}));

export const ALL_STILLS: readonly Still[] = [...HERO_STILLS, ...HEAD_STILLS];

/** File name of a still, as written by the generator. */
export function stillFileName(still: Still, format: StillFormat): string {
  return `${still.name}.${format}`;
}

/** Base-path-aware URL of a still, for `<img>`, `<source>`, and metadata. */
export function stillUrl(still: Still, format: StillFormat): string {
  return `${basePath}/mascot/${stillFileName(still, format)}`;
}

/**
 * `srcset` with the 1x and 2x hero stills in one format, by pixel width, so
 * the browser picks by the displayed size (`sizes`) and a phone does not
 * download the 2x render.
 */
export function heroSrcSet(format: StillFormat): string {
  return HERO_STILLS.map((still) => `${stillUrl(still, format)} ${still.width}w`).join(', ');
}

/** The head still of a given pixel size. */
export function headStill(size: 64 | 128 | 256): Still {
  const still = HEAD_STILLS.find((s) => s.width === size);
  if (!still) throw new Error(`no head still of ${size}px`);
  return still;
}

/** Text alternative for the mascot, used wherever it appears. */
export const MASCOT_DESCRIPTION =
  'The gobento mascot: a sky-blue gopher sitting in an open red bento box next to rice, an onigiri, tamagoyaki, and edamame, with the lid leaning behind it.';
