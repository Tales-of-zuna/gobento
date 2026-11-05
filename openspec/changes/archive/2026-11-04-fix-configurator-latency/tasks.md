## 1. Static scene on software WebGL

- [x] 1.1 Classify the WebGL renderer in `components/mascot/support.ts` (none, software, hardware) from the probe context's unmasked renderer name; `supportsWebGL` keeps its meaning. Verify: `support.test.ts` covers SwiftShader, llvmpipe, a GPU name, a hidden name, and no context.
- [x] 1.2 Expose the cached classification through `components/mascot/hooks.ts` and make `MascotHero` animate only on hardware WebGL without reduced motion. Verify: `pnpm test` and `pnpm typecheck` pass; in headless Chromium the scene still replaces the still and idle frames stop (no long tasks over 2 s idle).
- [x] 1.3 Update `website/DESIGN.md` where it says when the mascot animates. Verify: the text matches `MascotHero` behavior.

## 2. Stable scene subtree while typing

- [x] 2.1 In `Configurator.tsx`, memoize tray cells on the chosen components and `packed` on pack count and cells; make `PackingStage` a memo component. Verify: `Configurator.test.tsx` passes and typing a name no longer re-renders the stage (React render count in a unit test).

## 3. Regression guards

- [x] 3.1 Add a view-model budget test for every supported selection of the real catalog (average under 10 ms). Verify: `pnpm test` passes and prints no failure.
- [x] 3.2 Add e2e tests on a forced software rasterizer (`tests/software-webgl.spec.ts`): the home scene draws nothing while idle, and with the 3D bento showing, typing 16 characters into the module path draws nothing, takes under 2 s, and shows the path in the command. Verify: they pass after the fix and fail on the old code (draw counts keep rising).
- [x] 3.3 Run `pnpm test`, `pnpm lint`, `pnpm typecheck`, `npx prettier --check .`, and `CI= E2E_PORT=4185 pnpm e2e --workers=4`; record the keyboard-only test's duration and per-keystroke latency before and after. Verify: all pass.

## Workflow follow-up

- `openspec archive fix-configurator-latency -y`, then `openspec validate --specs --strict`.
