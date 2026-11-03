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

Versions verified on 2026-11-03 and pinned exactly:
- Next.js 16.4.4.
- React and React DOM **19.3.1, pinned exactly**, because React Three Fiber
  9.8.2 only supports React below 19.4.
- Fumadocs core/ui 16.17.10 (the Base UI build, `@fumadocs/base-ui`) and
  fumadocs-mdx 15.4.10.
- Tailwind CSS 4.3.4.
- shadcn CLI 4.23.0, initialized with `--base base`, so shadcn and Fumadocs
  share Base UI instead of shipping both Base UI and Radix.
- TypeScript ^5.9. TypeScript 7 support in Next.js is unverified.
- three 0.187, @react-three/fiber 9.8.2, @react-three/drei 10.7.9, and
  motion 14.9.
- pnpm 12.16.1, pinned in `packageManager`. Corepack is not relied on: it is
  no longer bundled from Node 25.

The baseline is Fumadocs' official "Next.js Static" template. One shared set of
CSS variables (design tokens, D6) themes shadcn and Fumadocs through
`fumadocs-ui/css/shadcn.css`, so docs and marketing pages look like one
product.

**Gotcha:** Fumadocs' static search ignores Next.js `basePath`. The client must
call `useStaticSearch({ from: "/gobento/api/search" })`, derived from the same
base-path constant as `next.config.ts`.

### D3. Catalog bridge: Go owns the logic, the site reads JSON

- `catalog.json` follows a schema defined in `website/lib/catalog/schema.ts`
  (zod) and mirrored by a Go struct. It holds:
  - kinds and components (name, title, description, kind);
  - provided and required traits;
  - the Go version range and default;
  - modules per component, with role, range, default, and constraints already
    normalized (each constraint as `{ when, requires: { module|go, range } }`);
  - the selections the CLI supports (`compose.Supported`), one component
    name per kind;
  - the CLI's name and module-path rules as regular expressions, exported
    from `internal/generate`: the project-name pattern with the reserved
    Windows device names, and a module-path pattern that approximates
    `module.CheckImportPath` (it accepts every valid path, but does not
    reject Windows-reserved path elements such as `con` or `abc~1`, which
    `gobento new` still reports; a Go fuzz test holds it to that).
- **Producer:** a small Go program `internal/tools/catalogexport`, which uses
  the real `internal/catalog`, `internal/compat`, and `internal/compose`
  packages, writes `website/src/generated/catalog.json` (`make catalog`). The
  file is committed, so website contributors do not need Go. The Go test
  `TestCommittedCatalogIsCurrent` (run by `go test` in CI) and
  `make catalog-check` re-run the exporter and fail if the file differs, the
  same freshness pattern as sqlc in `add-generator-foundation` (D6). Kind
  titles and descriptions, which the catalog has no text for, live in the
  exporter.
- **Before the generator landed,** `catalog.json` was a hand-written fixture
  of the first slice, validated by the same zod schema; task group 7 replaced
  it with the exporter's output.
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
- **Cross-check test** (spec: *Generated commands are accepted by the CLI*),
  split so that neither toolchain needs the other and the command builder is
  never duplicated: a Vitest suite (`lib/command/cli-commands.test.ts`)
  enumerates every supported selection from the real `catalog.json`, builds
  each command with `lib/command/build.ts` (with the defaults, and with every
  flag explicit at the range floors), and keeps them in the committed file
  `lib/command/supported-commands.txt` as a file snapshot, so a stale file
  fails the suite. The Go test `TestWebsiteCommands` (`internal/cli`) runs
  each command with `--dry-run` and checks that together they cover every
  supported selection. Driving `build.ts` from Go through Node was rejected:
  it would make `go test` depend on Node and `node_modules`.

### D5. Documentation

- Fumadocs serves `content/docs`. Sections follow the spec: Getting started,
  Concepts, CLI reference, Component reference, and Contributing.
- The Component reference is a server component that renders from
  `catalog.json` at build time, so it contains no hand-written version numbers (a test enforces this). Other pages may show clearly labelled illustrative versions.
- Search is Fumadocs' static index, built at export time and queried in the
  browser.
- "Edit this page" links to
  `https://github.com/Tales-of-zuna/gobento/edit/main/website/content/docs/<path>`.
- The existing `docs/components.md` from the generator change moves into
  `content/docs/contributing/adding-a-component.mdx` when both changes have
  landed. A stub at the old path keeps links working.

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
   source of the static stills and favicon (see D7).
3. **Pages:** home, configurator, and a docs page, at desktop and mobile
   sizes, in light and dark themes.

