# Proposal

## Why

gobento is meant to be a role-model open-source Go repository: anyone who opens
it should find the standard files, conventions, and tooling where they expect
them, and be able to contribute or use it without asking. Setting this up before
the first line of product code means every later commit already follows it.

## What Changes

- **Legal and community files:** an MIT `LICENSE`, `README.md`,
  `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md` (Contributor Covenant), `SECURITY.md`
  (GitHub private vulnerability reporting), and `SUPPORT.md`.
- **GitHub collaboration:** issue forms (bug report, feature request, component
  request), a pull request template, `CODEOWNERS`, and Dependabot for Go modules
  and GitHub Actions.
- **Editor and Git hygiene:** `.editorconfig`, `.gitattributes` (LF line
  endings everywhere), and a Go `.gitignore`.
- **Developer tooling:** a `Makefile` with self-documenting targets (`help`,
  `build`, `test`, `test-all`, `lint`, `fmt`, `tidy`, `cover`, `clean`) that thinly
  wraps the go commands, with the equivalent plain commands documented for
  Windows; and a golangci-lint v2 configuration.
- **CI baseline:** lint, `go mod tidy` drift check, unit tests on supported Go
  versions and all three operating systems, `govulncheck`, `actionlint`, and a
  Conventional Commits check on pull request titles. Actions are pinned by
  commit SHA.
- **Releases:** release-please turns Conventional Commits into a release pull
  request with a generated `CHANGELOG.md` and SemVer tag. GoReleaser then builds
  and publishes binaries for Linux, macOS, and Windows with checksums. Nothing
  is released until the user decides to merge the first release PR.
- **Module bootstrap:** `go.mod` for `github.com/Tales-of-zuna/gobento` and a
  package doc file, so the CI baseline has a module to check.
- **Maintainer guide:** `docs/maintainers.md` lists the GitHub settings to apply
  by hand (branch protection, squash merges, private vulnerability reporting).

## Capabilities

### New Capabilities
<!-- None: repository tooling and documentation only; no product behavior
     changes. The change sets skip_specs: true. -->

### Modified Capabilities
<!-- None. -->

## Impact

- Lands before `add-generator-foundation`, which extends the CI workflow and
  README added here.
- Publishing effects happen only on push or merge, and each needs the user's
  approval: the license and community files become public, and release
  workflows become active.
- Repository settings on GitHub are documented, not changed, by this change.
