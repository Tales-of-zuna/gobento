# Tasks

## 1. The ddd architecture

- [x] 1.1 Declare `web: { dir: internal/adapters/web, filled_by: frontend }` in `components/architecture/ddd/component.yaml`; verify `go test ./internal/catalog`
- [x] 1.2 Build the handler in `newHandler` in `files/entry/main.go.tmpl`, passing the application service through `domain.ItemStore`, with the `{{if .Slots.web.Filled}}` mux; verify the no-frontend `golden-ddd` diff only adds `newHandler`
- [x] 1.3 Add `files/entry/main_test.go.tmpl`, the routing test through `app.NewItemService` on an in-memory store
- [x] 1.4 Update the README template: the web adapter row, the browser hint, and `{{.APIPrefix}}` in the `curl` paths
- [x] 1.5 Add `testdata/golden-ddd-htmx` to the goldens and regenerate with `go test ./internal/compose -run TestGolden -update`

## 2. Tests

- [x] 2.1 Expect `ddd` selections with htmx in `TestShippedCatalogSupported`, run `TestShippedFrontendLeavesFrameworkUntouched` and `TestShippedFrontendMount` for every architecture; verify `go test -short ./...`
- [x] 2.2 Treat the web package as an adapter in the integration test's `checkLayers`; verify with `TestIntegration` against PostgreSQL

## 3. Website and docs

- [x] 3.1 Run `make catalog`; add the `ddd` web directory to the configurator's file-tree preview with a unit test
- [x] 3.2 Update `concepts/slots-and-contracts.mdx`, `cli/new.mdx`, `cli/list.mdx`, `getting-started/quick-start.mdx`, `contributing/adding-a-component.mdx`, and `README.md`; verify `TestDocsListComponents` and every documented command with the built binary

## 4. Verification

- [x] 4.1 `go build ./...`, `go test -short ./...`, `make lint`, `make catalog-check`, `openspec validate add-ddd-frontend-support --strict`, and the website checks pass
- [x] 4.2 Generate `ddd/chi/sqlite/gorm/htmx`, run it, and check `GET /`, `GET /api/health`, and an htmx `POST /items`
