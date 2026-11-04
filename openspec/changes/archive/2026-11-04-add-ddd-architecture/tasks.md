# Tasks

## 1. The ddd architecture component

- [x] 1.1 Add `components/architecture/ddd/` (manifest with the D2 slots, `entry/main.go`, `models/domain.go` and its test, `app/items.go` and its test, `root/README.md`, `root/dot-gitignore`); verify `go test ./internal/catalog` loads it
- [x] 1.2 Make the generated README explain the layers, the dependency direction, and how to add a second aggregate; verify against the golden README
- [x] 1.3 Add `testdata/golden-ddd` for `ddd/nethttp/sqlite/sqlc` to the `goldens` table, make `TestGoldenComponents` prefer the reference selection, and regenerate; verify `go test ./internal/compose -run TestGolden` passes and no existing golden file changes except the new `architecture/ddd` directory
- [x] 1.4 Extend `TestShippedFrameworksAreLayoutIndependent` to every architecture other than `mvc`; verify it passes for chi, gin, and nethttp under `ddd`
- [x] 1.5 Update the supported-selection and candidate lists (`internal/compose/golden_test.go`, `internal/catalog/shipped_test.go`) to the three architectures; verify `go test ./internal/...`
- [x] 1.6 Add the domain-independence check to the integration test; verify with `go test -count=1 -run 'TestIntegration/ddd' ./internal/generate/`

## 2. CLI

- [x] 2.1 Update the `--arch` flag text and an example in `gobento new --help`, and the CLI tests that list architectures (missing `--arch` error, prompts); verify `go test ./internal/cli` and `go run ./cmd/gobento new --help`
- [x] 2.2 Update the docs that describe the CLI and the architectures (`README.md` usage, tree example and roadmap tick, `website/content/docs/cli/new.mdx`, `getting-started/quick-start.mdx`, `index.mdx`, `concepts/components.mdx`, `concepts/slots-and-contracts.mdx`); verify every documented command with the built binary (`--dry-run`) and the documented error text against `gobento new shop --no-input`

## 3. Website

- [x] 3.1 Run `make catalog` to refresh `website/src/generated/catalog.json` and `website/lib/command/supported-commands.txt`; verify `go test ./internal/catalog/... ./internal/cli`
- [x] 3.2 Add the `ddd` directories to the configurator's file-tree preview (`view-model.ts`) with a unit test, list `ddd` in `components-reference.test.tsx`, and update the e2e expectations (home belt plates, architecture error message); verify `pnpm test`, `pnpm lint`, `pnpm typecheck`, `npx prettier --check .`, and `CI= E2E_PORT=4181 pnpm e2e --workers=4`

## 4. Verification

- [x] 4.1 Run `go test ./...` (with `TEST_DATABASE_URL` set to a PostgreSQL 18 container), `make lint`, and `openspec validate add-ddd-architecture --strict`; all pass

## Workflow follow-up

- Archive with `openspec archive add-ddd-architecture -y`, then run `openspec validate --specs --strict`.
- Commit on `feat/ddd` as `feat: add the domain-driven design architecture`.
