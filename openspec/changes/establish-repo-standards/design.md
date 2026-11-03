# Design

## Context

The repository is empty apart from OpenSpec scaffolding. The user chose MIT,
GitHub private vulnerability reporting, release-please plus GoReleaser, and a
Makefile. The goal is "a role-model repo", so every choice below prefers the
most widely recognized convention over a custom one. It must work for
contributors on Linux, macOS, and Windows.

## Goals / Non-Goals

**Goals:**
- A newcomer can go from clone to green local checks with only Go installed
  (plus `make` and golangci-lint where noted), following CONTRIBUTING alone.
- CI enforces the same checks as the Makefile, so "works locally" means
  "passes CI".
- Releases need no manual steps beyond merging a release pull request.

**Non-Goals:**
- Changing GitHub repository settings via API. They are documented for the
  maintainer to apply.
- A Developer Certificate of Origin (DCO) sign-off requirement. It adds friction
  for first-time contributors, and the MIT license needs none. We can revisit
  it if corporate contributors appear.
- A docs website, Discussions categories, or sponsorship files.

## Decisions

### D1. Community files follow GitHub's community-health conventions

GitHub detects them in the root or `.github/` and links them in the UI (the
Community Standards checklist).

| File | Content |
|---|---|
| `LICENSE` | MIT, "Copyright (c) 2026 The gobento authors". This form stays correct as contributors join, and is common in Go projects. |
| `README.md` | Badges (CI, Go Reference, Go Report Card, License), a one-paragraph pitch, a status notice ("early development, not released yet"), installation and usage placeholders filled by the generator change, plus contributing, security, and license sections |
| `CONTRIBUTING.md` | Prerequisites; first build; Make targets with plain go equivalents for Windows; the OpenSpec workflow (every behavior change starts as a proposal); Conventional Commits; PR process; how to add a component (links to docs/components.md once it exists) |
| `CODE_OF_CONDUCT.md` | The official Contributor Covenant text, latest version, fetched verbatim during apply, with the enforcement contact `tales.of.zuna@gmail.com` (chosen by the maintainer) |
| `SECURITY.md` | Supported versions (latest minor only, pre-1.0), reporting through GitHub's "Report a vulnerability" button, and the expected response time |
| `SUPPORT.md` | Where to ask questions (GitHub issues with the question label until Discussions exist) and what not to file as an issue |
| `.github/CODEOWNERS` | `* @Tales-of-zuna` |

### D2. Issue forms, not Markdown templates

YAML issue forms (`.github/ISSUE_TEMPLATE/*.yml`) make required fields
mandatory, such as the gobento version, OS, Go version, and the exact command run.

- `bug_report.yml`
- `feature_request.yml`
- `component_request.yml` is specific to gobento: it asks which kind
  (framework, database, data access, architecture), upstream links, and the
  maintenance status of the proposed library.
- `config.yml` sets `blank_issues_enabled: false` and adds a contact link to
  private security reporting.

`.github/pull_request_template.md` holds a short checklist: linked OpenSpec
change or "no behavior change", tests, docs, and a PR title in Conventional
Commits form.

### D3. Commit and merge convention

- PRs are squash-merged, and the PR title becomes the commit message on `main`.
  So the PR title is what must follow Conventional Commits, and CI checks it
  (`amannn/action-semantic-pull-request`). Contributors' intermediate commits
  stay free-form, which is friendlier to newcomers.
- Allowed types: feat, fix, docs, refactor, perf, test, build, ci, chore, revert.
- Breaking changes are marked with `!` or a `BREAKING CHANGE:` footer, which
  drives release-please's version bump.

### D4. Editor and Git hygiene

- `.gitattributes` with `* text=auto eol=lf`, plus `*.png binary`-style entries
  as needed. This is required by the generator (golden files and templates must
  be byte-stable on Windows checkouts).
- `.editorconfig`: tabs for `*.go` and `Makefile`, two spaces for YAML,
  Markdown, and JSON; a final newline; trimmed trailing whitespace (except in
  Markdown).
- `.gitignore`: built binaries (`/gobento`, `/dist/`), coverage output,
  `.gobento-tmp-*`, and OS/editor junk. No `vendor/`, because modules are not
  vendored.

### D5. Makefile is a thin, self-documenting wrapper

- The default target is `help`, which prints targets from `## comment`
  annotations.
- Targets:
  - `build` → `go build -o bin/gobento ./cmd/gobento` (a no-op while the main
    package does not exist yet)
  - `test` → `go test -short -race ./...`
  - `test-all` → `go test -race ./...`
  - `lint` → `golangci-lint run`
  - `fmt` → `golangci-lint fmt`
  - `tidy` → `go mod tidy`
  - `cover` → `go test -short -coverprofile=coverage.out ./...` and
    `go tool cover -func`
  - `clean`
- Every target is one or two plain commands, so Windows users without `make` can
  copy them from CONTRIBUTING.
