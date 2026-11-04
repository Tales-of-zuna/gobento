/**
 * Data for the home page's catalog-driven sections: the conveyor-belt menu,
 * the "Pick your stack" tray, and the "Versions you can trust" example. The
 * shipping components and every version come from the catalog, so the page
 * cannot drift from what the CLI offers; only the roadmap is listed here.
 */
import { KIND_IDS, type Catalog, type Component, type KindId } from '@/lib/catalog/schema';

/** Short kind labels that fit on a conveyor plate. */
export const KIND_SHORT: Readonly<Record<KindId, string>> = {
  architecture: 'Architecture',
  framework: 'HTTP',
  database: 'Database',
  data: 'Data',
  frontend: 'Frontend',
};

/** A component planned for a later release (shown dashed). */
export interface RoadmapItem {
  readonly kind: KindId;
  readonly name: string;
}

/** Components on the roadmap; they join once they pass the CI matrix. */
export const ROADMAP: readonly RoadmapItem[] = [{ kind: 'frontend', name: 'Tailwind CSS' }];

/** One plate on the belt. */
export interface Plate {
  readonly kind: KindId;
  readonly name: string;
  readonly status: 'ready' | 'roadmap';
}

/**
 * A component's display name for small spaces: its title, or the title's
 * initials when the title is long (`Model-View-Controller` → `MVC`).
 */
export function shortTitle(component: Pick<Component, 'title'>): string {
  const { title } = component;
  if (title.length <= 12) return title;
  return title
    .split(/[\s-]+/)
    .filter(Boolean)
    .map((word) => word[0]!.toUpperCase())
    .join('');
}

/** The belt's plates: every shipping component in kind order, then the roadmap. */
export function beltPlates(catalog: Catalog, roadmap: readonly RoadmapItem[] = ROADMAP): Plate[] {
  const ready = KIND_IDS.flatMap((kind) =>
    catalog.components
      .filter((c) => c.kind === kind)
      .map((c): Plate => ({ kind, name: shortTitle(c), status: 'ready' })),
  );
  const planned = roadmap.map((item): Plate => ({ ...item, status: 'roadmap' }));
  return [...ready, ...planned];
}

/** One compartment of the "Pick your stack" tray. */
export interface StackCell {
  readonly kind: KindId;
  readonly title: string;
  /** Whether the kind may be left out (the frontend: none means API only). */
  readonly optional: boolean;
  readonly options: readonly { readonly name: string; readonly description: string }[];
  readonly roadmap: readonly string[];
}

/** One compartment per kind, with its shipping options and roadmap names. */
export function stackCells(
  catalog: Catalog,
  roadmap: readonly RoadmapItem[] = ROADMAP,
): StackCell[] {
  return KIND_IDS.map((kind) => ({
    kind,
    title: catalog.kinds.find((k) => k.id === kind)?.title ?? kind,
    optional: catalog.kinds.find((k) => k.id === kind)?.optional ?? false,
    options: catalog.components
      .filter((c) => c.kind === kind)
      .map((c) => ({ name: shortTitle(c), description: c.description })),
    roadmap: roadmap.filter((item) => item.kind === kind).map((item) => item.name),
  }));
}

/** The versions example: a rule from the catalog and the Go choice it rules out. */
export interface VersionShowcase {
  readonly module: string;
  readonly version: string;
  readonly role: string;
  /** The rule as written in the catalog, for example `>=1.26`. */
  readonly goRange: string;
  /** Minimum Go version the rule allows, for example `1.26`. */
  readonly goMinimum: string;
  /** The default Go version, which meets every rule. */
  readonly goDefault: string;
  /** The Go release just below the minimum, which the rule rules out. */
  readonly goDisabled: string;
  /** Why {@link goDisabled} cannot be chosen. */
  readonly reason: string;
}

/**
 * Picks the first library whose default version requires a minimum Go
 * version, and derives the example from it. Returns undefined when no
 * library has such a rule.
 */
export function versionShowcase(catalog: Catalog): VersionShowcase | undefined {
  for (const mod of catalog.modules) {
    if (mod.role !== 'library') continue;
    for (const constraint of mod.constraints) {
      if (constraint.requires.target !== 'go') continue;
      const match = /^>=\s*(\d+)\.(\d+)$/.exec(constraint.requires.range.trim());
      if (!match) continue;
      const [, major, minor] = match;
      const goMinimum = `${major}.${minor}`;
      const goDisabled = `${major}.${Number(minor) - 1}`;
      return {
        module: mod.path,
        version: mod.default,
        role: mod.role,
        goRange: constraint.requires.range,
        goMinimum,
        goDefault: catalog.go.default,
        goDisabled,
        reason: `${mod.path} ${mod.default} needs Go ${goMinimum} or newer`,
      };
    }
  }
  return undefined;
}
