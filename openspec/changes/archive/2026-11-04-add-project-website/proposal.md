# Proposal

## Why

A CLI is only as approachable as its front door. gobento needs one place
where people understand it in seconds, browse documentation, and assemble the
exact `gobento new` command for the stack they want, without reading flag
references. go-blueprint shows how much a web configurator helps adoption. We
want the same ease, with the warmth of a memorable mascot and a distinct
Japanese bento identity.

## What Changes

- A new website in `website/` (Next.js App Router, static export) published to
  GitHub Pages at `https://tales-of-zuna.github.io/gobento/`.
- **Landing page:** a hero with an animated 3D mascot (an original character
  combining a Go-gopher-inspired figure with a bento box), the value
  proposition, a copyable quick-start command, and paths to the configurator
  and docs.
- **Configurator:** choose architecture, framework, database, data access, Go
  version, and module versions, and get the exact `gobento new` command, with
  incompatible choices disabled and explained. Options come from the same
  component catalog and version rules the CLI uses, so the site can never
  offer something the CLI does not support.
- **Documentation:** MDX docs with static search covering getting started,
  concepts (slots, contracts, traits, version rules), a CLI reference, a
  component reference generated from the catalog, and contributing a
  component.
- **Visual identity:** designed first in Claude Design (moodboard, mascot,
  page designs) and approved by the maintainer before implementation. It
  covers light and dark themes, motion that respects reduced-motion
  preferences, and a non-WebGL fallback.
- **CI and deployment:** website lint, type-check, test, and build on every
  pull request, and deployment to GitHub Pages on pushes to `main`.
  Dependabot covers the website's npm dependencies.

## Capabilities

### New Capabilities
- `project-website`: site-wide behavior: landing page, themes, motion and
  reduced motion, WebGL fallback, accessibility and performance targets,
  mascot attribution, and publishing.
- `website-configurator`: building a valid `gobento new` command from
  catalog-driven choices, compatibility feedback, sharing, and copying.
- `website-docs`: documentation content structure, catalog-generated
  reference pages, search, and edit links.

### Modified Capabilities
<!-- None. The CLI's behavior is unchanged; the site reads the catalog that
     add-generator-foundation defines. -->

## Impact

- New `website/` directory with its own Node toolchain (pnpm). Go contributors
  do not need Node unless they work on the site.
- Reads `components/` (manifests and `versions.yaml`) at build time. Until
  `add-generator-foundation` lands, the site uses a fixture catalog in the same
  format, and switches to the real catalog when it exists.
- Extends `.github/workflows/` (a website CI job and a Pages deploy workflow)
  and `.github/dependabot.yml` (npm ecosystem). Builds on
  `establish-repo-standards`.
- Public assets: an original mascot crediting the Go gopher's designer, Renée
  French (CC BY 4.0), with a statement that gobento is not affiliated with or
  endorsed by the Go project.
- Out of scope: translations (English only for now), versioned docs (until
  1.0), a blog, analytics, and a custom domain.
