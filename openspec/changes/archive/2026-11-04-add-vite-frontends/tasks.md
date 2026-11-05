# Tasks

## 1. Rules data for Node.js and npm

- [x] 1.1 Parse and validate `node` (range, release line inside it) and `npm` (components, role `scaffolder`/`dependency`/`dev-dependency`, exact version, `create-` prefix for scaffolders, Node.js rule required) in `internal/compat`; resolve the selected components' npm packages, reject their pins as not pinnable, replace `{node}` in image tags; verify with `go test ./internal/compat`
- [x] 1.2 Add `node`, `npm.create-vite` 9.3.3, the `node` image, and `actions/setup-node` v7.1.0 to `components/versions.yaml`, with the header comment explaining them; verify with `go test ./internal/catalog`
- [x] 1.3 Document `node` and `npm` in `website/content/docs/concepts/version-rules.mdx`; verify the documented keys against the parser

## 2. Catalog and composition

- [x] 2.1 Add the manifest `scaffold` block (with, template, dir, remove) to `internal/catalog`, checked against the rules' scaffolders; verify with `go test ./internal/catalog`
- [x] 2.2 Give component templates the `contrib` function, pass `.Node` and `.NPM` to templates, and declare the points `make-build`, `docker-stages`, `dependabot`, `gitignore`, `vite-imports`, `vite-plugins`, and `vite-styles`; include the first three in the base templates and `gitignore` in the architectures' `.gitignore`, and move htmx's `*_templ.txt` to `contrib/gitignore.tmpl`; add `frontend-build: go` to htmx; verify with `go test ./internal/compose -run TestGolden -update` and an unchanged golden diff for API-only and htmx projects
- [x] 2.3 Document contributions in component templates, the new points, `.Node`/`.NPM`, scaffolded components, and the `frontend-build` trait in `contributing/adding-a-component.mdx`, `concepts/traits.mdx`, and `concepts/project-tooling.mdx`; verify against `base.yaml` and `compose.Data`

## 3. Scaffold step in generation

- [x] 3.1 Add scaffold steps to `generate.Prepared` and `Request`, the Node.js preflight (`node`, `npm`, version range) with clear errors, the scaffold step (staging, remove list, merge with overlays winning, lock file), and process-tree cancellation; verify with fake node/npm unit tests for success, missing node, old node, npm failure, and interruption (`go test ./internal/generate`)
- [x] 3.2 Make `--dry-run` describe the scaffold step and next steps build the frontend in `internal/cli`; verify with `go test ./internal/cli`
- [x] 3.3 Document the scaffold step, the Node.js requirement and its errors, and the dry-run note in `cli/new.mdx`, `getting-started/installation.mdx`, and `README.md`; verify against a dry run and a missing-node run

## 4. React, Vue, and Svelte frontends

- [x] 4.1 Add `components/frontend/react`, `vue`, and `svelte`: manifests, the shared web package (`handler.go`, `handler_test.go`, `dist/.gitignore`), overlays, notes, and contributions (`make`, `make-build`, `ci-steps`, `docker-stages`, `docker-build`, `dockerignore`, `dependabot`, `readme-docker`); verify with `go test ./internal/catalog ./internal/compose`
- [x] 4.2 Add `TestViteFrontendsAlike` and a full-project golden for `mvc/nethttp/sqlite/sqlc/react`; verify with `go test ./internal/compose ./internal/catalog`
- [x] 4.3 Generate a project with each frontend, run `make build`, start the binary, and check `/`, a client route, `/api/items` GET and POST from the page in a headless browser, and `docker build` of one; record the results
- [x] 4.4 Document the frontends in `README.md`, `website/content/docs/index.mdx`, `getting-started/quick-start.mdx` ("With React, Vue, or Svelte"), `cli/new.mdx` examples, `concepts/components.mdx`, and `concepts/slots-and-contracts.mdx`; verify the documented commands against a generated project

## 5. Listing, CI, and coverage

- [x] 5.1 Show the Node.js rule and npm packages in `gobento list` and its JSON, and export them from `internal/catalogjson`; verify with `go test ./internal/cli -run TestListSnapshots -update` and `go test ./internal/catalogjson`
- [x] 5.2 Build Vite frontends in `TestIntegration` (npm ci, checks, build, then Go checks), run the templ check only for `frontend-build: go`, and set up Node.js in the integration job of `.github/workflows/ci.yml`; verify with actionlint and `go test -count=1 -run TestIntegration ./internal/generate/`
- [x] 5.3 Document the listing in `cli/list.mdx` and CI with Node.js in `CONTRIBUTING.md`; verify against `gobento list versions`

## 6. Website

- [x] 6.1 Regenerate the catalog (`make catalog`), extend the schema and the frozen fixture with `node` and `npm`, show the scaffolded `frontend/` in the configurator's file preview, list the three frontends on the home page and drop them from the roadmap lists, and show npm packages in the component reference; verify with `pnpm test`, `pnpm lint`, `pnpm typecheck`, `npx prettier --check .`, and `CI= E2E_PORT=4188 pnpm e2e --workers=4`
- [x] 6.2 Verify the CLI accepts every website command: `go test -count=1 -run 'TestCommittedCatalogIsCurrent|TestWebsiteCommands' ./internal/tools/catalogexport ./internal/cli`

## 7. Verification

- [x] 7.1 Run `go test -short ./...`, `go test ./...` with a local PostgreSQL (`TEST_DATABASE_URL`), `make lint`, and actionlint; record the results

## Workflow follow-up

- Archive with `openspec archive add-vite-frontends -y` and validate with `openspec validate --specs --strict`.
