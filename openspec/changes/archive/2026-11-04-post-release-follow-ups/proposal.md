# Proposal

## Why

gobento v0.1.0 is released (GitHub release with GoReleaser binaries and
`checksums.txt`), and the website is live. Several things still assume that
nothing is released: the repository and website docs say "not released yet",
the home page says "There is nothing to install yet", end-to-end tests
hard-code the pre-release notice (they failed in CI after the release),
release-please still forces `0.1.0`, and `go install ...@v0.1.0` reports
`gobento version dev`.

## What Changes

- **Release config:** remove the one-off `"release-as": "0.1.0"` from
  `release-please-config.json`.
- **Version reporting:** when no version was set with `-ldflags`, `gobento
  --version` falls back to the main module version Go recorded
  (`debug.ReadBuildInfo`), so `go install ...@vX.Y.Z` reports `X.Y.Z`.
- **Release-aware home page:** the hero note and the closing call to action
  are worded from the release manifest: before a release they say there is
  nothing to install yet; after it the hero names the version and links to the
  installation guide and the prebuilt binaries.
- **Release-aware tests:** end-to-end expectations about the pre-release
  notice and install copy derive from the real manifest (parsed with
  `lib/release`), covering every page; unit tests keep fixtures for both
  states.
- **Install docs:** real install instructions (`go install`, prebuilt binaries,
  checksum verification, releases page) replace "not released yet".
- **CI evidence:** the website workflow uploads the Playwright report and
  test results when the end-to-end job fails.
- **Lighthouse robustness:** more Lighthouse runs per URL, so the median that
  is asserted against the unchanged 0.9 budget is less noisy on CI runners.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `project-generation`: adds the version reporting requirement.
- `project-website`: adds release-aware install guidance (the requirement is
  added here because `add-project-website` is not archived yet).

## Impact

- Code: `cmd/gobento/main.go`, `website/components/home/`,
  `website/components/site/links.ts`, `website/app/(home)/page.tsx`,
  `website/tests/`, `website/lighthouserc.cjs`,
  `.github/workflows/website.yml`, `release-please-config.json`.
- Documentation updated in the same change: `README.md`, `SECURITY.md`,
  `docs/maintainers.md`, and `website/content/docs` (`index`,
  `getting-started/installation`, `getting-started/quick-start`, `cli/new`).
- The GitHub repository description is changed by a maintainer by hand (see
  `docs/maintainers.md`).
