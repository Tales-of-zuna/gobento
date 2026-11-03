# Tasks

## 1. Legal and community files

- [ ] 1.1 Add `LICENSE` (MIT, "Copyright (c) 2026 The gobento authors"); verify GitHub's licensee detection would match it by comparing it with the SPDX MIT text (only the copyright line differs)
- [ ] 1.2 Write `README.md` (badges, pitch, status notice, install and usage placeholders, contributing, security, license); verify every link resolves to a file in the repo or a valid URL
- [ ] 1.3 Add `CODE_OF_CONDUCT.md` with the official latest Contributor Covenant text and the enforcement contact `tales.of.zuna@gmail.com`; verify the text matches the official source apart from the contact line
- [ ] 1.4 Write `SECURITY.md` (private vulnerability reporting, supported versions, response time) and `SUPPORT.md`; verify they are linked from README and CONTRIBUTING
- [ ] 1.5 Write `CONTRIBUTING.md` (setup, Make targets with plain-command equivalents, Windows notes, the OpenSpec workflow, Conventional Commit PR titles, review process); verify a fresh clone in a temp dir follows it to a passing `make lint test`

## 2. GitHub collaboration files

- [ ] 2.1 Add issue forms `bug_report.yml`, `feature_request.yml`, and `component_request.yml`, plus `config.yml` (blank issues off, security contact link); verify each parses as YAML and has the top-level keys GitHub requires (`name`, `description`, `body`)
- [ ] 2.2 Add `.github/pull_request_template.md` and `.github/CODEOWNERS`; verify after the first push that GitHub's CODEOWNERS view reports no errors (until then, check each line is a pattern followed by `@owner`)
- [ ] 2.3 Add `.github/dependabot.yml` (gomod and github-actions, weekly, grouped, conventional prefixes); verify it parses and matches the documented schema keys

## 3. Editor, Git, and module bootstrap

- [ ] 3.1 Add `.gitattributes`, `.editorconfig`, and `.gitignore`; verify `git ls-files --eol` reports `i/lf` for all text files after `git add --renormalize .`
- [ ] 3.2 Add `go.mod` (`github.com/Tales-of-zuna/gobento`, `go 1.26.0`) and `doc.go`; verify `go build ./...`, `go vet ./...`, and `go test ./...` succeed

## 4. Developer tooling

- [ ] 4.1 Add `.golangci.yml` (v2 format) per design D6; verify `golangci-lint config verify` and `golangci-lint run` pass
- [ ] 4.2 Add the `Makefile` per design D5 with `help` as the default target; verify every target runs successfully and `make` prints the target list

## 5. CI and release automation

- [ ] 5.1 Add `.github/workflows/ci.yml` (lint, tidy check, test matrix, govulncheck, actionlint) with SHA-pinned actions and least-privilege permissions; verify `actionlint` passes locally
- [ ] 5.2 Add `.github/workflows/pr-title.yml` (semantic PR title check, no checkout, read-only); verify `actionlint` passes and the type list matches CONTRIBUTING
- [ ] 5.3 Add `release-please-config.json`, `.release-please-manifest.json`, `.goreleaser.yaml`, and `.github/workflows/release.yml` per design D9; verify `actionlint` passes and `goreleaser check` validates the config

## 6. Maintainer guide and final check

- [ ] 6.1 Write `docs/maintainers.md` (manual GitHub settings with reasons, how to release, how to bump golangci-lint); verify every setting it names exists in GitHub's current repository settings documentation
- [ ] 6.2 Run the full local gate (`make lint test`, `actionlint`, `goreleaser check`) on a clean checkout; verify everything passes and `git status` is clean

## Workflow follow-up

- After review, push to `origin` and open the first PR only with the user's approval; then the user applies the settings in `docs/maintainers.md`.
- Archive the change with `/opsx:archive` after it is merged.
