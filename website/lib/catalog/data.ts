/**
 * The catalog the site is built from. It is validated when the module loads,
 * so an invalid `catalog.json` fails the build instead of rendering bad data.
 */
import data from '@/src/generated/catalog.json';
import { parseCatalog, type Catalog } from './schema';

export const catalog: Catalog = parseCatalog(data);
