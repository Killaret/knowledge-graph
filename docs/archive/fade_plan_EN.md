# fade_plan — Restoring "Fog Veil" Effect for 2D GraphCanvas

TL;DR — add opacity for nodes and links in `GraphCanvas`, start animation 0→1 via `requestAnimationFrame`, tying progress to the proportion of stabilized nodes (like in 3D). Add tests and ensure existing tests pass.

## Brief Steps

1. Add `Map` for storing opacity: `nodeOpacity = new Map()` and `linkOpacity = new Map()`, initialize when loading data.

2. In `startSimulation` (`frontend/src/components/organisms/GraphCanvas/simulation.ts`) on start:
   - Start RAF-loop and subscribe to `simulation.on('tick')`.
   - Calculate proportion of stabilized nodes → `progress`.
   - Calculate `targetOpacity = ease(progress)` and interpolate values in `nodeOpacity`/`linkOpacity` to `targetOpacity`.
   - On simulation completion, finish to `1` over 2400 ms.

3. In the renderer (`frontend/src/components/organisms/GraphCanvas/renderer.ts`) consider opacity from `Map` when drawing (`ctx.globalAlpha` or `rgba(...)`) and trigger redraw on changes.

4. Control RAF: store id and cancel via `cancelAnimationFrame` on stop/unmount.

5. Tests: add `GraphCanvas.fade.spec.ts` — mock simulation, simulate ticks, check opacity growth and final finish to ~1.

6. Validation: run tests and perform visual verification.

## Relevant Files

- `frontend/src/components/organisms/GraphCanvas.svelte`
- `frontend/src/components/organisms/GraphCanvas/simulation.ts`
- `frontend/src/components/organisms/GraphCanvas/renderer.ts`
- `frontend/src/components/organisms/Graph3D.svelte` (3D frozen/removed, reference)
- `frontend/src/components/organisms/GraphCanvas.rendering.spec.ts`
- `frontend/src/components/organisms/GraphCanvas.fade.spec.ts`

## Status

- ✅ Implemented: 2D fog veil effect now smoothly progresses from `0` to `1` along with node stabilization.
- ✅ Tests: added `frontend/src/components/organisms/GraphCanvas.fade.spec.ts`, asserting initial zero opacity, progression, and final approach to `1`.
- ✅ Validation: target test suite `GraphCanvas` passed successfully.
