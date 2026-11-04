# Proposal

## Why

The project tooling (base layer with contribution points) and the htmx
frontend were built on separate branches. After both were merged, a project
generated with `--frontend htmx` gets the Makefile, CI workflow, and
Dockerfile, but none of them knows about templ: `make generate` regenerates
only sqlc code, CI never checks that the committed `pages_templ.go` matches
`pages.templ`, and the README's Docker section does not say where the pages
and the API are served. The base layer was designed so that a component of a
kind added later contributes without base layer changes; the frontend kind
is the first to prove it.

## What Changes

- `frontend/htmx` gets contribution templates:
  - `make-generate`: `$(GO) generate ./<web dir>`, which runs the
    `//go:generate` line and thus the templ version from the rules file.
  - `ci-steps`: a step "Check that the templ code is current" that runs
    `go generate ./<web dir>` and fails, showing the diff, when the web
    directory then differs from the commit.
  - `dockerignore`: `*_templ.txt`, the files `templ generate --watch` writes.
  - `readme-docker`: the pages and static files are embedded in the binary;
    the pages are at `/` and the API, its health check included, under
    `{{.APIPrefix}}/`.
- The architectures' `.gitignore` lists `*_templ.txt` when the `web` slot is
  filled.
- The htmx notes point to `make generate` and the CI check, and name
  `go generate ./<web dir>` as the command without make.
- No base layer change and no change to the Dockerfile: `COPY . .` already
  copies the embedded static files.
- Tests: a compose test for the htmx tooling in every architecture, the
  goldens (`golden-htmx`, `golden-ddd-htmx`, `golden-components/frontend/htmx`),
  a templ freshness check in `TestIntegration` for every covered selection
  with a frontend, and an htmx image in `TestIntegrationDocker` that serves
  `/`, `/static/htmx.min.js`, and `/api/health`.
- Docs: project tooling page (targets, CI, image, points table), quick start.

## Capabilities

### Modified Capabilities

- `project-tooling`: the tooling regenerates and checks a frontend's
  generated code and serves its pages from the image.

## Impact

- `components/frontend/htmx/` (new `contrib/`, notes), the three
  architectures' `dot-gitignore.tmpl`.
- `internal/compose` and `internal/generate` tests and goldens.
- `website/content/docs/concepts/project-tooling.mdx`,
  `website/content/docs/getting-started/quick-start.mdx`.
- No change to the CLI, the catalog format, or the rules file.
