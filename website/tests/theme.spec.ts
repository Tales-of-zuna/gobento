import { expect, test, type Page } from '@playwright/test';

/** Records whether <html> had the dark class on the first animation frame, before first paint. */
async function recordFirstFrame(page: Page) {
  await page.addInitScript(() => {
    requestAnimationFrame(() => {
      (window as unknown as { __firstFrameDark: boolean }).__firstFrameDark =
        document.documentElement.classList.contains('dark');
    });
  });
}

const firstFrameDark = (page: Page) =>
  page.waitForFunction(() => {
    const value = (window as unknown as { __firstFrameDark?: boolean }).__firstFrameDark;
    return value === undefined ? undefined : { value };
  });

test.describe('theme', () => {
  test.describe('System preference', () => {
    test.use({ colorScheme: 'dark' });

    test('a first-time visitor whose system prefers dark gets dark with no light flash', async ({
      page,
    }) => {
      await recordFirstFrame(page);
      await page.goto('./');
      expect(await (await firstFrameDark(page)).jsonValue()).toEqual({ value: true });
      await expect(page.locator('html')).toHaveClass(/\bdark\b/);
    });
  });

  test.describe('Explicit choice persists', () => {
    test.use({ colorScheme: 'dark' });

    test('switching to light survives a reload', async ({ page }) => {
      await page.goto('./');
      await expect(page.locator('html')).toHaveClass(/\bdark\b/);
      await page.getByRole('button', { name: 'Switch to light theme' }).click();
      await expect(page.locator('html')).not.toHaveClass(/\bdark\b/);

      await recordFirstFrame(page);
      await page.reload();
      expect(await (await firstFrameDark(page)).jsonValue()).toEqual({ value: false });
      await expect(page.locator('html')).not.toHaveClass(/\bdark\b/);
      await expect(page.getByRole('button', { name: 'Switch to dark theme' })).toBeVisible();
    });
  });

  test('follows the system preference when no choice is stored', async ({ page }) => {
    await page.emulateMedia({ colorScheme: 'light' });
    await page.goto('./');
    await expect(page.locator('html')).not.toHaveClass(/\bdark\b/);
    await page.emulateMedia({ colorScheme: 'dark' });
    await expect(page.locator('html')).toHaveClass(/\bdark\b/);
  });
});
