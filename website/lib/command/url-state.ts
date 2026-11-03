/**
 * Configurator state in the page URL (design D4), so a shared link restores
 * the same choices. Parameters use the CLI's flag names: `name`, `module`,
 * `arch`, `framework`, `db`, `data`, `go`, and repeated `pin=<path>=<version>`.
 * Values equal to their default are left out to keep links short.
 */
import { KIND_IDS, type Catalog } from '@/lib/catalog/schema';
import { parseGoVersion, parseModuleVersion, sameVersion, satisfies } from '@/lib/catalog/semver';
import { KIND_FLAGS, type ConfigState } from './build';

/** Project name a fresh configurator starts with. */
export const DEFAULT_PROJECT_NAME = 'myapp';

/** The state a first-time visitor sees: single-option kinds and default versions preselected. */
export function defaultState(catalog: Catalog): ConfigState {
  const components: ConfigState['components'] = Object.fromEntries(
    KIND_IDS.flatMap((kind) => {
      const options = catalog.components.filter((c) => c.kind === kind);
      return options.length === 1 ? [[kind, options[0]!.name]] : [];
    }),
  );
  return { name: DEFAULT_PROJECT_NAME, module: '', components, go: catalog.go.default, pins: {} };
}

/** Encodes the non-default parts of a state as a query string (without `?`). */
export function encodeState(catalog: Catalog, state: ConfigState): string {
  const defaults = defaultState(catalog);
  const params = new URLSearchParams();
  if (state.name !== defaults.name) params.set('name', state.name);
  if (state.module !== '') params.set('module', state.module);
  for (const kind of KIND_IDS) {
    const name = state.components[kind];
    if (name !== undefined && name !== defaults.components[kind])
      params.set(KIND_FLAGS[kind], name);
  }
  if (!sameVersion(state.go, catalog.go.default, 'go')) params.set('go', state.go);
  for (const [path, version] of Object.entries(state.pins).sort(([a], [b]) => (a < b ? -1 : 1))) {
    params.append('pin', `${path}=${version}`);
  }
  return params.toString();
}

/** A decoded state, with a notice for each value that was replaced by its default. */
export interface DecodedState {
  readonly state: ConfigState;
  readonly notices: readonly string[];
}

/**
 * Restores a state from a query string. Components, Go versions, and module
 * versions the catalog does not contain fall back to their defaults, each with
 * a notice. The name and module path are kept as given, so the form can show
 * why they are invalid.
 */
export function decodeState(catalog: Catalog, query: string | URLSearchParams): DecodedState {
  const params = new URLSearchParams(query);
  const defaults = defaultState(catalog);
  const notices: string[] = [];

  const components = { ...defaults.components };
  for (const kind of KIND_IDS) {
    const name = params.get(KIND_FLAGS[kind]);
    if (name === null) continue;
    if (catalog.components.some((c) => c.kind === kind && c.name === name)) {
      components[kind] = name;
    } else {
      const title = catalog.kinds.find((k) => k.id === kind)?.title ?? kind;
      notices.push(
        `The link names ${title.toLowerCase()} "${name}", which is not available; the default is used.`,
      );
    }
  }

  let go = defaults.go;
  const requestedGo = params.get('go');
  if (requestedGo !== null) {
    if (parseGoVersion(requestedGo) && satisfies(requestedGo, catalog.go.range, 'go'))
      go = requestedGo;
    else
      notices.push(
        `The link names Go ${requestedGo}, which is not supported; Go ${defaults.go} is used.`,
      );
  }

  const pins: Record<string, string> = {};
  for (const pin of params.getAll('pin')) {
    const separator = pin.lastIndexOf('=');
    const path = pin.slice(0, Math.max(separator, 0));
    const version = pin.slice(separator + 1);
    const target = catalog.modules.find((m) => m.path === path);
    if (!target || target.role !== 'library') {
      notices.push(`The link pins "${path || pin}", which cannot be pinned; it is ignored.`);
    } else if (!parseModuleVersion(version) || !satisfies(version, target.range, 'module')) {
      notices.push(
        `The link pins ${path} ${version}, which is not supported; the default ${target.default} is used.`,
      );
    } else {
      pins[path] = version;
    }
  }

  return {
    state: {
      name: params.get('name') ?? defaults.name,
      module: params.get('module') ?? '',
      components,
      go,
      pins,
    },
    notices,
  };
}
