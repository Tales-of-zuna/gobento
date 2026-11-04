# Tasks

## 1. Optional frontend kind

- [ ] 1.1 Add `KindFrontend`, `Kind.Optional()`, `Selection.Frontend`, `Selection.Component(kind)`, nil-aware `Components`/`String`, and frontend-aware `Candidates` in `internal/catalog`; reject a component named `none` in an optional kind; verify with `go test ./internal/catalog`
- [ ] 1.2 Accept a nil optional kind in `internal/compose` (`checkSelection`), add `SlotData.Filled` and `Data.APIPrefix`; verify with new compose tests for a filled and an unfilled optional slot (`go test ./internal/compose`)
- [ ] 1.3 Resolve `""` and `none` to no frontend in `generate.Prepare`, list `none` among the supported names of an optional kind, and never require an optional kind's flag; verify with `go test ./internal/generate`
- [ ] 1.4 Add `--frontend` to `gobento new`, the frontend question with "None (API only)" first, nil-aware `matching`, and the equivalent command rule for none; verify with `go test ./internal/cli`
- [ ] 1.5 Export `optional` kinds and selections without a frontend key from `internal/tools/catalogexport`; verify with `go test ./internal/tools/catalogexport`
- [ ] 1.6 Document the optional kind in `website/content/docs/concepts/components.mdx`, `cli/new.mdx` (flag, questions, examples), and `README.md` usage; verify every documented command with `--dry-run`

## 2. Vendored assets in the rules

- [ ] 2.1 Parse and validate `assets` in `internal/compat` (components, version, https source, path, 64-hex sha256), resolve the assets of selected components, reject pins of assets as not pinnable; verify with `go test ./internal/compat`
- [ ] 2.2 Check asset components in `catalog.Load`, pass `.Assets` to templates, and export assets for the website; verify with `go test ./internal/catalog ./internal/compose ./internal/tools/catalogexport`
- [ ] 2.3 Document `assets` in `concepts/version-rules.mdx`, `contributing/adding-a-component.mdx`, and the header of `components/versions.yaml`; verify the documented YAML keys match the parser

## 3. Architectures support frontends

- [ ] 3.1 Declare the optional `web` slot in `mvc` (`internal/views`) and `vertical-slice` (`internal/web`); move wiring into `newHandler`, mounting the API under `{{.APIPrefix}}/` and the web handler at `/` when `.Slots.web.Filled`; verify the API-only goldens are unchanged (`go test ./internal/compose -run TestGolden`)
- [ ] 3.2 Add the entry routing test `main_test.go.tmpl` to both architectures and the web row, browser hint, and `/api` curl paths to both READMEs; prefix the API tables in the framework notes with `{{.APIPrefix}}`; verify with `go test ./internal/compose -run TestGolden -update` and a review of the diff
- [ ] 3.3 Document the frontend contract, serving, and the architecture checklist in `concepts/slots-and-contracts.mdx` and `contributing/adding-a-component.mdx`; verify the documented signatures against the templates

## 4. htmx + templ frontend

- [ ] 4.1 Add `components/frontend/htmx`: manifest, `handler.go.tmpl`, `pages.templ.tmpl`, shipped `pages_templ.go.tmpl`, `static/htmx.min.js.tmpl` (htmx 2.0.11 verbatim), `static/styles.css.tmpl`, `handler_test.go.tmpl`, and `notes.md.tmpl`; register `github.com/a-h/templ` (library, `=v0.3.1100`) and the `htmx.org` asset in `components/versions.yaml`; verify with `go test ./internal/catalog ./internal/compose`
- [ ] 4.2 Add a full-project golden for `mvc/nethttp/sqlite/sqlc/htmx`, the shipped-asset checksum test, the upstream asset test (network), and a test that framework files are identical with and without a frontend; verify with `go test ./internal/...`
- [ ] 4.3 Add the `templ-fresh` CI job; verify with actionlint and by running its steps locally (no diff after `go generate ./...`)
- [ ] 4.4 Document htmx in `README.md`, `website/content/docs/index.mdx`, `getting-started/quick-start.mdx` ("With htmx"), `getting-started/installation.mdx`, and the contributor guide (regenerating templ output, updating htmx); verify the documented commands and output against a generated project

## 5. CI coverage

- [ ] 5.1 Add `compose.Coverage` (every selection without an optional component plus a greedy pairwise covering set of the rest) and `TestCoverage` proving the pair property on the shipped catalog; verify with `go test ./internal/compose`
- [ ] 5.2 Make `TestIntegration` generate `Coverage(Supported(cat))`, check that the web package depends on neither the database nor the store package nor a framework module; verify with `go test -count=1 -run TestIntegration ./internal/generate/`
- [ ] 5.3 Update the CI coverage statements in `CONTRIBUTING.md`, `README.md`, `concepts/components.mdx`, `contributing/adding-a-component.mdx`, `openspec/config.yaml`, and the website's CI copy; verify by searching for "every supported" across the repository

## 6. Website

- [ ] 6.1 Regenerate `website/src/generated/catalog.json` and `lib/command/supported-commands.txt` (`make catalog`), and update the schema (`optional`, `assets`, selections without optional kinds), `KIND_IDS`, the frozen fixture catalog, the configurator ("None (API only)"), URL state, command builder, home page kinds and stack, roadmap lists (React, Vue, Svelte, Tailwind CSS planned; htmx shipped), and the component reference (assets, optional); verify with `pnpm test`, `pnpm lint`, `pnpm typecheck`, `npx prettier --check .`, and `pnpm e2e`
- [ ] 6.2 Verify the CLI accepts every website command: `go test -count=1 -run 'TestCommittedCatalogIsCurrent|TestWebsiteCommands' ./internal/tools/catalogexport ./internal/cli`

## 7. Verification

- [ ] 7.1 Generate `gobento new shop --arch mvc --framework nethttp --db sqlite --data sqlc --frontend htmx`, run it, and check `curl /`, an htmx `POST /items`, a plain form post, and `curl /api/items`; also run `govulncheck` in it; verify the responses and that no vulnerability is reported
- [ ] 7.2 Run `go test -short ./...`, `go test ./...` (with a local PostgreSQL on `TEST_DATABASE_URL`), `make lint`, and `openspec validate add-frontend-htmx-templ --strict`; verify all pass

## Workflow follow-up

- Archive with `openspec archive add-frontend-htmx-templ -y`, then `openspec validate --specs --strict`.
- The DDD architecture change follows the frontend checklist in the contributor guide to support frontends.
- Follow-up changes: Vite frontends (scaffolder hook, npm tool versions) and Tailwind CSS, as in design D10.