The maintainer approves the direction (task 2.4) before implementation starts.
Approved tokens become CSS variables in `app/globals.css`, consumed by
Tailwind, shadcn, and Fumadocs.

### D7. Mascot implementation

- **3D-first (maintainer's choice):**
  - The mascot is designed directly as a 3D model in code, from primitives
    (rounded boxes for the bento; capsules and spheres for the gopher), with
    toon shading and inverted-hull outlines in direction A's palette.
  - Concepts are prototyped in plain three.js and compared as renders. The
    winner is ported to React Three Fiber.
  - Claude Design is not used to draw the mascot, because its guidelines
    reserve complex illustration for real assets. Page designs show a
    labelled placeholder.
  - This keeps the mascot reviewable as code, animatable (idle bob, blink, lid
    opening on hover), and small, with no binary model pipeline.
  *Alternative:* a GLB modelled in Blender, kept as the fallback if the
  procedural version cannot reach the approved look. That decision is made at
  task 4.2, with the maintainer.
- **Chosen concept:** the maintainer picked concept 1, "In the box" (the
  gopher sunk into the big compartment of an open bento, paws over the rim,
  the lid leaning behind as a backrest, four food compartments beside it), in
  sora blue `#A9C6E6`. Its framing is landscape, 3:2.
- **Static fallback:** rendered stills of the approved 3D model (WebP/PNG at
  1× and 2×, 720×480 and 1440×960), plus a head-only render (the round head
  with ears and sprout, slight three-quarter) at 64, 128, and 256 px. The
  hero stills are used for reduced motion, without WebGL, and while loading;
  the head renders are the favicon (64), the apple-touch icon (128), and the
  avatar or social image (256).
  `pnpm mascot:stills` renders them from the R3F scene itself (Vite plus
  headless Chromium, transparent background) into `public/mascot/`, so the
  stills cannot drift from the 3D model.
- **Attribution:** the footer credits Renée French (CC BY 4.0) for the original
  Go gopher design and states non-affiliation with the Go project. The mascot
  is our own drawing. The Go logo is not used.

### D8. 3D loading and runtime

- The scene is a client component loaded with `next/dynamic` (`ssr: false`).
  The static still (D7) is in the HTML first, sized to avoid layout shift, so
  the scene never blocks content or LCP; the canvas cross-fades in after its
  first frame.
- WebGL support is detected before mounting. On failure, on an error
  boundary catch, or on a lost WebGL context, the still stays.
- `prefers-reduced-motion`: when it is set on load, the still is shown and
  the 3D code is never downloaded (the still is a render of the same static
  pose). If it is turned on while the scene runs, the scene switches to a
  single static pose (`frameloop="demand"`). Motion's reduced-motion setting
  disables decorative transitions.
- Device pixel ratio is clamped to [1, 2]. Rendering pauses when the canvas is
  off-screen or the tab is hidden.
- **3D and motion across the site:** the maintainer wants clean but impressive
  animation used often. This is the motion language:
  - Home: the mascot hero, plus subtle ambient particles.
  - **Configurator:** a 3D bento box that *packs itself*. Each chosen
    component drops into its compartment with a spring, which shows the slot
    model.
  - UI-wide: staggered reveals, spring presses, crossfades on command
    changes, and copy feedback, implemented with Motion.
  - Docs stay calm: an active-indicator slide, TOC highlight, and a search
    dialog scale-in.
- Every 3D scene is its own lazily loaded client chunk, shares one small
  material and asset module, and obeys the same rules as above: SSR off,
  static still first, WebGL fallback, and reduced motion means a static
  frame. The Lighthouse budget (spec *Performance*) is the guardrail that keeps
  "often" from becoming "heavy".

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
  - The catalog freshness check and the Go half of the CLI cross-check run
    with the Go tests in `ci.yml`; Vitest checks the commands file.
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
- [three.js bundle size (~150 kB gzipped), now used on several pages] → Each
  scene is a lazy client chunk, loaded after content and only on pages that use
  it. They share one three.js chunk, so it is cached across pages. Guarded by
  the Lighthouse budget on every page.
- [Catalog logic diverges between Go and TypeScript] → Pre-normalized JSON, a
  freshness check, and the CLI cross-check over every supported selection.
- [GitHub Pages `basePath` bugs (broken assets or links)] → Playwright runs
  against the exported files served under `/gobento/`, not against the dev
  server.
- [Fumadocs and shadcn style conflicts] → A single token source. Visual check
  of the docs page during the design review (task 2.4) and in the e2e
  screenshots.
