import { expect, test, type Page } from '@playwright/test';

/**
 * The 3D mascot on a software rasterizer (no GPU, a blocklisted driver, a VM):
 * there, every frame the scene draws blocks the page's main thread, and so
 * its input, until the frame is drawn. These tests force Chromium's software
 * rasterizer whatever GPU the machine has.
 */
test.use({ launchOptions: { args: ['--use-angle=swiftshader'] } });

test.beforeEach(async ({ page }) => {
  // Count WebGL draw calls, so a test can tell whether frames are drawn.
  await page.addInitScript(() => {
    const counter = window as unknown as { __webglDraws: number };
    counter.__webglDraws = 0;
    for (const proto of [WebGLRenderingContext.prototype, WebGL2RenderingContext.prototype]) {
      for (const name of ['drawArrays', 'drawElements'] as const) {
        const draw = proto[name] as (...args: unknown[]) => void;
        Object.defineProperty(proto, name, {
          configurable: true,
          writable: true,
          value(this: unknown, ...args: unknown[]) {
            counter.__webglDraws++;
            return draw.apply(this, args);
          },
        });
      }
    }
  });
});

const draws = (page: Page) =>
  page.evaluate(() => (window as unknown as { __webglDraws: number }).__webglDraws);

/**
 * Interacts so the scene loads, waits until the mascot shows it, then lets
 * its first frames and cross-fade settle.
 */
async function showScene(page: Page) {
  await page.waitForLoadState('networkidle');
  await page.mouse.move(200, 200);
  await expect(page.locator('[data-mascot]').first()).toHaveAttribute('data-mascot', 'scene', {
    timeout: 30_000,
  });
  await page.waitForTimeout(1_000);
  expect(await draws(page)).toBeGreaterThan(0);
}

test('the home mascot holds a static pose that draws nothing while idle', async ({ page }) => {
  const errors: string[] = [];
  page.on('console', (message) => {
    if (message.type() === 'error') errors.push(message.text());
  });
  page.on('pageerror', (error) => errors.push(error.message));
  await page.goto('./');
  await showScene(page);
  const settled = await draws(page);
  await page.waitForTimeout(1_000);
  expect(await draws(page)).toBe(settled);
  expect(errors).toEqual([]);
});

test('typing on the configurator neither draws the scene nor waits on it', async ({ page }) => {
  await page.goto('configurator/?arch=mvc&framework=nethttp&db=sqlite&data=sqlc');
  await expect(page.locator('[data-hydrated="true"]')).toBeAttached();
  await showScene(page);
  await page.getByLabel('Module path').focus();
  const settled = await draws(page);
  const start = Date.now();
  await page.keyboard.type('example.com/shop');
  expect(Date.now() - start).toBeLessThan(2_000);
  await expect(page.locator('[data-command]').locator('visible=true')).toContainText(
    '--module example.com/shop',
  );
  expect(await draws(page)).toBe(settled);
});
