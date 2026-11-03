/**
 * Compatibility evaluation for a selection of components and versions: unmet
 * traits, versions outside their range, and violated constraints, each with a
 * human-readable reason. Mirrors the CLI's resolution order (design D4 of
 * add-generator-foundation): defaults, then overrides, then range checks,
 * then every matching constraint.
 */
import type { Catalog, Component, KindId, Module } from './schema';
import { KIND_IDS } from './kinds';
import { parseGoVersion, parseModuleVersion, satisfies } from './semver';

/** Chosen components and versions. */
export interface Selection {
  /** Chosen component name per kind; a kind without a choice is undefined. */
  readonly components: Readonly<Partial<Record<KindId, string>>>;
  /** Chosen Go version, for example `1.27`. */
  readonly go: string;
  /**
   * Version overrides by module path. Overrides for modules no selected
   * component uses are inert and ignored.
   */
  readonly pins: Readonly<Record<string, string>>;
}

/** Where a problem is shown: a kind's choice, the Go version, or one module's version. */
export type ProblemField = KindId | 'go' | `pin:${string}`;

/** One reason a selection is invalid. */
export interface Problem {
  readonly field: ProblemField;
  readonly message: string;
  /** Component ids, module paths, and `go` that the problem is about. */
  readonly involves: readonly string[];
}

/** Returns the components chosen in a selection, in kind order. */
export function selectedComponents(catalog: Catalog, selection: Selection): Component[] {
  return KIND_IDS.flatMap((kind) => {
    const name = selection.components[kind];
    const component = catalog.components.find((c) => c.kind === kind && c.name === name);
    return component ? [component] : [];
  });
}

/** Returns the modules used by the chosen components. */
export function selectedModules(catalog: Catalog, selection: Selection): Module[] {
  const ids = new Set(selectedComponents(catalog, selection).map((c) => c.id));
  return catalog.modules.filter((m) => m.components.some((id) => ids.has(id)));
}

/** Returns the version each selected module gets: its override, or else its default. */
export function resolvedVersions(catalog: Catalog, selection: Selection): Map<string, string> {
  return new Map(
    selectedModules(catalog, selection).map((m) => [m.path, selection.pins[m.path] ?? m.default]),
  );
}

function traitProblems(catalog: Catalog, selection: Selection): Problem[] {
  const components = selectedComponents(catalog, selection);
  const complete = KIND_IDS.every((kind) => selection.components[kind] !== undefined);
  const problems: Problem[] = [];
  for (const requirer of components) {
    for (const [trait, allowed] of Object.entries(requirer.requires)) {
      const providers = components.filter((c) => c !== requirer && trait in c.provides);
      const wanted = allowed.join(' or ');
      if (providers.length === 0) {
        // Until every kind is chosen, a provider may still be added.
        if (complete) {
          problems.push({
            field: requirer.kind,
            message: `${requirer.title} requires a component that provides ${trait} ${wanted}.`,
            involves: [requirer.id],
          });
        }
        continue;
      }
      for (const provider of providers) {
        const value = provider.provides[trait]!;
        if (allowed.includes(value)) continue;
        problems.push({
          field: requirer.kind,
          message: `${requirer.title} requires ${trait} ${wanted}, but ${provider.title} provides ${value}.`,
          involves: [requirer.id, provider.id],
        });
      }
    }
  }
  return problems;
}

function versionProblems(catalog: Catalog, selection: Selection): Problem[] {
  const problems: Problem[] = [];
  const { go } = selection;
  if (!parseGoVersion(go)) {
    problems.push({
      field: 'go',
      message: `"${go}" is not a Go version such as 1.27.`,
      involves: ['go'],
    });
  } else if (!satisfies(go, catalog.go.range, 'go')) {
    problems.push({
      field: 'go',
      message: `Go ${go} is not supported; choose a version in ${catalog.go.range}.`,
      involves: ['go'],
    });
  }

  const versions = resolvedVersions(catalog, selection);
  for (const mod of selectedModules(catalog, selection)) {
    const field = `pin:${mod.path}` as const;
    const version = versions.get(mod.path)!;
    const pinned = mod.path in selection.pins;
    if (pinned && mod.role === 'tool' && version !== mod.default) {
      problems.push({
        field,
        message: `${mod.path} is a tool; its version is fixed per gobento release.`,
        involves: [mod.path],
      });
      continue;
    }
    if (!parseModuleVersion(version)) {
      problems.push({
        field,
        message: `"${version}" is not a module version such as v1.2.3.`,
        involves: [mod.path],
      });
      continue;
    }
    if (!satisfies(version, mod.range, 'module')) {
      problems.push({
        field,
        message: `${mod.path} ${version} is outside the supported range ${mod.range}.`,
        involves: [mod.path],
      });
      continue;
    }
    for (const { when, requires } of mod.constraints) {
      if (!satisfies(version, when, 'module')) continue;
      if (requires.target === 'go') {
        if (parseGoVersion(go) && !satisfies(go, requires.range, 'go')) {
          problems.push({
            field,
            message: `${mod.path} ${version} needs Go ${requires.range}, but Go ${go} is selected.`,
            involves: [mod.path, 'go'],
          });
        }
        continue;
      }
      const other = versions.get(requires.target);
      if (other !== undefined && !satisfies(other, requires.range, 'module')) {
        problems.push({
          field,
          message: `${mod.path} ${version} needs ${requires.target} ${requires.range}, but ${requires.target} ${other} is selected.`,
          involves: [mod.path, requires.target],
        });
      }
    }
  }
  return problems;
}

/** Lists every problem with a selection; an empty list means the CLI accepts it. */
export function selectionProblems(catalog: Catalog, selection: Selection): Problem[] {
  return [...traitProblems(catalog, selection), ...versionProblems(catalog, selection)];
}

function reasonsInvolving(problems: Problem[], subjects: readonly string[]): string[] {
  return problems.filter((p) => p.involves.some((s) => subjects.includes(s))).map((p) => p.message);
}

/**
 * Explains why choosing a component would make the selection invalid. An
 * empty list means the option can be offered; otherwise it is disabled.
 */
export function componentReasons(
  catalog: Catalog,
  selection: Selection,
  kind: KindId,
  name: string,
): string[] {
  const component = catalog.components.find((c) => c.kind === kind && c.name === name);
  if (!component) return [`No ${kind} named "${name}" is supported.`];
  const candidate: Selection = {
    ...selection,
    components: { ...selection.components, [kind]: name },
  };
  const modules = catalog.modules.filter((m) => m.components.includes(component.id));
  return reasonsInvolving(selectionProblems(catalog, candidate), [
    component.id,
    ...modules.map((m) => m.path),
  ]);
}

/** Explains why a module version cannot be chosen with the rest of the selection. */
export function moduleVersionReasons(
  catalog: Catalog,
  selection: Selection,
  path: string,
  version: string,
): string[] {
  const candidate: Selection = { ...selection, pins: { ...selection.pins, [path]: version } };
  return reasonsInvolving(selectionProblems(catalog, candidate), [path]);
}

/** Explains why a Go version cannot be chosen with the rest of the selection. */
export function goVersionReasons(catalog: Catalog, selection: Selection, go: string): string[] {
  return reasonsInvolving(selectionProblems(catalog, { ...selection, go }), ['go']);
}
