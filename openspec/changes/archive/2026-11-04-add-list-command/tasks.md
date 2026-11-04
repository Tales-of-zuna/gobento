# Tasks

## 1. Shared export package

- [x] 1.1 Create `internal/catalogjson` with the document types, kind titles, `FromCatalog`, `Load`, and `Encode` moved from `internal/tools/catalogexport`, and move their unit tests with them; verify `go test ./internal/catalogjson`
- [x] 1.2 Reduce `internal/tools/catalogexport` to flag parsing and the file write around `catalogjson`; verify `make catalog-check` reports no difference
- [x] 1.3 Add `internal/catalogjson/` to the website CI path filters in `.github/workflows/pages.yml` and `.github/workflows/website.yml`
- [x] 1.4 Update contributor docs that name the exporter's code (`CONTRIBUTING.md`, `docs/maintainers.md`, comments in `website/lib/catalog/schema.ts` and `website/lib/test/catalog.ts`, `website/.prettierignore`); verify every named path exists

## 2. The list command

- [x] 2.1 Add `gobento list` with `components` (`--kind`), `selections` (`--arch`, `--framework`, `--db`, `--data`), `versions`, and the persistent `--json` flag, rendering text and JSON from `catalogjson.Document`
- [x] 2.2 Point `gobento --help` and `gobento new --help` at `gobento list`
- [x] 2.3 Snapshot tests of the text output from a frozen document in `internal/cli/testdata/list/` (with `-update`), tests against the embedded catalog comparing with computed values, error tests for unknown kinds and components, and a JSON round-trip test
- [x] 2.4 Document the command: new `website/content/docs/cli/list.mdx` wired into `website/content/docs/cli/meta.json` and the docs e2e page list, pointers in `website/content/docs/cli/new.mdx` and `website/content/docs/components.mdx`, and `README.md` usage; verify every documented command and output against the built binary

## 3. Verification

- [x] 3.1 `go test -short ./...`, `go test ./...` once, and `make lint`
- [x] 3.2 Website: `pnpm test`, `pnpm lint`, `pnpm typecheck`, `npx prettier --check .`, and the e2e tests
- [x] 3.3 `openspec validate add-list-command --strict`
