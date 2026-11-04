# Proposal

## Why

GORM is the most widely used Go ORM, and the roadmap lists it as the next
data-access option after sqlc. Today the `store` slot can only be filled by
sqlc, so users who prefer an ORM have to replace the generated store by hand.

## What Changes

- **New component `data/gorm`:** fills the `store` slot with a GORM-backed
  implementation of `models.ItemStore`. It requires `conn: database/sql`,
  `sql-dialect: sqlite or postgresql`, and `sql-schema: schema.sql`, wraps the
  `*sql.DB` the database slot opens (it never opens its own connection), and
  uses the database's `schema.sql` as the only schema: no `AutoMigrate`.
  `gobento new shop --db sqlite --data gorm` generates a project that builds
  without cgo and ships a store test against a real SQLite file.
- **Dialect from traits:** the store selects its GORM dialector from the
  `sql-dialect` trait in its templates (a pure-Go SQLite dialector for
  `sqlite`, `gorm.io/driver/postgres` for `postgresql`), so it works with a
  PostgreSQL database component without changes once one is added.
- **Trait-conditioned modules in the rules file:** a module entry in
  `components/versions.yaml` may add `traits:`, restricting it to selections
  whose traits have one of the listed values. The GORM PostgreSQL driver is
  thus only required in `go.mod` when the selected database provides
  `sql-dialect: postgresql`, and pinning it in a SQLite project is rejected
  as "not used by the selection". The website's catalog export, schema, and
  compatibility logic understand the new field.
- **Behavior change of `gobento new` without `--data`:** the data-access kind
  now has two options, so the existing rule "several options require a
  choice" applies: `gobento new shop` without `--data` exits non-zero and
  lists `gorm, sqlc`. Examples, CI, and docs pass `--data` explicitly.
- **Integration test:** every supported selection is additionally built with
  `CGO_ENABLED=0`; the PostgreSQL path of the GORM store is compiled and
  vetted against a test-only PostgreSQL database fixture.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `component-composition`: data-access components use the database's schema
  as the single source of truth, and components that support several values
  of a trait choose dialect-specific code from the trait.
- `version-compatibility`: rules-file modules may be conditioned on traits.
- `project-generation`: generated projects build without cgo; the reference
  examples name `--data` now that the kind has several options.
- `website-docs`: the component reference shows trait conditions of modules.

## Impact

- Code: new `components/data/gorm/`; `components/versions.yaml`;
  `internal/compat` (trait conditions on modules, `Request.Traits`);
  `internal/catalog` (trait names in the rules are checked against the
  catalog); `internal/compose` and `internal/generate` (pass the selection's
  traits to version resolution); `internal/generate` integration test
  (`CGO_ENABLED=0` build, PostgreSQL fixture);
  `internal/tools/catalogexport` (exports `traits`);
  `.github/workflows/ci.yml` (the sqlc freshness job passes `--data sqlc`);
  website `lib/catalog/schema.ts`, `lib/catalog/compat.ts`,
  `lib/command/cli-commands.test.ts`, `lib/command/supported-commands.txt`,
  `components/components-reference.tsx`, the roadmap lists, the home page's
  quick-start command, the website unit and end-to-end tests, and the
  regenerated `src/generated/catalog.json`.
- Dependencies of generated projects: `gorm.io/gorm`, `github.com/libtnb/sqlite`
  (SQLite) or `gorm.io/driver/postgres` (PostgreSQL), only when GORM is chosen.
- CLI: `gobento new --help` examples pass `--data`; the data kind's selection
  error says "data-access library".
- Documentation updated in the same change: `README.md` (usage, layout,
  roadmap), `website/content/docs/index.mdx`, `getting-started/quick-start.mdx`,
  `cli/new.mdx`, `concepts/components.mdx`, `concepts/traits.mdx`,
  `concepts/version-rules.mdx` (the `traits:` field),
  `contributing/adding-a-component.mdx` (the `traits:` field and the
  dialect-switch pattern), and the GORM component's `notes.md.tmpl` in the
  generated README.
