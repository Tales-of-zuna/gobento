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
      'data/gorm',
      'data/sqlc',
      'frontend/htmx',
      'css/tailwind',
    ]);
    expect(result.data?.kinds.filter((k) => k.optional).map((k) => k.id)).toEqual([
      'frontend',
      'css',
    ]);
    expect(result.data?.assets.map((a) => [a.name, a.version])).toEqual([['htmx.org', '2.0.11']]);
    expect(result.data?.binaries.map((b) => [b.name, b.version, b.components])).toEqual([
      ['tailwindcss', '4.3.4', ['css/tailwind']],
    ]);
    expect(result.data?.binaries[0]?.platforms).toContain('linux/amd64');
  });

  it('accepts supported selections with and without an optional kind', () => {
    const result = catalogSchema.safeParse(fixtureData());
    const optional = result.data?.supported.map(
      (s) => `${s.frontend ?? 'none'}+${s.css ?? 'none'}`,
    );
    expect(optional?.filter((f) => f === 'none+none')).toHaveLength(4);
    expect(optional?.filter((f) => f === 'htmx+none')).toHaveLength(4);
    expect(optional?.filter((f) => f === 'htmx+tailwind')).toHaveLength(4);
    expect(optional).toHaveLength(12);
  });

  it('defaults a kind to required and the assets and binaries to none when they are left out', () => {
    const data = fixtureData();
    for (const kind of data.kinds) if (!kind.optional) delete kind.optional;
    delete data.assets;
    delete data.binaries;
    const result = catalogSchema.safeParse(data);
    expect(result.error?.issues).toBeUndefined();
    expect(result.data?.kinds.filter((k) => !k.optional)).toHaveLength(4);
    expect(result.data?.assets).toEqual([]);
    expect(result.data?.binaries).toEqual([]);
  });

  it('accepts the tooling versions and defaults them to none', () => {
    const data = fixtureData();
    expect(catalogSchema.safeParse(data).data?.tooling).toEqual({
      images: [],
      actions: [],
      tools: [],
    });
    data.tooling = {
      images: [{ key: 'go', image: 'golang', tag: '{go}', digest: '' }],
      actions: [
        {
          name: 'actions/checkout',
          version: 'v7.0.1',
          sha: '3d3c42e5aac5ba805825da76410c181273ba90b1',
        },
      ],
      tools: [{ name: 'golangci-lint', version: 'v2.15.0' }],
    };
    const result = catalogSchema.safeParse(data);
    expect(result.error?.issues).toBeUndefined();
    expect(result.data?.tooling.images[0]?.tag).toBe('{go}');
    data.tooling.tools = [{ name: 'golangci-lint', version: '' }];
    expect(catalogSchema.safeParse(data).success).toBe(false);
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
      name: 'module trait condition with no allowed values',
      edit: (d) => (sqlite(d).traits = { 'sql-dialect': [] }),
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
      name: 'supported selection without a required kind',
      edit: (d) => delete d.supported[0]!.data,
      expected: 'supported selection lacks a data',
    },
    {
      name: 'supported selection without a kind that is no longer optional',
      edit: (d) => {
        d.kinds.find((k) => k.id === 'frontend')!.optional = false;
      },
      expected: 'supported selection lacks a frontend',
    },
    {
      name: 'supported selection naming an unknown optional component',
      edit: (d) => (d.supported[0]!.frontend = 'react'),
      expected: 'supported selection names unknown component "frontend/react"',
    },
    {
      name: 'asset of an unknown component',
      edit: (d) => (d.assets![0]!.components = ['frontend/react']),
      expected: 'htmx.org: unknown component "frontend/react"',
    },
    {
      name: 'asset without a version',
      edit: (d) => (d.assets![0]!.version = ''),
      expected: 'Too small',
    },
    {
      name: 'asset without a source',
      edit: (d) => (d.assets![0]!.source = ''),
      expected: 'Too small',
    },
    {
      name: 'asset without components',
      edit: (d) => (d.assets![0]!.components = []),
      expected: 'Too small',
    },
    {
      name: 'duplicate asset',
      edit: (d) => d.assets!.push({ ...d.assets![0]! }),
      expected: 'duplicate asset name',
    },
    {
      name: 'binary of an unknown component',
      edit: (d) => (d.binaries![0]!.components = ['css/bootstrap']),
      expected: 'tailwindcss: unknown component "css/bootstrap"',
    },
    {
      name: 'binary without a version',
      edit: (d) => (d.binaries![0]!.version = ''),
      expected: 'Too small',
    },
    {
      name: 'binary without platforms',
      edit: (d) => (d.binaries![0]!.platforms = []),
      expected: 'Too small',
    },
    {
      name: 'binary with a malformed platform',
      edit: (d) => (d.binaries![0]!.platforms = ['linux-amd64']),
      expected: 'Invalid string',
    },
    {
      name: 'binary without components',
      edit: (d) => (d.binaries![0]!.components = []),
      expected: 'Too small',
    },
    {
      name: 'duplicate binary',
      edit: (d) => d.binaries!.push({ ...d.binaries![0]! }),
      expected: 'duplicate binary name',
    },
    {
      name: 'supported selection naming an unknown CSS framework',
      edit: (d) => (d.supported[0]!.css = 'bootstrap'),
      expected: 'supported selection names unknown component "css/bootstrap"',
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
