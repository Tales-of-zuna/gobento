import { readdirSync } from 'node:fs';
import { join, sep } from 'node:path';
import type { Page } from '@playwright/test';

/**
 * Waits until the page has no finite animation running, so audits such as
 * axe measure settled colors instead of an element halfway through a fade.
 *
 * Infinite decorative loops (the conveyor belt, floating stickers) never
 * finish and are ignored. The page counts as settled once no finite animation
 * has run for a few consecutive frames, which also covers entrance animations
 * that only start after hydration. Gives up after `timeoutMs`.
 */
export async function waitForAnimationsToSettle(page: Page, timeoutMs = 4000): Promise<void> {
  await page.waitForLoadState('load');
  await page.evaluate(async (timeout) => {
    const frame = () => new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));
    const finiteRunning = () =>
      document.getAnimations().some((animation) => {
        const timing = animation.effect?.getComputedTiming();
        return (
          animation.playState === 'running' &&
          timing !== undefined &&
          timing.iterations !== Infinity &&
          timing.endTime !== Infinity
        );
      });
    const deadline = performance.now() + timeout;
    let quietFrames = 0;
    while (quietFrames < 5 && performance.now() < deadline) {
      await frame();
      quietFrames = finiteRunning() ? 0 : quietFrames + 1;
    }
  }, timeoutMs);
}

/** Where the docs pages live; every `.mdx` file there is one page. */
const DOCS_CONTENT = join(import.meta.dirname, '..', 'content', 'docs');

/**
 * The site path of every docs page, derived from the content files the way
 * Fumadocs maps them (`a/b.mdx` and `a/b/index.mdx` are `docs/a/b/`), so a
 * new page is covered without editing a list.
 */
export function docsPagePaths(): string[] {
  return readdirSync(DOCS_CONTENT, { recursive: true, encoding: 'utf8' })
    .filter((file) => file.endsWith('.mdx'))
    .map((file) => {
      const slug = file
        .split(sep)
        .join('/')
        .replace(/\.mdx$/, '')
        .replace(/(^|\/)index$/, '');
      return slug ? `docs/${slug}/` : 'docs/';
    })
    .sort();
}
