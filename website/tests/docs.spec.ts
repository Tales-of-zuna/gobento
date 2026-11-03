import AxeBuilder from '@axe-core/playwright';
import { expect, test, type Page } from '@playwright/test';
import { docsPagePaths, waitForAnimationsToSettle } from './helpers';

/** Every docs page in sidebar order, as the previous/next links walk them. */
const pages = [
  { path: 'docs/', title: 'Introduction', source: 'index.mdx' },
  {
    path: 'docs/getting-started/installation/',
    title: 'Installation',
    source: 'getting-started/installation.mdx',
  },
  {
    path: 'docs/getting-started/quick-start/',
    title: 'Quick start',
    source: 'getting-started/quick-start.mdx',
  },
  {
    path: 'docs/concepts/components/',
    title: 'Components',
    source: 'concepts/components.mdx',
  },
  {
    path: 'docs/concepts/slots-and-contracts/',
    title: 'Slots and contracts',
    source: 'concepts/slots-and-contracts.mdx',
  },
  { path: 'docs/concepts/traits/', title: 'Traits', source: 'concepts/traits.mdx' },
  {
    path: 'docs/concepts/version-rules/',
    title: 'Version rules',
    source: 'concepts/version-rules.mdx',
  },
  { path: 'docs/cli/new/', title: 'gobento new', source: 'cli/new.mdx' },
  { path: 'docs/components/', title: 'Component reference', source: 'components.mdx' },
  {
    path: 'docs/contributing/adding-a-component/',
    title: 'Adding a component',
    source: 'contributing/adding-a-component.mdx',
  },
];

/** Top-level sidebar entries: one per documentation section. */
const sections = [
  'Introduction',
  'Getting started',
  'Concepts',
  'CLI reference',
  'Component reference',
  'Contributing',
];

const sidebar = (page: Page) => page.locator('#nd-sidebar');
/** The previous/next cards at the end of the page. */
const pageFooter = (page: Page) =>
  page.getByRole('navigation', { name: 'Previous and next pages' });
const url = (path: string) => `/gobento/${path}`;

test.describe('Navigation', () => {
  for (const [i, current] of pages.entries()) {
    test(`${current.path} shows every section, the current page, and previous/next links`, async ({
      page,
    }) => {
      await page.goto(current.path);
      await expect(page.getByRole('heading', { level: 1, name: current.title })).toBeVisible();

      // Top-level entries only: folder contents are nested one level deeper.
      const topLevel = sidebar(page).locator(
        '[role="presentation"] > a, [role="presentation"] > div > button',
      );
      await expect(topLevel).toHaveText(sections);

      const active = sidebar(page).locator('a[data-active="true"]');
      await expect(active).toHaveCount(1);
      await expect(active).toHaveAttribute('href', url(current.path));
      await expect(active).toHaveText(current.title);

      const expected = [pages[i - 1], pages[i + 1]].filter((p) => p !== undefined);
      const links = page
        .getByRole('navigation', { name: 'Previous and next pages' })
        .getByRole('link');
      await expect(links).toHaveCount(expected.length);
      for (const [j, neighbour] of expected.entries()) {
        await expect(links.nth(j)).toHaveAttribute('href', url(neighbour.path));
      }
    });
  }

  test('previous and next links open the neighbouring pages', async ({ page }) => {
    await page.goto('docs/concepts/traits/');
    await pageFooter(page)
      .getByRole('link', { name: /^Version rules/ })
      .click();
    await expect(page).toHaveURL(url('docs/concepts/version-rules/'));
    await expect(page.getByRole('heading', { level: 1, name: 'Version rules' })).toBeVisible();

    await pageFooter(page)
      .getByRole('link', { name: /^Traits/ })
      .click();
    await expect(page).toHaveURL(url('docs/concepts/traits/'));
    await expect(page.getByRole('heading', { level: 1, name: 'Traits' })).toBeVisible();
  });
});

test('Searching for "traits" finds and opens the Concepts page on traits', async ({ page }) => {
  await page.goto('docs/');
  const dialog = page.getByRole('dialog', { name: 'Search' });
  // A click that lands before hydration is lost, as it would be for a person: click again.
  await expect(async () => {
    await page.getByRole('button', { name: 'Search' }).first().click();
    await expect(dialog).toBeVisible({ timeout: 1_000 });
  }).toPass();
  await dialog.getByRole('combobox', { name: 'Search' }).fill('traits');
  // Results show the page's breadcrumb (Docs > Concepts) before the title.
  const result = dialog
    .getByRole('option')
    .filter({ hasText: /Concepts\s*Traits/ })
    .first();
  await expect(result).toBeVisible();
  await result.click();
  await expect(page).toHaveURL(/\/gobento\/docs\/concepts\/traits\/(#.*)?$/);
  await expect(page.getByRole('heading', { level: 1, name: 'Traits' })).toBeVisible();
});

test.describe('Suggesting an edit', () => {
  for (const { path, source } of pages) {
    test(`${path} links to its source on GitHub's editor`, async ({ page }) => {
      await page.goto(path);
      await expect(page.getByRole('link', { name: 'Edit this page' })).toHaveAttribute(
        'href',
        `https://github.com/Tales-of-zuna/gobento/edit/main/website/content/docs/${source}`,
      );
    });
  }
});

test('Copy a command puts the block on the clipboard and confirms it', async ({
  page,
  context,
}) => {
  await context.grantPermissions(['clipboard-read', 'clipboard-write']);
  await page.goto('docs/getting-started/installation/');
  const block = page
    .locator('figure')
    .filter({ hasText: 'go install github.com/Tales-of-zuna/gobento' });
  const copy = block.getByRole('button', { name: 'Copy Text' });
  await copy.click();

  await expect(block.getByRole('button', { name: 'Copied Text' })).toBeVisible();
  await expect(block.locator('button svg.lucide-check')).toBeVisible();
  const clipboard = await page.evaluate(() => navigator.clipboard.readText());
  expect(clipboard).toContain('go install github.com/Tales-of-zuna/gobento/cmd/gobento@latest');
});

test('the Components reference is generated from the catalog', async ({ page }) => {
  await page.goto('docs/components/');
  await expect(page.getByRole('heading', { level: 3, name: /\(sqlite\)/ })).toBeVisible();
  await expect(page.getByRole('table', { name: 'Modules of sqlite' })).toContainText(
    'modernc.org/sqlite',
  );
});

test('the accessibility audit covers every docs page', () => {
  expect(docsPagePaths()).toEqual(pages.map((p) => p.path).sort());
});

for (const colorScheme of ['light', 'dark'] as const) {
  test.describe(`Accessibility in the ${colorScheme} theme`, () => {
    test.use({ colorScheme });

    for (const path of docsPagePaths()) {
      test(`${path} has no automated accessibility violations`, async ({ page }) => {
        await page.goto(path);
        await expect(page.locator('html')).toHaveClass(new RegExp(`\\b${colorScheme}\\b`));
        await waitForAnimationsToSettle(page);
        const results = await new AxeBuilder({ page }).analyze();
        expect(results.violations).toEqual([]);
      });
    }
  });
}
