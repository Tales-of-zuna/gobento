# Proposal

## Why

The CI run for pull request #13 (`next` into `main`) failed two checks that
pass on `main`:

- The three integration jobs stopped at `go test`'s default 10-minute timeout
  (`panic: test timed out after 10m0s`). `TestIntegration` now generates and
  checks many more selections (324 supported, with Vite frontends that run
  `npm`), and a full run takes about 11 minutes even on a fast local machine.
- Lighthouse scored the quick-start docs page 0.79–0.83 for mobile
  performance, below the 0.9 budget. The cause is a cumulative layout shift
  of 0.173: the new opening paragraph has many inline code spans, and the
  code font (not preloaded, `font-display: swap`) arrives after first paint
  and reflows that paragraph. The home and configurator pages have no shift.

## What Changes

- The code font uses `font-display: optional`. On a slow first load, code
  stays in the fallback monospace font instead of swapping late and moving
  the text. Measured locally, quick-start goes from CLS 0.173 and a 0.81–0.86
  score to CLS 0 and 0.93–0.94.
- The Lighthouse CI assertions also cap cumulative layout shift at 0.1 per
  page, so a font or layout regression fails with its real cause.
- CI's integration job and the documented integration commands
  (`make test-all`, CONTRIBUTING.md, docs/maintainers.md, the
  adding-a-component guide) pass `-timeout 60m`.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
- `project-website`: the Performance requirement adds a layout-stability
  scenario.

## Impact

- `website/app/layout.tsx`, `website/lighthouserc.cjs`
- `.github/workflows/ci.yml`, `Makefile`, `CONTRIBUTING.md`,
  `docs/maintainers.md`, `website/content/docs/contributing/adding-a-component.mdx`
- No change to generated projects or the CLI.
