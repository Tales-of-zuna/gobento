# Tasks

## 1. Shared API test

- [ ] 1.1 Change the Content-Type check in `components/framework/nethttp/files/handlers/router_test.go.tmpl` to compare the parsed media type; verify `go test ./internal/compose -run TestGolden -update` changes only that file and the integration test still passes for nethttp
- [ ] 1.2 Add the catalog test that every shipped framework component has a `handlers/router_test.go` template and that all copies are byte-identical, naming the differing components; verify it fails on a deliberately edited copy

## 2. Components

- [ ] 2.1 Add `components/framework/chi` (manifest, `router.go`, `items.go`, shared test, `notes.md.tmpl` with the API table and chi-specific notes)
- [ ] 2.2 Add `components/framework/gin` (manifest, `router.go` with `HandleMethodNotAllowed`, `items.go` per design D1, shared test, `notes.md.tmpl` including `GIN_MODE`)
- [ ] 2.3 Add gin and chi to `components/versions.yaml` with the ranges and defaults of design D3 and an updated "last checked" date; verify the rules load and `compose.Supported` lists three selections

## 3. Tests

- [ ] 3.1 Add `TestGoldenComponents` and `testdata/golden-components/` per design D4; verify it passes and `-update` reproduces the committed files
- [ ] 3.2 Update tests that assumed one framework (`internal/catalog/shipped_test.go`, `internal/cli`, `internal/generate/prepare_test.go`, fixtures that copy nethttp under the name `chi`); verify `go test ./...`
- [ ] 3.3 Run the integration test (`go test -count=1 -run TestIntegration ./internal/generate/`) for all three selections with defaults and floors; verify build, vet, test, gofmt, `go mod verify`, and the handlers-independence check pass

## 4. CI, website data, and documentation

- [ ] 4.1 Pass `--framework nethttp` in the `sqlc-fresh` job of `.github/workflows/ci.yml`
- [ ] 4.2 Regenerate `website/src/generated/catalog.json` (`make catalog`) and `website/lib/command/supported-commands.txt`; verify `make catalog-check` and the website unit tests pass
- [ ] 4.3 Update docs: `README.md` (usage examples with `--framework`, roadmap tick for gin and chi), `website/content/docs/index.mdx`, `getting-started/quick-start.mdx`, `cli/new.mdx` (flag table, examples, error output), `concepts/components.mdx`, `contributing/adding-a-component.mdx` (shared API test rule); verify every documented command and error message against a real run of the CLI
