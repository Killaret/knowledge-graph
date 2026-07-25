# 3D Graph Visualization

This document describes the SvelteKit/Three.js 3D graph view in the Knowledge Graph frontend.

## Overview

The 3D graph provides an alternative, spatial way to explore the note graph. It renders notes as glowing celestial bodies and links as light rays in a 3D force-directed simulation. The 3D scene is lazy-loaded so the 2D graph bundle is not penalized.

## User-facing behavior

- A **3D** toggle is available in `FloatingControls` alongside **2D** and **List**.
- Selecting **3D** on the home page renders `Graph3DViewer` inside the existing `fullscreen-graph` container.
- The 3D view respects the same `FilterState` (type filter and search) as the 2D graph.
- Clicking a node fires `onNodeClick` and opens the existing `NoteSidePanel`.
- `/graph/3d` and `/graph/3d/[id]` render the full or centered 3D graph.

## Architecture (Feature-Sliced Design)

```
widgets/graph-3d-viewer/Graph3DViewer.svelte  # lazy wrapper (loading / error / WebGL check)
features/graph-3d/
  index.ts                                    # public API exports
  config.ts                                   # runtime config + toSimulationNodes helper
  model/
    layout-provider.ts                        # D3 / graph-service layout adapter
    types.ts                                  # Graph3DConfig, SimulationNode, callbacks
  lib/
    engine.ts                                 # orchestration: scene, simulation, managers, loop
    scene.ts                                  # Three.js scene, renderers, lights, starfield
    nodes.ts                                  # InstancedMesh-based node rendering
    links.ts                                  # line-based link rendering
    labels.ts                                 # CSS2D labels
    simulation.ts                             # d3-force-3d wiring
    camera.ts                                 # auto-fit and focus-on-node camera helpers
  ui/
    Graph3DScene.svelte                       # Svelte 5 facade over the engine
shared/
  stores/graph.svelte.ts                      # cross-view graph state (selectedNodeId, currentView)
  lib/webgl-detector.ts                       # browser WebGL availability check
  lib/performance-monitor.ts                  # FPS/frame-time helper
```

## Domain alignment

- Node colors, emissive glow, and emoji come from `CelestialBody`.
- Link colors and weights come from `LinkType`.
- Node/link data flows through `filterValidLinks` from `$shared/utils/graphUtils`.

## Shared graph state

`shared/stores/graph.svelte.ts` owns `selectedNodeId` and `currentView`. Both `routes/+page.svelte`, `routes/graph/3d/[id]/+page.svelte` and `features/graph-canvas/canvas-state.svelte.ts` proxy through this store, so selection and active view are consistent between 2D and 3D modes.

## Layout providers

`features/graph-3d/model/layout-provider.ts` defines the `GraphLayoutProvider` adapter and two implementations:

- `D3ForceLayoutProvider` — returns API data and lets the client-side `d3-force-3d` engine seed positions.
- `GraphServiceLayoutProvider` — calls `getFullGraphData` or `getGraphData(..., layout="3d")` to retrieve server-computed 3D positions.

The backend `graph-service` HTTP server supports `?layout=3d` on `GET /api/v1/graph/note/:id` and invokes `engine.Layout3D` when requested. 2D results are still cached; 3D results bypass the 2D cache key.

`Graph3DEngine` detects when all nodes already carry x/y/z coordinates and shortens `warmStartTicks` from the default 80 to 10, preserving the service layout while still allowing a brief physical relaxation.

## Performance notes

- `Graph3DScene.svelte` is loaded only when the 3D view is active (`Graph3DViewer` uses dynamic `import()`).
- `Graph3DViewer` checks `isWebGLAvailable()` before loading the heavy Three.js bundle and surfaces a user-facing error if WebGL is missing.
- Nodes are batched into one `THREE.InstancedMesh` per node type.
- Links are `THREE.Line` objects with shared geometry patterns.
- The animation loop caps at ~30 fps and stops the internal `d3-force-3d` timer once stable.
- `max_nodes` is read from `graphConfig3D` (`knowledge-graph.config.json`).
- `shared/lib/performance-monitor.ts` exposes FPS/frame-time tracking that future quality presets can use.

## Configuration

Key values live in `frontend/src/features/graph-3d/model/types.ts` and are bounded by `graphConfig3D.max_nodes`:

- `baseNodeScale` — base size multiplier for celestial bodies.
- `linkOpacity` — base link transparency.
- `fogDensityInitial` / `fogDensityFinal` — fog transition while the simulation stabilizes.
- `warmStartTicks` — default number of force ticks before the first frame.
- `disableAnimation` — used by tests to run synchronously.

## Testing

- Unit tests:
  - `frontend/src/features/graph-3d/ui/Graph3DScene.spec.ts`
  - `frontend/src/widgets/graph-3d-viewer/Graph3DViewer.spec.ts` (todo)
- Mocks:
  - `frontend/src/__mocks__/d3-force-3d.ts` provides a deterministic `d3-force-3d` mock.
  - `vitest-setup.ts` mocks `three` (`WebGLRenderer`), `OrbitControls`, and `CSS2DRenderer`.
- Run: `cd frontend && npm run test:unit`.
- Backend tests: `cd services/graph-service && go test ./...`.

## Frozen archive

The original 3D sources were snapshotted in `docs/3d-archive/` before the refactor. New code follows Svelte 5 runes and current FSD/Atomic Design boundaries, using the archive only as a reference.

## Future improvements

- GPU picking for large graphs instead of `Raycaster`.
- Spatial/LOD culling for graphs above the `max_nodes` threshold.
- Dashed link patterns and animated link creation.
- Keyboard-accessible node focus and ARIA live regions.
