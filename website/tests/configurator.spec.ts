import AxeBuilder from '@axe-core/playwright';
import { expect, test, type Page } from '@playwright/test';

import { releaseStatus } from './helpers';

const { released } = releaseStatus();

/**
 * Opens the configurator and waits until it has read the URL (after
 * hydration). The architecture kind has several options, so tests start from
 * mvc unless they pass their own query.
 */
async function open(page: Page, query = '?arch=mvc') {
  await page.goto(`configurator/${query}`);
  await expect(page.locator('[data-hydrated="true"]')).toBeAttached();
}

const command = (page: Page) => page.locator('[data-command]').locator('visible=true');
const group = (page: Page, name: string) => page.getByRole('group', { name });

test.describe('configurator', () => {
  test.use({ permissions: ['clipboard-read', 'clipboard-write'] });

  test('Defaults: single options and default versions are preselected', async ({ page }) => {
    await open(page, '');
    await expect(page.getByRole('heading', { level: 1, name: 'Pack your project' })).toBeVisible();
    // The pre-release notice follows the release manifest (spec "Pre-release notice").
    const notice = page.getByText('Early development · not released yet');
    if (released) await expect(notice).toHaveCount(0);
    else await expect(notice).toBeVisible();
    for (const [kind, option] of [
      ['HTTP framework', 'net/http'],
      ['Database', 'SQLite'],
      ['Data access', 'sqlc'],
      ['Go version', 'Go 1.27'],
    ]) {
      await expect(group(page, kind!).getByRole('radio', { name: option })).toBeChecked();
    }
    await expect(group(page, 'HTTP framework').getByRole('radio', { name: 'gin' })).toBeDisabled();
    // Several architectures: none is preselected, and the CLI needs --arch.
    const architecture = group(page, 'Architecture');
    await expect(architecture.getByRole('radio', { checked: true })).toHaveCount(0);
    await expect(
      architecture.getByText('Choose an architecture: mvc, vertical-slice.'),
    ).toBeVisible();
    await expect(
      page.getByText('No command yet. Fix the architecture choice').locator('visible=true'),
    ).toBeVisible();
    await architecture.getByText('Model-View-Controller', { exact: true }).click();
    await expect(command(page)).toHaveText('gobento new myapp --arch mvc');
    await page.getByRole('button', { name: /Advanced/ }).click();
    await expect(page.getByLabel('modernc.org/sqlite')).toHaveValue('v1.61.1');
    await expect(page.getByText('Tool versions are fixed per gobento release.')).toBeVisible();
  });

  test('Changing a choice updates the command with defaults omitted', async ({ page }) => {
    await open(page);
    await page.getByLabel('Project name').fill('shop');
    await page.getByLabel('Module path').fill('example.com/shop');
    await expect(command(page)).toHaveText('gobento new shop --module example.com/shop --arch mvc');
    await group(page, 'Go version').getByText('Go 1.26').click();
    await expect(command(page)).toHaveText(
      'gobento new shop --module example.com/shop --arch mvc --go 1.26',
    );
    const files = page.getByRole('list', { name: 'Generated files' });
    await expect(files).toContainText('cmd/shop/main.go');
    await expect(files).toContainText('controllers/');
    await group(page, 'Architecture').getByText('Vertical slice', { exact: true }).click();
    await expect(command(page)).toHaveText(
      'gobento new shop --module example.com/shop --arch vertical-slice --go 1.26',
    );
    await expect(files).toContainText('items/itemshttp/');
    await expect(files).not.toContainText('controllers/');
  });

  test('Copy command puts the displayed command on the clipboard', async ({ page }) => {
    await open(page);
    await page.getByLabel('Project name').fill('blog');
    const shown = await command(page).textContent();
    await page.getByRole('button', { name: 'Copy command' }).click();
    // Check the visible feedback first: it deliberately reverts after 2 seconds.
    await expect(page.getByRole('button', { name: 'Copied' })).toHaveText('Copied ✓');
    expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(shown);
    await expect(
      page.getByText('Command copied to clipboard').locator('visible=true'),
    ).toBeVisible();
  });

  test('Invalid project name explains the rule and offers no command', async ({ page }) => {
    await open(page);
    const name = page.getByLabel('Project name');
    await name.fill('My App');
    await expect(name).toHaveAttribute('aria-invalid', 'true');
    await expect(name).toHaveAccessibleDescription(
      /Start with a lowercase letter and use only lowercase letters, digits/,
    );
    await expect(page.getByRole('button', { name: 'Copy command' })).toHaveCount(0);
    await expect(
      page.getByText('No command yet. Fix the project name').locator('visible=true'),
    ).toBeVisible();
    await expect(page.getByRole('button', { name: 'Share this configuration' })).toHaveAttribute(
      'aria-disabled',
      'true',
    );
  });

  test('Version too new for Go: module versions are validated live', async ({ page }) => {
    // The shipped catalog has no module version that needs a newer Go than
    // its supported range, so no Go option can be disabled here, and URL
    // state cannot add a constraint. The disabling and its explanation are
    // covered against a test catalog in components/configurator/
    // Configurator.test.tsx and view-model.test.ts. This test checks the live
    // range validation the same field uses.
    await open(page);
    await page.getByRole('button', { name: /Advanced/ }).click();
    const sqlite = page.getByLabel('modernc.org/sqlite');
    await sqlite.fill('v2.1.0');
    await expect(sqlite).toHaveAttribute('aria-invalid', 'true');
    await expect(sqlite).toHaveAccessibleDescription(/outside the supported range >=v1\.61\.0/);
    await sqlite.fill('v1.62.0');
    await expect(sqlite).toHaveAttribute('aria-invalid', 'false');
    await expect(command(page)).toHaveText(
      'gobento new myapp --arch mvc --pin modernc.org/sqlite=v1.62.0',
    );
    await expect(group(page, 'Go version').getByRole('radio', { name: 'Go 1.26' })).toBeEnabled();
  });

  test('Share a link restores the same choices and command', async ({ page, context }) => {
    await open(page);
    await page.getByLabel('Project name').fill('shop');
    await page.getByLabel('Module path').fill('example.com/shop');
    await group(page, 'Go version').getByText('Go 1.26').click();
    const shown = await command(page).textContent();
    await page.getByRole('button', { name: 'Share this configuration' }).click();
    await expect(
      page.getByText('Link to this configuration copied').locator('visible=true'),
    ).toBeVisible();
    const url = await page.evaluate(() => navigator.clipboard.readText());
    expect(url).toContain('/gobento/configurator/?');

    const other = await context.newPage();
    await other.goto(url);
    await expect(other.locator('[data-hydrated="true"]')).toBeAttached();
    await expect(other.getByLabel('Project name')).toHaveValue('shop');
    await expect(other.getByLabel('Module path')).toHaveValue('example.com/shop');
    await expect(group(other, 'Go version').getByRole('radio', { name: 'Go 1.26' })).toBeChecked();
    await expect(command(other)).toHaveText(shown!);
  });

  test('Unknown value in a shared link falls back with a notice', async ({ page }) => {
    await open(page, '?name=blog&arch=mvc&framework=fiber&go=1.19');
    const notice = page.getByRole('status').filter({ hasText: 'shared link was reset' });
    await expect(notice).toContainText('framework "fiber"');
    await expect(notice).toContainText('Go 1.19');
    await expect(
      group(page, 'HTTP framework').getByRole('radio', { name: 'net/http' }),
    ).toBeChecked();
    await expect(command(page)).toHaveText('gobento new blog --arch mvc');
    await notice.getByRole('button', { name: 'Dismiss' }).click();
    await expect(notice).toHaveCount(0);
  });

  test('Keyboard only: fill, choose, expand, and copy without a pointer', async ({ page }) => {
    await open(page);
    const name = page.getByLabel('Project name');
    for (let i = 0; i < 30 && !(await name.evaluate((el) => el === document.activeElement)); i++) {
      await page.keyboard.press('Tab');
    }
    await expect(name).toBeFocused();
    await page.keyboard.press('ControlOrMeta+a');
    await page.keyboard.type('shop');
    await page.keyboard.press('Tab');
    await expect(page.getByLabel('Module path')).toBeFocused();
    await page.keyboard.type('example.com/shop');

    // Each radio group is one tab stop; the Go group is the last one.
    const go127 = group(page, 'Go version').getByRole('radio', { name: 'Go 1.27' });
    for (let i = 0; i < 10 && !(await go127.evaluate((el) => el === document.activeElement)); i++) {
      await page.keyboard.press('Tab');
    }
    await expect(go127).toBeFocused();
    await page.keyboard.press('ArrowLeft');
    await expect(group(page, 'Go version').getByRole('radio', { name: 'Go 1.26' })).toBeChecked();
    await expect(command(page)).toHaveText(
      'gobento new shop --module example.com/shop --arch mvc --go 1.26',
    );

    await page.keyboard.press('Tab');
    const advanced = page.getByRole('button', { name: /Advanced/ });
    await expect(advanced).toBeFocused();
    await page.keyboard.press('Enter');
    await expect(advanced).toHaveAttribute('aria-expanded', 'true');
    await page.keyboard.press('Tab');
    await expect(page.getByLabel('modernc.org/sqlite')).toBeFocused();

    const copy = page.getByRole('button', { name: 'Copy command' });
    for (let i = 0; i < 10 && !(await copy.evaluate((el) => el === document.activeElement)); i++) {
      await page.keyboard.press('Tab');
    }
    await expect(copy).toBeFocused();
    const outline = await copy.evaluate((el) => getComputedStyle(el).outlineWidth);
    expect(outline).toBe('3px');
    await page.keyboard.press('Enter');
    expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(
      'gobento new shop --module example.com/shop --arch mvc --go 1.26',
    );
  });

  for (const colorScheme of ['light', 'dark'] as const) {
    test(`Automated audit: no axe violations in the ${colorScheme} theme`, async ({ page }) => {
      await page.emulateMedia({ colorScheme, reducedMotion: 'reduce' });
      await open(page, '?name=My%20App');
      await expect(page.locator('html')).toHaveClass(
        colorScheme === 'dark' ? /dark/ : /^(?!.*dark)/,
      );
      await page.getByRole('button', { name: /Advanced/ }).click();
      const results = await new AxeBuilder({ page }).analyze();
      expect(results.violations).toEqual([]);
      await page.getByLabel('Project name').fill('shop');
      expect((await new AxeBuilder({ page }).analyze()).violations).toEqual([]);
    });
  }

  test.describe('on a phone', () => {
    test.use({ viewport: { width: 390, height: 844 } });

    test('shows the stage first and pins the command to the bottom', async ({ page }) => {
      await open(page);
      const stage = page.locator('[data-stage]');
      const nameBox = await page.getByLabel('Project name').boundingBox();
      expect((await stage.boundingBox())!.y).toBeLessThan(nameBox!.y);
      const bar = command(page);
      await expect(bar).toHaveText('gobento new myapp --arch mvc');
      const box = await bar.boundingBox();
      expect(box!.y + box!.height).toBeGreaterThan(844 - 100);
      await page.mouse.wheel(0, 1500);
      await expect(bar).toBeInViewport();
    });
  });

  test('the 3D bento loads on interaction without moving the tray', async ({ page }) => {
    await open(page);
    const tray = page.getByRole('list', { name: 'Bento: chosen parts' });
    // The mascot still holds the scene's place from the first paint.
    await expect(page.locator('[data-stage] img[data-mascot-still]')).toBeVisible();
    await page.waitForLoadState('networkidle');
    await expect(page.locator('canvas')).toHaveCount(0);
    // Let the arrival packing finish first: filling a compartment may change its text.
    await expect(tray.locator('[data-filled="true"]')).toHaveCount(4);
    const before = await tray.boundingBox();
    await page.mouse.move(400, 300);
    await expect(page.locator('[data-stage]')).toHaveAttribute('data-stage', 'scene', {
      timeout: 30_000,
    });
    expect(await tray.boundingBox()).toEqual(before);
  });

  test('reduced motion shows the static tray without loading 3D', async ({ page }) => {
    await page.emulateMedia({ reducedMotion: 'reduce' });
    await open(page);
    await expect(page.locator('[data-stage]')).toHaveAttribute('data-stage', 'tray');
    await expect(page.locator('canvas')).toHaveCount(0);
    await expect(page.getByRole('list', { name: 'Bento: chosen parts' })).toContainText('SQLite');
  });
});
