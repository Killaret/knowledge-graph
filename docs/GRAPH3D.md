# 3D Graph Visualization

This document describes the SvelteKit/Three.js 3D graph view in the Knowledge Graph frontend.

## Overview

The 3D graph provides an alternative, spatial way to explore the note graph. It renders notes as glowing celestial bodies and links as light rays in a 3D force-directed simulation. The 3D scene is lazy-loaded so the 2D graph bundle is not penalized.

## User-facing behavior

- A **3D** toggle is available in `FloatingControls` alongside **2D** and **List**.
- Selecting **3D** on the home page renders `Graph3DViewer` inside the existing `fullscreen-graph` container.
- The 3D view respects the same `FilterState` (type filter and search) as the 2D graph.
- Clicking a node fires `onNodeClick` and opens the existing `NoteSidePanel` via `graphStore.selectedNodeId`.
- Double-clicking a node focuses the camera on it.
- `/graph/3d` and `/graph/3d/[id]` render the full graph or a centered ego-network.

## Architecture (Feature-Sliced Design)

```
widgets/graph-3d-viewer/
  Graph3DViewer.svelte         # lazy wrapper (loading / error / WebGL check)

features/graph-3d/
  index.ts                     # public API: Graph3DEngine, providers, types, defaults
  config.ts                    # Graph3DRuntimeConfig + toSimulationNodes helper
  model/
    layout-provider.ts         # GraphLayoutProvider adapter (D3 / graph-service)
    types.ts                   # Graph3DConfig, SimulationNode, callbacks
  lib/
    engine.ts                  # orchestration: scene, simulation, managers, loop
    scene.ts                   # Three.js scene, renderer, lights, starfield, fog
    nodes.ts                   # InstancedMesh-based node rendering + selection
    links.ts                   # line-based link rendering
    labels.ts                  # CSS2D labels
    simulation.ts              # d3-force-3d wiring
    camera.ts                  # auto-fit and focus-on-node camera helpers
  ui/
    Graph3DScene.svelte        # Svelte 5 facade over the engine

shared/
  stores/graph.svelte.ts       # cross-view graph state (selectedNodeId, currentView, etc.)
  lib/webgl-detector.ts        # browser WebGL availability check
  lib/performance-monitor.ts   # FPS/frame-time helper
```

### Domain alignment

- Node colors, emissive glow, and emoji come from `CelestialBody`.
- Link colors and weights come from `LinkType`.
- Node/link data flows through `filterValidLinks` from `$shared/utils/graphUtils`.

## Shared graph state

`shared/stores/graph.svelte.ts` owns `selectedNodeId`, `currentView`, `searchQuery`, `selectedType`, `graphData`, and `hoveredNodeId`. Both `routes/+page.svelte`, `routes/graph/3d/[id]/+page.svelte`, and `features/graph-canvas/canvas-state.svelte.ts` proxy through this store, so selection and active view are consistent between 2D and 3D modes.

## Graph API unification ✅

`routes/+page.svelte` now treats `graph-service` as the single source of truth for graph data on the home page:

- `loadData()` fetches notes and graph in parallel with `Promise.all([getNotes(), getGraphWithPreload()])` for authenticated users, or just `getGraphWithPreload()` for anonymous users.
- `getGraphWithPreload()` returns cached data from `PreloadService` when available; otherwise it calls `getFullGraphData()` once and seeds the result via `PreloadService.seedGraph(graphData)`.
- After note mutations, `refreshAfterMutation()` first tries `PreloadService.updateWithDelta()` using `lastHash`; if no hash/preloaded data exists, or the delta call fails, it falls back to `loadData()`.
- The old `loadGraphData()`, sequential `loadDataParallel()`, and defensive "rebuild graph from notes" fallback were removed.

This means the 2D canvas, the 3D viewer, and the list view all consume the same normalized graph payload from `graph-service`.

## Layout providers

`features/graph-3d/model/layout-provider.ts` defines the `GraphLayoutProvider` adapter and two implementations:

- `D3ForceLayoutProvider` — calls the same endpoints as the 2D view and lets the client-side `d3-force-3d` engine seed and refine positions.
- `GraphServiceLayoutProvider` — calls `getFullGraphData` or `getGraphData(..., layout="3d")` to retrieve server-computed 3D positions.

`createLayoutProvider(runtime: Graph3DRuntimeConfig)` selects the active provider at runtime based on `frontend.graph.3d.layout_provider` in `knowledge-graph.config.json`. Routes `/graph/3d` and `/graph/3d/[id]` use `toRuntimeConfig()` + `createLayoutProvider()` instead of hard-coding a provider.

On the home page, `FloatingControls` exposes a **D3 ↔ Graph-service** layout provider toggle when the 3D view is active. The toggle re-fetches graph data through the selected provider and updates the 3D scene.

The backend `graph-service` HTTP server supports `?layout=3d` on `GET /api/v1/graph/note/:id` and invokes `engine.Layout3D` when requested. 2D results are still cached; 3D results bypass the 2D cache key.

