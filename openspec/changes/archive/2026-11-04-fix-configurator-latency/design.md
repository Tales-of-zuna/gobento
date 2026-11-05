# Design

## Context

See proposal.md, Why, for the symptom. Hypotheses, ranked before measuring,
and what each measurement showed (static export served locally, Playwright
Chromium 1.64 headless, WebGL renderer
`ANGLE (Google, Vulkan 1.3.0 (SwiftShader Device (Subzero)))`):

| Hypothesis | Measurement | Verdict |
| --- | --- | --- |
| View-model recomputation (compat per option, `selectionProblems`) | vitest timing on the real catalog: `kindGroups` 0.65 ms, `buildCommand` 0.18 ms, `goOptions` 0.14 ms, `encodeState` 0.12 ms, others < 0.04 ms; about 1.2 ms per render in all | Not the cause |
| `history.replaceState` on every keystroke | `encodeState` 0.12 ms; the trace shows no navigation or history cost | Not the cause |
| The 3D scene | Typing 16 characters: 4132 ms with the scene, 109 ms with WebGL disabled, 145 ms with reduced motion. Idle with the scene: 14 long tasks, 2.3 s in a 2 s window. Trace of 2 s idle: about 1.85 s of the main thread in `ProxyMain::BeginMainFrame > Commit > LayerTreeHost::DoUpdateLayers > GLES2::ReadPixels > WaitForGetOffset`; all JavaScript about 45 ms | Root cause |
| The packing animation or motion springs | Gone with the scene alone (WebGL disabled keeps all motion springs) | Not the cause |

With software compositing, Chromium reads the WebGL canvas back every frame,
and the main thread blocks until SwiftShader has drawn it: about 100 ms per
frame for the configurator's 390x260 canvas and 50 ms for the home page's
600x400 one. With `frameloop="always"` a new frame is always queued, so
every input event waits one or more frames. Playwright's `type` sends three
events per character, which matches the measured ~250 ms per character.

The catalog's growth did not cause this; a larger catalog only adds well
under a millisecond per render.

## Goals / Non-Goals

**Goals:** input on the configurator never waits on the scene; the scene
keeps working wherever it works today; a guard that fails if either the
derivations or the scene make typing slow again.

**Non-Goals:** making the scene cheaper to draw on a GPU; changing how or
when the scene loads; restructuring the view-model.

## Decisions

### D1. Detect a software rasterizer and use the static, on-demand pose

The WebGL probe that already runs once per page load also reads the
unmasked renderer name (`WEBGL_debug_renderer_info`) and classifies it as
software when it names a known software rasterizer (SwiftShader, llvmpipe,
softpipe, Microsoft Basic Render Driver, or any renderer calling itself
"Software", such as Apple Software Renderer). On
software WebGL the mascot mounts as before but with `animate` off, the path
reduced motion already takes for a running scene: the static pose and
`frameloop="demand"`, which draws a frame only when something in the scene
changes. A GPU-backed browser, or one that hides the renderer name, behaves
exactly as before.

Alternatives considered:

- *Treat software WebGL like no WebGL (still only).* Simplest for visitors,
  but the e2e suite runs on SwiftShader and checks that the scene replaces
  the still (`mascot.spec.ts`, `configurator.spec.ts`); those checks would
  have to go. Rejected: it weakens tests and drops the 3D bento for these
  visitors entirely.
- *`failIfMajorPerformanceCaveat` probe.* Chromium creates the context on
  SwiftShader anyway (measured), so it detects nothing here. Rejected.
- *Measure frame times and degrade adaptively.* Catches slow GPUs too, but
  needs a sampling window of janky frames before it acts, which is exactly
  when a visitor starts typing, and adds timing-dependent behavior that is
  hard to test. Rejected for now.
- *Lower resolution or a frame-rate cap on software WebGL.* Still blocks the
  main thread on every frame it draws; it only shrinks the lag. Rejected.

### D2. Keep the scene's React subtree stable while typing

`trayCells` is memoized on the chosen components rather than the whole
state, the `packed` tuple on the pack count and cells, and `PackingStage`
is a memo component, so typing a name or module path no longer re-renders
the stage, the mascot wrapper, or the R3F tree. In demand mode a re-render
that touched scene props could queue a frame; this rules that out, and it
saves work on GPU machines too.

Rejected: memoizing compatibility (`kindGroups`, `componentReasons`) on the
selection alone. It measured under 1 ms per render, and the kind errors it
shows come from the full build result, so splitting it would add coupling
for no measurable gain.

### D3. URL sync stays immediate

`history.replaceState` costs about 0.1 ms per keystroke. Debouncing it would
let the address bar, and so a copied link, briefly lag the visible command.
Rejected.

### D4. Regression guards

- A vitest test enumerates every supported selection of the real catalog
  (`src/generated/catalog.json`) and computes the whole view model for each
  (`buildCommand`, `kindGroups`, `goOptions`, `moduleFields`, `trayCells`,
  `fileTree`, `encodeState`), asserting an average under 10 ms per
  selection. The 36 selections measure about 0.15 ms each once warm (1.2 ms
  cold), so the budget holds on slow CI runners and still catches an accidental blow-up, for example when a new
  kind multiplies the selections. The frontend kind being added elsewhere
  is covered automatically.
- An e2e test waits for the 3D bento, then types 16 characters into the
  module path and asserts it takes under 2 s and that the command shows the
  path. Before the fix it took about 4 s; after it, about 0.2 s, which
  leaves room for loaded CI runners.
- Unit tests for the renderer classification.

## Risks / Trade-offs

- [A browser hides the renderer name, so software WebGL goes undetected] →
  It behaves as today; nothing gets worse.
- [A hardware renderer's name matches the software patterns] → It gets the
  static pose, which is the reduced-motion look and stays fully usable. The
  patterns are specific rasterizer names, not broad words.
- [Visitors with software WebGL lose the idle animation] → Intended: the
  animation is what blocked their input. Packing still updates the scene.

## Migration Plan

Website-only change, deployed with the next site publish. Rollback is a
revert of the commit.
