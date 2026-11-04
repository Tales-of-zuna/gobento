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
  trayCells,
  withPin,
} from './view-model';

const catalog = catalogWith();
const defaults = defaultState(catalog);

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

  it('lets a catalog component replace its roadmap card and explains incompatibility', () => {
    const withPostgres = catalogWith(addPostgres);
    const database = kindGroups(withPostgres, defaultState(withPostgres), new Map()).find(
      (g) => g.kind === 'database',
    )!;
    const postgres = database.options.filter((o) => o.value === 'postgres');
    expect(postgres).toHaveLength(1);
    expect(postgres[0]!.status).toBe('incompatible');
    expect(postgres[0]!.reasons.join(' ')).toMatch(/sqlc requires sql-dialect sqlite/);
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
    const fields = moduleFields(catalog, defaults, new Map());
    expect(fields.map((f) => [f.path, f.fixed])).toEqual([
      ['github.com/sqlc-dev/sqlc', true],
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
    expect(trayCells(catalog, defaults).map((c) => c.value)).toEqual([
      'Model-View-Controller',
      'net/http',
      'SQLite',
      'sqlc',
    ]);
    expect(trayCells(catalog, { ...defaults, components: {} }).every((c) => !c.value)).toBe(true);
  });

  it('names the tree after a valid project only', () => {
    const tree = fileTree(catalog, { ...defaults, name: 'shop' }, true).map((l) => l.label);
    expect(tree).toContain('cmd/shop/main.go');
    expect(tree).toContain('sqlc.yaml');
    expect(fileTree(catalog, { ...defaults, name: 'My App' }, false)[0]!.label).toBe('<name>/');
  });
});

describe('blockedMessage', () => {
  it('names the first invalid field', () => {
    expect(blockedMessage(catalog, buildCommand(catalog, { ...defaults, name: 'My App' }))).toBe(
      'No command yet. Fix the project name to get a copyable command.',
    );
    expect(
      blockedMessage(
        catalog,
        buildCommand(catalog, { ...defaults, pins: { 'modernc.org/sqlite': 'x' } }),
      ),
    ).toBe(
      'No command yet. Fix the modernc.org/sqlite version under Advanced to get a copyable command.',
    );
    expect(blockedMessage(catalog, buildCommand(catalog, defaults))).toBe('');
  });
});
