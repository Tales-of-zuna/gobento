import AxeBuilder from '@axe-core/playwright';
import { expect, test, type Page } from '@playwright/test';

/**
 * Records every request that answered with an HTTP error or never got a
 * response. Chromium reports Next's prefetch HEAD probes as aborted once the
 * response arrives unread, so a failure that did receive a successful
 * response is not a broken request.
 */
function collectFailedRequests(page: Page): () => Promise<string[]> {
  const failed: Promise<string | undefined>[] = [];
  page.on('response', (response) => {
    if (response.status() >= 400) {
      failed.push(Promise.resolve(`${response.url()} (${response.status()})`));
    }
  });
  page.on('requestfailed', (request) => {
    failed.push(
      request
        .response()
        .then((response) =>
          response ? undefined : `${request.url()} (${request.failure()?.errorText})`,
        ),
    );
  });
  return async () => (await Promise.all(failed)).filter((entry) => entry !== undefined);
}

const pages = [
  { name: 'home', path: './', heading: 'gobento' },
  { name: 'docs', path: 'docs/', heading: 'Introduction' },
];

for (const { name, path, heading } of pages) {
  test.describe(`${name} page`, () => {
    test('loads under /gobento/ with every request succeeding', async ({ page }) => {
      const failedRequests = collectFailedRequests(page);
      await page.goto(path);
      await expect(page.getByRole('heading', { level: 1, name: heading })).toBeVisible();
      await page.waitForLoadState('networkidle');
      expect(await failedRequests()).toEqual([]);
    });

    test('maps the shadcn theme onto Fumadocs', async ({ page }, testInfo) => {
      await page.goto(path);
      // shadcn components use --primary; Fumadocs' primary must be mapped onto it.
      const colors = await page.evaluate(() => {
        const root = getComputedStyle(document.documentElement);
        return {
          primary: root.getPropertyValue('--primary').trim(),
          fdPrimary: root.getPropertyValue('--color-fd-primary').trim(),
        };
      });
      expect(colors.primary).not.toBe('');
      expect(colors.fdPrimary).toBe(colors.primary);
      await page.screenshot({ path: testInfo.outputPath(`${name}.png`), fullPage: true });
    });

    test('has no automated accessibility violations', async ({ page }) => {
      await page.goto(path);
      const results = await new AxeBuilder({ page }).analyze();
      expect(results.violations).toEqual([]);
    });
  });
}

test('renders the shadcn Button with the shared theme', async ({ page }) => {
  await page.goto('./');
  const button = page.locator('[data-slot="button"]').first();
  await expect(button).toBeVisible();
  expect(await button.evaluate((el) => getComputedStyle(el).backgroundColor)).not.toBe(
    'rgba(0, 0, 0, 0)',
  );
});

test('static search loads its index from under the base path', async ({ page }) => {
  await page.goto('docs/');
  const dialog = page.getByRole('dialog', { name: 'Search' });
  // A click that lands before hydration is lost, as it would be for a person: click again.
  await expect(async () => {
    await page.getByRole('button', { name: 'Search' }).first().click();
    await expect(dialog).toBeVisible({ timeout: 1_000 });
  }).toPass();
  const index = page.waitForResponse((response) => response.url().endsWith('/gobento/api/search'));
  await dialog.getByRole('combobox', { name: 'Search' }).fill('go-blueprint');
  expect((await index).status()).toBe(200);
  await expect(dialog.getByRole('option').first()).toContainText('Introduction');
});
