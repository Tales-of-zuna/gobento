import { expect, test, type Locator, type Page } from '@playwright/test';

/**
 * Progressive enhancement: the server-rendered home page is complete and
 * visible without JavaScript and before hydration (content, LCP, SEO).
 * Animations may only add to it once scripts run.
 */

const SECTION_HEADINGS = [
  "Today's menu, and what comes next.",
  'Three steps, one box.',
  'One option per compartment today.',
  'Compatibility is data, not guesswork.',
  'Built on all three, before you ever run it.',
  'Fill your box.',
];

/** Opacity as painted: the element's own opacity times every ancestor's. */
function paintedOpacity(locator: Locator): Promise<number> {
  return locator.evaluate((element) => {
    let opacity = 1;
    for (let node: Element | null = element; node; node = node.parentElement) {
      opacity *= Number(getComputedStyle(node).opacity);
    }
    return opacity;
  });
}

/** Asserts the element is laid out and painted fully opaque. */
async function expectShown(locator: Locator) {
  await expect(locator).toBeVisible();
  expect(await paintedOpacity(locator)).toBe(1);
}

/** The lowest painted opacity among every reveal item on the page. */
function minRevealOpacity(page: Page): Promise<number> {
  return page.evaluate(() => {
    let min = 1;
    for (const item of document.querySelectorAll('[data-reveal-item]')) {
      let opacity = 1;
      for (let node: Element | null = item; node; node = node.parentElement) {
        opacity *= Number(getComputedStyle(node).opacity);
      }
      min = Math.min(min, opacity);
    }
    return min;
  });
}

test.describe('without JavaScript', () => {
  test.use({ javaScriptEnabled: false, viewport: { width: 1280, height: 800 } });

  test('the hero, quick start and every section heading are visible', async ({ page }) => {
    await page.goto('./');
    await expectShown(
      page.getByRole('heading', { level: 1, name: 'Pack your Go project like a bento.' }),
    );
    await expectShown(page.getByText(/gobento picks versions that are known to work/));
    await expectShown(
      page.locator('main code').filter({ hasText: 'gobento new shop --module example.com/shop' }),
    );
    for (const name of SECTION_HEADINGS) {
      await expectShown(page.getByRole('heading', { level: 2, name }));
    }
  });

  test('a full-page screenshot without scrolling shows every section', async ({
    page,
  }, testInfo) => {
    await page.goto('./');
    expect(await page.locator('[data-reveal-item]').count()).toBeGreaterThan(0);
    expect(await minRevealOpacity(page)).toBe(1);
    // The compartments of every tray are painted, not just their headings.
    for (const title of ['Choose', 'Pack', 'Prove']) {
      await expectShown(page.getByRole('heading', { level: 3, name: title }));
    }
    for (const kind of ['Architecture', 'HTTP framework', 'Database', 'Data access']) {
      await expectShown(page.getByRole('list', { name: `${kind} roadmap` }));
    }
    await expectShown(page.getByRole('img', { name: /^Stamp: passed/ }));
    await expectShown(page.getByRole('list', { name: 'Windows, Go oldstable' }));
    await page.screenshot({
      path: testInfo.outputPath('home-no-js.png'),
      fullPage: true,
      animations: 'disabled',
    });
  });
});

test.describe('with JavaScript', () => {
  test.use({ viewport: { width: 1280, height: 800 } });

  test('sections below the fold stay visible until they scroll into view', async ({ page }) => {
    await page.goto('./');
    await page.waitForLoadState('networkidle');
    // In-view reveals only start as a group enters the window, never hiding it beforehand.
    for (const name of SECTION_HEADINGS.slice(1)) {
      await expectShown(page.getByRole('heading', { level: 2, name }));
    }
    // The closing call to action, far below the fold, is painted as well.
    expect(await paintedOpacity(page.locator('[data-reveal-item]').last())).toBe(1);
  });

  test('the hero settles fully visible after its entrance', async ({ page }) => {
    await page.goto('./');
    await expect(page.locator('html')).toHaveClass(/\bmotion-ready\b/);
    const h1 = page.getByRole('heading', { level: 1 });
    // Wait for the entrance itself to finish (a busy machine may draw it slowly).
    await h1.evaluate(async (element) => {
      const item = element.closest('[data-reveal-item]');
      await Promise.all((item?.getAnimations() ?? []).map((animation) => animation.finished));
    });
    expect(await paintedOpacity(h1)).toBe(1);
  });

  test('the hero heading is painted at once and is the largest contentful paint', async ({
    page,
  }) => {
    await page.goto('./');
    await expect(page.locator('html')).toHaveClass(/\bmotion-ready\b/);
    // The entrance only moves the hero text; it never hides it, even mid-animation.
    expect(await paintedOpacity(page.getByRole('heading', { level: 1 }))).toBe(1);
    const lcp = await page.evaluate(
      () =>
        new Promise<string | undefined>((resolve) => {
          new PerformanceObserver((list) => {
            const entries = list.getEntries() as (PerformanceEntry & { element?: Element })[];
            resolve(entries.at(-1)?.element?.tagName);
          }).observe({ type: 'largest-contentful-paint', buffered: true });
        }),
    );
    expect(lcp).toBe('H1');
  });

  test('reduced motion shows the hero at once, with no entrance', async ({ page }) => {
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await page.goto('./');
    expect(await paintedOpacity(page.getByRole('heading', { level: 1 }))).toBe(1);
    expect(await minRevealOpacity(page)).toBe(1);
  });
});
