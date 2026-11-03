# Tasks

## 1. Module setup

- [x] 1.1 Add `cmd/gobento/main.go` (with the `version` variable GoReleaser sets) and the package directories from design D1, on top of the `go.mod` from `establish-repo-standards`; verify `go build ./...` and `make lint` succeed
- [x] 1.2 Add cobra v1.10.2, `golang.org/x/mod` v0.42.0, and `go.yaml.in/yaml/v3` v3.0.6 (or newer patch releases, re-checked on the proxy); verify `go mod tidy` leaves no diff and `go mod verify` succeeds

## 2. Version compatibility (`internal/compat`)

- [x] 2.1 Implement range parsing (ANDed comparators) and Go-version normalization on `x/mod/semver`; verify table tests for each comparator, malformed ranges, and `1.27` → `v1.27.0`
- [x] 2.2 Implement the rules model and loader (go range, modules with components list, role, range, default, constraints) with consistency checks; verify one test per error in spec *Rules file is checked for consistency*
- [x] 2.3 Implement resolution (defaults → `--go`/`--pin` → ranges → constraints, collecting all errors; tools are not pinnable); verify tests for every scenario in *Default versions*, *User version overrides*, and *Reject incompatible combinations*

## 3. Component catalog (`internal/catalog`, `components/`)

- [x] 3.1 Define manifest types (slots with `filled_by`/`required`, provided and required traits) and a loader over `fs.FS` that validates manifests and enforces the `.tmpl`-only file rule and the `dot-` mapping; verify with invalid fixtures in `testdata/`
- [x] 3.2 Implement listing of supported selections (cross product filtered by traits); verify with a fixture catalog containing a trait mismatch (spec *Trait mismatch is not supported*)
- [x] 3.3 Create `components/embed.go` and `components/versions.yaml` with the researched versions and confirmed range floors; verify a test loads the embedded catalog and rules and that every rules entry references an existing component

## 4. Composition (`internal/compose`)

- [x] 4.1 Implement slot resolution and template data (`.Project`, `.Slots`, `.Traits`, `.Versions`), with CRLF normalization; verify with a two-architecture fixture that one framework file lands in different directories with the correct package and imports
- [x] 4.2 Implement Plan building with all composition errors collected (unknown slot, unfilled required slot, two owners, root-path conflict, unmet trait); verify one test per scenario in spec *One owner per slot*, *Traits connect compatible components*, and *Reject selections that cannot be composed*
- [x] 4.3 Run `.go` outputs through `go/format`, reporting component and file on failure, and append component `notes.md.tmpl` into the README; verify with a broken fixture template and a notes fixture
- [x] 4.4 Add a second fixture data component and verify that the framework's planned files are byte-identical across both (spec *Swapping one component leaves others untouched*)

## 5. First slice components

- [x] 5.1 Write `architecture/mvc`: slots, `main.go` per the D2 contract (env config, graceful shutdown), `models` contract, README, and `dot-gitignore`
- [x] 5.2 Write `framework/nethttp`: `NewRouter`, health and items handlers, and handler tests with a fake `ItemStore`
- [x] 5.3 Write `database/sqlite`: `Open` with `modernc.org/sqlite` (confirm the driver name in source), embedded `schema.sql` applied at startup, traits `sql-dialect: sqlite`, `conn: database/sql`
- [x] 5.4 Write `data/sqlc`: `sqlc.yaml` (`omit_sqlc_version: true`, schema from the `sql-schema` trait, output in the `sqlcgen` subpackage of the store slot), `queries.sql`, sqlc v1.31.1 output, the `New` adapter mapping `sqlcgen.Item` to `models.Item`, `notes.md.tmpl` with the regenerate command, and a real-SQLite test that closes the DB in `t.Cleanup`
- [x] 5.5 Add the golden test for mvc/nethttp/sqlite/sqlc with `-update`; verify it passes and `testdata/golden/` is committed

## 6. Generation (`internal/generate`)

- [x] 6.1 Implement preflight (`go` on PATH, `go env GOVERSION` vs the selected Go version under `GOTOOLCHAIN`, target directory state); verify tests for the *Go toolchain preflight* and *Output location* scenarios using a fake `go` binary in a temp PATH
- [x] 6.2 Implement writing to `.gobento-tmp-*`, rendering `go.mod`, the single `go get` plus `go mod tidy` under the controlled environment, the modfile re-check, and `toolchain` removal; verify a test where a fake `go` raises a version and the error names it as a missing constraint
- [x] 6.3 Implement placement (rename, move into an existing empty target, Windows retry) and cleanup on error and interrupt via `signal.NotifyContext`; verify tests for *No partial projects* (a failing fake `go get`, and a cancelled context mid-run) that assert no leftovers and an untouched pre-existing target
- [x] 6.4 Add the integration test (skipped under `-short`): for each supported selection, generate with defaults and with range floors, then run build, vet, test, `gofmt -l`, `go mod verify`, the pin assertions, and the `go list -deps` contract check; verify it passes locally

## 7. CLI (`internal/cli`)

- [x] 7.1 Implement `gobento new <name>` with `--module`, `--arch`, `--framework`, `--db`, `--data`, `--go`, a repeatable `--pin`, `--dir`, and `--dry-run`, with auto-selection of single-option kinds and next-steps output; verify CLI tests for *Fully specified by flags*, *Module path defaults to the project name*, *Automatic selection of the only option*, *Project name and module path rules*, *Report every input problem at once*, *Valid dry run*, and *Invalid dry run*
- [x] 7.2 Write help text and error messages that state the fix for each problem; verify `gobento new --help` lists every flag with an example, and that a snapshot test covers the multi-error output

## 8. Documentation and CI

- [x] 8.1 Write `docs/components.md` (layout, slots, contracts, traits, template data, the `.tmpl` and `dot-` rules, adding a component, bumping versions); verify that its example manifest loads with the catalog loader in a test
- [x] 8.2 Add the Usage section to `README.md`; verify every documented command runs as written
- [x] 8.3 Extend `.github/workflows/ci.yml` with the integration job (newest Go in range, ubuntu/macos/windows, module cache) and the sqlc freshness job; verify with `actionlint` locally (the first real run happens after a push the user approves)

## 9. End-to-end check

- [ ] 9.1 Build gobento, run `gobento new shop --module example.com/shop`, start the app, and verify `GET /health` returns 200 and that `POST /items` then `GET /items` round-trips through SQLite (spec *Example functionality is present*)

## Workflow follow-up

- Apply only after `establish-repo-standards` is applied.
- Review with the user, then archive the change with `/opsx:archive`.
- Push to `origin` only after the user approves.
