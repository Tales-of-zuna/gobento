/**
 * Builds the `gobento new` command for a configurator state (design D4). The
 * flag order is fixed to match the CLI reference, and every flag whose value
 * equals what gobento would choose by itself is omitted.
 */
import {
  selectedModules,
  selectionProblems,
  type Problem,
  type Selection,
} from '@/lib/catalog/compat';
import { KIND_IDS } from '@/lib/catalog/kinds';
import type { Catalog, KindId } from '@/lib/catalog/schema';
import { sameVersion } from '@/lib/catalog/semver';

/** The CLI flag for each component kind. */
export const KIND_FLAGS: Readonly<Record<KindId, string>> = {
  architecture: 'arch',
  framework: 'framework',
  database: 'db',
  data: 'data',
};

/** Everything the configurator lets a visitor choose. */
export interface ConfigState extends Selection {
  /** Project name, the positional argument of `gobento new`. */
  readonly name: string;
  /** Go module path; empty means the CLI's default, the project name. */
  readonly module: string;
}

/** A problem with one field of the state. */
export interface FieldError {
  readonly field: 'name' | 'module' | Problem['field'];
  readonly message: string;
}

/** The command, or the reasons no command can be offered. */
export type BuildResult =
  | { readonly ok: true; readonly command: string }
  | { readonly ok: false; readonly errors: readonly FieldError[] };

/** Validates a value against one of the catalog's input rules; returns the rule's message on failure. */
function checkRule(rule: Catalog['rules']['modulePath'], value: string): string | undefined {
  return new RegExp(rule.pattern).test(value) ? undefined : rule.message;
}

/**
 * Validates the project name and module path with the CLI's rules. The module
 * path pattern approximates the CLI's check (see `ModulePathPattern` in
 * internal/generate): it does not reject Windows-reserved path elements such
 * as `con`, which `gobento new` still reports.
 */
export function inputErrors(catalog: Catalog, state: ConfigState): FieldError[] {
  const errors: FieldError[] = [];
  const { projectName } = catalog.rules;
  const nameError =
    checkRule(projectName, state.name) ??
    (projectName.reserved.includes(state.name) ? projectName.reservedMessage : undefined);
  if (nameError) errors.push({ field: 'name', message: nameError });
  const moduleError =
    state.module === '' ? undefined : checkRule(catalog.rules.modulePath, state.module);
  if (moduleError) errors.push({ field: 'module', message: moduleError });
  return errors;
}

/** Returns the components of a kind, which decides whether its flag is needed. */
function optionsOf(catalog: Catalog, kind: KindId) {
  return catalog.components.filter((c) => c.kind === kind);
}

/**
 * Builds the command for a state, or lists every problem that prevents one:
 * invalid name or module path, a kind with several options left unchosen,
 * or an incompatible selection.
 */
export function buildCommand(catalog: Catalog, state: ConfigState): BuildResult {
  const errors: FieldError[] = [...inputErrors(catalog, state)];
  for (const kind of KIND_IDS) {
    const options = optionsOf(catalog, kind);
    if (options.length > 1 && state.components[kind] === undefined) {
      const title = (catalog.kinds.find((k) => k.id === kind)?.title ?? kind).toLowerCase();
      const article = /^[aeiou]/.test(title) ? 'an' : 'a';
      errors.push({
        field: kind,
        message: `Choose ${article} ${title}: ${options.map((o) => o.name).join(', ')}.`,
      });
    }
  }
  errors.push(...selectionProblems(catalog, state));
  if (errors.length > 0) return { ok: false, errors };

  const args = ['gobento', 'new', state.name];
  if (state.module !== '' && state.module !== state.name) args.push('--module', state.module);
  for (const kind of KIND_IDS) {
    const name = state.components[kind];
    if (name !== undefined && optionsOf(catalog, kind).length > 1) {
      args.push(`--${KIND_FLAGS[kind]}`, name);
    }
  }
  if (!sameVersion(state.go, catalog.go.default, 'go')) args.push('--go', state.go);
  const libraries = selectedModules(catalog, state)
    .filter((m) => m.role === 'library')
    .sort((a, b) => (a.path < b.path ? -1 : 1));
  for (const { path, default: fallback } of libraries) {
    const version = state.pins[path];
    if (version !== undefined && !sameVersion(version, fallback, 'module')) {
      args.push('--pin', `${path}=${version}`);
    }
  }
  return { ok: true, command: args.join(' ') };
}
