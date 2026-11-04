/**
 * Schema of `src/generated/catalog.json`, the component catalog and version
 * rules the CLI uses, exported for the website (design D3) by
 * `internal/tools/catalogexport` (`make catalog`). The Go types in
 * `internal/catalogjson` mirror this schema; `gobento list --json` prints the
 * same document. This schema is the contract between them, and both change
 * together.
 */
import { z } from 'zod';
import { KIND_IDS } from './kinds';
import { parseRange, parseVersion, satisfies, type VersionKind } from './semver';

export { KIND_IDS, type KindId } from './kinds';

const text = z.string().min(1);

const kindSchema = z.object({
  id: z.enum(KIND_IDS),
  title: text,
  description: text,
  /**
   * An optional kind may be left without a component (its flag defaults to
   * none); every other kind needs exactly one.
   */
  optional: z.boolean().default(false),
});

const componentSchema = z.object({
  /** `kind/name`, as referenced by modules (for example `database/sqlite`). */
  id: text,
  kind: z.enum(KIND_IDS),
  /** Value of the kind's flag (for example `sqlite` for `--db`). */
  name: z.string().regex(/^[a-z][a-z0-9-]*$/),
  title: text,
  /** One-sentence summary; optional in component.yaml, so it may be empty. */
  description: z.string(),
  /** Traits this component provides, as trait → value. */
  provides: z.record(text, text),
  /** Traits this component needs from the others, as trait → allowed values. */
  requires: z.record(text, z.array(text).min(1)),
});

const constraintSchema = z.object({
  /** Range of the owning module's version for which the constraint applies. */
  when: text,
  requires: z.object({
    /** `go`, or the path of another module. */
    target: text,
    range: text,
  }),
});

const moduleSchema = z.object({
  path: text,
  /** Ids of the components that use this module. */
  components: z.array(text).min(1),
  /**
   * Trait conditions as trait → allowed values: the module is used only when
   * the selection provides one of the values for every listed trait (for
   * example a dialect-specific driver). Empty means no condition.
   */
  traits: z.record(text, z.array(text).min(1)).default({}),
  /** Libraries are pinned in go.mod and overridable; tools are fixed per release. */
  role: z.enum(['library', 'tool']),
  range: text,
  default: text,
  constraints: z.array(constraintSchema),
});

/**
 * A file vendored into generated projects (for example htmx.org's script),
 * with a version fixed per gobento release: not a Go module and not pinnable.
 */
const assetSchema = z.object({
  name: text,
  /** Ids of the components that vendor this asset. */
  components: z.array(text).min(1),
  version: text,
  /** Where the release's copy was taken from. */
  source: text,
});

const ruleSchema = z.object({
  /** Regular expression the value must match. */
  pattern: text,
  /** Explanation shown when the value does not match. */
  message: text,
});

const projectNameRuleSchema = ruleSchema.extend({
  /** Names that match the pattern but are reserved (Windows device names). */
  reserved: z.array(text),
  /** Explanation shown for a reserved name. */
  reservedMessage: text,
});

/**
 * One component name per kind, as `kind → name`. An optional kind without a
 * component is left out; every other kind is present (checked against the
 * catalog's kinds in {@link checkConsistency}).
 */
const supportedSelectionSchema = z.partialRecord(z.enum(KIND_IDS), text);

const rawCatalogSchema = z.object({
  kinds: z.array(kindSchema),
  components: z.array(componentSchema),
  go: z.object({ range: text, default: text }),
  modules: z.array(moduleSchema),
  assets: z.array(assetSchema).default([]),
  /**
   * The selections the CLI supports (compose.Supported): traits met, default
   * versions resolved, and composition without errors. CI builds every one
   * without a frontend, and those with a frontend pairwise.
   */
  supported: z.array(supportedSelectionSchema).min(1),
  rules: z.object({ projectName: projectNameRuleSchema, modulePath: ruleSchema }),
});

type RawCatalog = z.infer<typeof rawCatalogSchema>;

