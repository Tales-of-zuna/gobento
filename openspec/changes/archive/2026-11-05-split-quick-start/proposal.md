# Proposal

## Why

After the layout-shift fix, the quick-start docs page still scored 0.88–0.89
for mobile Lighthouse performance on a slower CI runner (budget 0.9; the run
before scored at least 0.9). Its LCP element, the intro paragraph, waits 3.3s
in render delay. The page grew with every new option, and its HTML is now
294 KB raw (34 KB compressed) with 1,659 DOM elements, against 72 KB for the
home page. It sits at the edge of the budget, so whether CI passes depends on
the runner.

The page has also stopped being a quick start. More than half of it walks
through what changes with each other architecture, database, data access,
frontend, and CSS framework.

## What Changes

- Quick start keeps the path a newcomer follows: install, generate `shop`,
  run its tests, start the server, call the API, and what gets generated. Its
  intro links to the new page for other choices.
- A new Getting started page, **Choose your stack**
  (`/docs/getting-started/choose-your-stack/`), holds the sections on the
  vertical-slice and ddd architectures, PostgreSQL, GORM, htmx, Tailwind CSS,
  and React, Vue, or Svelte, unchanged except for the links between them.
- The docs e2e page list includes the new page, so its previous/next links,
  axe audit, and edit link are tested like the others.

## Capabilities

### New Capabilities
<!-- None. -->

### Modified Capabilities
<!-- None: a content reorganization. The Performance requirement already
     covers the budget this restores. -->

## Impact

- `website/content/docs/getting-started/{quick-start,choose-your-stack}.mdx`, `meta.json`
- `website/tests/docs.spec.ts`
- No change to the CLI, generated projects, or other pages' URLs.
