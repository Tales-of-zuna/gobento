# Design

## Context

The catalog has one component per kind: `mvc`, `nethttp`, `sqlite`, `sqlc`.
Three things are SQLite-specific today, outside the `sqlite` component:

- `mvc`'s `main.go` defaults `DATABASE_URL` to `<name>.db`, and its README
  says the data lives in that file.
- `sqlc` requires `sql-dialect: [sqlite]`; its `sqlc.yaml` hard-codes
  `engine: sqlite`, its queries use `?`, and its pre-generated `sqlcgen`
  code holds those `?` queries.
- `sqlc`'s store test opens `database.Open(ctx, <tempdir>/test.db)`, a path
  that only a file database accepts.

Templates see the selection only through `.Traits`, `.Slots`, `.Versions`,
and `.Project` (`internal/compose/data.go`); every file of a component is
always rendered, so a component cannot ship a file for one dialect only.
Generated projects are tested by `TestIntegration` on Linux, macOS, and
Windows runners; only the Linux runner can run service containers.

## Goals / Non-Goals

**Goals:**

- `--db postgres` with `--data sqlc` generates a project that builds, vets,
  and passes `go test ./...` on all three OSes, without cgo.
- No component names another: everything SQLite- or PostgreSQL-specific
  lives in the database component or is chosen by a trait value.
- The PostgreSQL code path is exercised for real in CI at least on Linux.
- A contract the parallel GORM component can use unchanged: `*sql.DB` from
  `Open`, `dbtest.Open(t)` for its tests, `DefaultURL` for the entry point.

**Non-Goals:**

- Migrations. Both databases apply an idempotent `schema.sql` on start, as
  SQLite does today.
- pgx's native interface (`conn: pgx`), connection-pool tuning, TLS setup.
- Docker Compose files or other deployment tooling in generated projects.
- A notion of a default component when a kind has several (the interactive
  prompts are a separate change); `--db` simply becomes required.

## Decisions

### D1. Driver: pgx v5 through `database/sql` (`pgx/v5/stdlib`)

`github.com/jackc/pgx/v5` is the maintained, pure-Go PostgreSQL driver; its
`stdlib` package registers the `pgx` driver for `database/sql`, so the
component provides `conn: database/sql` and returns `*sql.DB` like SQLite.
`lib/pq` is in maintenance mode and recommends pgx. pgx's native `pgxpool`
would need a new `conn` value that no data component accepts yet.

Version range `>=v5.12.0 <v6.0.0`, default `v5.12.0` (checked against
proxy.golang.org and vuln.go.dev on 2026-11-04):

- pgx itself has advisories fixed in v5.5.2, v5.5.4, v5.9.0 (two), and
  v5.9.2 (GO-2026-5004, SQL injection through dollar-quoted literals), so
  the floor is at least v5.9.2.
- v5.9.2 through v5.11.0 require `golang.org/x/text` v0.29.0, which has
  GO-2026-6629 (panic in `secure/precis`, fixed in v0.41.0). pgx uses
  `precis` for SCRAM authentication, so the advisory is reachable. v5.12.0
  is the first release that requires x/text v0.41.0, so it is the floor:
  the range floor is what CI's "floors" run and users' `--pin` can select,
  and it must not pull in a vulnerable dependency either.
- pgx v5.12.0 declares `go 1.25.0`, below gobento's go floor (1.26), so no
  `require go` constraint is needed; resolution cannot raise the go
  directive. If a later floor needs a newer Go, a constraint is added then,
  as for modernc.org/sqlite.

Alternative considered: floor v5.9.2 (no pgx advisory). Rejected because
the floor project would then build with a vulnerable x/text.

### D2. Default connection string: `DefaultURL` in the database slot

The entry point must default `DATABASE_URL` to something that depends on the
database. Options:

1. A trait, e.g. `default-url`. Trait values are fixed strings in
   `component.yaml`, but the SQLite default contains the project name, and
   the architecture would have to paste a trait into Go source.
2. A conditional on `sql-dialect` in `main.go`: the architecture would
   depend on knowing every database. Rejected.
3. **Chosen:** every database package exports
   `const DefaultURL = "<connection string>"`, and `main.go` calls
   `getenv("DATABASE_URL", database.DefaultURL)`. It is plain Go, checked by
   the compiler like the rest of the slot contract, and the database
   component owns both the value and its documentation (its notes).

Values: SQLite `"<name>.db"` (unchanged behavior);
PostgreSQL `"postgres://postgres:postgres@localhost:5432/<name>?sslmode=disable"`,
matching the `docker run` command in the generated README. The password is a
well-known local-development default, documented as such; production sets
`DATABASE_URL`. `mvc`'s README no longer names a file; the database notes
state the default.

### D3. Test databases: a `dbtest` subpackage of the database slot

Tests in other slots (sqlc's store test, a future GORM store test) need an
empty database without knowing how to make one. The database slot gains a
test-support subpackage, like `net/http/httptest`:

```go
package dbtest // <database slot>/dbtest
func URL(t testing.TB) string  // isolated connection string, cleaned up by t
func Open(t testing.TB) *sql.DB // database.Open(URL(t)), closed by t
```

- SQLite: `URL` is `<t.TempDir()>/test.db`. `Open` registers `Close` after
  `TempDir`, so the file is closed before Windows removes the directory.
- PostgreSQL: `URL` reads `TEST_DATABASE_URL` and calls `t.Skip` when it is
  unset. Otherwise it creates a schema `test_<random>` and returns the URL
  with `search_path=<schema>` (pgx passes unknown URL parameters to the
  server as run-time parameters); a cleanup drops the schema with
  `CASCADE`. `go test ./...` runs packages in parallel against one server,
  so a schema per test keeps them apart; a schema needs only the `CREATE`
  privilege on the database, not `CREATEDB`.

The subpackage is part of the database slot's contract (documented in
"Slots and contracts"), so importing it from another slot's tests does not
break "import other slots only through the contract". Code outside tests
never imports it. The PostgreSQL `database` package's own tests use
`package database_test` to import `dbtest` without an import cycle; the
SQLite package's internal tests stay as they are.