/** Reports every inconsistency the CLI's rules loader would also reject. */
function checkConsistency(catalog: RawCatalog, ctx: z.RefinementCtx): void {
  const issue = (path: (string | number)[], message: string) =>
    ctx.addIssue({ code: 'custom', path, message });

  const checkRange = (path: (string | number)[], range: string, kind: VersionKind) => {
    if (!parseRange(range, kind)) issue(path, `malformed range "${range}"`);
  };
  const checkDefault = (
    path: (string | number)[],
    value: string,
    range: string,
    kind: VersionKind,
    owner: string,
  ) => {
    if (!parseVersion(value, kind)) issue(path, `${owner}: malformed version "${value}"`);
    else if (!satisfies(value, range, kind)) {
      issue(path, `${owner}: default ${value} is outside its range ${range}`);
    }
  };

  const kindIds = catalog.kinds.map((kind) => kind.id);
  for (const id of KIND_IDS) {
    if (!kindIds.includes(id)) issue(['kinds'], `missing kind "${id}"`);
  }
  if (new Set(kindIds).size !== kindIds.length) issue(['kinds'], 'duplicate kind');

  const componentIds = new Set<string>();
  catalog.components.forEach((component, i) => {
    if (component.id !== `${component.kind}/${component.name}`) {
      issue(['components', i, 'id'], `id must be "${component.kind}/${component.name}"`);
    }
    if (componentIds.has(component.id)) {
      issue(['components', i, 'id'], `duplicate component "${component.id}"`);
    }
    componentIds.add(component.id);
  });

  checkRange(['go', 'range'], catalog.go.range, 'go');
  checkDefault(['go', 'default'], catalog.go.default, catalog.go.range, 'go', 'go');

  const modulePaths = new Set(catalog.modules.map((module) => module.path));
  if (modulePaths.size !== catalog.modules.length) issue(['modules'], 'duplicate module path');

  catalog.modules.forEach((module, i) => {
    const at = (...rest: (string | number)[]) => ['modules', i, ...rest];
    module.components.forEach((id, j) => {
      if (!componentIds.has(id))
        issue(at('components', j), `${module.path}: unknown component "${id}"`);
    });
    checkRange(at('range'), module.range, 'module');
    checkDefault(at('default'), module.default, module.range, 'module', module.path);
    if (module.role === 'tool' && !/^=\S+$/.test(module.range.trim())) {
      issue(at('range'), `${module.path}: a tool's range must be one exact version (=vX.Y.Z)`);
    }
    module.constraints.forEach((constraint, j) => {
      const { target, range } = constraint.requires;
      checkRange(at('constraints', j, 'when'), constraint.when, 'module');
      if (target !== 'go' && !modulePaths.has(target)) {
        issue(
          at('constraints', j, 'requires', 'target'),
          `${module.path}: unknown target "${target}"`,
        );
      }
      checkRange(
        at('constraints', j, 'requires', 'range'),
        range,
        target === 'go' ? 'go' : 'module',
      );
    });
  });

  const assetNames = new Set(catalog.assets.map((asset) => asset.name));
  if (assetNames.size !== catalog.assets.length) issue(['assets'], 'duplicate asset name');
  catalog.assets.forEach((asset, i) => {
    asset.components.forEach((id, j) => {
      if (!componentIds.has(id))
        issue(['assets', i, 'components', j], `${asset.name}: unknown component "${id}"`);
    });
  });

  const optional = new Set(catalog.kinds.filter((k) => k.optional).map((k) => k.id));
  catalog.supported.forEach((selection, i) => {
    for (const kind of KIND_IDS) {
      const name = selection[kind];
      if (name === undefined) {
        if (!optional.has(kind)) issue(['supported', i], `supported selection lacks a ${kind}`);
      } else if (!componentIds.has(`${kind}/${name}`)) {
        issue(
          ['supported', i, kind],
          `supported selection names unknown component "${kind}/${name}"`,
        );
      }
    }
  });

  for (const key of ['projectName', 'modulePath'] as const) {
    try {
      new RegExp(catalog.rules[key].pattern);
    } catch {
      issue(['rules', key, 'pattern'], `${key}: invalid regular expression`);
    }
  }
}

/** Validates the structure of a catalog and its internal consistency. */
export const catalogSchema = rawCatalogSchema.superRefine(checkConsistency);

/** A validated catalog. */
export type Catalog = z.infer<typeof catalogSchema>;
/** A component of the catalog. */
export type Component = Catalog['components'][number];
/** A Go module used by one or more components. */
export type Module = Catalog['modules'][number];
/** A file vendored into generated projects, with a version fixed per release. */
export type Asset = Catalog['assets'][number];
/** A component kind as described by the catalog. */
export type Kind = Catalog['kinds'][number];
/** A selection the CLI supports; an optional kind without a component is absent. */
export type SupportedSelection = Catalog['supported'][number];

/** Parses and validates catalog data, throwing a ZodError that lists every problem. */
export function parseCatalog(data: unknown): Catalog {
  return catalogSchema.parse(data);
}
