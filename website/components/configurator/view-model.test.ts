import { describe, expect, it } from 'vitest';
import { KIND_IDS } from '@/lib/catalog/kinds';
import { buildCommand } from '@/lib/command/build';
import { defaultState } from '@/lib/command/url-state';
import {
  addPostgres,
  addViteFrontend,
  addViteTailwind,
  catalogWith,
  moduleOf,
} from '@/lib/test/catalog';
import {
  blockedMessage,
  errorsByField,
  fileTree,
  goOptions,
  goVersions,
  kindGroups,
  moduleFields,
  NONE,
  trayCells,
  withChoice,
  withPin,
} from './view-model';

const catalog = catalogWith();
const defaults = defaultState(catalog);
/** The defaults with an architecture and sqlc chosen, the kinds with several options. */
const withArch = (architecture: string) => ({
  ...defaults,
  components: { ...defaults.components, architecture, data: 'sqlc' },
});
const mvc = withArch('mvc');
/** A state with the htmx frontend added. */
const htmx = (state: typeof defaults) => ({
  ...state,
  components: { ...state.components, frontend: 'htmx' },
});
/** A state with Tailwind CSS added. */
const tailwind = (state: typeof defaults) => ({
  ...state,
  components: { ...state.components, css: 'tailwind' },
});

/** The labels inside internal/ of the tree. */
const internal = (state: typeof defaults, from = catalog) => {
  const tree = fileTree(from, state, true);
  const start = tree.findIndex((l) => l.label === 'internal/');
  const end = tree.findIndex((l, i) => i > start && l.depth < 2);
  return tree.slice(start + 1, end).map((l) => l.label);
};

describe('goVersions', () => {
  it('lists every minor in the range', () => {
    expect(goVersions(catalog)).toEqual(['1.26', '1.27']);
  });

  it('falls back to the default for an open upper bound', () => {
    const open = catalogWith((data) => {
      data.go = { range: '>=1.25', default: '1.27' };
    });
    expect(goVersions(open)).toEqual(['1.25', '1.26', '1.27']);
  });
});

describe('kindGroups', () => {
  it('preselects single options and adds roadmap cards', () => {
    const groups = kindGroups(catalog, defaults, new Map(), {
      architecture: [],
      framework: [
        { name: 'echo', title: 'Echo' },
        { name: 'fiber', title: 'Fiber' },
      ],
      database: [],
      data: [],
    });
    const framework = groups.find((g) => g.kind === 'framework')!;
    expect(framework.description).toBe('One option today, preselected.');
    expect(framework.options.map((o) => [o.value, o.status])).toEqual([
      ['nethttp', 'selected'],
      ['echo', 'soon'],
      ['fiber', 'soon'],
    ]);
  });

  it('offers every option of a kind with several, none preselected', () => {
    const data = kindGroups(catalog, defaults, new Map()).find((g) => g.kind === 'data')!;
    expect(data.description).toBe('How the code reads and writes the database.');
    expect(data.options.map((o) => [o.value, o.status])).toEqual([
      ['gorm', 'available'],
      ['sqlc', 'available'],
    ]);
  });

  it('lets a catalog component replace its roadmap card and explains incompatibility', () => {
    const withPostgres = catalogWith(addPostgres);
    const state = {
      ...defaultState(withPostgres),
      components: { ...defaultState(withPostgres).components, data: 'sqlc' },
    };
    const database = kindGroups(withPostgres, state, new Map()).find((g) => g.kind === 'database')!;
    const postgres = database.options.filter((o) => o.value === 'postgres');
    expect(postgres).toHaveLength(1);
    expect(postgres[0]!.status).toBe('incompatible');
    expect(postgres[0]!.reasons.join(' ')).toMatch(/sqlc requires conn database\/sql/);
  });
});

