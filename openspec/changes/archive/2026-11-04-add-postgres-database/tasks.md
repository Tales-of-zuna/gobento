# Tasks

## 1. Database slot contract (SQLite and the architecture)

- [x] 1.1 Add `DefaultURL` to `components/database/sqlite/files/database/database.go.tmpl` and use it in `components/architecture/mvc/files/entry/main.go.tmpl`; verify with `go test ./internal/compose -run TestGolden -update` that only `main.go`, `database.go`, and the README change in the golden files
- [x] 1.2 Add the SQLite `dbtest` subpackage (`files/database/dbtest/dbtest.go.tmpl`) with `URL` and `Open`; verify it is planned at `internal/database/dbtest/dbtest.go` in the golden files
- [x] 1.3 Move the database description out of `mvc`'s README into the SQLite notes (default file, `DATABASE_URL`); verify the golden README
- [x] 1.4 Document the extended database contract (`DefaultURL`, `dbtest`) in `website/content/docs/concepts/slots-and-contracts.mdx` and `website/content/docs/contributing/adding-a-component.mdx`; verify every identifier named there exists in the templates

## 2. sqlc for both dialects

- [x] 2.1 Accept `sql-dialect: [sqlite, postgresql]` in `components/data/sqlc/component.yaml`, take the engine from the trait in `sqlc.yaml.tmpl`, use `sqlc.arg(name)` in `queries.sql.tmpl`, and choose the placeholder by trait in `sqlcgen/queries.sql.go.tmpl`; verify the SQLite output matches `go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate` in a generated project
- [x] 2.2 Switch the store test to `dbtest.Open(t)`; verify `go test ./...` in a generated SQLite project
- [x] 2.3 Turn the `sqlc-fresh` CI job into a matrix over `sqlite` and `postgres`; verify with actionlint and by running its script locally for both dialects
- [x] 2.4 Update the sqlc notes and `website/content/docs/concepts/traits.mdx` (sqlc requirements, dialect values) and the regeneration instructions in `adding-a-component.mdx`; verify the documented regenerate command matches CI

## 3. PostgreSQL component

- [x] 3.1 Add `github.com/jackc/pgx/v5` to `components/versions.yaml` (range `>=v5.12.0 <v6.0.0`, default `v5.12.0`, with the advisory reasoning); verify `go test ./internal/catalog ./internal/compat`
- [x] 3.2 Add `components/database/postgres/` (manifest with the three traits, `database.go`, `schema.sql`, `dbtest/dbtest.go`, `database_test.go`, notes); verify `catalog.Load` accepts it and `compose.Supported` lists `mvc/nethttp/postgres/sqlc`
- [x] 3.3 Add a compose test for the PostgreSQL selection (engine, placeholder, `pgx` import, identical framework and architecture Go files across databases); verify `go test ./internal/compose`
- [x] 3.4 Run a generated PostgreSQL project's tests with and without `TEST_DATABASE_URL` against a local `postgres:18-alpine`; verify they skip without it and pass with it, and that no `test_` schema remains
- [x] 3.5 Run the shipped sqlc regeneration for a generated PostgreSQL project; verify no diff
- [x] 3.6 Write the PostgreSQL notes (local server command, `DATABASE_URL`, `TEST_DATABASE_URL`, schema) and check every command in them against a running server

## 4. Selection, tests, and CI

- [x] 4.1 Update tests that relied on the database being chosen automatically or on a single supported selection (`internal/catalog`, `internal/compose`, `internal/cli`, `internal/generate`); verify `go test -short ./...`
- [x] 4.2 Add a PostgreSQL service and `TEST_DATABASE_URL` to the Linux integration job in `.github/workflows/ci.yml`; verify with `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12`
- [x] 4.3 Run `make catalog` and update `website/lib/command/supported-commands.txt` to the configurator's output for the new selection (node_modules are not installed here); verify `go test -count=1 -run 'TestCommittedCatalogIsCurrent|TestWebsiteCommands' ./internal/tools/catalogexport ./internal/cli`
- [x] 4.4 Update user documentation for the new option and the now-required `--db`: `README.md`, `website/content/docs/index.mdx`, `getting-started/quick-start.mdx`, `cli/new.mdx`, `concepts/components.mdx`, `concepts/version-rules.mdx`, and `CONTRIBUTING.md` (running PostgreSQL tests); verify each documented command with `go run ./cmd/gobento ... --dry-run`

## 5. Verification

- [x] 5.1 `go test ./...`, `make lint`, and `go mod tidy` leave no diff
- [x] 5.2 `TEST_DATABASE_URL=<local server> go test -count=1 -run TestIntegration ./internal/generate/` passes for every selection and variant
- [x] 5.3 `openspec validate add-postgres-database --strict` passes

## Workflow follow-up

- Archive with `openspec archive add-postgres-database -y` and check `openspec validate --specs --strict`.
- Commit on `feat/postgres` with Conventional Commits messages; do not push.
