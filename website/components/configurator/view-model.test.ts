import { describe, expect, it } from 'vitest';
import { buildCommand } from '@/lib/command/build';
import { defaultState } from '@/lib/command/url-state';
import { addPostgres, catalogWith, moduleOf } from '@/lib/test/catalog';
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
    ]);
    expect(trayCells(catalog, htmx(mvc)).at(-1)).toMatchObject({
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

  /** The labels inside internal/ of the tree. */
  const internal = (state: typeof defaults, from = catalog) => {
    const tree = fileTree(from, state, true);
    const start = tree.findIndex((l) => l.label === 'internal/');
    const end = tree.findIndex((l, i) => i > start && l.depth < 2);
    return tree.slice(start + 1, end).map((l) => l.label);
  };

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
      ['react', 'React via Vite', 'soon'],
      ['vue', 'Vue via Vite', 'soon'],
      ['svelte', 'Svelte via Vite', 'soon'],
      ['tailwind', 'Tailwind CSS', 'soon'],
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
