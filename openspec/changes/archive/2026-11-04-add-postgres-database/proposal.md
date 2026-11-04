# Proposal

## Why

gobento generates only SQLite projects, while most Go web services run on
PostgreSQL. PostgreSQL is the next item on the roadmap, and it is the first
second option of any kind: it proves that the composition model lets one data
access component (sqlc) serve two databases without per-combination
templates.

## What Changes

- Add the database component `database/postgres`: PostgreSQL through
  `github.com/jackc/pgx/v5` and its `database/sql` adapter (pure Go, no cgo).
  It provides the traits `sql-dialect: postgresql`, `conn: database/sql`, and
  `sql-schema: schema.sql`, and fills the `database` slot with the same
  `Open(ctx, url) (*sql.DB, error)` contract as `database/sqlite`.
- Extend the database slot's contract, for every database component:
  - an exported `DefaultURL` constant, the connection string the entry point
    uses when `DATABASE_URL` is unset (`<name>.db` for SQLite, a local
    PostgreSQL URL for PostgreSQL);
  - a test-support subpackage `dbtest` that gives a test an isolated, empty
    database with the schema applied, or skips the test when the database
    needs a server and `TEST_DATABASE_URL` is unset.
- `mvc` uses `database.DefaultURL` instead of a hard-coded `<name>.db`, and its
  README describes the database through the database component's notes.
- `data/sqlc` accepts `sql-dialect: [sqlite, postgresql]`: `sqlc.yaml` takes
  its engine from the trait, `queries.sql` uses dialect-neutral `sqlc.arg()`
  parameters, the pre-generated code differs only in the placeholder, chosen
  by the trait, and the store test runs through `dbtest`.
- Pin pgx in `components/versions.yaml` with a floor that excludes every
  release with a known advisory, including advisories in the dependencies it
  requires.
- CI: the integration job on Linux starts a PostgreSQL server and sets
  `TEST_DATABASE_URL`, so the PostgreSQL tests run for real there; macOS and
  Windows skip them. The `sqlc-fresh` job regenerates the shipped sqlc code
  for both dialects.
- **BREAKING** (CLI behavior): with two databases, `gobento new` no longer
  picks the database automatically; `--db` is required, as the existing
  "Several options require a choice" rule says.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `component-composition`: the database contract gains a default connection
  string and test support, so data-access components and their tests work
  with any database.
- `project-generation`: a PostgreSQL project builds and passes its tests
  everywhere, running the database tests where a test server is configured;
  the database is now a choice the user must make.

## Impact

- New: `components/database/postgres/` (manifest, notes, `database.go`,
  `schema.sql`, tests, `dbtest/`).
- Changed: `components/database/sqlite/` (`DefaultURL`, `dbtest/`, notes),
  `components/data/sqlc/` (manifest, `sqlc.yaml`, queries, generated code,
  store test), `components/architecture/mvc/` (`main.go`, README),
  `components/versions.yaml` (pgx).
- Tests: shipped-catalog and supported-selection expectations, CLI and prepare
  tests that relied on the database being chosen automatically, the golden
  files, a new PostgreSQL composition test; the integration test covers the
  new selection automatically.
- CI: `.github/workflows/ci.yml` (PostgreSQL service for the Linux integration
  job, `sqlc-fresh` for both dialects).
- Website data: `website/src/generated/catalog.json` (`make catalog`),
  `website/lib/command/supported-commands.txt`, and the home page's
  quick-start command (`website/components/home/links.ts`), which now needs
  `--db`. The website's TypeScript tests, which run against the real
  catalog, need `pnpm test` (and snapshot updates) after this change.
- Documentation updated in this change: `README.md` (examples, layout,
  roadmap); `website/content/docs/index.mdx` (component table),
  `getting-started/quick-start.mdx`, `cli/new.mdx`,
  `concepts/components.mdx`, `concepts/traits.mdx`,
  `concepts/slots-and-contracts.mdx`, `concepts/version-rules.mdx`,
  `contributing/adding-a-component.mdx`; `CONTRIBUTING.md` (running the
  PostgreSQL tests locally); the generated README through the notes of
  `database/sqlite`, `database/postgres`, and `data/sqlc`.
