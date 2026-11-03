import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';

import { waitForAnimationsToSettle } from './helpers';

const QUICK_START = 'gobento new shop --module example.com/shop';
const released =
  JSON.parse(
    readFileSync(join(import.meta.dirname, '..', '..', '.release-please-manifest.json'), 'utf8'),
  )['.'] !== '0.0.0';

test.describe('home page', () => {
  test.describe('First visit', () => {
    test.use({ viewport: { width: 1280, height: 800 } });

    test('shows the name, value proposition, quick start and links without scrolling', async ({
      page,
    }) => {
      await page.goto('./');
      const header = page.getByRole('banner');
      await expect(header.getByRole('link', { name: 'gobento' })).toBeInViewport();
      await expect(
        page.getByRole('heading', { level: 1, name: 'Pack your Go project like a bento.' }),
      ).toBeInViewport();
      await expect(
        page.getByText(/gobento picks versions that are known to work/),
      ).toBeInViewport();
      await expect(page.locator('main code').filter({ hasText: QUICK_START })).toBeInViewport();
      await expect(page.getByRole('button', { name: 'Copy quick-start command' })).toBeInViewport();
      const nav = page.getByRole('navigation', { name: 'Main' });
      await expect(nav.getByRole('link', { name: 'Docs' })).toBeInViewport();
      await expect(nav.getByRole('link', { name: 'Configurator' })).toBeInViewport();
      await expect(nav.getByRole('link', { name: 'GitHub' })).toHaveAttribute(
        'href',
        'https://github.com/Tales-of-zuna/gobento',
      );
      await expect(nav.getByRole('link', { name: 'GitHub' })).toBeInViewport();
      await expect(
        page.getByRole('link', { name: 'Open the configurator' }).first(),
      ).toBeInViewport();
      await expect(page.getByRole('link', { name: 'Read the docs' })).toBeInViewport();
    });
  });

  test('Copy quick start: puts the command on the clipboard and confirms', async ({
    page,
    context,
  }) => {
    await context.grantPermissions(['clipboard-read', 'clipboard-write']);
    await page.goto('./');
    await page.getByRole('button', { name: 'Copy quick-start command' }).click();
    await expect(page.getByRole('button', { name: 'Copied ✓' })).toBeVisible();
    await expect(page.getByRole('status')).toHaveText('Copied to clipboard');
    expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(QUICK_START);
  });

  test('No release yet: shows the early-development notice while unreleased', async ({ page }) => {
    await page.goto('./');
    const notice = page.getByText('Early development · not released yet');
    const planned = page.getByText('Planned interface. There is nothing to install yet.');
    if (released) {
      await expect(notice).toHaveCount(0);
      await expect(planned).toHaveCount(0);
    } else {
      await expect(notice).toBeVisible();
      await expect(planned).toBeVisible();
    }
  });

  test('Footer credits: attribution with license link and non-affiliation', async ({ page }) => {
    await page.goto('./');
    const footer = page.getByRole('contentinfo');
    await expect(footer).toContainText(
      'The gobento mascot is an original character inspired by the Go gopher, designed by Renée French and licensed under CC BY 4.0. gobento is not affiliated with or endorsed by the Go project.',
    );
    await expect(footer.getByRole('link', { name: 'CC BY 4.0' })).toHaveAttribute(
      'href',
      'https://creativecommons.org/licenses/by/4.0/',
    );
    for (const [name, file] of [
      ['Code of Conduct', 'CODE_OF_CONDUCT.md'],
      ['Security', 'SECURITY.md'],
    ]) {
      await expect(footer.getByRole('link', { name })).toHaveAttribute(
        'href',
        `https://github.com/Tales-of-zuna/gobento/blob/main/${file}`,
      );
    }
    await expect(footer).toContainText('MIT licensed');
  });

  test('lists the catalog components on the belt and the Go rule from the catalog', async ({
    page,
  }) => {
    await page.goto('./');
    const belt = page.getByRole('list', { name: 'Components' });
    for (const name of ['MVC', 'net/http', 'SQLite', 'sqlc']) {
      await expect(belt.getByText(name, { exact: true }).first()).toBeAttached();
    }
    await expect(
      page.getByText('modernc.org/sqlite v1.61.1 needs Go 1.26 or newer'),
    ).toBeAttached();
  });

  test.describe('mobile', () => {
    test.use({ viewport: { width: 390, height: 844 }, hasTouch: true });

    test('opens the menu with the site links and closes it with Escape', async ({ page }) => {
      await page.goto('./');
      const menuButton = page.getByRole('button', { name: 'Menu' });
      await expect(menuButton).toHaveAttribute('aria-expanded', 'false');
      const box = await menuButton.boundingBox();
      expect(box!.width).toBeGreaterThanOrEqual(44);
      expect(box!.height).toBeGreaterThanOrEqual(44);
      await menuButton.click();
      const menu = page.getByRole('navigation', { name: 'Mobile' });
      await expect(menu.getByRole('link', { name: 'Docs' })).toBeVisible();
      await expect(menu.getByRole('link', { name: 'Configurator' })).toBeVisible();
      await page.keyboard.press('Escape');
      await expect(menu).toBeHidden();
      await expect(menuButton).toBeFocused();
    });

    test('has no horizontal scroll', async ({ page }) => {
      await page.goto('./');
      expect(
        await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
      ).toBe(true);
    });
  });

  for (const colorScheme of ['light', 'dark'] as const) {
    test(`Automated audit: axe reports no violations (${colorScheme})`, async ({ page }) => {
      await page.emulateMedia({ colorScheme });
      await page.goto('./');
      // Play every in-view reveal, then audit the settled page.
      for (let y = 0; y < (await page.evaluate(() => document.body.scrollHeight)); y += 400) {
        await page.mouse.wheel(0, 400);
        await page.waitForTimeout(60);
      }
      await waitForAnimationsToSettle(page);
      const results = await new AxeBuilder({ page }).analyze();
      expect(results.violations).toEqual([]);
    });
  }
});
