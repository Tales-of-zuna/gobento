# Design

## Context

See proposal.md for the motivation. The relevant current state:

- The `mvc` architecture's entry point calls `database.Open(ctx, dsn)`, which
  returns a `*sql.DB` with `schema.sql` already applied, and passes it to
  `store.New(db)`, which returns a `models.ItemStore`. The handlers package
  only sees `models`.
- `database/sqlite` uses `modernc.org/sqlite` (pure Go), which registers the
  `database/sql` driver name `sqlite`. It provides the traits
  `sql-dialect: sqlite`, `conn: database/sql`, `sql-schema: schema.sql`. A
  parallel change adds `database/postgres` with `sql-dialect: postgresql` and
  the same `conn` and `sql-schema` values.
- Templates see `.Traits` (the selection's trait values) and `.Versions`.
- `components/versions.yaml` binds modules to components only. Generation
  runs `go get` for every chosen library, then `go mod tidy`, then `recheck`
  fails if a chosen library is not a direct requirement afterwards
  ("no generated code imports it"). So `go mod tidy` dropping an unused
  driver does not make listing both dialect drivers acceptable: the
  generation of every SQLite project would fail.

## Goals / Non-Goals

**Goals:**

- A GORM store for the `store` slot that works for SQLite now and for
  PostgreSQL once a database providing `sql-dialect: postgresql` exists,
  without editing the component then.
- Generated GORM projects build with `CGO_ENABLED=0` on Linux, macOS, and
  Windows, and depend only on the modules their dialect needs.

**Non-Goals:**

- Migrations or `AutoMigrate`; the schema stays owned by the database.
- GORM features beyond what `ItemStore` needs (associations, hooks, soft
  delete).
- Changing the architecture's contract (`store.New(db)` returns an
  `ItemStore`, with no error).

## Decisions

### D1. SQLite dialector: `github.com/libtnb/sqlite`

The dialector must accept an existing `*sql.DB` opened with modernc's driver
and must not need cgo. Candidates, each tried in a scratch module with
`modernc.org/sqlite` v1.61.1 and `gorm.io/gorm` v1.31.2:

| Candidate | Result |
|---|---|
| `gorm.io/driver/sqlite` with `Config{Conn: db}` | Works, and builds with `CGO_ENABLED=0` only because `github.com/mattn/go-sqlite3` (imported unconditionally) has a non-cgo stub. Wherever a C compiler is present (the default on most Linux and macOS machines), `go build` compiles mattn's C SQLite and links a second, unused SQLite engine: slow builds and a cgo dependency. Rejected. |
| `github.com/glebarez/sqlite` | Imports `github.com/glebarez/go-sqlite`, which registers the driver name `sqlite` too: the program panics with `sql: Register called twice for driver sqlite` next to `modernc.org/sqlite`. Last release 2024. Rejected. |
| `github.com/ncruces/go-sqlite3/gormlite` | Pure Go, but brings its own SQLite (WebAssembly runtime) and driver, so the project would carry two SQLite engines. Rejected. |
| `github.com/libtnb/sqlite` | Pure Go, built directly on `modernc.org/sqlite` (the same driver the database component registers, so no second registration or engine), accepts `Config{Conn}`, runs GORM's test suite on Linux, macOS, and Windows, maintained (v1.2.2, 2026-08). `go list -deps` with cgo enabled shows no cgo package from it. **Chosen.** |

Risk: it is a younger, smaller project than the official driver (see Risks).

### D2. PostgreSQL dialector: `gorm.io/driver/postgres` with `Config{Conn}`

`postgres.New(postgres.Config{Conn: db})` wraps an existing `*sql.DB`
regardless of which `database/sql` driver opened it. A dry-run of the store's
queries through it renders `INSERT INTO "items" ("name") VALUES ($1)
RETURNING "created_at","id"`, matching what the SQLite path does.

### D3. Dialect chosen in templates from `sql-dialect`

`store.go.tmpl` switches on `{{index .Traits "sql-dialect"}}` for the import
and the dialector constructor; everything else is shared. The component never
names a database component (spec: dialect-specific code is chosen from
traits). An unsupported dialect cannot reach the template, because the
component's `requires` limits `sql-dialect` to `[sqlite, postgresql]`.

### D4. Dialect-specific modules: trait conditions in the rules file

A module entry may add `traits: {<trait>: [<values>]}`. The module is used
when one of its components is selected *and* every listed trait has one of
the listed values in the selection:

```yaml
github.com/libtnb/sqlite:
  components: [data/gorm]
  traits: { sql-dialect: [sqlite] }
gorm.io/driver/postgres:
  components: [data/gorm]
  traits: { sql-dialect: [postgresql] }
```

`compat.Request` gains `Traits map[string]string`. Every caller that resolves
a complete selection passes `Selection.Traits()`. A nil map means "traits not
known yet" and disables the filter; `generate.Prepare` uses it when some kind
is unresolved, matching its existing rule that every component of an
unresolved kind counts as possibly selected (pins are then still checked
against all modules that could apply). Catalog loading checks that every
trait named in a condition is provided by some component, catching typos;
values are not checked, so `postgresql` can be listed before a PostgreSQL
database exists (that is the point).

Alternatives considered:

- *List both drivers under `data/gorm` and rely on `go mod tidy`:* tidy does
  drop the unused one, but `recheck` then fails every SQLite generation
  (`not a direct requirement after go mod tidy`). Relaxing `recheck` would
  weaken the "chosen versions are exactly what the project gets" guarantee
  and silently accept pins of unused modules. Rejected.
