import AxeBuilder from '@axe-core/playwright';
import { expect, test, type Page } from '@playwright/test';

/** Collects console errors and uncaught exceptions for the whole test. */
function collectErrors(page: Page): string[] {
  const errors: string[] = [];
  page.on('console', (message) => {
    if (message.type() === 'error') errors.push(message.text());
  });
  page.on('pageerror', (error) => errors.push(error.message));
  return errors;
}

const mascot = (page: Page) => page.locator('[data-mascot]');
const still = (page: Page) => page.locator('img[data-mascot-still]');

/** The still is shown, has loaded, and no 3D canvas ever appears. */
async function expectStillOnly(page: Page) {
  await page.waitForLoadState('networkidle');
  await expect(still(page)).toBeVisible();
  expect(await still(page).evaluate((img: HTMLImageElement) => img.naturalWidth)).toBeGreaterThan(
    0,
  );
  await expect(mascot(page)).toHaveAttribute('data-mascot', 'still');
  await expect(page.locator('canvas')).toHaveCount(0);
}

test('uses the mascot head as favicon and apple-touch-icon', async ({ page, request }) => {
  await page.goto('./');
  const icon = page.locator('link[rel="icon"][type="image/png"]');
  const apple = page.locator('link[rel="apple-touch-icon"]');
  await expect(icon).toHaveAttribute('href', '/gobento/mascot/gobento-head-64.png');
  await expect(apple).toHaveAttribute('href', '/gobento/mascot/gobento-head-128.png');
  for (const link of [icon, apple]) {
    const response = await request.get(
      new URL((await link.getAttribute('href')) ?? '', page.url()).href,
    );
    expect(response.status()).toBe(200);
    expect(response.headers()['content-type']).toBe('image/png');
  }
});

test.describe('mascot', () => {
  test('swaps the still for the live 3D scene without console errors', async ({ page }) => {
    const errors = collectErrors(page);
    await page.goto('./');
    // The still is in the server-rendered HTML, sized up front.
    await expect(still(page)).toHaveAttribute('width', '720');
    await expect(still(page)).toHaveAttribute('height', '480');
    // The 3D code waits for the visitor: nothing loads while the page itself does.
    await page.waitForLoadState('networkidle');
    await expect(page.locator('canvas')).toHaveCount(0);
    await page.mouse.move(200, 200);
    await expect(mascot(page)).toHaveAttribute('data-mascot', 'scene', { timeout: 30_000 });
    const canvas = page.locator('[data-mascot] canvas');
    await expect(canvas).toBeVisible();
    await expect(canvas.locator('xpath=ancestor::*[@aria-hidden="true"]').first()).toBeAttached();
    await expect(page.getByText(/The gobento mascot:/)).toHaveClass(/sr-only/);
    expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
    expect(errors).toEqual([]);
  });

  test.describe('Reduced motion requested', () => {
    test.use({ reducedMotion: 'reduce' });

    test('shows the mascot in a static pose with no canvas animation', async ({ page }) => {
      const errors = collectErrors(page);
      await page.goto('./');
      await expectStillOnly(page);
      expect(errors).toEqual([]);
    });
  });

  test.describe('WebGL unavailable', () => {
    test.beforeEach(async ({ page }) => {
      await page.addInitScript(() => {
        const getContext = HTMLCanvasElement.prototype.getContext;
        HTMLCanvasElement.prototype.getContext = function (
          this: HTMLCanvasElement,
          id: string,
          ...rest: unknown[]
        ) {
          if (id === 'webgl' || id === 'webgl2' || id === 'experimental-webgl') return null;
          return (getContext as (...args: unknown[]) => RenderingContext | null).call(
            this,
            id,
            ...rest,
          );
        } as typeof getContext;
      });
    });

    test('shows the static illustration and no error', async ({ page }) => {
      const errors = collectErrors(page);
      await page.goto('./');
      await expectStillOnly(page);
      expect(errors).toEqual([]);
    });
  });
});
