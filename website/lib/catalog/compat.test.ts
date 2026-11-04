import { describe, expect, it } from 'vitest';
import {
  componentReasons,
  goVersionReasons,
  moduleVersionReasons,
  resolvedVersions,
  selectedModules,
  selectionProblems,
  type Selection,
} from './compat';
import type { Catalog } from './schema';
import { addPostgres, catalogWith, type CatalogData, moduleOf } from '@/lib/test/catalog';

const SQLITE = 'modernc.org/sqlite';
const SQLC = 'github.com/sqlc-dev/sqlc';

const firstSlice: Selection = {
  components: { architecture: 'mvc', framework: 'nethttp', database: 'sqlite', data: 'sqlc' },
  go: '1.27',
  pins: {},
};

/** Widens the Go range so that a too-old Go version can be selected. */
const allowGo125 = (d: CatalogData) => (d.go.range = '>=1.25 <1.28');

describe('selectionProblems', () => {
  const fixture = catalogWith();

  const cases: {
    name: string;
    edit?: (d: CatalogData) => void;
    selection: Selection;
    expected: string[];
  }[] = [
    { name: 'defaults (spec: No version flags)', selection: firstSlice, expected: [] },
    {
      name: 'override within range (spec: Override within range)',
      selection: { ...firstSlice, go: '1.26', pins: { [SQLITE]: 'v1.62.0' } },
      expected: [],
    },
    {
      name: 'override outside range (spec: Override outside range)',
      selection: { ...firstSlice, pins: { [SQLITE]: 'v2.1.0' } },
      expected: [`${SQLITE} v2.1.0 is outside the supported range >=v1.61.0 <v2.0.0.`],
    },
    {
      name: 'override of a tool version (spec: Override of a tool version)',
      selection: { ...firstSlice, pins: { [SQLC]: 'v1.30.0' } },
      expected: [`${SQLC} is a tool; its version is fixed per gobento release.`],
    },
    {
      name: 'override with a malformed version',
      selection: { ...firstSlice, pins: { [SQLITE]: '1.62.0' } },
      expected: ['"1.62.0" is not a module version such as v1.2.3.'],
    },
    {
      name: 'Go version too old for a module (spec: Go version too old for a module)',
      edit: allowGo125,
      selection: { ...firstSlice, go: '1.25' },
      expected: [`${SQLITE} v1.61.1 needs Go >=1.26, but Go 1.25 is selected.`],
    },
    {
      name: 'constraint that does not apply to the chosen version',
      edit: (d) => {
        allowGo125(d);
        moduleOf(d, 'modernc.org/sqlite').constraints[0]!.when = '>=v1.62.0';
      },
      selection: { ...firstSlice, go: '1.25' },
      expected: [],
    },
    {
      name: 'constraint on another module',
      edit: (d) =>
        moduleOf(d, 'modernc.org/sqlite').constraints.push({
          when: '>=v1.62.0',
          requires: { target: SQLC, range: '>=v1.32.0' },
        }),
      selection: { ...firstSlice, pins: { [SQLITE]: 'v1.62.0' } },
      expected: [`${SQLITE} v1.62.0 needs ${SQLC} >=v1.32.0, but ${SQLC} v1.31.1 is selected.`],
    },
    {
      name: 'Go version outside the supported range',
      selection: { ...firstSlice, go: '1.28' },
      expected: ['Go 1.28 is not supported; choose a version in >=1.26 <1.28.'],
    },
    {
      name: 'malformed Go version',
      selection: { ...firstSlice, go: 'latest' },
      expected: ['"latest" is not a Go version such as 1.27.'],
    },
    {
      name: 'unmet trait (provided with another value)',
      edit: addPostgres,
      selection: { ...firstSlice, components: { ...firstSlice.components, database: 'postgres' } },
      expected: [
        'sqlc requires conn database/sql, but PostgreSQL provides pgx.',
        'sqlc requires sql-dialect sqlite, but PostgreSQL provides postgresql.',
      ],
    },
    {
      name: 'unmet trait (no provider in a complete selection)',
      edit: (d) => (d.components.find((c) => c.name === 'sqlite')!.provides = {}),
      selection: firstSlice,
      expected: [
        'sqlc requires a component that provides conn database/sql.',
        'sqlc requires a component that provides sql-dialect sqlite.',
        'sqlc requires a component that provides sql-schema schema.sql.',
      ],
    },
    {
      name: 'missing provider while a kind is still unchosen',
      edit: addPostgres,
      selection: { ...firstSlice, components: { ...firstSlice.components, database: undefined } },
      expected: [],
    },
    {
      name: 'override for a module no selected component uses is inert',
      edit: addPostgres,
      selection: {
        ...firstSlice,
        components: { architecture: 'mvc', framework: 'nethttp', database: 'postgres' },
        pins: { [SQLITE]: 'v9.9.9' },
      },
      expected: [],
    },
  ];

  it.each(cases)('$name', ({ edit, selection, expected }) => {
    const catalog = edit ? catalogWith(edit) : fixture;
    expect(selectionProblems(catalog, selection).map((p) => p.message)).toEqual(expected);
  });

  it('attributes a Go constraint to the module version field', () => {
    const [problem] = selectionProblems(catalogWith(allowGo125), { ...firstSlice, go: '1.25' });
    expect(problem).toMatchObject({ field: `pin:${SQLITE}`, involves: [SQLITE, 'go'] });
  });
});