- *Bind the PostgreSQL driver to `database/postgres`:* that component is not
  in this branch (unknown component references fail catalog loading), and
  `sqlc` with PostgreSQL would then fail the same `recheck`. Rejected.
- *Combination ids such as `data/gorm+database/sqlite`:* works, but ties a
  module to specific components, which is exactly what traits exist to
  avoid; each new SQLite-compatible database would need a rules edit.
  Rejected.

The website mirrors the rule: `catalog.json` exports `traits` per module, the
schema accepts it (default `{}`), and `selectedModules` filters by the traits
the selected components provide. In the website a missing trait means the
module does not apply (the configurator always evaluates the components the
visitor has chosen so far).

### D5. Schema: the database's `schema.sql`, no `AutoMigrate`

The store maps a private `item` struct to the existing `items` table and
never calls `AutoMigrate`. `AutoMigrate` would make the Go struct a second
schema definition that can silently diverge from `schema.sql` (types,
defaults, constraints differ by dialect) and would alter tables at startup,
which the database notes promise not to do. `created_at` is tagged read-only
with the database default (`gorm:"->;default:CURRENT_TIMESTAMP"`), so GORM
omits it from `INSERT` and reads it back with `RETURNING`, exactly like sqlc;
`id` comes from the database the same way. The PostgreSQL schema's own
default (for example `now()`) is used there; the tag only tells GORM the
column has a default.

### D6. Wrapping the connection and reporting setup errors

`New(db *sql.DB)` opens GORM with the dialector's `Conn` set to `db`,
`DisableAutomaticPing` (the database slot already verified the connection by
applying the schema), and the GORM logger discarded (errors are returned to
callers, which log them with `slog`; GORM's default logger would write to
stdout in its own format). The SQLite dialector queries `sqlite_version()`
during setup, which can fail. Because the architecture's contract has no
error return, `New` then returns a store whose methods return that setup
error, wrapped as `set up gorm: ...`, instead of panicking: the failure is
reported on the first request and in the server log. Changing the contract
would touch the architecture and every data component and is out of scope.

### D7. Versions and floors

Researched on proxy.golang.org and vuln.go.dev on 2026-11-04:

| Module | Range | Default | Notes |
|---|---|---|---|
| `gorm.io/gorm` | `>=v1.31.1 <v2.0.0` | `v1.31.2` | No advisories. |
| `github.com/libtnb/sqlite` | `>=v1.1.0 <v2.0.0` | `v1.2.2` | No advisories. v1.2.0+ requires gorm v1.31.2 (constraint), all need Go ≤ 1.25 and modernc ≤ v1.56.0. |
| `gorm.io/driver/postgres` | `>=v1.6.1 <v2.0.0` | `v1.6.3` | v1.6.1 is the first release requiring `pgx/v5` v5.10.0, past the pgx advisories fixed up to v5.9.2 (GO-2026-5004). v1.6.2+ requires gorm v1.31.2 (constraint). |

Constraints keep floor pins from being raised by MVS (`recheck` would report
them). The integration test exercises both floors and defaults for the SQLite
path.

### D8. Verifying the PostgreSQL path without a PostgreSQL component

The integration test adds a test-only `database/postgres` fixture (pgx's
`database/sql` driver, a PostgreSQL `schema.sql`) overlaid on the shipped
catalog, generates `mvc/nethttp/postgres/gorm` with defaults and floors, and
runs `go build`, `go vet`, `CGO_ENABLED=0 go build`, `gofmt -l`, the `go.mod`
pin check, and the handlers-contract check. Its `go test` runs too: the
store test skips without a server (it reads `DATABASE_URL`). The fixture is
deleted once the real PostgreSQL component merges and covers the same path.

While writing this change, the fixture project was also run against a
throwaway `postgres:18-alpine` container: the generated store test passed
(twice on the same database, which exercises the relative assertions of D9),
and `POST /items` / `GET /items` worked through the server.

### D9. Store test in the generated project

Like sqlc's, it opens the real database through the database slot. For
`sqlite` it uses a file in `t.TempDir()`; for other dialects it uses
`DATABASE_URL` and skips when it is unset. Assertions are relative to the
rows present before the test, so they also hold on a shared server.

## Risks / Trade-offs

- [libtnb/sqlite is maintained by a small team] → It is a thin dialector over
  modernc (the engine we already depend on), its range is open within v1,
  and the integration test runs on all three OSes. If it stalls, switching to
  another dialector only touches `store.go.tmpl`'s SQLite branch and the
  rules file.
- [Setup errors surface on first use, not at startup] → Documented in the
  code and notes; the database slot already proved the connection works, so
  only a broken SQLite build could fail here.
- [The `--data` flag becomes mandatory] → It is the documented rule for kinds
  with several options; the interactive-prompt work on the roadmap removes
  the need to type it. Docs, examples, and CI are updated. The selection error
  for the data kind now reads "choose a data-access library" instead of the
  ungrammatical "choose a data".
- [Website tests cannot be run in this change's environment] → The website
  edits are small and mirror the Go logic; the maintainer runs `pnpm test -u`
  and `make web-check` before merging.

## Migration Plan

Additive for generated projects. After the PostgreSQL database component
merges: delete the test fixture (D8), make sure that component's
`database.Open` accepts a DSN string and returns `*sql.DB`, and run the
integration test, which then covers `mvc/nethttp/postgres/gorm` from the real
catalog; run the generated store test against a server with `DATABASE_URL`.
