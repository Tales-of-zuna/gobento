/**
 * Test helpers for building catalogs from `fixture-catalog.json`, a frozen
 * catalog in the format of `internal/catalogjson` (`make catalog`), filtered to
 * mvc and vertical-slice, nethttp, sqlite, gorm and sqlc, htmx, and Tailwind
 * CSS: architecture and data access have several options, framework and
 * database one, gorm's driver modules carry trait conditions, and the two
 * optional kinds (frontend, css) are linked by the `frontend-build` trait; no
 * component uses an npm package. Tests edit copies of it
 * for the cases they need, so they do not change whenever a component ships.
 * Tests about the real catalog import `@/lib/catalog/data` instead.
 */
import type { z } from 'zod';
import fixture from './fixture-catalog.json';
import { parseCatalog, type catalogSchema, type Catalog } from '@/lib/catalog/schema';

/** Catalog data before validation, as found in catalog.json. */
export type CatalogData = z.input<typeof catalogSchema>;

/** Returns a fresh, mutable copy of the fixture data. */
export function fixtureData(): CatalogData {
  // The JSON's inferred type is too wide to compare with the schema's input
  // (optional keys of heterogeneous records); the schema validates it.
  return structuredClone(fixture) as unknown as CatalogData;
}

/** Returns the validated fixture, optionally changed by `edit` first. */
export function catalogWith(edit: (data: CatalogData) => void = () => {}): Catalog {
  const data = fixtureData();
  edit(data);
  return parseCatalog(data);
}

/** Returns the module with the given path; modules are listed by path, not by component. */
export function moduleOf(data: CatalogData, path: string): CatalogData['modules'][number] {
  const found = data.modules.find((m) => m.path === path);
  if (!found) throw new Error(`catalog.json has no module ${path}`);
  return found;
}

/**
 * A PostgreSQL database that connects through pgx rather than database/sql, so
 * its traits do not satisfy sqlc or gorm: for unmet-trait tests and for a
 * database kind with several options.
 */
export function addPostgres(data: CatalogData): void {
  data.components.push({
    id: 'database/postgres',
    kind: 'database',
    name: 'postgres',
    title: 'PostgreSQL',
    description: 'A PostgreSQL server through pgx.',
    provides: { 'sql-dialect': 'postgresql', conn: 'pgx', 'sql-schema': 'schema.sql' },
    requires: {},
  });
}

/** The Vite frontends, by name, with the titles catalog.json gives them. */
const VITE_TITLES = { react: 'React', vue: 'Vue', svelte: 'Svelte' } as const;

/**
 * A TypeScript single-page app scaffolded by create-vite, the way catalog.json
 * lists the Vite frontends: the component, the scaffolder in `npm` (shared by
 * every Vite frontend added), the Node.js rule, and a supported selection with
 * it for every selection without a frontend. The frozen fixture predates them.
 */
export function addViteFrontend(data: CatalogData, name: keyof typeof VITE_TITLES): void {
  const id = `frontend/${name}`;
  const title = VITE_TITLES[name];
  data.components.push({
    id,
    kind: 'frontend',
    name,
    title,
    description: `A ${title} single-page app in TypeScript, scaffolded with create-vite, built with Vite, and embedded in the Go binary.`,
    provides: { 'frontend-build': 'vite' },
    requires: {},
  });
  data.node = { range: '>=v22.12.0', default: '26' };
  data.npm ??= [];
  const scaffolder = data.npm.find((pkg) => pkg.name === 'create-vite');
  if (scaffolder) scaffolder.components.push(id);
  else
    data.npm.push({ name: 'create-vite', components: [id], role: 'scaffolder', version: '9.3.3' });
  const backends = data.supported.filter((selection) => selection.frontend === undefined);
  data.supported.push(...backends.map((selection) => ({ ...selection, frontend: name })));
}

/**
 * Tailwind CSS for the Vite frontends already added with `addViteFrontend`,
 * the way catalog.json lists it: its npm packages (used only with
 * `frontend-build: vite`) and a supported selection with Tailwind for every
 * selection with a Vite frontend.
 */
export function addViteTailwind(data: CatalogData): void {
  data.npm ??= [];
  for (const name of ['@tailwindcss/vite', 'tailwindcss']) {
    data.npm.push({
      name,
      components: ['css/tailwind'],
      traits: { 'frontend-build': ['vite'] },
      role: 'dev-dependency',
      version: '4.3.4',
    });
  }
  const vite = new Set(
    data.components.filter((c) => c.provides?.['frontend-build'] === 'vite').map((c) => c.name),
  );
  const withVite = data.supported.filter(
    (selection) => selection.frontend !== undefined && vite.has(selection.frontend),
  );
  data.supported.push(...withVite.map((selection) => ({ ...selection, css: 'tailwind' })));
}
