# Tasks

## 1. Export and listing

- [x] 1.1 Export `tooling` in `internal/catalogjson` (sorted, `{go}` kept, empty digest for unpinned images); verify with `TestLoadExportsTooling` and `make catalog`
- [x] 1.2 Print the "Project tooling" table in `gobento list versions` and add `tooling` to its JSON; add the tooling to the frozen fixture and update the snapshots with `go test ./internal/cli -run TestListSnapshots -update`; verify `go test ./internal/cli`
- [x] 1.3 Accept `tooling` in `website/lib/catalog/schema.ts`, with a schema test; verify `pnpm test` and `pnpm typecheck`

## 2. Docs and verification

- [x] 2.1 Update `website/content/docs/cli/list.mdx`; verify the sample against `gobento list versions` of the built binary
- [x] 2.2 `go test -short ./...`, `make lint`, `make catalog-check`, `openspec validate --strict`, and the website checks pass