Alternatives: a `testing`-importing helper in the database package itself
(puts test code in the production package); a conditional on `sql-dialect`
in the store test (every data component would repeat per-database logic);
`CREATE DATABASE` per test (needs `CREATEDB`, slower).

The variable is named `TEST_DATABASE_URL`, distinct from `DATABASE_URL`, so
running the tests never touches the application's database by accident.

### D4. sqlc for two dialects: trait-driven engine and placeholder

sqlc's output for the two engines differs only in the placeholder of
`CreateItem` (verified with the pinned v1.31.1: `?1` for sqlite, `$1` for
postgresql; `models.go` and `db.go` are identical because the PostgreSQL
schema uses `BIGINT`, `TEXT`, and `TIMESTAMPTZ`, which map to the same Go
types as the SQLite schema).

- `requires.sql-dialect: [sqlite, postgresql]`. The trait values are sqlc's
  engine names, so `sqlc.yaml` says `engine: {{index .Traits "sql-dialect"}}`.
- `queries.sql` uses `sqlc.arg(name)` instead of `?`, so the user-facing
  query file is the same for both dialects.
- `sqlcgen/queries.sql.go.tmpl` chooses the placeholder with
  `{{if eq (index .Traits "sql-dialect") "postgresql"}}$1{{else}}?1{{end}}`.
  This is the only conditional; it depends on a trait value, not on a
  component, as the composition model allows.
- The store test uses `dbtest.Open(t)`.

Alternatives: one sqlc component per dialect (duplicates the store and its
tests; a per-combination template in disguise); running sqlc at generation
time (needs cgo for the PostgreSQL parser and network access, and makes
generation slower); a `placeholder` trait provided by databases (leaks a
sqlc detail into every database component).

CI's `sqlc-fresh` job becomes a matrix over `db: [sqlite, postgres]`: it
generates the reference project with `--db <db>`, runs the regenerate
command from the README, and fails on any diff, so both conditional branches
are checked against real sqlc output.

### D5. PostgreSQL `Open` and schema

`Open(ctx, url)` calls `sql.Open("pgx", url)` and executes the embedded
`schema.sql` once. With no arguments pgx uses the simple query protocol, so
the schema may hold several statements. Errors never include the URL, which
may contain a password (the SQLite component includes its file path; a
PostgreSQL URL is a secret). No `SetMaxOpenConns(1)`: PostgreSQL handles
concurrent connections, and the `database/sql` defaults suit an example.

Schema: `id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY`, `name TEXT NOT
NULL`, `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`, with
`CREATE TABLE IF NOT EXISTS` as for SQLite.

### D6. CI and the integration test

`TestIntegration` already covers every supported selection; it inherits the
environment, so `TEST_DATABASE_URL` reaches the generated projects' tests.
On Linux, the integration job starts `postgres:18-alpine` with `docker run`
in a step guarded by `runner.os == 'Linux'`, waits until `pg_isready` answers
over TCP, and exports `TEST_DATABASE_URL` through `$GITHUB_ENV`; on macOS and
Windows the variable is unset and the server tests skip. A step is used
instead of a `services:` container because service containers are
Linux-only and the job is a matrix over three OSes; a conditional empty
service image would work but is harder to read and to run locally. Unit
tests (`go test -short`) are unaffected.

### D7. Selecting a database becomes mandatory

With two databases, the existing rule "Several options require a choice"
makes `--db` required. Every documented command and test that relied on the
automatic choice adds `--db sqlite`. A default-component mechanism is out of
scope (D-Non-Goals).

## Risks / Trade-offs

- [pgx v5.12.0 is days old when this change lands] → It is the only release
  with a clean dependency graph; the range lets users pin any later v5, and
  Dependabot-style review of `versions.yaml` will raise it.
- [Tests skip silently on machines without PostgreSQL] → The skip message
  names `TEST_DATABASE_URL`, `go test -v` shows it, the generated README and
  CONTRIBUTING explain how to run them, and CI on Linux always runs them.
- [`search_path` in a URL is ignored by poolers in transaction mode] → Only
  the test helper uses it, against a direct test server.
- [The development password in `DefaultURL`] → It only applies when
  `DATABASE_URL` is unset, the notes say so, and it matches the local server
  command; no production default exists that would be safer.
- [Parallel tracks (GORM, vertical slice, gin/chi) copy today's patterns]
  → On merge, a store test written against a SQLite path and a `main.go`
  with `"<name>.db"` must switch to `dbtest.Open(t)` and
  `database.DefaultURL`; TestIntegration fails loudly if they do not.
- [`--db` becomes required] → Matches the existing spec; the error lists
  the values. The interactive-prompt change can later offer a choice.

## Migration Plan

No migration for existing generated projects: their files are theirs. For
gobento users, scripts that ran `gobento new <name>` without `--db` must add
`--db sqlite` to keep the previous result.
