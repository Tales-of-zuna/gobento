/**
 * Pre-rendered stills of the mascot in its static pose, generated from the
 * R3F scene by `pnpm mascot:stills` into `public/mascot/`.
 */
import { basePath } from '@/lib/site';

/** Pixel sizes rendered: 1x and 2x of the hero slot, plus a favicon. */
export const STILL_SIZES = [480, 960, 64] as const;
export type StillSize = (typeof STILL_SIZES)[number];
export type StillFormat = 'webp' | 'png';

/** CSS size of the hero still (its 1x render). */
export const HERO_STILL_SIZE = 480;

/** File name of a still, as written by the generator. */
export function stillFileName(size: StillSize, format: StillFormat): string {
  return `gobento-mascot-${size}.${format}`;
}

/** Base-path-aware URL of a still, for `<img>` and `<source>`. */
export function stillUrl(size: StillSize, format: StillFormat): string {
  return `${basePath}/mascot/${stillFileName(size, format)}`;
}

/** `srcset` with the 1x and 2x hero stills in one format. */
export function heroSrcSet(format: StillFormat): string {
  return `${stillUrl(480, format)} 1x, ${stillUrl(960, format)} 2x`;
}

/** Text alternative for the mascot, used wherever it appears. */
export const MASCOT_DESCRIPTION =
  'The gobento mascot: a sky-blue gopher whose body is a red bento box filled with rice, tamagoyaki, edamame, and an onigiri, waving hello.';
