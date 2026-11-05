# Tasks

## 1. Contributions

- [x] 1.1 Add `contrib/make-generate.tmpl`, `ci-steps.tmpl`, `dockerignore.tmpl`, and `readme-docker.tmpl` to `components/frontend/htmx`, and list `*_templ.txt` in the architectures' `.gitignore` when the `web` slot is filled; verify in a generated htmx project that `make generate` changes nothing and that the CI step passes when fresh and fails after committing an edited `pages.templ`
- [x] 1.2 Point the htmx notes to `make generate` and the CI check; regenerate the goldens with `go test ./internal/compose -run TestGolden -update` and review the diff

## 2. Tests

- [x] 2.1 Add `TestShippedToolingFollowsFrontend` to `internal/compose` (every architecture with htmx, and none of it without a frontend); verify `go test ./internal/compose`
- [x] 2.2 Check templ freshness and `make generate` in `TestIntegration` for selections with a frontend, and build and run an htmx image in `TestIntegrationDocker`; verify with `go test -count=1 -run TestIntegration ./internal/generate/` and a local PostgreSQL

## 3. Docs and verification

- [x] 3.1 Document the htmx contributions in `website/content/docs/concepts/project-tooling.mdx` and `getting-started/quick-start.mdx`; verify against a generated project
- [x] 3.2 `go test -short ./...`, `make lint`, `make catalog-check`, `openspec validate --strict`, and the website checks pass
