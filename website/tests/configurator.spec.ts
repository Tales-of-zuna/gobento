import AxeBuilder from '@axe-core/playwright';
import { expect, test, type Page } from '@playwright/test';

import { releaseStatus } from './helpers';

const { released } = releaseStatus();

/**
 * Opens the configurator and waits until it has read the URL (after
 * hydration). Every component kind has several options, so tests start from
 * the reference selection unless they pass their own query.
 */
async function open(page: Page, query = '?arch=mvc&framework=nethttp&db=sqlite&data=sqlc') {
  await page.goto(`configurator/${query}`);
  await expect(page.locator('[data-hydrated="true"]')).toBeAttached();
}

const command = (page: Page) => page.locator('[data-command]').locator('visible=true');
const group = (page: Page, name: string) => page.getByRole('group', { name });

/** Chooses a data access by clicking its card; with several options, none is preselected. */
const chooseData = (page: Page, title: string) =>
  group(page, 'Data access').getByText(title, { exact: true }).click();

test.describe('configurator', () => {
  test.use({ permissions: ['clipboard-read', 'clipboard-write'] });

  test('Defaults: single options and default versions are preselected', async ({ page }) => {
    await open(page, '');
    await expect(page.getByRole('heading', { level: 1, name: 'Pack your project' })).toBeVisible();
    // The pre-release notice follows the release manifest (spec "Pre-release notice").
    const notice = page.getByText('Early development · not released yet');
    if (released) await expect(notice).toHaveCount(0);
    else await expect(notice).toBeVisible();
    await expect(group(page, 'Go version').getByRole('radio', { name: 'Go 1.27' })).toBeChecked();
    // Several options in every kind: none is preselected, and the CLI needs a
    // flag for each.
    for (const [kind, prompt] of [
      ['Architecture', 'Choose one of: ddd, mvc, vertical-slice.'],
      ['HTTP framework', 'Choose one of: chi, gin, nethttp.'],
      ['Database', 'Choose one of: postgres, sqlite.'],
      ['Data access', 'Choose one of: gorm, sqlc.'],
    ]) {
      await expect(group(page, kind!).getByRole('radio', { checked: true })).toHaveCount(0);
      await expect(group(page, kind!).getByText(prompt!)).toBeVisible();
    }
    // The frontend is optional: None (API only) is preselected (spec "Frontend is optional").
    const frontend = group(page, 'Frontend');
    await expect(frontend.getByRole('radio', { name: 'None (API only)' })).toBeChecked();
    await expect(frontend.getByRole('radio', { name: 'htmx + templ' })).toBeEnabled();
    await expect(
      page.getByText('No command yet. Fix the architecture choice').locator('visible=true'),
    ).toBeVisible();
    for (const [kind, option] of [
      ['Architecture', 'Model-View-Controller'],
      ['HTTP framework', 'net/http'],
      ['Database', 'SQLite'],
      ['Data access', 'sqlc'],
    ]) {
      await group(page, kind!).getByText(option!, { exact: true }).click();
    }
    await expect(command(page)).toHaveText(
      'gobento new myapp --arch mvc --framework nethttp --db sqlite --data sqlc',
    );
    await page.getByRole('button', { name: /Advanced/ }).click();
    await expect(page.getByLabel('modernc.org/sqlite')).toHaveValue('v1.61.1');
    await expect(page.getByText('Tool versions are fixed per gobento release.')).toBeVisible();
  });

  test('Changing a choice updates the command with defaults omitted', async ({ page }) => {
    await open(page);
    await page.getByLabel('Project name').fill('shop');
    await page.getByLabel('Module path').fill('example.com/shop');
    await expect(command(page)).toHaveText(
      'gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc',
    );
    await group(page, 'Go version').getByText('Go 1.26').click();
    await expect(command(page)).toHaveText(
      'gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc --go 1.26',
    );
    await expect(page.getByRole('list', { name: 'Generated files' })).toContainText(
      'cmd/shop/main.go',
    );
    const files = page.getByRole('list', { name: 'Generated files' });
    await expect(files).toContainText('cmd/shop/main.go');
    await expect(files).toContainText('controllers/');
    await group(page, 'Architecture').getByText('Vertical slice', { exact: true }).click();
    await expect(command(page)).toHaveText(
      'gobento new shop --module example.com/shop --arch vertical-slice --framework nethttp --db sqlite --data sqlc --go 1.26',
    );
    await expect(files).toContainText('items/itemshttp/');
    await expect(files).not.toContainText('controllers/');
  });

  test('Choosing a frontend adds --frontend and None removes it', async ({ page }) => {
    await open(page);
    const frontend = group(page, 'Frontend');
    const base = 'gobento new myapp --arch mvc --framework nethttp --db sqlite --data sqlc';
    await expect(command(page)).toHaveText(base);
    await frontend.getByText('htmx + templ', { exact: true }).click();
    await expect(command(page)).toHaveText(`${base} --frontend htmx`);
    await expect(page).toHaveURL(/[?&]frontend=htmx(&|$)/);
    // The other kinds stay compatible with htmx.
    await expect(page.getByText('Incompatible', { exact: true })).toHaveCount(0);
    const files = page.getByRole('list', { name: 'Generated files' });
    await expect(files).toContainText('views/pages.templ');
    await expect(page.getByRole('list', { name: 'Bento: chosen parts' })).toContainText(
      'htmx + templ',
    );
    await page.getByRole('button', { name: /Advanced/ }).click();
    await expect(page.getByLabel('github.com/a-h/templ')).toHaveValue('v0.3.1100');
    await expect(page.getByText('htmx.org (htmx + templ)')).toBeVisible();
    await expect(page.getByText('Vendored', { exact: true })).toBeVisible();

    await frontend.getByText('None (API only)', { exact: true }).click();
    await expect(command(page)).toHaveText(base);
    await expect(page).not.toHaveURL(/frontend=/);
    await expect(page.getByLabel('github.com/a-h/templ')).toHaveCount(0);
    await expect(files).not.toContainText('pages.templ');
  });

  test('A shared link restores the frontend, and an unknown one falls back to None', async ({
    page,
  }) => {
    await open(page, '?arch=mvc&framework=nethttp&db=sqlite&data=sqlc&frontend=htmx');
    const frontend = group(page, 'Frontend');
    await expect(frontend.getByRole('radio', { name: 'htmx + templ' })).toBeChecked();
    await expect(command(page)).toHaveText(
      'gobento new myapp --arch mvc --framework nethttp --db sqlite --data sqlc --frontend htmx',
    );

    await open(page, '?arch=mvc&framework=nethttp&db=sqlite&data=sqlc&frontend=react');
    const notice = page.getByRole('status').filter({ hasText: 'shared link was reset' });
    await expect(notice).toContainText('frontend "react", which is not available; none is used');
    await expect(frontend.getByRole('radio', { name: 'None (API only)' })).toBeChecked();
    await expect(command(page)).toHaveText(
      'gobento new myapp --arch mvc --framework nethttp --db sqlite --data sqlc',
    );
  });

  test('CSS framework needs a frontend: None (plain CSS) is preselected and Tailwind CSS disabled', async ({
    page,
  }) => {
    await open(page);
    const css = group(page, 'CSS framework');
    await expect(css.getByRole('radio', { name: 'None (plain CSS)' })).toBeChecked();
    const tailwind = css.getByRole('radio', { name: 'Tailwind CSS' });
    await expect(tailwind).toBeDisabled();
    await expect(css).toContainText('Incompatible');
    await expect(css).toContainText(
      'Tailwind CSS requires a frontend that provides frontend-build go or vite, such as htmx + templ.',
    );
    await expect(command(page)).not.toContainText('--css');
    // The tray has a compartment for every kind, the CSS framework's empty.
    const tray = page.getByRole('list', { name: 'Bento: chosen parts' });
    await expect(tray.getByRole('listitem')).toHaveCount(6);
    await expect(tray.locator('[data-kind="css"]')).toHaveAttribute('data-filled', 'false');
  });

  test('CSS framework with a frontend: htmx and Tailwind CSS add --frontend htmx --css tailwind', async ({
    page,
  }) => {
    await open(page);
    const base = 'gobento new myapp --arch mvc --framework nethttp --db sqlite --data sqlc';
    await group(page, 'Frontend').getByText('htmx + templ', { exact: true }).click();
    const css = group(page, 'CSS framework');
    await expect(css.getByRole('radio', { name: 'Tailwind CSS' })).toBeEnabled();
    await css.getByText('Tailwind CSS', { exact: true }).click();
    await expect(command(page)).toHaveText(`${base} --frontend htmx --css tailwind`);
    await expect(page).toHaveURL(/[?&]frontend=htmx&css=tailwind(&|$)/);
    await expect(page.getByText('Incompatible', { exact: true })).toHaveCount(0);
    const tray = page.getByRole('list', { name: 'Bento: chosen parts' });
    await expect(tray.locator('[data-kind="css"]')).toHaveAttribute('data-filled', 'true');
    await expect(tray).toContainText('Tailwind CSS');
    const files = page.getByRole('list', { name: 'Generated files' });
    await expect(files).toContainText('tailwind.css');
    await page.getByRole('button', { name: /Advanced/ }).click();
    await expect(page.getByText('tailwindcss (Tailwind CSS)')).toBeVisible();
    await expect(page.getByText('On demand', { exact: true })).toBeVisible();

    // The shared link restores both choices.
    await open(page, `?arch=mvc&framework=nethttp&db=sqlite&data=sqlc&frontend=htmx&css=tailwind`);
    await expect(css.getByRole('radio', { name: 'Tailwind CSS' })).toBeChecked();
    await expect(command(page)).toHaveText(`${base} --frontend htmx --css tailwind`);

    // Removing the frontend keeps Tailwind CSS but blocks the command until it is fixed.
    await group(page, 'Frontend').getByText('None (API only)', { exact: true }).click();
    await expect(css.getByRole('radio', { name: 'Tailwind CSS' })).toBeChecked();
    await expect(
      page.getByText('No command yet. Fix the CSS framework choice').locator('visible=true'),
    ).toBeVisible();
    await css.getByText('None (plain CSS)', { exact: true }).click();
    await expect(command(page)).toHaveText(base);
    await expect(page).not.toHaveURL(/css=/);
  });

  test('Copy command puts the displayed command on the clipboard', async ({ page }) => {
    await open(page);
    await page.getByLabel('Project name').fill('blog');
    await chooseData(page, 'sqlc');
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
    await chooseData(page, 'sqlc');
    await page.getByRole('button', { name: /Advanced/ }).click();
    const sqlite = page.getByLabel('modernc.org/sqlite');
    await sqlite.fill('v2.1.0');
    await expect(sqlite).toHaveAttribute('aria-invalid', 'true');
    await expect(sqlite).toHaveAccessibleDescription(/outside the supported range >=v1\.61\.0/);
    await sqlite.fill('v1.62.0');
    await expect(sqlite).toHaveAttribute('aria-invalid', 'false');
    await expect(command(page)).toHaveText(
      'gobento new myapp --arch mvc --framework nethttp --db sqlite --data sqlc --pin modernc.org/sqlite=v1.62.0',
    );
    await expect(group(page, 'Go version').getByRole('radio', { name: 'Go 1.26' })).toBeEnabled();
  });

  test('Share a link restores the same choices and command', async ({ page, context }) => {
    await open(page);
    await page.getByLabel('Project name').fill('shop');
    await page.getByLabel('Module path').fill('example.com/shop');
    await chooseData(page, 'sqlc');
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
    await expect(group(other, 'Data access').getByRole('radio', { name: 'sqlc' })).toBeChecked();
    await expect(group(other, 'Go version').getByRole('radio', { name: 'Go 1.26' })).toBeChecked();
    await expect(command(other)).toHaveText(shown!);
  });

  test('Unknown value in a shared link falls back with a notice', async ({ page }) => {
    await open(page, '?name=blog&arch=mvc&framework=fiber&go=1.19');
    const notice = page.getByRole('status').filter({ hasText: 'shared link was reset' });
    await expect(notice).toContainText('framework "fiber"');
    await expect(notice).toContainText('Go 1.19');
    // There is no single framework to fall back to, so a choice is required.
    const framework = group(page, 'HTTP framework');
    await expect(framework.getByRole('radio', { checked: true })).toHaveCount(0);
    await expect(framework.getByText('Choose one of: chi, gin, nethttp.')).toBeVisible();
    await notice.getByRole('button', { name: 'Dismiss' }).click();
    await expect(notice).toHaveCount(0);
  });

  test('Keyboard only: fill, choose, expand, and copy without a pointer', async ({ page }) => {
    // Data access is left unchosen so the keyboard makes that choice.
    await open(page, '?arch=mvc&framework=nethttp&db=sqlite');
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

    // Each radio group is one tab stop. Data access has no choice yet, so Tab
    // lands on its first card and an arrow key moves to and checks the next.
    const gorm = group(page, 'Data access').getByRole('radio', { name: 'GORM' });
    for (let i = 0; i < 10 && !(await gorm.evaluate((el) => el === document.activeElement)); i++) {
      await page.keyboard.press('Tab');
    }
    await expect(gorm).toBeFocused();
    await page.keyboard.press('ArrowRight');
    await expect(group(page, 'Data access').getByRole('radio', { name: 'sqlc' })).toBeChecked();

    // The Go group is the last one.
    const go127 = group(page, 'Go version').getByRole('radio', { name: 'Go 1.27' });
    for (let i = 0; i < 10 && !(await go127.evaluate((el) => el === document.activeElement)); i++) {
      await page.keyboard.press('Tab');
    }
    await expect(go127).toBeFocused();
    await page.keyboard.press('ArrowLeft');
    await expect(group(page, 'Go version').getByRole('radio', { name: 'Go 1.26' })).toBeChecked();
    await expect(command(page)).toHaveText(
      'gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc --go 1.26',
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
      'gobento new shop --module example.com/shop --arch mvc --framework nethttp --db sqlite --data sqlc --go 1.26',
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
      await expect(bar).toHaveText(
        'gobento new myapp --arch mvc --framework nethttp --db sqlite --data sqlc',
      );
      const box = await bar.boundingBox();
      expect(box!.y + box!.height).toBeGreaterThan(844 - 100);
      await page.mouse.wheel(0, 1500);
      await expect(bar).toBeInViewport();
    });
  });

  test('the 3D bento loads on interaction without moving the tray', async ({ page }) => {
    // The reference selection fills every compartment.
    await open(page);
    const tray = page.getByRole('list', { name: 'Bento: chosen parts' });
    // The mascot still holds the scene's place from the first paint.
    await expect(page.locator('[data-stage] img[data-mascot-still]')).toBeVisible();
    await page.waitForLoadState('networkidle');
    await expect(page.locator('canvas')).toHaveCount(0);
    // Let the arrival packing finish first: filling a compartment may change its text.
    await expect(tray.locator('[data-filled="true"]')).toHaveCount(4);
    await expect(tray.getByRole('listitem')).toHaveCount(6);
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
