import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { expect, test } from '@playwright/test';

/**
 * The shared site chrome (bar and footer) and the pre-release notice, on
 * every kind of page.
 */

const released =
  JSON.parse(
    readFileSync(join(import.meta.dirname, '..', '..', '.release-please-manifest.json'), 'utf8'),
  )['.'] !== '0.0.0';

const PAGES = [
  { name: 'home', path: './', active: undefined },
  { name: 'configurator', path: 'configurator/', active: 'Configurator' },
  { name: 'docs index', path: 'docs/', active: 'Docs' },
  { name: 'docs page', path: 'docs/concepts/traits/', active: 'Docs' },
  { name: 'components reference', path: 'docs/components/', active: 'Docs' },
] as const;

for (const { name, path, active } of PAGES) {
  test.describe(`${name}`, () => {
    test('has the site bar with the current section marked', async ({ page }) => {
      await page.goto(path);
      const bar = page.locator('header[data-site-nav]');
      await expect(bar).toHaveCount(1);
      await expect(bar.getByRole('link', { name: 'gobento' })).toBeVisible();
      const nav = bar.getByRole('navigation', { name: 'Main' });
      await expect(nav.getByRole('link', { name: 'GitHub' })).toHaveAttribute(
        'href',
        'https://github.com/Tales-of-zuna/gobento',
      );
      await expect(bar.locator('[data-theme-toggle]')).toBeVisible();
      const current = nav.locator('[aria-current="page"]');
      if (active) {
        await expect(current).toHaveCount(1);
        await expect(current).toHaveText(active);
      } else {
        await expect(current).toHaveCount(0);
      }
    });

    test('has the site footer with the mascot credit', async ({ page }) => {
      await page.goto(path);
      const footer = page.getByRole('contentinfo');
      await expect(footer).toHaveCount(1);
      await expect(footer).toContainText('not affiliated with or endorsed by the Go project');
      await expect(footer.getByRole('link', { name: 'CC BY 4.0' })).toBeAttached();
    });

    test('Pre-release notice: shows while nothing is released', async ({ page }) => {
      await page.goto(path);
      const notice = page.locator('[data-prerelease-notice]');
      if (released) {
        await expect(notice).toHaveCount(0);
      } else {
        await expect(notice).toHaveCount(1);
        await expect(notice).toBeVisible();
        await expect(notice).toHaveText('Early development · not released yet');
      }
    });
  });
}

test.describe('docs chrome', () => {
  test('the bar sticks above the sidebar, which stays below it', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 });
    // A long page, scrolled well short of the footer (which ends the sticky range).
    await page.goto('docs/concepts/slots-and-contracts/');
    const bar = page.locator('header[data-site-nav]');
    const sidebar = page.locator('#nd-sidebar');
    await expect(sidebar).toBeVisible();
    await page.mouse.wheel(0, 800);
    await expect.poll(() => page.evaluate(() => window.scrollY)).toBeGreaterThan(400);
    const barBox = (await bar.boundingBox())!;
    const sideBox = (await sidebar.boundingBox())!;
    expect(barBox.y).toBe(0);
    expect(sideBox.y).toBeGreaterThanOrEqual(barBox.height - 1);
  });

  test('Fumadocs keeps its search; the theme switch and GitHub link are not doubled', async ({
    page,
  }) => {
    await page.goto('docs/');
    await expect(page.locator('#nd-sidebar [data-search-full]')).toBeVisible();
    await expect(page.locator('[data-theme-toggle]')).toHaveCount(1);
    await expect(page.locator('a[href="https://github.com/Tales-of-zuna/gobento"]')).toHaveCount(
      // the bar, its mobile menu, and the footer
      3,
    );
  });

  test.describe('mobile', () => {
    test.use({ viewport: { width: 390, height: 844 }, hasTouch: true });

    test('the site menu and the docs drawer both open', async ({ page }) => {
      await page.goto('docs/concepts/traits/');
      const menuButton = page.getByRole('button', { name: 'Menu' });
      await expect(async () => {
        await menuButton.click();
        await expect(menuButton).toHaveAttribute('aria-expanded', 'true', { timeout: 1_000 });
      }).toPass();
      const menu = page.getByRole('navigation', { name: 'Mobile' });
      await expect(menu.getByRole('link', { name: 'Docs' })).toHaveAttribute(
        'aria-current',
        'page',
      );
      await page.keyboard.press('Escape');
      await expect(menu).toBeHidden();

      await page.locator('#nd-subnav').getByRole('button', { name: 'Open Sidebar' }).click();
      await expect(page.locator('#nd-sidebar-mobile')).toBeVisible();
    });

    test('has no horizontal scroll', async ({ page }) => {
      for (const path of ['docs/concepts/traits/', 'configurator/']) {
        await page.goto(path);
        expect(
          await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth),
          path,
        ).toBe(true);
      }
    });
  });
});
