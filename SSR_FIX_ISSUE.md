# GitHub Issue: Fix SSR compatibility for three-forcegraph

## Title
Fix SSR compatibility for three-forcegraph in Graph3D.svelte

---

## Problem

The `Graph3D.svelte` component uses `three-forcegraph` which relies on browser-specific APIs (WebGL, Canvas, Three.js). This causes SSR (Server-Side Rendering) errors in SvelteKit:

```
Error: Cannot find name 'OrbitControls'
Error: Cannot find name 'ThreeForceGraph'
[vite] Error when evaluating SSR module /src/lib/components/Graph3D.svelte: failed to import "three-forcegraph"
```

### Current Code Problems

In `Graph3D.svelte`:
```typescript
import ThreeForceGraph from 'three-forcegraph';  // ❌ SSR fails here
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';  // ❌ SSR fails here
```

### Affected Tests

The following E2E tests are currently **skipped** due to this issue:

**`tests/ui-components.spec.ts`**:
```typescript
// Line 122
test.skip('should show note popup on node click', async ({ page, request }) => {
  // ... test code
});

// Line 147  
test.skip('should show empty state for isolated node', async ({ page, request }) => {
  // ... test code
});
```

**`tests/graph-3d.spec.ts`**: All graph visualization tests are failing

---

## Current Workaround

Currently using `@ts-ignore` comments which suppress TypeScript errors but don't solve the underlying SSR issue:

```typescript
// @ts-ignore - dynamic import
const newControls = new OrbitControls(newCamera, newRenderer.domElement);

// @ts-ignore - dynamic import  
const newGraph = new ThreeForceGraph()
```

---

## Proposed Solution

Use dynamic imports with browser check from `$app/environment`:

```svelte
<script lang="ts">
  import { browser } from '$app/environment';
  import * as THREE from 'three';
  import type { GraphData, GraphNode } from '$lib/api/graph';
  
  // Types for dynamic imports
  type ThreeForceGraphType = any;
  type OrbitControlsType = any;
  
  let { 
    data,
    onNodeClick,
    onNodeRightClick
  }: { 
    data: GraphData;
    onNodeClick?: (node: GraphNode, event: MouseEvent) => void;
    onNodeRightClick?: (node: GraphNode, event: MouseEvent) => void;
  } = $props();
  
  // DOM refs
  let containerRef: HTMLDivElement;
  
  // Three.js refs
  let renderer: THREE.WebGLRenderer | null = null;
  let scene: THREE.Scene | null = null;
  let camera: THREE.PerspectiveCamera | null = null;
  let controls: OrbitControlsType | null = null;
  let graphInstance: ThreeForceGraphType | null = null;
  let animationId: number = 0;
  let resizeObserver: ResizeObserver | null = null;
  
  onMount(async () => {
    if (!browser || !containerRef) return;
    
    // Dynamic imports for SSR compatibility
    const [{ default: ThreeForceGraph }, { OrbitControls }] = await Promise.all([
      import('three-forcegraph'),
      import('three/examples/jsm/controls/OrbitControls.js')
    ]);
    
    // Now initialize graph with imported classes
    const width = containerRef.clientWidth;
    const height = containerRef.clientHeight;
    
    // Renderer
    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    renderer.setSize(width, height);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    containerRef.appendChild(renderer.domElement);
    
    // Scene, Camera, Controls
    scene = new THREE.Scene();
    camera = new THREE.PerspectiveCamera(60, width / height, 0.1, 2000);
    controls = new OrbitControls(camera, renderer.domElement);
    
    // Graph
    graphInstance = new ThreeForceGraph()
      .graphData(data)
      .nodeRelSize(6)
      .nodeColor((node: any) => getColorByType(node.type));
    
    scene.add(graphInstance);
    
    // Animation loop
    const animate = () => {
      animationId = requestAnimationFrame(animate);
      graphInstance.tickFrame();
      controls.update();
      renderer.render(scene, camera);
    };
    animate();
    
    // Resize observer
    resizeObserver = new ResizeObserver((entries) => {
      const { width: w, height: h } = entries[0].contentRect;
      camera.aspect = w / h;
      camera.updateProjectionMatrix();
      renderer.setSize(w, h);
    });
    resizeObserver.observe(containerRef);
    
    // Cleanup
    return () => {
      cancelAnimationFrame(animationId);
      resizeObserver?.disconnect();
      controls?.dispose();
      renderer?.dispose();
      if (renderer?.domElement.parentNode === containerRef) {
        containerRef.removeChild(renderer.domElement);
      }
    };
  });
</script>

<div bind:this={containerRef} class="graph-container">
  {#if !browser}
    <div class="ssr-placeholder">Loading 3D visualization...</div>
  {/if}
</div>
```

---

## Alternative Solutions

### Option 1: SvelteKit `browser` check (Recommended)
Use dynamic imports with `$app/environment` browser check as shown above.

### Option 2: Vite `ssr.noExternal`
Add to `vite.config.ts`:
```typescript
export default {
  ssr: {
    noExternal: ['three-forcegraph', 'three']
  }
}
```

### Option 3: Separate server/browser components
Create `Graph3D.svelte` (browser-only) and `Graph3DSSR.svelte` (placeholder for SSR).

---

## Acceptance Criteria

- [ ] Graph3D.svelte loads without SSR errors
- [ ] Component renders correctly in browser
- [ ] All skipped graph tests are re-enabled and passing
- [ ] No `@ts-ignore` comments needed for imports
- [ ] Build passes without errors (`npm run build`)
- [ ] TypeScript check passes (`npm run check`)

---

## Files to Modify

1. **`frontend/src/lib/components/Graph3D.svelte`**
   - Convert to dynamic imports
   - Add browser check
   - Remove @ts-ignore comments

2. **`frontend/tests/ui-components.spec.ts`**
   - Re-enable skipped tests (remove `.skip()`)

3. **`frontend/tests/graph-3d.spec.ts`**
   - Fix and re-enable all graph tests

---

## Environment

- Svelte 5.54.0 with runes mode
- SvelteKit 2.5.0
- three-forcegraph 1.43.2
- three 0.160.1
- TypeScript 5.9.3
- Vite 5.4.11

## Related Issues/PRs

- Part of UI/UX improvements in branch `feature/ui-ux-improvements`
- Related to: #123 (if there's a related issue)

---

## Labels

`bug`, `frontend`, `ssr`, `three.js`, `testing`, `help wanted`