describe('resolvedVersions', () => {
  it('uses overrides, then defaults, for the selected modules only', () => {
    const versions = resolvedVersions(catalogWith(), {
      ...firstSlice,
      pins: { [SQLITE]: 'v1.62.0' },
    });
    expect(Object.fromEntries(versions)).toEqual({ [SQLITE]: 'v1.62.0', [SQLC]: 'v1.31.1' });
  });
});

describe('selectedModules', () => {
  const paths = (catalog: Catalog, selection: Selection) =>
    selectedModules(catalog, selection).map((m) => m.path);
  const gorm = (database: string): Selection => ({
    ...firstSlice,
    components: { ...firstSlice.components, database, data: 'gorm' },
  });

  it('includes modules whose trait conditions the selection meets', () => {
    const selected = paths(catalogWith(), gorm('sqlite'));
    expect(selected).toEqual(expect.arrayContaining(['github.com/libtnb/sqlite', 'gorm.io/gorm']));
    expect(selected).not.toContain('gorm.io/driver/postgres');
  });

  it('follows the dialect the database provides', () => {
    const selected = paths(catalogWith(addPostgres), gorm('postgres'));
    expect(selected).toEqual(expect.arrayContaining(['gorm.io/driver/postgres', 'gorm.io/gorm']));
    expect(selected).not.toContain('github.com/libtnb/sqlite');
  });
});

describe('option reasons', () => {
  it('disables a module version that needs a newer Go and names it (spec: Version too new for Go)', () => {
    const catalog = catalogWith((d) => {
      moduleOf(d, 'modernc.org/sqlite').constraints.push({
        when: '>=v1.62.0',
        requires: { target: 'go', range: '>=1.27' },
      });
    });
    const selection = { ...firstSlice, go: '1.26' };
    expect(moduleVersionReasons(catalog, selection, SQLITE, 'v1.62.0')).toEqual([
      `${SQLITE} v1.62.0 needs Go >=1.27, but Go 1.26 is selected.`,
    ]);
    expect(moduleVersionReasons(catalog, selection, SQLITE, 'v1.61.1')).toEqual([]);
  });

  it('disables a Go version too old for a chosen module version', () => {
    const catalog = catalogWith(allowGo125);
    expect(goVersionReasons(catalog, firstSlice, '1.25')).toEqual([
      `${SQLITE} v1.61.1 needs Go >=1.26, but Go 1.25 is selected.`,
    ]);
    expect(goVersionReasons(catalog, firstSlice, '1.26')).toEqual([]);
  });

  it('disables components with unmet traits, from either side', () => {
    const catalog = catalogWith(addPostgres);
    expect(componentReasons(catalog, firstSlice, 'database', 'postgres')).toEqual([
      'sqlc requires conn database/sql, but PostgreSQL provides pgx.',
      'sqlc requires sql-dialect sqlite, but PostgreSQL provides postgresql.',
    ]);
    expect(componentReasons(catalog, firstSlice, 'database', 'sqlite')).toEqual([]);
    const onPostgres = {
      ...firstSlice,
      components: { ...firstSlice.components, database: 'postgres' },
    };
    expect(componentReasons(catalog, onPostgres, 'data', 'sqlc')).toHaveLength(2);
  });

  it('rejects a component the catalog does not contain', () => {
    expect(componentReasons(catalogWith(), firstSlice, 'framework', 'express')).toEqual([
      'No framework named "express" is supported.',
    ]);
  });
});
