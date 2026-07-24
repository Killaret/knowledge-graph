# 3D Graph Visualization

This document describes the SvelteKit/Three.js 3D graph view that was unfrozen and re-integrated into the Knowledge Graph frontend.

## Overview

The 3D graph provides an alternative, spatial way to explore the note graph. It renders notes as glowing celestial bodies and links as light rays in a 3D force-directed simulation. The view is lazily loaded so the 2D graph bundle is not penalized.

## User-facing behavior

- A **3D** toggle is available in `FloatingControls` alongside **2D** and **List**.
- Selecting **3D** on the home page renders `GraphCanvas3D` inside the existing `fullscreen-graph` container.
- The 3D view respects the same `FilterState` (type filter and search) as the 2D graph.
- Clicking a node fires `onNodeClick` and opens the existing `NoteSidePanel`.
- `/graph/3d` and `/graph/3d/[id]` are no longer redirects; they render the full or centered 3D graph.

## Architecture

```
components/organisms/GraphCanvas3D.svelte   # lazy wrapper (loading / error states)
components/organisms/Graph3D.svelte         # Svelte 5 facade over the engine
features/graph-canvas/graph3d/              # engine and helpers
  engine.ts      # orchestration: scene, simulation, managers, loop
  scene.ts       # Three.js scene, renderers, lights, starfield
  nodes.ts       # InstancedMesh-based node rendering
  links.ts       # line-based link rendering
  labels.ts      # CSS2D labels
  simulation.ts  # d3-force-3d wiring
  camera.ts      # auto-fit and focus-on-node camera helpers
  types.ts       # config and type aliases
```

## Domain alignment

- Node colors, emissive glow, and emoji come from `CelestialBody`.
- Link colors and weights come from `LinkType`.
- Node/link data flows through `filterValidLinks` from `$shared/utils/graphUtils`.

## Performance notes

- `Graph3D.svelte` is loaded only when the 3D view is active (`GraphCanvas3D` uses dynamic `import()`).
- Nodes are batched into one `THREE.InstancedMesh` per node type.
- Links are `THREE.Line` objects with shared geometry patterns.
- The animation loop caps at ~30 fps and stops the internal `d3-force-3d` timer once stable.
- `max_nodes` is read from `graphConfig3D` (`knowledge-graph.config.json`).

## Configuration

Key values live in `frontend/src/features/graph-canvas/graph3d/types.ts` and are bounded by `graphConfig3D.max_nodes`:

- `baseNodeScale` — base size multiplier for celestial bodies.
- `linkOpacity` — base link transparency.
- `fogDensityInitial` / `fogDensityFinal` — fog transition while the simulation stabilizes.
- `warmStartTicks` — number of force ticks before the first frame.
- `disableAnimation` — used by tests to run synchronously.

## Testing

- Unit tests: `frontend/src/components/organisms/Graph3D.spec.ts`.
- Mocks:
  - `frontend/src/__mocks__/d3-force-3d.ts` provides a deterministic `d3-force-3d` mock.
  - `vitest-setup.ts` mocks `three` (`WebGLRenderer`), `OrbitControls`, and `CSS2DRenderer`.
- Run: `cd frontend && npm run test:unit`.

## Frozen archive

The original 3D sources were snapshotted in `docs/3d-archive/` before the refactor. New code was written to follow Svelte 5 runes and current FSD/Atomic Design boundaries, using the archive only as a reference.

## Future improvements

- GPU picking for large graphs instead of `Raycaster`.
- Spatial/LOD culling for graphs above the `max_nodes` threshold.
- Dashed link patterns and animated link creation.
- Keyboard-accessible node focus and ARIA live regions.
