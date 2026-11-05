# Tasks

## 1. Tooling versions as data

- [x] 1.1 Add the `tooling` section (images, actions, tools) to `internal/compat` parsing and validation, and its resolution for the chosen Go version; verify with new cases in `go test ./internal/compat` (valid file, malformed SHA, digest, tag, placeholder, version, unknown fields)
- [x] 1.2 Record the golang, distroless, and postgres images, `actions/checkout`, `actions/setup-go`, `golangci/golangci-lint-action`, and golangci-lint in `components/versions.yaml` with the check date; verify the SHAs with `gh api repos/<repo>/commits/<tag>` and the digest with `docker buildx imagetools inspect`, and `go test ./internal/catalog`
- [x] 1.3 Document the tooling section in `website/content/docs/concepts/version-rules.mdx`; verify every key named there exists in the parser

## 2. Base layer and contribution points

- [x] 2.1 Load `components/base/` (`base.yaml`, `notes.md.tmpl`, `files/root/` only) and `contrib/<point>.tmpl` in components, rejecting undeclared points; verify with new cases in `go test ./internal/catalog`
- [x] 2.2 Compose the base layer into every selection, add the `contrib`, `indent`, and `prefix` functions and the `.Images`, `.Actions`, `.Tools` data, append the base notes last, and expose rendered contributions in the plan; verify with new fixture tests in `go test ./internal/compose` (order, placement, unknown point, path conflict with the base layer, contrib outside the base layer)
- [x] 2.3 Pass the base layer and tooling versions from `Prepare` and `compose.Check`; verify `go test -short ./...`
- [x] 2.4 Document the base layer and contributions in `website/content/docs/contributing/adding-a-component.mdx`, `website/content/docs/concepts/components.mdx`, and `website/content/docs/concepts/slots-and-contracts.mdx`; verify the documented points match `components/base/base.yaml`

## 3. Generated tooling files

- [x] 3.1 Write the base layer templates (Makefile, Dockerfile, .dockerignore, CI workflow, Dependabot, golangci-lint config, Development notes); verify by generating a project and running `make help build lint`, `actionlint`, and `docker build`
- [x] 3.2 Add the SQLite, PostgreSQL, and sqlc contributions, move the PostgreSQL image in its notes to `.Images.postgres`, and silence the G101 false positive on PostgreSQL's `DefaultURL`; verify the CI and Dockerfile differences with compose tests and a generated project of each database
- [x] 3.3 Regenerate the goldens with `go test ./internal/compose -run TestGolden -update`, adding `contrib/<point>` pseudo-files to the component goldens; review the diff and verify `go test ./internal/compose`
- [x] 3.4 Extend `TestIntegration` (actionlint on every workflow, `make build` and `make lint` where make exists) and add `TestIntegrationDocker`; verify with `go test -count=1 -run 'TestIntegration' ./internal/generate/` with a local PostgreSQL
- [x] 3.5 Document the generated tooling: new page `website/content/docs/concepts/project-tooling.mdx`, the "What gets generated" tree in `getting-started/quick-start.mdx`, the dry-run list in `cli/new.mdx`, the tree and "How it works" in `README.md`, and `CONTRIBUTING.md`; verify every documented file list and command against the built binary's `--dry-run` output and a generated project

## 4. Website

- [x] 4.1 Add the tooling files to the configurator's file-tree preview (`website/components/configurator/view-model.ts`) and its tests, and run `make catalog`; verify `pnpm test`, `pnpm lint`, `pnpm typecheck`, `npx prettier --check .`, and `CI= E2E_PORT=4184 pnpm e2e --workers=4` in `website/`

## 5. Verification

- [x] 5.1 `go test -short ./...`, `go test ./...` with `TEST_DATABASE_URL` set to a local PostgreSQL, `make lint`, and `go mod tidy` leave no diff
- [x] 5.2 `openspec validate add-project-tooling --strict` passes

## Workflow follow-up

- Archive with `openspec archive add-project-tooling -y` and check `openspec validate --specs --strict`.
- Commit on `feat/project-tooling` with Conventional Commits messages; do not push.
