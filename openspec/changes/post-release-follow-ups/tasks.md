# Tasks

## 1. Release configuration

- [x] 1.1 Remove `"release-as": "0.1.0"` from `release-please-config.json`; update the "First release" paragraph and the suggested repository description in `docs/maintainers.md`

## 2. Version reporting

- [x] 2.1 Fall back to the main module version from `debug.ReadBuildInfo` in `cmd/gobento/main.go` when no `-ldflags` version is set, with a table test covering release, `go install`, pseudo-version, `(devel)`, and missing build info
- [x] 2.2 Update docs: `website/content/docs/cli/new.mdx` (`--version`) and the installation guide (v0.1.0 via `go install` still prints `dev`); verify `go install ...@v0.1.0` prints `gobento version dev` and the release binary prints `gobento version 0.1.0`

## 3. Release-aware website

- [x] 3.1 Word the hero note and the closing call to action from the release manifest (`components/home/ReleaseCopy.tsx`), with unit tests for both states
- [x] 3.2 Derive end-to-end expectations from the real manifest through `lib/release` (`tests/helpers.ts`): the notice on every page including all docs pages, the home install note, and the configurator defaults; verify against a released build and a build with the manifest set to `0.0.0`
- [x] 3.3 Update docs: `README.md`, `SECURITY.md`, and `website/content/docs` (`index`, `getting-started/installation`, `getting-started/quick-start`, `cli/new`); verify `go install ...@v0.1.0`, `gobento new shop --dry-run`, `gh release download` and the `sha256sum`/`shasum` checksum commands as documented

## 4. CI evidence and Lighthouse

- [x] 4.1 Upload `playwright-report/` and `test-results/` as the `playwright-report` artifact when `website-e2e` fails (7 days); add the HTML reporter on CI; `actionlint` clean; document it in `docs/maintainers.md`
- [x] 4.2 Run Lighthouse five times per URL (median run asserted, budget unchanged at 0.9); update `docs/maintainers.md`; record the local five-run medians