- `-race` needs cgo. CONTRIBUTING notes that Windows contributors without a C
  compiler should run `go test -short ./...`. CI runs `-race` on Linux and macOS.

*Alternative:* Taskfile. The user picked Makefile, the most recognized option in
Go repositories.

### D6. Lint configuration

`.golangci.yml` in v2 format (`version: "2"`):

- Default linters: govet, staticcheck, errcheck, ineffassign, unused.
- Plus: errorlint, misspell, revive (with exported-comment rules), gosec,
  bodyclose, nolintlint (requires a reason on every `//nolint`), usestdlibvars.
- Formatters: gofmt, goimports (with the module path as the local prefix).
- The golangci-lint version is pinned in two places, the CI action input and
  CONTRIBUTING's install command. A comment in each points to the other.

### D7. CI baseline (`.github/workflows/ci.yml`)

Triggers: pull requests, and pushes to `main`. `concurrency` cancels superseded
runs. `permissions: contents: read` by default.

| Job | Runs on | Steps |
|---|---|---|
| lint | ubuntu | golangci-lint action; `go mod tidy` then `git diff --exit-code` |
| test | {ubuntu, macos, windows} × Go {1.26, 1.27} | `go test -short ./...` (`-race` on ubuntu/macos) |
| vuln | ubuntu | `govulncheck ./...` |
| actions | ubuntu | `actionlint` |
| pr-title | ubuntu, PRs only (separate workflow `pr-title.yml`, on `pull_request_target` with read-only permissions, no checkout) | semantic PR title check |

- Every third-party action is pinned by full commit SHA with a version comment,
  and Dependabot updates them. This is the OpenSSF hardening recommendation.
- `setup-go` reads the Go version from the matrix and uses its built-in module
  cache.
- The generator change later adds the integration and sqlc freshness jobs to
  this file.

### D8. Dependabot

`.github/dependabot.yml`: `gomod` and `github-actions` ecosystems, weekly, with
minor and patch updates grouped into one PR per ecosystem. Commit messages use
the `build(deps):` / `ci(deps):` prefixes, so they satisfy D3.

Note that `components/versions.yaml` (the generator's version data) is *not*
covered by Dependabot. A dedicated bump workflow is future work for the
generator.

### D9. Releases: release-please plus GoReleaser in one workflow

- `release-please-config.json`: release type `go`, `bump-minor-pre-major: true`
  (breaking changes bump the minor version while pre-1.0), and
  `include-v-in-tag: true`. `.release-please-manifest.json` starts at `0.0.0`,
  so the first release is `v0.1.0`.
- `.github/workflows/release.yml` runs on pushes to `main`:
  - release-please opens or updates the release PR.
  - When that PR is merged, it tags and creates the GitHub release, and *the
    same workflow* then runs GoReleaser on the new tag. Tags created with the
    default `GITHUB_TOKEN` do not trigger other workflows, so a separate
    tag-triggered workflow would never fire.
- `.goreleaser.yaml` (v2):
  - builds `./cmd/gobento` with `CGO_ENABLED=0` for linux, darwin, and windows,
    on amd64 and arm64;
  - packages tar.gz archives (zip on Windows) with LICENSE and README;
  - writes `checksums.txt`;
  - sets the version via `-ldflags -X main.version={{.Version}}`;
  - leaves the changelog to release-please (`changelog.disable: true`).
- `go install github.com/Tales-of-zuna/gobento/cmd/gobento@latest` keeps working
  alongside the binaries.
- Signing and SBOMs (cosign, syft) are deferred until the first real release.

### D10. Module bootstrap

`go.mod` (`module github.com/Tales-of-zuna/gobento`, `go 1.26.0`) plus `doc.go`
with the package documentation shown on pkg.go.dev. CI then has a real package
to build, lint, and test before any product code exists. The generator change
adds `cmd/gobento` and `internal/...` on top.

### D11. Maintainer guide

`docs/maintainers.md` lists the settings to apply by hand, and why:

- Branch protection on `main`: require PRs, require the CI jobs, linear
  history.
- Squash merge only, with the default commit message set to the PR title.
- Automatically delete merged branches.
- Enable private vulnerability reporting, Dependabot alerts, and security
  updates.
- Allow GitHub Actions to create pull requests, which release-please needs.

It also covers how to cut a release (merge the release PR) and how to bump the
golangci-lint version.

## Risks / Trade-offs

- [Contributors on Windows lack `make`] → Every target is documented as plain go
  commands.
- [SHA-pinned actions are harder to read] → A version comment on each line, with
  Dependabot keeping them current.
- [`pull_request_target` is a known injection vector] → The PR-title workflow
  never checks out code, has only `pull-requests: read`, and runs a single,
  SHA-pinned action.
- [Release automation publishes binaries] → Nothing is published until the
  maintainer merges a release PR. Until the generator change lands, there is
  nothing to build, and GoReleaser is not reached.

