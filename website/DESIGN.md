# gobento website design

The approved visual direction for the website (OpenSpec change
`add-project-website`, task 2.4). Approved by the maintainer on 2026-11-03.
Designs live in the Claude Design project **gobento website**:

| Page                  | File                                                                                                        |
| --------------------- | ----------------------------------------------------------------------------------------------------------- |
| Home                  | [Home v2](https://claude.ai/design/p/4e0480c1-c7ec-40d2-8f4e-f8bcc42c3752?file=Home+v2.dc.html)             |
| Configurator          | [Configurator](https://claude.ai/design/p/4e0480c1-c7ec-40d2-8f4e-f8bcc42c3752?file=Configurator.dc.html)   |
| Docs                  | [Docs v2](https://claude.ai/design/p/4e0480c1-c7ec-40d2-8f4e-f8bcc42c3752?file=Docs+v2.dc.html)             |
| Components and states | [Design System](https://claude.ai/design/p/4e0480c1-c7ec-40d2-8f4e-f8bcc42c3752?file=Design+System.dc.html) |

## Direction: "Playful ekiben"

Warm, toy-like and colorful, with a Japanese feeling carried by aesthetics
rather than text:

- **Bento structure.** Content is organised in compartments: trays with rounded
  cells, one color per cell.
- **Washi warmth.** Off-white paper background, sumi-ink text and outlines,
  hard "toy" offset shadows.
- **Patterns.** Seigaiha waves, ichimatsu checks and mizutama dots, drawn in
  CSS, on wide bands only and never behind body text.
- **Food palette.** Umeboshi, tamagoyaki, edamame, sky and sakura pastels for
  section bands and compartments, with vermilion and matcha as strong accents.
- **English only.** No Japanese characters in the interface. The wordmark is
  plain `gobento`.
- **Light-first.** A warm "evening bento" dark theme is derived from the light
  one: cocoa-brown surfaces (page lightness about 0.28, not near-black), lightened
  on the maintainer's request on 2026-11-03.

## Tokens

Defined in [`app/tokens.css`](app/tokens.css), with the same names as the
design files.

| Group             | Tokens                                                                         | Use                                                                           |
| ----------------- | ------------------------------------------------------------------------------ | ----------------------------------------------------------------------------- |
| Surfaces          | `--wash`, `--paper`, `--tray`                                                  | page, raised surfaces, compartments                                           |
| Text              | `--ink`, `--ink-2`, `--line`                                                   | text and outlines, secondary text, hairlines                                  |
| Strong accents    | `--shu`, `--matcha`, `--sora` (+ `--on-*`)                                     | buttons, selected states, display type                                        |
| Text-safe accents | `--shu-text`, `--matcha-ink`, `--sora-ink`, `--tamago-ink`, `--danger-ink`     | links, labels, errors (at least 4.5:1)                                        |
| Food pastels      | `--shu-soft`, `--tamago-soft`, `--matcha-soft`, `--sora-soft`, `--sakura-soft` | section bands, compartments                                                   |
| Illustration      | `--tamago`, `--sakura`, `--nori`                                               | stickers and decoration only, never text                                      |
| Recessed          | `--well`                                                                       | command blocks, code snippets, tray frames (ink in light, deep cocoa in dark) |
| Shape             | `--r-sm/md/lg/xl`, `--stroke`, `--toy-1/2/3`                                   | radii, outline width, offset shadows                                          |
| Kind identity     | `--kind-architecture/framework/database/data`                                  | the same color per component kind everywhere                                  |
| Patterns          | `--pat-seigaiha`, `--pat-ichimatsu`, `--pat-mizutama`, `--scallop`             | band backgrounds and scalloped edges                                          |

Type: **Zen Maru Gothic** (display), **Figtree** (text), **JetBrains Mono**
(code, ligatures off). All three are on Google Fonts, under the SIL Open Font
License.

Font loading (`app/layout.tsx`), all with `font-display: swap`:

- **Zen Maru Gothic** is a Japanese font: Google Fonts serves it as about 120
  slices per weight, and `next/font/google` preloaded every slice (about 4 MB).
  The site shows no Japanese text, so only its Latin slice (weights 500, 700,
  900, about 11 KB each) is self-hosted in `app/fonts/` with
  `next/font/local`, next to its license (`OFL.txt`). Text outside that slice
  falls back to the system's rounded sans-serif.
- **Figtree** loads its `latin` subset from `next/font/google`, preloaded.
- **JetBrains Mono** loads its `latin` subset from `next/font/google`, not
  preloaded: code is never the first thing on screen.

Component-kind colors: Architecture is sora, HTTP framework is shu, Database
is tamago, and Data access is matcha.

## Mascot

The official mascot is **concept 1, "In the box"**. It is an original
gopher-inspired character sitting in an open red-lacquer bento box, with the
lid behind it as a backrest and four food compartments beside it, in sora blue.

- Built in code with React Three Fiber (`components/mascot/`), using toon
  shading and outlines.
- Rendered stills are the fallback for reduced motion, missing WebGL, and the
  loading state.
- A head-only render is used at small sizes: the favicon and avatar.
- Its four compartments are the configurator's packing animation (`packed`).

Attribution, shown in the site footer: _"The gobento mascot is an original
character inspired by the Go gopher, designed by Renée French and licensed
under CC BY 4.0. gobento is not affiliated with or endorsed by the Go
project."_ The Go logo is not used.

## Motion

Clean but frequent, and always disabled under `prefers-reduced-motion`. The
final state is the base style.

| Area         | Motion                                                                                                                                                                                                          |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Home         | Staggered hero reveal, floating food stickers, a conveyor-belt plate loop that pauses on hover, compartments popping in on scroll, the CI stamp press. The 3D mascot waves, blinks and turns toward the cursor. |
| Configurator | A spring press on selection, a crossfade when the command changes, copy feedback. The 3D bento packs a compartment for each choice.                                                                             |
| Docs         | Calm: an active-indicator slide, a TOC highlight, a search dialog scale-in.                                                                                                                                     |
| Everywhere   | Hover lift with the toy shadow; press-in on click.                                                                                                                                                              |

## Accessibility and quality bar

- WCAG 2.2 AA: text at least 4.5:1, a visible 3px focus ring, keyboard
  operable, hit targets at least 44px on mobile.
- Patterns and stickers are decorative (`aria-hidden`). The 3D canvas has a
  text alternative.
- The Lighthouse mobile budget is at least 90 for performance and
  accessibility. 3D loads lazily, after content, on the visitor's first
  interaction; until then the mascot still holds its place.
- The home hero's entrance only moves its text (no fade), so the heading is
  painted at once and is the page's largest contentful paint.
