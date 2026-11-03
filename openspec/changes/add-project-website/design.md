# Design

## Context

See proposal.md for motivation and `specs/` for required behavior. The
repository is a Go module with the standards from `establish-repo-standards`
(CI, Dependabot, Makefile). The component catalog (`components/`) is defined by
`add-generator-foundation`, which may land before or after this change. The
maintainer chose: the site lives in `website/` in this repository, Fumadocs for
docs, GitHub Pages for hosting, and Claude Design for the visual design. The
site's tone takes inspiration from playful, clean product sites built around a
mascot (for example, zustand's), with an original Japanese bento identity.
Nothing is copied from those sites.

## Goals / Non-Goals

**Goals:**
- One static site serving three jobs: landing page, configurator, and docs.
- Zero drift between the configurator, the docs reference, and the CLI. All
  three read the same catalog data.
- A visual identity that is designed and approved before it is coded.

**Non-Goals:**
- Server features. The site is a static export, and the configurator runs
  entirely in the browser.
- Running gobento in the browser (for example, generating a zip). The site
  produces a command, not a project.
- Translations, versioned docs, analytics, and a custom domain (see proposal).

## Decisions

### D1. Layout and toolchain

```
website/
  app/                    Next.js App Router: (home)/, configurator/, docs/[[...slug]]/
  components/ui/          shadcn/ui components (copied in, owned by us)
  components/mascot/      3D scene + static fallback
  content/docs/           MDX documentation
  lib/catalog/            catalog types, validation, compatibility evaluation
  lib/command/            command builder + URL state
  src/generated/catalog.json   catalog data (see D3)
  tests/                  Playwright e2e + axe
  package.json, pnpm-lock.yaml, next.config.ts, source.config.ts, ...
```

- pnpm, pinned with the `packageManager` field in `package.json`, on Node 24
  LTS (`.nvmrc`). It is fast, strict about dependencies, and the norm for
  Next.js and shadcn projects.
- TypeScript in strict mode. ESLint (the Next.js config) plus Prettier. Vitest
  with Testing Library for unit and component tests. Playwright with
  `@axe-core/playwright` for end-to-end and accessibility tests. Lighthouse CI
  for the performance budget.
- `website/` contains no Go files. A CI assertion checks that
  `go list ./...` reports nothing under `website/`, so `node_modules` can never
  leak into Go tooling.
- Makefile targets `web-dev`, `web-check`, and `web-build` wrap the pnpm
  scripts, keeping one entry point for contributors.

### D2. Stack

| Concern | Choice | Why |
|---|---|---|
| Framework | Next.js App Router, `output: "export"` | The maintainer's choice. A static export fits GitHub Pages. |
| UI kit | shadcn/ui on Tailwind CSS | Components are copied in and fully themeable, with accessible Radix primitives. |
| Docs | Fumadocs (core, ui, mdx) | Native to the Next.js App Router, MDX, static search, and its theme shares Tailwind CSS variables with shadcn |
| 3D | three + @react-three/fiber + @react-three/drei | Declarative scenes as React components. The pmndrs ecosystem behind the zustand demo site's style. |
| Animation | Motion | Spring-based UI animation with built-in reduced-motion support |

Exact versions are pinned during apply from the version research (see task
1.1). One shared set of CSS variables (design tokens, D6) themes shadcn and
Fumadocs together, so docs and marketing pages look like one product.

### D3. Catalog bridge: Go owns the logic, the site reads JSON

- `catalog.json` follows a schema defined in `website/lib/catalog/schema.ts`
  (zod) and mirrored by a Go struct. It holds:
  - kinds and components (name, title, description, kind);
  - provided and required traits;
  - the Go version range and default;
  - modules per component, with role, range, default, and constraints already
    normalized (each constraint as `{ when, requires: { module|go, range } }`);
  - the CLI's name and module-path rules as regular expressions.
- **Producer:** a small Go program `internal/tools/catalogexport`, which uses
  the real `internal/catalog` and `internal/compat` packages, writes
  `website/src/generated/catalog.json`. The file is committed, so website
  contributors do not need Go. A CI job re-runs the exporter and fails if the
  file changes, the same freshness pattern as sqlc in
  `add-generator-foundation` (D6).
- **Before the generator lands:** `catalog.json` is a hand-written fixture
  describing the first slice (mvc, nethttp, sqlite, sqlc), validated by the same
  zod schema. Task group 7 switches to the exporter.
- **In the browser:** `lib/catalog` evaluates ranges and constraints with a
  small semver comparison. The logic is deliberately tiny, because the rules
  arrive pre-normalized. The cross-check test (D4) guards against any
  divergence from Go.
- **Version selection:** defaults are preselected. An "Advanced" section lets
  users type a version, which is validated live against the range. Ranges do
  not enumerate versions, and fetching version lists from the Go proxy would
  make the site depend on the network.

*Alternatives considered:*
- Compiling `internal/compat` to WebAssembly. That would be one
  implementation, but it means 2+ MB of WASM for a page whose logic is a few
  comparisons.
- Reading the YAML directly in the website build. That would duplicate the
  normalization in TypeScript, and the two could drift.

### D4. Configurator

- State: `{ name, module, arch, framework, db, data, go, pins }`, kept in sync
  with URL query parameters through `URLSearchParams` (no extra dependency).
  Unknown values fall back to defaults with a notice (spec: *Unknown value in a
  shared link*).
- `lib/command/build.ts` is a pure function from state to command string. It
  uses a fixed flag order matching the CLI reference and omits default
  values. It is unit-tested with table tests.
- Incompatible options are rendered disabled, with a tooltip and an
  `aria-describedby` explanation built from the failed trait or constraint.
- **Cross-check test** (spec: *Generated commands are accepted by the CLI*): a
  Vitest suite enumerates every supported selection from the real
  `catalog.json`, builds each command, and runs
  `go run ./cmd/gobento new … --dry-run` in a CI job that has Go. It runs only
  when the real exporter output is present (task group 7).

### D5. Documentation

- Fumadocs serves `content/docs`. Sections follow the spec: Getting started,
  Concepts, CLI reference, Components, and Contributing a component.
- The Components section is a server component that renders from
  `catalog.json` at build time, so no MDX contains a version number.
- Search is Fumadocs' static index, built at export time and queried in the
  browser.
- "Edit this page" links to
  `https://github.com/Tales-of-zuna/gobento/edit/main/website/content/docs/<path>`.
- The existing `docs/components.md` from the generator change moves into
  `content/docs/contributing/components.mdx` when both changes have landed. A
  stub at the old path keeps links working.

### D6. Visual design in Claude Design, approved before code

A Claude Design project, "gobento website", is created through the Claude
Design tools and shared with the maintainer. It contains:

1. **Design direction and tokens:**
   - A Japanese bento palette: washi off-white, sumi ink, vermilion
     lacquer, matcha, kohaku gold, and dark-mode equivalents.
   - Type: a clean sans for text, a rounded display face for headings, and
     Japanese glyph accents such as 弁当, with license-compatible fonts only.
   - Spacing, radius, and shadow tokens.
   - Motifs (bento compartments as a grid system, subtle wave or asanoha
     patterns), used respectfully and not as decoration clichés.
2. **Mascot sheet:** an original gopher-inspired character in or as a bento
   box, shown in front, three-quarter, and side views, with expressions and a
   static hero pose. This sheet is the reference for the 3D build and the
   source of the SVG fallback and favicon.
3. **Pages:** home, configurator, and a docs page, at desktop and mobile
   sizes, in light and dark themes.

The maintainer approves the direction (task 2.4) before implementation starts.
Approved tokens become CSS variables in `app/globals.css`, consumed by
Tailwind, shadcn, and Fumadocs.

### D7. Mascot implementation

- **3D:** built in code from primitives in React Three Fiber (rounded boxes for
  the bento, capsules and spheres for the gopher), with toon shading and
  outlines matching the approved 2D sheet. This keeps the mascot reviewable as
  code, animatable (idle bob, blink, lid opening on hover), and small, with no
  binary model pipeline.
  *Alternative:* a GLB modelled in Blender. Kept as the fallback if the
  procedural version cannot reach the approved look. That decision is made at
  task 4.2, with the maintainer.
- **Static fallback:** an SVG exported from the approved sheet. It is used for
  reduced motion, without WebGL, while loading, in the favicon, and in social
  images.
- **Attribution:** the footer credits Renée French (CC BY 4.0) for the original
  Go gopher design and states non-affiliation with the Go project. The mascot
  is our own drawing. The Go logo is not used.

### D8. 3D loading and runtime

- The scene is a client component loaded with `next/dynamic` (`ssr: false`).
  The static SVG renders first, so the scene never blocks content or LCP.
- WebGL support is detected before mounting. On failure, or on an error
  boundary catch, the SVG stays.
- `prefers-reduced-motion`: the scene renders a single static pose
  (`frameloop="demand"`) and Motion's reduced-motion setting disables
  decorative transitions.
- Device pixel ratio is clamped to [1, 2]. Rendering pauses when the canvas is
  off-screen or the tab is hidden.

### D9. Build, CI, and deployment

- `next.config.ts`: `output: "export"`, `basePath: "/gobento"`,
  `images.unoptimized: true`, `trailingSlash: true`. `.nojekyll` goes in the
  export.
- `.github/workflows/website.yml` runs on pull requests touching `website/**`
  or `components/**`:
  - `pnpm install --frozen-lockfile`;
  - lint, type-check, Vitest, `next build`;
  - Playwright e2e with axe in both themes;
  - Lighthouse CI against the export with the budget from the spec.
  - Once the exporter exists, also the catalog freshness check and the CLI
    cross-check.
- `.github/workflows/pages.yml` runs on pushes to `main` touching the same
  paths. It builds and deploys with `actions/configure-pages`,
  `actions/upload-pages-artifact`, and `actions/deploy-pages`, with permissions
  `pages: write` and `id-token: write`. Actions are SHA-pinned, as in
  `establish-repo-standards`.
- Dependabot gets an `npm` entry for `/website`: weekly, grouped, 7-day
  cooldown, `build(deps)` prefix.
- The pre-release notice is driven by `.release-please-manifest.json`: version
  `0.0.0` means unreleased, read at build time.

## Risks / Trade-offs

- [The visual result drifts toward the inspiration site] → The design
  direction is written as our own identity (D6). The maintainer reviews it
  explicitly for originality before approval.
- [Japanese motifs used as clichés] → Motifs are limited to a few authentic
  elements (bento grid, washi texture, restrained patterns), with typography
  and color doing most of the work.
- [The procedural 3D mascot looks worse than a modelled one] → A decision
  point at task 4.2, with GLB as the fallback.
- [three.js bundle size (~150 kB gzipped)] → Loaded lazily after content, and
  only on the home page, guarded by the Lighthouse budget.
- [Catalog logic diverges between Go and TypeScript] → Pre-normalized JSON, a
  freshness check, and the CLI cross-check over every supported selection.
- [GitHub Pages `basePath` bugs (broken assets or links)] → Playwright runs
  against the exported files served under `/gobento/`, not against the dev
  server.
- [Fumadocs and shadcn style conflicts] → A single token source. Visual check
  of the docs page during the design review (task 2.4) and in the e2e
  screenshots.
