# Tasks

## 1. Trait-conditioned modules in the rules

- [x] 1.1 Add `traits` to rules-file modules and `Traits` to `compat.Request` in `internal/compat`; a module is used when a component matches and every condition holds, nil request traits disable the filter; parsing rejects empty value lists; verify with new `internal/compat` tests (`go test ./internal/compat`)
- [x] 1.2 Check in `internal/catalog` that every trait named in a module condition is provided by some component; verify with a load test naming the module and the trait
- [x] 1.3 Pass `Selection.Traits()` to resolution in `compose.Check`, `generate.Prepare` (nil while a kind is unresolved), and the tests that resolve shipped selections; verify a pin of a trait-excluded module fails with "not used by the selection" in a `generate` test
- [x] 1.4 Export module `traits` from `internal/tools/catalogexport`, accept it in `website/lib/catalog/schema.ts` (default `{}`), filter `selectedModules` in `website/lib/catalog/compat.ts` by the selected components' traits, use it in `cli-commands.test.ts`, and show conditions in `components-reference.tsx`; verify with `go test ./internal/tools/catalogexport`
- [x] 1.5 Document the `traits:` field in `website/content/docs/concepts/version-rules.mdx` and `contributing/adding-a-component.mdx`, and the header comment of `components/versions.yaml`; verify the documented YAML matches the parser's field names

## 2. GORM component

- [x] 2.1 Add `components/data/gorm` (manifest, `store.go.tmpl` choosing the dialector from `sql-dialect`, `store_test.go.tmpl` against a real SQLite file, `notes.md.tmpl`) and register `gorm.io/gorm`, `github.com/libtnb/sqlite`, and `gorm.io/driver/postgres` with ranges, defaults, trait conditions, and constraints in `components/versions.yaml`; verify `go test ./internal/catalog ./internal/compose`
- [x] 2.2 Extend `internal/compose` tests: the GORM plan for SQLite and for a PostgreSQL fixture database is gofmt-formatted and imports only its dialect; framework files are identical for sqlc and gorm; verify `go test ./internal/compose`
- [x] 2.3 Extend the integration test: build every supported selection with `CGO_ENABLED=0`, and generate, build, vet, and test `mvc/nethttp/postgres/gorm` against a test-only PostgreSQL database fixture; verify `go test -count=1 -run TestIntegration ./internal/generate/`
- [x] 2.4 Update CLI tests, examples, and help text that relied on `--data` being optional, and the `sqlc-fresh` CI job (`--data sqlc`); verify `go test ./internal/cli`
- [x] 2.5 Regenerate `website/src/generated/catalog.json` with `make catalog`, update `website/lib/command/supported-commands.txt` to what `cli-commands.test.ts` produces, remove GORM from the website roadmap lists, and update website unit tests whose fixtures change; verify `go test -count=1 -run 'TestCommittedCatalogIsCurrent|TestWebsiteCommands' ./internal/tools/catalogexport ./internal/cli` (pnpm steps are left to the maintainer: no node_modules here)
- [x] 2.6 Update all affected docs: `README.md` (usage, layout note, roadmap), `website/content/docs/index.mdx`, `getting-started/quick-start.mdx`, `cli/new.mdx`, `concepts/components.mdx`, `concepts/traits.mdx`, `contributing/adding-a-component.mdx` (dialect switch pattern, component table), and `CONTRIBUTING.md` where it names data components; verify every documented `gobento new` command with `--dry-run`

## 3. Verification

- [x] 3.1 Generate `gobento new shop --db sqlite --data gorm`, then run `CGO_ENABLED=0 go build ./...`, `go vet ./...`, and `go test ./...` in it, and cross-compile with `CGO_ENABLED=0` for `GOOS=darwin` and `GOOS=windows`; verify all succeed and `go.mod` does not mention `gorm.io/driver/postgres`
- [x] 3.2 Run `go test ./...`, `make lint`, and `openspec validate add-gorm-data-access --strict`; verify all pass

## Workflow follow-up

- Archive with `openspec archive add-gorm-data-access -y`, then `openspec validate --specs --strict`.
- After the PostgreSQL database component merges, the maintainer deletes the test fixture, runs the integration test, and runs the generated store test with `DATABASE_URL` set (design D8, Migration Plan).
