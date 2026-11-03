/**
 * Test helpers for building catalogs from `catalog.json`, the real catalog
 * exported by `internal/tools/catalogexport`.
 */
import type { z } from 'zod';
import fixture from '@/src/generated/catalog.json';
import { parseCatalog, type catalogSchema, type Catalog } from '@/lib/catalog/schema';

/** Catalog data before validation, as found in catalog.json. */
export type CatalogData = z.input<typeof catalogSchema>;

/** Returns a fresh, mutable copy of the fixture data. */
export function fixtureData(): CatalogData {
  return structuredClone(fixture) as CatalogData;
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

/** A database whose traits do not satisfy sqlc, for unmet-trait tests. */
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
