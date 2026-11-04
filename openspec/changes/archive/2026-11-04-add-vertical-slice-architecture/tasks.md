# Tasks

## 1. Nested slot ownership (core)

- [x] 1.1 In `internal/compose`, reject a slot file that lies inside another slot's directory nested within its own, naming the component, the file, and the slot; verify with a fixture test for the scenario "File inside a nested slot" and `go test ./internal/compose`
- [x] 1.2 Document nested slot directories and the new composition error in `website/content/docs/concepts/slots-and-contracts.mdx` and `website/content/docs/contributing/adding-a-component.mdx`; verify the wording against the error message produced by the test in 1.1

## 2. The vertical-slice architecture component

- [x] 2.1 Add `components/architecture/vertical-slice/` (manifest with the D2 slots, `entry/main.go`, `models/items.go`, `root/README.md`, `root/dot-gitignore`); verify `go test ./internal/catalog` loads it
- [x] 2.2 Make the generated README explain the layout, the dependency direction, and how to add a second feature; verify against the golden README
- [x] 2.3 Check the shipped framework, database, and data templates for layer-specific wording and make it layout-neutral if needed; verify by reading the vertical-slice golden output
- [x] 2.4 Make `TestGolden` table-driven and add `testdata/golden-vertical-slice` for `vertical-slice/nethttp/sqlite/sqlc`; verify `go test ./internal/compose -run TestGolden` passes and the `mvc` golden is unchanged
- [x] 2.5 Add a shipped-catalog test that the framework files under `vertical-slice` equal those under `mvc` once package names and imports are swapped; verify it passes
- [x] 2.6 Update the supported-selection expectations (`internal/compose` and `internal/catalog/shipped_test.go`) to both architectures; verify `go test ./internal/...`

## 3. Architecture becomes a required choice

- [x] 3.1 Update tests that relied on automatic architecture selection (`internal/cli`, `internal/generate`) and add a CLI test that omitting `--arch` names `--arch`, `mvc`, and `vertical-slice`; verify `go test ./internal/cli ./internal/generate`
- [x] 3.2 Update `gobento new --help` (flag text and examples) so no example relies on a single architecture; verify `TestNewHelp` and `go run ./cmd/gobento new --help`
- [x] 3.3 Pass `--arch mvc` in the CI `sqlc-fresh` job in `.github/workflows/ci.yml`; verify by running its generate command locally
- [x] 3.4 Update every documented `gobento new` example that omits `--arch` (README.md, CONTRIBUTING.md, `website/content/docs/**`); verify with a grep for `gobento new` lines without `--arch` outside archived changes

## 4. Website catalog and configurator

- [x] 4.1 Run `make catalog` to refresh `website/src/generated/catalog.json`; verify `TestCommittedCatalogIsCurrent`
- [x] 4.2 Update `website/lib/command/supported-commands.txt` to the commands the configurator builds for both selections (website `node_modules` is not installed here, so write it by following `build.ts`); verify `TestWebsiteCommands`
- [x] 4.3 Add the vertical-slice directories to the configurator's file-tree preview (`view-model.ts`) and extend website unit tests that list architectures; verify by reading the tests (pnpm checks skipped without `node_modules`)
- [x] 4.4 Update the component and architecture docs (`README.md` component list and roadmap tick, `website/content/docs/index.mdx`, `concepts/components.mdx`, `concepts/slots-and-contracts.mdx` with the vertical-slice layout, `cli/new.mdx`, `getting-started/quick-start.mdx`); verify each documented command with `go run ./cmd/gobento new ... --dry-run`

## 5. Verification

- [x] 5.1 Run `go test ./...`, `make lint`, and `openspec validate add-vertical-slice-architecture --strict`; all pass
- [x] 5.2 Run `go test -count=1 -run TestIntegration ./internal/generate/` (network) and verify both architectures build, vet, test, and keep handlers independent of database and store

## Workflow follow-up

- Archive with `openspec archive add-vertical-slice-architecture -y`, then run `openspec validate --specs --strict`.
- Commit on `feat/vslice` as `feat: add the vertical slice architecture`.
