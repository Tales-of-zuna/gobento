/**
 * Test helpers for building catalogs from `fixture-catalog.json`, a frozen
 * catalog in the format of `internal/catalogjson` (`make catalog`), filtered to
 * mvc and vertical-slice, nethttp, sqlite, and gorm and sqlc: architecture and
 * data access have several options, framework and database one, and gorm's
 * driver modules carry trait conditions. Tests edit copies of it
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