### Backend fallback

`$shared/api/graph.ts` always attempts the `graph-service` first. If it returns a 5xx, 408, 429, timeout, or network error, `getFullGraphData`/`getGraphData` transparently fall back to the main backend endpoints (`/api/v1/graph/all` and `/api/v1/notes/:id/graph`). The graph-service HTTP server supports `?layout=3d` on `GET /api/v1/graph/note/:id` and invokes `engine.Layout3D` when requested. 2D results are still cached; 3D results bypass the 2D cache key.

`Graph3DEngine` detects when all nodes already carry `x/y/z` coordinates and shortens `warmStartTicks` from the default 80 to 10, preserving the service layout while still allowing a brief physical relaxation.

## Fog and performance presets

Named fog presets and FPS-based performance presets are implemented and driven by `frontend.graph.3d` in `knowledge-graph.config.json`.

## Fog presets

`features/graph-3d/lib/fog.ts` exposes named fog presets driven by `frontend.graph.3d.fog.presets`:

- `birth` — dense initial fog (`density_initial`) that thins to `density_final` as the simulation stabilizes.
- `nebula` — medium fog.
- `deep-space` — no fog.

`applyFogPreset(scene, presetName, config)` returns `{ initial, final }` so `Graph3DEngine` can smoothly interpolate density as the graph settles. The default preset is configured via `frontend.graph.3d.fog.default_preset`.

## Performance presets

`Graph3DEngine` uses `shared/lib/performance-monitor.ts` to track FPS. When FPS stays below `fps_threshold_low` for `low_fps_sample_count` frames, it:

- switches to a lighter fog preset (`birth` → `nebula` → `deep-space`);
- reduces the starfield particle count (`high` → `medium` → `low`);
- disables `autoRotate`.

When FPS rises above `fps_threshold_high` for the same sample count, quality steps back up. Thresholds and starfield counts are configured in `frontend.graph.3d.performance`.

## Performance notes

- `Graph3DScene.svelte` is loaded only when the 3D view is active (`Graph3DViewer` uses dynamic `import()`).
- `Graph3DViewer` checks `isWebGLAvailable()` before loading the heavy Three.js bundle and surfaces a user-facing error if WebGL is missing.
- Nodes are batched into one `THREE.InstancedMesh` per node type.
- Links are `THREE.Line` objects with shared geometry patterns.
- The animation loop caps at ~30 fps (`frameInterval = 33 ms`) and stops the internal `d3-force-3d` timer once the simulation is stable.
- `max_nodes` is read from `graphConfig3D` (`knowledge-graph.config.json`) and merged into `Graph3DConfig`.
- `Graph3DEngine.dispose()` tears down the renderer, managers, simulation, and animation frame to avoid GPU/CPU leaks when switching views.

## Configuration

Key values live in `frontend/src/features/graph-3d/model/types.ts` and are bounded by `graphConfig3D.max_nodes`:

- `baseNodeScale` — base size multiplier for celestial bodies.
- `labelScale` — size multiplier for CSS2D labels.
- `linkOpacity` — base link transparency.
- `fogDensityInitial` / `fogDensityFinal` — fog transition while the simulation stabilizes.
- `warmStartTicks` — default number of force ticks before the first frame.
- `enableLabels` — show/hide node labels.
- `autoRotate` — enable automatic camera rotation when idle.
- `disableAnimation` — used by tests to run synchronously.

Runtime provider selection:

- `useGraphServiceLayout` — hint whether the service should be used for the initial 3D layout.
- `layoutProvider` — explicit provider name (`"d3"` or `"graph-service"`).

## Testing

- Unit tests:
  - `frontend/src/features/graph-3d/ui/Graph3DScene.spec.ts`
  - `frontend/src/features/graph-3d/config.test.ts`
  - `frontend/src/widgets/graph-3d-viewer/Graph3DViewer.spec.ts`
  - `frontend/src/features/graph-3d/model/layout-provider.test.ts`
  - `frontend/src/features/graph-3d/lib/fog.test.ts`
  - `frontend/src/features/graph-3d/lib/engine.performance.test.ts`
- Mocks:
  - `frontend/src/__mocks__/d3-force-3d.ts` provides a deterministic `d3-force-3d` mock.
  - `vitest-setup.ts` mocks `three` (`WebGLRenderer`), `OrbitControls`, and `CSS2DRenderer`.
- Run: `cd frontend && npm run test:unit`.
- Backend tests: `cd services/graph-service && go test ./...`.

## Frozen archive

The original 3D sources were snapshotted in `docs/3d-archive/` before the refactor. New code follows Svelte 5 runes and current FSD/Atomic Design boundaries, using the archive only as a reference.

## Future improvements

- Named fog and performance presets selected automatically from FPS telemetry.
- GPU picking for large graphs instead of `Raycaster`.
- Spatial/LOD culling for graphs above the `max_nodes` threshold.
- Dashed link patterns and animated link creation.
- Keyboard-accessible node focus and ARIA live regions.