describe('Version too new for Go', () => {
  // The shipped fixture has no module version that needs a newer Go, so the
  // scenario adds one: modernc.org/sqlite v1.62.0 and later need Go 1.27.
  const strict = catalogWith((data) => {
    moduleOf(data, 'modernc.org/sqlite').constraints.push({
      when: '>=v1.62.0',
      requires: { target: 'go', range: '>=1.27' },
    });
  });

  it('disables the older Go version with a reason naming the required version', () => {
    const state = { ...defaultState(strict), pins: { 'modernc.org/sqlite': 'v1.62.0' } };
    const go126 = goOptions(strict, state).find((o) => o.value === '1.26')!;
    expect(go126.status).toBe('incompatible');
    expect(go126.reasons.join(' ')).toContain('needs Go >=1.27');
  });

  it('flags the module version when the selected Go is too old', () => {
    const state = {
      ...defaultState(strict),
      go: '1.26',
      pins: { 'modernc.org/sqlite': 'v1.62.0' },
    };
    const errors = errorsByField(buildCommand(strict, state));
    const field = moduleFields(strict, state, errors).find((f) => f.path === 'modernc.org/sqlite')!;
    expect(field.error).toContain('needs Go >=1.27, but Go 1.26 is selected');
    expect(field.hint).toMatch(/needs Go .*>=1\.27/);
  });
});

describe('moduleFields', () => {
  it('shows tools as fixed', () => {
    const fields = moduleFields(catalog, mvc, new Map());
    expect(fields.map((f) => [f.path, f.fixed])).toEqual([
      ['github.com/sqlc-dev/sqlc', true],
      ['modernc.org/sqlite', false],
    ]);
  });

  it('lists only the modules whose trait conditions the selection meets', () => {
    const gorm = { ...mvc, components: { ...mvc.components, data: 'gorm' } };
    expect(moduleFields(catalog, gorm, new Map()).map((f) => [f.path, f.fixed])).toEqual([
      ['github.com/libtnb/sqlite', false],
      ['gorm.io/gorm', false],
      ['modernc.org/sqlite', false],
    ]);
  });
});

describe('withPin', () => {
  it('clears a pin equal to the default and keeps anything else as typed', () => {
    expect(withPin(catalog, {}, 'modernc.org/sqlite', 'v1.61.1')).toEqual({});
    expect(withPin(catalog, {}, 'modernc.org/sqlite', ' v1.62.0 ')).toEqual({
      'modernc.org/sqlite': 'v1.62.0',
    });
    expect(withPin(catalog, {}, 'modernc.org/sqlite', 'nope')).toEqual({
      'modernc.org/sqlite': 'nope',
    });
  });
});

describe('trayCells and fileTree', () => {
  it('fills a compartment per chosen kind', () => {
    expect(trayCells(catalog, mvc).map((c) => c.value)).toEqual([
      'Model-View-Controller',
      'net/http',
      'SQLite',
      'sqlc',
      undefined,
      undefined,
    ]);
    expect(trayCells(catalog, mvc).map((c) => c.kind)).toEqual(KIND_IDS);
    expect(trayCells(catalog, htmx(mvc)).at(-2)).toMatchObject({
      kind: 'frontend',
      label: 'Frontend',
      value: 'htmx + templ',
    });
    expect(trayCells(catalog, { ...defaults, components: {} }).every((c) => !c.value)).toBe(true);
  });

  it('names the tree after a valid project only', () => {
    const tree = fileTree(catalog, { ...mvc, name: 'shop' }, true).map((l) => l.label);
    expect(tree).toContain('cmd/shop/main.go');
    expect(tree).toContain('sqlc.yaml');
    expect(fileTree(catalog, { ...defaults, name: 'My App' }, false)[0]!.label).toBe('<name>/');
  });

  it('lists the tooling every project gets', () => {
    const tree = fileTree(catalog, { ...mvc, name: 'shop' }, true).map((l) => [l.depth, l.label]);
    expect(tree).toEqual([
      [0, 'shop/'],
      [1, '.github/'],
      [2, 'dependabot.yml'],
      [2, 'workflows/ci.yml'],
      [1, 'cmd/shop/main.go'],
      [1, 'internal/'],
      [2, 'controllers/'],
      [2, 'database/'],
      [2, 'models/'],
      [2, 'store/'],
      [1, '.dockerignore'],
      [1, '.golangci.yml'],
      [1, 'Dockerfile'],
      [1, 'Makefile'],
      [1, 'sqlc.yaml'],
      [1, 'go.mod'],
      [1, 'README.md'],
    ]);
  });

  it('shows the layered packages for mvc', () => {
    expect(internal(mvc)).toEqual(['controllers/', 'database/', 'models/', 'store/']);
  });

  it('shows the feature directory and platform for vertical-slice', () => {
    expect(internal(withArch('vertical-slice'))).toEqual([
      'items/',
      'items/itemshttp/',
      'items/itemsstore/',
      'platform/database/',
    ]);
  });

  it('shows the domain, application, and adapter packages for ddd', () => {
    // The frozen fixture predates ddd, so add it the way catalog.json lists it.
    const withDDD = catalogWith((data) => {
      data.components.unshift({
        id: 'architecture/ddd',
        kind: 'architecture',
        name: 'ddd',
        title: 'Domain-driven design',
        description: 'A domain package, an application service, and adapters.',
        provides: {},
        requires: {},
      });
    });
    const ddd = withArch('ddd');
    expect(internal(ddd, withDDD)).toEqual([
      'adapters/database/',
      'adapters/httpapi/',
      'adapters/persistence/',
      'app/',
      'domain/',
    ]);
  });
});

