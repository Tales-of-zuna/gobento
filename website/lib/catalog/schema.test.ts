import { describe, expect, it } from 'vitest';
import { catalogSchema } from './schema';
import { fixtureData, type CatalogData } from '@/lib/test/catalog';

describe('catalog schema', () => {
  it('accepts the exported catalog', () => {
    const result = catalogSchema.safeParse(fixtureData());
    expect(result.error?.issues).toBeUndefined();
    expect(result.data?.components.map((c) => c.id)).toEqual([
      'architecture/mvc',
      'architecture/vertical-slice',
      'framework/nethttp',
      'database/sqlite',
      'data/sqlc',
    ]);
  });

  const sqlite = (data: CatalogData) => data.modules.find((m) => m.path === 'modernc.org/sqlite')!;
  const sqlc = (data: CatalogData) => data.components.find((c) => c.id === 'data/sqlc')!;

  // Each malformed variant must be rejected with an issue mentioning `expected`.
  const malformed: { name: string; edit: (data: CatalogData) => void; expected: string }[] = [
    {
      name: 'module default outside its range (spec: Default outside range)',
      edit: (d) => {
        sqlite(d).range = '>=v2.0.0';
        sqlite(d).default = 'v1.5.0';
      },
      expected: 'modernc.org/sqlite: default v1.5.0 is outside its range >=v2.0.0',
    },
    {
      name: 'Go default outside the Go range',
      edit: (d) => (d.go.default = '1.25'),
      expected: 'default 1.25 is outside its range',
    },
    {
      name: 'malformed module range',
      edit: (d) => (sqlite(d).range = '~v1.61'),
      expected: 'malformed range "~v1.61"',
    },
    {
      name: 'malformed Go range',
      edit: (d) => (d.go.range = '>=go1.26'),
      expected: 'malformed range ">=go1.26"',
    },
    {
      name: 'empty range',
      edit: (d) => (sqlite(d).range = ' '),
      expected: 'malformed range',
    },
    {
      name: 'module version without the v prefix',
      edit: (d) => (sqlite(d).default = '1.61.1'),
      expected: 'malformed version "1.61.1"',
    },
    {
      name: 'unknown component referenced by a module',
      edit: (d) => (sqlite(d).components = ['database/postgres']),
      expected: 'unknown component "database/postgres"',
    },
    {
      name: 'unknown constraint target',
      edit: (d) => (sqlite(d).constraints[0]!.requires.target = 'example.com/missing'),
      expected: 'unknown target "example.com/missing"',
    },
    {
      name: 'malformed constraint range for a Go target',
      edit: (d) => (sqlite(d).constraints[0]!.requires.range = '>=v1.26.0'),
      expected: 'malformed range ">=v1.26.0"',
    },
    {
      name: 'tool without an exact version',
      edit: (d) => (d.modules.find((m) => m.role === 'tool')!.range = '>=v1.31.1'),
      expected: "a tool's range must be one exact version",
    },
    {
      name: 'component id not matching kind/name',
      edit: (d) => (sqlc(d).id = 'data/other'),
      expected: 'id must be "data/sqlc"',
    },
    {
      name: 'duplicate component',
      edit: (d) => d.components.push({ ...sqlc(d) }),
      expected: 'duplicate component "data/sqlc"',
    },
    {
      name: 'unknown kind',
      edit: (d) => ((sqlc(d) as { kind: string }).kind = 'cache'),
      expected: 'Invalid option',
    },
    {
      name: 'missing kind',
      edit: (d) => (d.kinds = d.kinds.filter((k) => k.id !== 'data')),
      expected: 'missing kind "data"',
    },
    {
      name: 'unknown module role',
      edit: (d) => ((sqlite(d) as { role: string }).role = 'plugin'),
      expected: 'Invalid option',
    },
    {
      name: 'trait requirement with no allowed values',
      edit: (d) => (sqlc(d).requires['sql-dialect'] = []),
      expected: 'Too small',
    },
    {
      name: 'invalid rule pattern',
      edit: (d) => (d.rules.projectName.pattern = '^[a-z'),
      expected: 'projectName: invalid regular expression',
    },
    {
      name: 'supported selection naming an unknown component',
      edit: (d) => (d.supported[0]!.database = 'postgres'),
      expected: 'supported selection names unknown component "database/postgres"',
    },
    {
      name: 'supported selection without every kind',
      edit: (d) => delete (d.supported[0] as Partial<CatalogData['supported'][number]>).data,
      expected: 'Invalid input',
    },
    {
      name: 'no supported selection',
      edit: (d) => (d.supported = []),
      expected: 'Too small',
    },
    {
      name: 'project name rule without reserved names',
      edit: (d) =>
        delete (d.rules.projectName as Partial<CatalogData['rules']['projectName']>).reserved,
      expected: 'Invalid input',
    },
    {
      name: 'missing rules',
      edit: (d) => delete (d as Partial<CatalogData>).rules,
      expected: 'Invalid input',
    },
  ];

  it.each(malformed)('rejects $name', ({ edit, expected }) => {
    const data = fixtureData();
    edit(data);
    const result = catalogSchema.safeParse(data);
    expect(result.success).toBe(false);
    expect(result.error?.issues.map((issue) => issue.message).join('\n')).toContain(expected);
  });
});
