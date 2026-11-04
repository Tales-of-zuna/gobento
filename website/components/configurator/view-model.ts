/**
 * Pure helpers that turn the catalog and a configurator state into what the
 * page shows: option cards and their states, Go versions, module version
 * fields, the bento tray, and the file tree. All compatibility logic comes
 * from `lib/catalog/compat` and `lib/command/build`; nothing is re-derived here.
 */
import {
  componentReasons,
  goVersionReasons,
  selectedModules,
  type Selection,
} from '@/lib/catalog/compat';
import { KIND_IDS, titleInSentence } from '@/lib/catalog/kinds';
import type { Catalog, KindId, Module } from '@/lib/catalog/schema';
import {
  compareVersions,
  parseGoVersion,
  parseRange,
  sameVersion,
  satisfies,
} from '@/lib/catalog/semver';
import type { BuildResult, ConfigState, FieldError } from '@/lib/command/build';
import { ROADMAP, type PlannedComponent } from './roadmap';

/** How an option card is drawn. */
export type OptionStatus = 'selected' | 'available' | 'incompatible' | 'soon';

/** One choice card. */
export interface OptionView {
  /** Value of the CLI flag (the component name, or the Go version). */
  readonly value: string;
  readonly title: string;
  readonly description: string;
  readonly status: OptionStatus;
  /** Why the option cannot be chosen; empty unless incompatible. */
  readonly reasons: readonly string[];
}

/** One kind's group of choice cards. */
export interface KindGroupView {
  readonly kind: KindId;
  readonly title: string;
  readonly description: string;
  readonly options: readonly OptionView[];
  /** A problem with the current choice, shown under the group. */
  readonly error?: string;
}

/** Errors by field, as returned by {@link buildCommand}; only the first message per field is kept. */
export function errorsByField(result: BuildResult): ReadonlyMap<FieldError['field'], string> {
  const map = new Map<FieldError['field'], string>();
  if (result.ok) return map;
  for (const { field, message } of result.errors) if (!map.has(field)) map.set(field, message);
  return map;
}

/** The choice cards for each component kind, in flag order. */
export function kindGroups(
  catalog: Catalog,
  state: Selection,
  errors: ReadonlyMap<FieldError['field'], string>,
  roadmap: Readonly<Record<KindId, readonly PlannedComponent[]>> = ROADMAP,
): KindGroupView[] {
  return KIND_IDS.map((kind) => {
    const meta = catalog.kinds.find((k) => k.id === kind)!;
    const components = catalog.components.filter((c) => c.kind === kind);
    const options: OptionView[] = components.map((component) => {
      const selected = state.components[kind] === component.name;
      const reasons = selected ? [] : componentReasons(catalog, state, kind, component.name);
      return {
        value: component.name,
        title: component.title,
        description: component.description,
        status: selected ? 'selected' : reasons.length > 0 ? 'incompatible' : 'available',
        reasons,
      };
    });
    const known = new Set(components.map((c) => c.name));
    for (const planned of roadmap[kind]) {
      if (known.has(planned.name)) continue;
      options.push({
        value: planned.name,
        title: planned.title,
        description: 'On the roadmap',
        status: 'soon',
        reasons: [],
      });
    }
    return {
      kind,
      title: meta.title,
      description: components.length === 1 ? 'One option today, preselected.' : meta.description,
      options,
      error: errors.get(kind),
    };
  });
}

/**
 * Every Go minor version the catalog's range allows, oldest first. Ranges do
 * not list versions, so the minors between the range's bounds (falling back
 * to the default for an open side) are tried one by one.
 */
