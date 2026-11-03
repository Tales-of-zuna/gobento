# Tasks

Groups 1, 2, and 3 can run in parallel. Group 7 needs `add-generator-foundation`
to be applied.

## 1. Scaffold

- [x] 1.1 Scaffold `website/` (Next.js App Router, TypeScript strict, Tailwind, Fumadocs) with exact versions from the version research, `packageManager` (pnpm) and `.nvmrc` (Node 24); verify `pnpm install --frozen-lockfile` and `pnpm build` succeed and produce a static export
- [x] 1.2 Configure the static export for GitHub Pages (`basePath: "/gobento"`, `images.unoptimized`, `trailingSlash`, `.nojekyll`); verify a test server on the `out/` folder at `/gobento/` serves the home page with all assets loading
- [ ] 1.3 Add ESLint, Prettier, Vitest with Testing Library, and Playwright with axe as pnpm scripts (`lint`, `typecheck`, `test`, `e2e`), and Makefile targets `web-dev`, `web-check`, `web-build`; verify each script runs green on the scaffold
- [x] 1.4 Initialise shadcn/ui against the Fumadocs Tailwind setup; verify a shadcn Button and the Fumadocs docs layout render together without style conflicts
- [ ] 1.5 Add a CI assertion that `go list ./...` reports no package under `website/`; verify it passes with `node_modules` installed

## 2. Visual design in Claude Design

- [x] 2.1 Create the Claude Design project "gobento website" and produce the design direction: palette (light and dark), typography with licensed fonts, spacing, radius and shadow tokens, motifs; verify the maintainer can open it
- [x] 2.2 Prototype the mascot 3D-first: at least three original gopher-and-bento concepts in three.js (toon shading, direction A palette), rendered as front and three-quarter views on one contact sheet; the maintainer picks one; verify it contains no copied artwork and record the attribution text
- [x] 2.3 Design the home, configurator, and docs pages at desktop and mobile sizes in light and dark themes; verify every state named in the specs (disabled option with reason, invalid field, copied confirmation, pre-release notice, WebGL fallback) is drawn
- [x] 2.4 Review with the maintainer and record the approved direction, with the token values, in `website/DESIGN.md`; verify the maintainer's explicit approval before group 4 starts

## 3. Catalog and command logic (`website/lib`)

- [x] 3.1 Define the catalog schema (zod) and types, and write the first-slice fixture `src/generated/catalog.json`; verify schema tests accept the fixture and reject malformed variants
- [x] 3.2 Implement range and constraint evaluation and compatibility explanations; verify table tests mirroring the Go compat tests (each comparator, Go-version constraint, unmet trait)
- [x] 3.3 Implement the command builder (fixed flag order, defaults omitted) and the name and module validation from the catalog's rules; verify table tests for the scenarios *Changing a choice* and *Invalid project name*
- [x] 3.4 Implement URL state encoding and decoding with fallback for unknown values; verify tests for *Share a link* and *Unknown value in a shared link*

## 4. Theme, landing page, and mascot

- [ ] 4.1 Apply the approved tokens as CSS variables for Tailwind, shadcn, and Fumadocs, with a theme switch that follows the system preference and remembers a choice without a flash; verify e2e tests for *System preference* and *Explicit choice persists*
- [ ] 4.2 Build the procedural 3D mascot in React Three Fiber from the approved sheet, and review it with the maintainer; if it does not reach the approved look, switch to a GLB model (record the decision in design.md); verify the maintainer's approval of the result
- [x] 4.3 Add lazy loading, WebGL detection with static-still fallback, an error boundary, reduced-motion static pose, DPR clamp, and off-screen pause; verify e2e tests for *Reduced motion requested* and *WebGL unavailable*
- [ ] 4.4 Build the landing page (value proposition, quick-start copy, links), the pre-release notice driven by the release manifest, and the footer with gopher attribution and non-affiliation; verify e2e tests for *First visit*, *Copy quick start*, *No release yet*, and *Footer credits*

## 5. Configurator page

- [ ] 5.1 Build the configurator UI on the logic from group 3 (choices, Advanced versions, live command, copy, disabled options with explanations, URL sync); verify e2e tests for *Defaults*, *Changing a choice*, *Copy command*, *Version too new for Go*, and *Share a link*
- [ ] 5.2 Verify keyboard-only operation and axe in both themes on the configurator and home pages (spec *Keyboard only*, *Automated audit*)

## 6. Documentation

- [x] 6.1 Write the docs sections (Getting started, Concepts, CLI reference, Contributing a component) in `content/docs`, with sidebar and previous/next navigation; verify the e2e test for *Navigation*
- [x] 6.2 Render the Components reference from `catalog.json`; verify a test that changes a fixture default and finds the new value on the rendered page (spec *Version bump*)
- [x] 6.3 Enable static search, "Edit this page" links, and code-block copy controls; verify e2e tests for *Searching*, *Suggesting an edit*, and *Copy a command*

## 7. Real catalog integration (after add-generator-foundation)

- [ ] 7.1 Write `internal/tools/catalogexport` on the real `internal/catalog` and `internal/compat`, emitting the D3 schema; verify its output passes the zod schema and replaces the fixture
- [ ] 7.2 Add the CI freshness check (re-export, fail on diff) and the CLI cross-check (every supported selection's command passes `gobento new --dry-run`); verify both run green in CI
- [ ] 7.3 Move `docs/components.md` into `content/docs/contributing/adding-a-component.mdx`, leaving a stub that links to the new page; verify links in the repository still resolve

## 8. CI, deployment, and maintenance

- [ ] 8.1 Add `.github/workflows/website.yml` (path-filtered; install, lint, typecheck, test, build, e2e with axe, Lighthouse CI with the spec budget) with SHA-pinned actions; verify `actionlint` passes and a PR run is green
- [ ] 8.2 Add `.github/workflows/pages.yml` (build and deploy to GitHub Pages on `main`) and the `npm` entry in `.github/dependabot.yml` (weekly, grouped, 7-day cooldown); verify `actionlint` passes
- [ ] 8.3 Document in `docs/maintainers.md` the Pages setting (Settings → Pages → Source: GitHub Actions) and the website commands in `CONTRIBUTING.md`; verify the documented commands run as written

## 9. Launch check

- [ ] 9.1 After the user approves merging, verify the deployed site at `https://tales-of-zuna.github.io/gobento/`: home, configurator, docs, and search work, and Lighthouse mobile scores meet the budget on the live URL (spec *Merge to main*, *Lighthouse budget*)

## Workflow follow-up

- Merge order: PR #1 (`establish-repo-standards`) first; this change's pull request after it.
- Archive the change with `/opsx:archive` after it is merged and verified live.
