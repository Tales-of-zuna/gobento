import { readFileSync } from 'node:fs';
import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { CatalogReference, ComponentsReference } from '@/components/components-reference';
import { catalogWith } from '@/lib/test/catalog';

const sqliteModules = () => screen.getByRole('table', { name: 'Modules of sqlite' });

describe('CatalogReference', () => {
  it('lists every component with its kind, traits, and module versions', () => {
    render(<ComponentsReference />);
    for (const heading of ['Architecture', 'HTTP framework', 'Database', 'Data access']) {
      expect(screen.getByRole('heading', { level: 2, name: heading })).toBeInTheDocument();
    }
    for (const name of ['ddd', 'mvc', 'vertical-slice', 'nethttp', 'sqlite', 'sqlc']) {
      expect(
        screen.getByRole('heading', { level: 3, name: new RegExp(`\\(${name}\\)`) }),
      ).toBeInTheDocument();
    }
    const sqlite = screen.getByRole('article', { name: /SQLite/ });
    expect(sqlite).toHaveTextContent('sql-dialect: sqlite');
    const sqlc = screen.getByRole('article', { name: /sqlc/ });
    expect(sqlc).toHaveTextContent('sql-dialect in [sqlite, postgresql]');
    const row = within(sqliteModules()).getByRole('row', { name: /modernc\.org\/sqlite/ });
    expect(row).toHaveTextContent('library');
    expect(row).toHaveTextContent('>=v1.61.0 <v2.0.0');
    expect(row).toHaveTextContent('v1.61.1');
    expect(screen.getByRole('table', { name: 'Modules of sqlc' })).toHaveTextContent(
      'tool (fixed per release)',
    );
  });

  it('marks the frontend kind optional and lists htmx with templ and its vendored asset (spec: Vendored asset)', () => {
    render(<CatalogReference catalog={catalogWith()} />);
    const kind = screen.getByRole('heading', { level: 2, name: /^Frontend/ });
    expect(kind).toHaveTextContent('optional');
    expect(screen.getByText(/without/, { selector: 'p' })).toHaveTextContent(
      'without --frontend (or with --frontend none) the project has none',
    );
    expect(screen.getByRole('table', { name: 'Modules of htmx' })).toHaveTextContent(
      'github.com/a-h/templ',
    );
    const assets = screen.getByRole('table', { name: 'Vendored assets of htmx' });
    const row = within(assets).getByRole('row', { name: /htmx\.org/ });
    expect(row).toHaveTextContent('vendored');
    expect(row).toHaveTextContent('2.0.11');
    expect(screen.getAllByRole('heading', { level: 2, name: /optional/ })).toHaveLength(1);
  });

  // Spec "Version bump": a changed default appears without editing docs content.
  it('shows a changed module default from the catalog', () => {
    const catalog = catalogWith((data) => {
      const sqlite = data.modules.find((m) => m.path === 'modernc.org/sqlite');
      if (!sqlite) throw new Error('fixture has no modernc.org/sqlite');
      sqlite.default = 'v1.62.3';
      data.go.default = '1.26';
    });
    render(<CatalogReference catalog={catalog} />);
    const row = within(sqliteModules()).getByRole('row', { name: /modernc\.org\/sqlite/ });
    expect(row).toHaveTextContent('v1.62.3');
    expect(row).not.toHaveTextContent('v1.61.1');
    expect(screen.getByRole('table', { name: 'Go versions' })).toHaveTextContent('1.26');
  });

  it('is the only source of versions on the Components page', () => {
    const mdx = readFileSync('content/docs/components.mdx', 'utf8');
    expect(mdx).toContain('<ComponentsReference />');
    expect(mdx).not.toMatch(/\bv?\d+\.\d+(\.\d+)?\b/);
  });
});
