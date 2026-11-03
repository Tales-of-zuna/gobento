# gobento website design

The approved visual direction for the website (OpenSpec change
`add-project-website`, task 2.4). Approved by the maintainer on 2026-11-03.
The designs were made in Claude Design and are kept here as full-canvas renders
(desktop light, desktop dark, mobile), so anyone can see them without an
account:

| Page                         | Render                                               |
| ---------------------------- | ---------------------------------------------------- |
| Home                         | [design/home.jpg](design/home.jpg)                   |
| Configurator                 | [design/configurator.jpg](design/configurator.jpg)   |
| Docs                         | [design/docs.jpg](design/docs.jpg)                   |
| Components and states        | [design/design-system.jpg](design/design-system.jpg) |
| The three directions offered | [design/directions.jpg](design/directions.jpg)       |

> [!NOTE]
> The renders show the approved **layout, components, states and motion
> notes**. Their colors are the original "washi" palette: the site switched to
> **Catppuccin** afterwards (see [Catppuccin palette](#direction-playful-ekiben)
> below), so for colors, `app/tokens.css` and this document are the source of
> truth.

## Direction: "Playful ekiben"

Warm, toy-like and colorful, with a Japanese feeling carried by aesthetics
rather than text:

- **Bento structure.** Content is organised in compartments: trays with rounded
  cells, one color per cell.
- **Soft surfaces.** Calm Catppuccin surfaces with text-colored outlines and
  hard "toy" offset shadows.
- **Patterns.** Seigaiha waves, ichimatsu checks and mizutama dots, drawn in
  CSS, on wide bands only and never behind body text.
- **Catppuccin palette.** [Catppuccin](https://catppuccin.com) (MIT): **Latte**
  for the light theme and **Frappé** for the dark theme, chosen by the maintainer
  on 2026-11-03. The bento roles map onto Catppuccin accents: peach is the
  primary action ("shu"), green is selected and success ("matcha"), blue is the
  mascot and architecture ("sora"), yellow and pink are the food stickers, and
  mauve is the focus ring. Pastel bands are Catppuccin accents mixed into base.
- **English only.** No Japanese characters in the interface. The wordmark is
  plain `gobento`.
- **Light-first.** Latte is the default light theme; Frappé follows the system
  dark preference. Compartments are recessed (darker than the page) in both.

## Tokens

Defined in [`app/tokens.css`](app/tokens.css). The raw Catppuccin colors are
`--ctp-*` (all 26 per flavor); the semantic tokens below point at them.

**Accessibility exception:** several Latte accents (peach, yellow, green, blue,
red) are below WCAG AA 4.5:1 _as text_ on Latte surfaces. Their text-safe tokens
(`--shu-text`, `--tamago-ink`, `--matcha-ink`, `--sora-ink`, `--danger-ink`) are
darkened versions of the same hue (at least 5:1 on mantle). Button text on peach
and green uses Frappé crust. Frappé accents pass as-is.

| Group             | Tokens                                                                         | Use                                                                                                              |
| ----------------- | ------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------- |
| Surfaces          | `--wash`, `--paper`, `--tray`                                                  | page, raised surfaces, compartments                                                                              |
| Text              | `--ink`, `--ink-2`, `--line`                                                   | text and outlines, secondary text, hairlines                                                                     |
| Strong accents    | `--shu`, `--matcha`, `--sora` (+ `--on-*`)                                     | buttons, selected states, display type                                                                           |
| Text-safe accents | `--shu-text`, `--matcha-ink`, `--sora-ink`, `--tamago-ink`, `--danger-ink`     | links, labels, errors (at least 4.5:1)                                                                           |
| Food pastels      | `--shu-soft`, `--tamago-soft`, `--matcha-soft`, `--sora-soft`, `--sakura-soft` | section bands, compartments                                                                                      |
| Illustration      | `--tamago`, `--sakura`, `--nori`                                               | stickers and decoration only, never text                                                                         |
| Recessed          | `--well`                                                                       | command blocks, code snippets, tray frames (Latte text color in light, Frappé mantle in dark)                    |
| Code              | `--code-bg`                                                                    | syntax-highlighted code blocks: Latte base / Frappé mantle, the surfaces Shiki's catppuccin themes are tuned for |
| Shape             | `--r-sm/md/lg/xl`, `--stroke`, `--toy-1/2/3`                                   | radii, outline width, offset shadows                                                                             |
| Kind identity     | `--kind-architecture/framework/database/data`                                  | the same color per component kind everywhere                                                                     |
| Patterns          | `--pat-seigaiha`, `--pat-ichimatsu`, `--pat-mizutama`, `--scallop`             | band backgrounds and scalloped edges                                                                             |

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