describe('blockedMessage', () => {
  it('names the first invalid field', () => {
    expect(blockedMessage(catalog, buildCommand(catalog, { ...defaults, name: 'My App' }))).toBe(
      'No command yet. Fix the project name to get a copyable command.',
    );
    expect(blockedMessage(catalog, buildCommand(catalog, defaults))).toBe(
      'No command yet. Fix the architecture choice to get a copyable command.',
    );
    expect(
      blockedMessage(
        catalog,
        buildCommand(catalog, { ...mvc, pins: { 'modernc.org/sqlite': 'x' } }),
      ),
    ).toBe(
      'No command yet. Fix the modernc.org/sqlite version under Advanced to get a copyable command.',
    );
    expect(blockedMessage(catalog, buildCommand(catalog, mvc))).toBe('');
  });
});

describe('optional frontend (spec: Frontend is optional)', () => {
  const frontend = (state: typeof defaults) =>
    kindGroups(catalog, state, new Map()).find((g) => g.kind === 'frontend')!;

  it('offers None (API only) first and preselected, then every frontend and the roadmap', () => {
    const group = frontend(mvc);
    expect(group.description).toBe(
      'Optional. Web pages served next to the API, which then moves under /api.',
    );
    expect(group.options.map((o) => [o.value, o.title, o.status])).toEqual([
      [NONE, 'None (API only)', 'selected'],
      ['htmx', 'htmx + templ', 'available'],
    ]);
  });

  it('offers the Vite frontends from the catalog before the roadmap', () => {
    const withVite = catalogWith((data) => {
      for (const name of ['react', 'svelte', 'vue'] as const) addViteFrontend(data, name);
    });
    const group = kindGroups(withVite, mvc, new Map()).find((g) => g.kind === 'frontend')!;
    expect(group.options.map((o) => [o.value, o.title, o.status])).toEqual([
      [NONE, 'None (API only)', 'selected'],
      ['htmx', 'htmx + templ', 'available'],
      ['react', 'React', 'available'],
      ['svelte', 'Svelte', 'available'],
      ['vue', 'Vue', 'available'],
    ]);
  });

  it('selects htmx and keeps None available', () => {
    expect(
      frontend(htmx(mvc))
        .options.slice(0, 2)
        .map((o) => o.status),
    ).toEqual(['available', 'selected']);
  });

  it('keeps every other option compatible when htmx is chosen', () => {
    for (const group of kindGroups(catalog, htmx(mvc), new Map())) {
      expect(group.options.filter((o) => o.status === 'incompatible')).toEqual([]);
    }
  });

  it('clears an optional kind on None and sets it otherwise', () => {
    expect(withChoice(catalog, htmx(mvc).components, 'frontend', NONE)).toEqual(mvc.components);
    expect(withChoice(catalog, mvc.components, 'frontend', 'htmx')).toEqual(htmx(mvc).components);
  });

  it('lists templ and the vendored htmx.org only when htmx is chosen', () => {
    const fields = moduleFields(catalog, htmx(mvc), new Map());
    expect(fields.map((f) => [f.path, f.fixed, f.tag])).toEqual([
      ['github.com/a-h/templ', false, undefined],
      ['github.com/sqlc-dev/sqlc', true, 'Fixed'],
      ['modernc.org/sqlite', false, undefined],
      ['htmx.org', true, 'Vendored'],
    ]);
    expect(fields.at(-1)).toMatchObject({ value: '2.0.11', usedBy: 'htmx + templ' });
    expect(moduleFields(catalog, mvc, new Map()).map((f) => f.path)).not.toContain('htmx.org');
  });

  it('previews the pages in the architecture web package', () => {
    const lines = (state: typeof defaults) =>
      fileTree(catalog, state, true).flatMap((l) => (l.depth === 2 ? [l.label] : []));
    expect(lines(htmx(mvc))).toEqual(
      expect.arrayContaining(['views/', 'views/pages.templ', 'views/static/']),
    );
    expect(lines(htmx(withArch('vertical-slice')))).toEqual(
      expect.arrayContaining(['web/', 'web/pages.templ', 'web/static/']),
    );
    expect(lines(mvc).some((l) => l.startsWith('views/'))).toBe(false);
  });

  describe('a Vite frontend', () => {
    const withVite = catalogWith((data) => {
      for (const name of ['react', 'svelte', 'vue'] as const) addViteFrontend(data, name);
    });
    const vite = (name: string, state: typeof defaults = mvc) => ({
      ...state,
      components: { ...state.components, frontend: name },
    });

    it('lists its scaffolder with the fixed version from the catalog', () => {
      const fields = moduleFields(withVite, vite('react'), new Map());
      expect(fields.at(-1)).toEqual({
        path: 'create-vite',
        usedBy: 'React, Svelte, Vue',
        fixed: true,
        tag: 'Scaffolder',
        value: '9.3.3',
        hint: 'Scaffolds the app at generation time and needs Node.js >=v22.12.0; its version is fixed per gobento release.',
      });
      expect(moduleFields(withVite, htmx(mvc), new Map()).map((f) => f.path)).not.toContain(
        'create-vite',
      );
    });

    it('previews the handler in the web package and gobento-owned files in frontend/', () => {
      const tree = fileTree(withVite, { ...vite('react'), name: 'shop' }, true);
      expect(internal(vite('react'), withVite)).toEqual([
        'controllers/',
        'database/',
        'models/',
        'store/',
        'views/',
        'views/dist/.gitignore',
        'views/handler.go',
        'views/handler_test.go',
      ]);
      const start = tree.findIndex((l) => l.label === 'frontend/');
      expect(tree[start]).toEqual({
        label: 'frontend/',
        depth: 1,
        dir: true,
        note: 'completed by create-vite 9.3.3 at generation time',
      });
      expect(tree.slice(start + 1, start + 5).map((l) => [l.depth, l.label])).toEqual([
        [2, 'src/App.tsx'],
        [2, 'src/api.ts'],
        [2, 'src/index.css'],
        [2, 'vite.config.ts'],
      ]);
      expect(tree[start + 5]!.depth).toBe(1);
    });

    it('names the page and stylesheet after the framework', () => {
      const app = (name: string) => {
        const tree = fileTree(withVite, vite(name, withArch('vertical-slice')), true);
        const start = tree.findIndex((l) => l.label === 'frontend/');
        return tree.slice(start + 1, start + 5).map((l) => l.label);
      };
      expect(app('vue')).toEqual(['src/App.vue', 'src/api.ts', 'src/style.css', 'vite.config.ts']);
      expect(app('svelte')).toEqual([
        'src/App.svelte',
        'src/api.ts',
        'src/app.css',
        'vite.config.ts',
      ]);
      expect(internal(vite('svelte', withArch('vertical-slice')), withVite)).toContain(
        'web/handler.go',
      );
    });

    it('shows no frontend/ directory without a Vite frontend', () => {
      for (const state of [mvc, htmx(mvc)]) {
        expect(fileTree(withVite, state, true).map((l) => l.label)).not.toContain('frontend/');
      }
    });
  });

  it('previews the pages in the web adapter for ddd', () => {
    // The frozen fixture predates ddd, so add it the way catalog.json lists it.
    const withDDD = catalogWith((data) => {
      data.components.unshift({
        id: 'architecture/ddd',
        kind: 'architecture',
        name: 'ddd',
        title: 'Domain-driven design',
        description: 'A domain package, an application service, and adapters.',
        provides: {},
        requires: {},
      });
    });
    const lines = fileTree(withDDD, htmx(withArch('ddd')), true).flatMap((l) =>
      l.depth === 2 ? [l.label] : [],
    );
    expect(lines).toEqual(
      expect.arrayContaining(['adapters/web/', 'adapters/web/pages.templ', 'adapters/web/static/']),
    );
  });
});