export function goVersions(catalog: Catalog): string[] {
  const fallback = parseGoVersion(catalog.go.default)!;
  const comparators = parseRange(catalog.go.range, 'go') ?? [];
  let low = fallback;
  let high = fallback;
  for (const { operator, version } of comparators) {
    if (version.major !== fallback.major) continue;
    if (operator !== '<' && operator !== '<=' && compareVersions(version, low) < 0) low = version;
    if (operator !== '>' && operator !== '>=' && compareVersions(version, high) > 0) high = version;
  }
  const versions: string[] = [];
  for (let minor = low.minor; minor <= high.minor; minor++) {
    const candidate = `${fallback.major}.${minor}`;
    if (satisfies(candidate, catalog.go.range, 'go')) versions.push(candidate);
  }
  return versions;
}

/** The Go version cards. */
export function goOptions(catalog: Catalog, state: Selection): OptionView[] {
  return goVersions(catalog).map((go) => {
    const selected = sameVersion(go, state.go, 'go');
    const reasons = goVersionReasons(catalog, state, go);
    const isDefault = sameVersion(go, catalog.go.default, 'go');
    return {
      value: go,
      title: `Go ${go}`,
      description: isDefault ? 'Default' : 'Supported',
      status: selected ? 'selected' : reasons.length > 0 ? 'incompatible' : 'available',
      reasons: selected ? [] : reasons,
    };
  });
}

/** A module version field in the Advanced section. */
export interface ModuleFieldView {
  readonly path: string;
  /** The component titles using the module, for the label. */
  readonly usedBy: string;
  readonly fixed: boolean;
  /** What the input shows: the override, or else the default. */
  readonly value: string;
  readonly hint: string;
  readonly error?: string;
}

function goNeeds(module: Module, version: string): string | undefined {
  const needs = module.constraints
    .filter((c) => c.requires.target === 'go' && satisfies(version, c.when, 'module'))
    .map((c) => c.requires.range);
  return needs.length > 0 ? needs.join(' ') : undefined;
}

/** The version fields for the modules the chosen components use. */
export function moduleFields(
  catalog: Catalog,
  state: Selection,
  errors: ReadonlyMap<FieldError['field'], string>,
): ModuleFieldView[] {
  return selectedModules(catalog, state).map((module) => {
    const usedBy = module.components
      .map((id) => catalog.components.find((c) => c.id === id)?.title ?? id)
      .join(', ');
    if (module.role === 'tool') {
      return {
        path: module.path,
        usedBy,
        fixed: true,
        value: module.default,
        hint: 'Tool versions are fixed per gobento release.',
      };
    }
    const value = state.pins[module.path] ?? module.default;
    const error = errors.get(`pin:${module.path}`);
    const needs = goNeeds(module, value);
    const hint =
      `Range ${module.range} · default ${module.default}` +
      (needs ? ` · this version needs Go ${needs}` : '');
    return { path: module.path, usedBy, fixed: false, value, hint, error };
  });
}

/** Returns the pins after the visitor typed `value` for `path`: a default clears the pin. */
export function withPin(
  catalog: Catalog,
  pins: Readonly<Record<string, string>>,
  path: string,
  value: string,
): Record<string, string> {
  const rest = Object.fromEntries(Object.entries(pins).filter(([p]) => p !== path));
  const target = catalog.modules.find((m) => m.path === path);
  if (target && sameVersion(value.trim(), target.default, 'module')) return rest;
  return { ...rest, [path]: value.trim() };
}

/** Bento food and colour for each kind's compartment, in tray order. */
export const COMPARTMENTS: Readonly<Record<KindId, { food: string; piece: string }>> = {
  architecture: { food: 'rice ball', piece: 'rice' },
  framework: { food: 'pickled plum', piece: 'ume' },
  database: { food: 'omelette', piece: 'tamago' },
  data: { food: 'edamame', piece: 'eda' },
};

/** One compartment of the bento tray. */
export interface CellView {
  readonly kind: KindId;
  readonly label: string;
  /** Title of the chosen component, if any. */
  readonly value?: string;
  readonly food: string;
  readonly piece: string;
}

