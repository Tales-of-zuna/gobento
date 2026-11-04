# Proposal

## Why

Typing on the configurator lags badly once the 3D bento is running. In
headless Chromium, `page.keyboard.type('example.com/shop')` (16 characters)
into the module path takes 4.1 s, against 0.11 s with WebGL disabled and
0.15 s with reduced motion. Pressing Enter on "Advanced" or "Copy command"
takes 3-5 s, and the keyboard-only e2e test needs about 30 s against a 30 s
timeout. A Chrome trace of 2 s of idle time with the scene mounted spends
about 1.85 s of the renderer main thread in `Commit > GLES2::ReadPixels`,
waiting for the software rasterizer (SwiftShader) to draw each frame, while
all JavaScript, React included, takes about 45 ms. The view-model derivations
for the real catalog take about 1.2 ms per render in total. The cause is the
continuously animated scene on software-rendered WebGL, not the
configurator's own logic. Visitors whose browser falls back to software WebGL
(no GPU, a blocklisted driver, a VM or remote desktop) get the same lag on
the configurator and the home page.

## What Changes

- When the browser's WebGL is software-rendered, the mascot scene still
  loads but holds its static pose and draws only when something changes, as
  it already does under reduced motion. With a GPU nothing changes.
- The configurator stops re-rendering the packing stage and the mascot
  scene for edits that do not change the tray, such as typing a project
  name or module path.
- Regression guards: a unit test that bounds the view-model's cost for every
  supported selection in the real catalog, and an e2e test that typing 16
  characters with the 3D bento running takes under 2 s.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `website-configurator`: adds a "Responsive input" requirement.
- `project-website`: adds a "Software-rendered WebGL" requirement for the
  mascot scene.

## Impact

- Code: `website/components/mascot/support.ts`, `hooks.ts`,
  `MascotHero.tsx`; `website/components/configurator/Configurator.tsx` and
  `PackingStage.tsx`; tests in `support.test.ts`, a new view-model budget
  test, and `website/tests/configurator.spec.ts`.
- No dependency, catalog, URL format, or CLI change.
- Documentation: `website/DESIGN.md` (the motion and quality-bar notes on
  when the mascot animates). No README, CONTRIBUTING, `docs/`, or
  `website/content/docs` page describes mascot motion or configurator
  performance, so none of them change.