describe('optional CSS framework (spec: CSS framework needs a frontend)', () => {
  const css = (state: typeof defaults) =>
    kindGroups(catalog, state, errorsByField(buildCommand(catalog, state))).find(
      (g) => g.kind === 'css',
    )!;
  const needsFrontend =
    'Tailwind CSS requires a frontend that provides frontend-build go or vite, such as htmx + templ.';

  it('comes after the frontend, offering None (plain CSS) first and preselected', () => {
    const groups = kindGroups(catalog, mvc, new Map());
    expect(groups.map((g) => g.kind).slice(-2)).toEqual(['frontend', 'css']);
    const group = css(mvc);
    expect(group.title).toBe('CSS framework');
    expect(group.description).toMatch(/^Optional\. /);
    expect(group.options[0]).toMatchObject({
      value: NONE,
      title: 'None (plain CSS)',
      status: 'selected',
    });
  });

  it('disables Tailwind CSS without a frontend, explaining that it needs one', () => {
    for (const state of [defaults, mvc]) {
      const option = css(state).options.find((o) => o.value === 'tailwind')!;
      expect(option.status).toBe('incompatible');
      expect(option.reasons).toEqual([needsFrontend]);
    }
  });

  it('offers Tailwind CSS with htmx and adds it to the command', () => {
    const option = css(htmx(mvc)).options.find((o) => o.value === 'tailwind')!;
    expect(option.status).toBe('available');
    const state = tailwind(htmx(mvc));
    expect(css(state).options.map((o) => o.status)).toEqual(['available', 'selected']);
    const result = buildCommand(catalog, state);
    expect(result.ok && result.command).toBe(
      'gobento new myapp --arch mvc --data sqlc --frontend htmx --css tailwind',
    );
  });

  it('keeps Tailwind CSS but blocks the command when the frontend is removed', () => {
    const state = {
      ...mvc,
      components: withChoice(catalog, tailwind(htmx(mvc)).components, 'frontend', NONE),
    };
    expect(state.components).toEqual({ ...mvc.components, css: 'tailwind' });
    const group = css(state);
    expect(group.options.find((o) => o.value === 'tailwind')!.status).toBe('selected');
    expect(group.error).toBe(needsFrontend);
    const result = buildCommand(catalog, state);
    expect(result.ok).toBe(false);
    expect(blockedMessage(catalog, result)).toBe(
      'No command yet. Fix the CSS framework choice to get a copyable command.',
    );
    // None (plain CSS) clears it again.
    expect(withChoice(catalog, state.components, 'css', NONE)).toEqual(mvc.components);
  });

  it('packs a sixth compartment and lists the binary and files Tailwind CSS adds', () => {
    const state = tailwind(htmx(mvc));
    expect(trayCells(catalog, state).at(-1)).toMatchObject({
      kind: 'css',
      label: 'CSS framework',
      value: 'Tailwind CSS',
      piece: 'shiba',
    });
    expect(moduleFields(catalog, state, new Map()).at(-1)).toMatchObject({
      path: 'tailwindcss',
      fixed: true,
      tag: 'On demand',
      value: '4.3.4',
      usedBy: 'Tailwind CSS',
    });
    expect(moduleFields(catalog, htmx(mvc), new Map()).map((f) => f.path)).not.toContain(
      'tailwindcss',
    );
    const labels = fileTree(catalog, state, true).map((l) => l.label);
    expect(labels).toContain('tailwind.css');
    expect(labels).toContain('tools/tailwindcss/');
    expect(fileTree(catalog, htmx(mvc), true).map((l) => l.label)).not.toContain('tailwind.css');
    expect(moduleFields(catalog, state, new Map()).map((f) => f.path)).not.toContain(
      '@tailwindcss/vite',
    );
  });

  it('lists the npm packages and no Tailwind files with a Vite frontend', () => {
    const withVite = catalogWith((data) => {
      addViteFrontend(data, 'react');
      addViteTailwind(data);
    });
    const state = tailwind({ ...mvc, components: { ...mvc.components, frontend: 'react' } });
    const fields = moduleFields(withVite, state, new Map());
    expect(fields.filter((f) => f.usedBy === 'Tailwind CSS').map((f) => [f.path, f.tag])).toEqual([
      ['@tailwindcss/vite', 'npm (dev)'],
      ['tailwindcss', 'npm (dev)'],
    ]);
    const labels = fileTree(withVite, state, true).map((l) => l.label);
    expect(labels).not.toContain('tailwind.css');
    expect(labels).not.toContain('tools/tailwindcss/');
    expect(labels).toContain('vite.config.ts');
  });
});