/** The four tray compartments, one per kind. */
export function trayCells(catalog: Catalog, state: Selection): CellView[] {
  return KIND_IDS.map((kind) => {
    const name = state.components[kind];
    const component = catalog.components.find((c) => c.kind === kind && c.name === name);
    return {
      kind,
      label: catalog.kinds.find((k) => k.id === kind)!.title,
      value: component?.title,
      ...COMPARTMENTS[kind],
    };
  });
}

/** The sentence shown in place of the command when none can be offered. */
export function blockedMessage(catalog: Catalog, result: BuildResult): string {
  if (result.ok) return '';
  const first = result.errors[0]!.field;
  let what: string;
  if (first === 'name') what = 'the project name';
  else if (first === 'module') what = 'the module path';
  else if (first === 'go') what = 'the Go version';
  else if (first.startsWith('pin:')) what = `the ${first.slice(4)} version under Advanced`;
  else
    what = `the ${titleInSentence(catalog.kinds.find((k) => k.id === first)?.title ?? first)} choice`;
  return `No command yet. Fix ${what} to get a copyable command.`;
}

/** One line of the "What you'll get" tree. */
export interface TreeLine {
  readonly label: string;
  readonly depth: 0 | 1 | 2;
  readonly dir: boolean;
}

/**
 * Files and folders each component adds to the generated project, by
 * component id. Illustrative: the generator owns the real layout, and a
 * component missing here simply adds nothing to the preview. The database
 * package sits where the architecture puts it, so it is listed there.
 */
const COMPONENT_FILES: Readonly<Record<string, readonly string[]>> = {
  'architecture/mvc': [
    'internal/models/',
    'internal/controllers/',
    'internal/store/',
    'internal/database/',
  ],
  'architecture/ddd': [
    'internal/domain/',
    'internal/app/',
    'internal/adapters/httpapi/',
    'internal/adapters/persistence/',
    'internal/adapters/database/',
  ],
  'architecture/vertical-slice': [
    'internal/items/',
    'internal/items/itemshttp/',
    'internal/items/itemsstore/',
    'internal/platform/database/',
  ],
  'data/sqlc': ['sqlc.yaml'],
};

/**
 * Files the base layer (components/base) adds to every project: the
 * developer tooling. Files under .github/ are listed inside that directory.
 */
const BASE_FILES: readonly string[] = [
  '.github/workflows/ci.yml',
  '.github/dependabot.yml',
  '.dockerignore',
  '.golangci.yml',
  'Dockerfile',
  'Makefile',
];

/** The project tree preview for the chosen components. */
export function fileTree(catalog: Catalog, state: ConfigState, nameValid: boolean): TreeLine[] {
  const name = nameValid ? state.name : '<name>';
  const ids = KIND_IDS.flatMap((kind) => {
    const component = catalog.components.find(
      (c) => c.kind === kind && c.name === state.components[kind],
    );
    return component ? [component.id] : [];
  });
  const files = ids.flatMap((id) => COMPONENT_FILES[id] ?? []);
  const internal = files
    .filter((f) => f.startsWith('internal/'))
    .map((f) => f.slice('internal/'.length))
    .sort();
  const top = [
    ...files.filter((f) => !f.startsWith('internal/')),
    ...BASE_FILES.filter((f) => !f.startsWith('.github/')),
  ].sort();
  const github = BASE_FILES.filter((f) => f.startsWith('.github/'))
    .map((f) => f.slice('.github/'.length))
    .sort();
  return [
    { label: `${name}/`, depth: 0, dir: true },
    { label: '.github/', depth: 1, dir: true },
    ...github.map((label) => ({ label, depth: 2 as const, dir: false })),
    { label: `cmd/${name}/main.go`, depth: 1, dir: false },
    ...(internal.length > 0
      ? [
          { label: 'internal/', depth: 1 as const, dir: true },
          ...internal.map((label) => ({ label, depth: 2 as const, dir: false })),
        ]
      : []),
    ...top.map((label) => ({ label, depth: 1 as const, dir: false })),
    { label: 'go.mod', depth: 1, dir: false },
    { label: 'README.md', depth: 1, dir: false },
  ];
}
