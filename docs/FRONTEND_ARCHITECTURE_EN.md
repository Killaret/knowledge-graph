# Frontend Architecture Documentation

## Overview

The Knowledge Graph frontend is a **note-centric Svelte 5 application** with graph visualization as a secondary view. The architecture supports both 2D and 3D visualization modes, with **3D Progressive Rendering (Fog of War)** as the primary graph interface for modern browsers.

### Key Features
- **Progressive Graph Loading**: 3D graph loads nodes incrementally with animated "fog of war" effect
- **Three.js Modular Architecture**: Organized core/simulation/rendering/camera modules
- **Device-Adaptive**: Automatic quality adjustment based on GPU capabilities
- **SSR-Safe**: All browser APIs properly guarded for server-side rendering

## Core Principles

1. **Graph-First Design**: The graph is the primary interface, not a separate page
2. **SSR-Safe**: All browser APIs are properly guarded to prevent server-side errors
3. **Progressive Enhancement**: 2D works everywhere, 3D is optional
4. **Accessibility**: Keyboard navigation, screen reader support, high contrast modes

## Technology Stack

- **Framework**: Svelte 5 (Runes mode)
- **Language**: TypeScript
- **Styling**: CSS with CSS variables for theming
- **Visualization**: 
  - **3D (Primary)**: Three.js + d3-force-3d physics
  - **2D (Fallback)**: D3-force + Canvas API
- **Testing**: Playwright + Cucumber (BDD)
- **Build**: Vite

## Three.js Module Architecture

The 3D graph visualization is organized in modular layers:

```
src/lib/three/
├── core/
│   ├── sceneSetup.ts         # Scene, camera, renderer initialization
│   └── controls.ts           # OrbitControls, mouse interactions
├── simulation/
│   └── forceSimulation.ts    # d3-force-3d physics engine integration
├── rendering/
│   ├── objectManager.ts      # Node/link mesh management
│   ├── nodeFactory.ts        # 3D node geometry creation
│   ├── linkFactory.ts        # Link line geometry creation
│   └── labelFactory.ts       # Text label management
└── camera/
    └── cameraUtils.ts        # Animation, zoom, positioning utilities
```

### Progressive Rendering (Fog of War)

**Concept**: Nodes gradually appear from "fog" with camera animation

```typescript
// Initialization
sceneSetup.init(scene, camera, renderer)
  └── scene.fog = new THREE.FogExp2(0x000000, 0.02)  // Dense fog

// Progressive loading
forceSimulation.addNodesToSimulation(nodes, incremental = true)
  └── Animate nodes from camera position to final position
  └── Fade in opacity over time
  └── Reduce fog density as more nodes appear

// Camera controls
cameraUtils.lerpCamera(targetPosition, duration)
cameraUtils.autoZoomToFit(nodes, padding)
```

**Key Functions**:
- `addNodesToSimulation()`: Adds nodes incrementally with animation
- `lerpCamera()`: Smooth camera transitions
- `animateFogDensity()`: Gradually clears fog during loading

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│  App Shell (/*) - All Pages                                      │
│  ┌─────────┬──────────────────────────────────────────────────┐  │
│  │Sidebar  │              Main Content Area                    │  │
│  │(CCC)    │                                                  │  │
│  │[v2.0]   │  ┌───────────────────────────────────────────┐  │  │
│  │width:0  │  │           Main Page (/)                    │  │  │
│  │(hidden) │  │                                            │  │  │
│  │         │  │  ┌─────────────────────────────────────┐   │  │  │
│  │         │  │  │      NoteListView (Primary)         │   │  │  │
│  │         │  │  │   Grid of note cards with search    │   │  │  │
│  │         │  │  └─────────────────────────────────────┘   │  │  │
│  │         │  │         │                                   │  │  │
│  │         │  │    ┌────┴────┐        ┌────┴────┐          │  │  │
│  │         │  │    │Floating │        │  Note   │          │  │  │
│  │         │  │    │Controls │        │SidePanel│          │  │  │
│  │         │  │    └─────────┘        └─────────┘          │  │  │
│  │         │  └───────────────────────────────────────────┘  │  │
│  │         │                                                  │  │
│  │         │  ┌─────────────────┐  ┌─────────────────────┐   │  │
│  │         │  │ 2D Graph View   │  │  3D Graph View      │   │  │
│  │         │  │ /graph/:id      │  │ /graph/3d/:id       │   │  │
│  │         │  │                 │  │                     │   │  │
│  │         │  │ GraphCanvas     │  │ Graph3D.svelte      │   │  │
│  │         │  └─────────────────┘  └─────────────────────┘   │  │
│  └─────────┴──────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## Component Hierarchy

### 1. GraphCanvas.svelte (Core Component)

The main 2D graph visualization component with the following features:

**Props**:
```typescript
{
  nodes: Array<{ id: string; title: string; type?: string }>;
  links: Array<{ source: string; target: string; weight?: number }>;
  selectedNodeId: string | null;
  highlightedNodeId: string | null;
}
```

**Events**:
- `nodeClick`: User clicked on a node
- `nodeDragStart`: User started dragging from a node (for link creation)
- `nodeDragEnd`: User released drag (contains sourceId and optional targetId)
- `canvasClick`: User clicked on empty canvas

**Public Methods**:
- `focusNode(nodeId: string)`: Animate zoom to node with highlight
- `addNode(node)`: Add new node to simulation
- `removeNode(nodeId)`: Remove node from simulation
- `addLink(link)`: Add new link to simulation

**SSR Safety**:
```typescript
onMount(() => {
  if (!browser) return;
  // Dynamic import of d3-force
  import('d3-force').then(d3 => {
    // Initialize simulation
  });
});
```

### 2. FloatingControls.svelte

Persistent floating control panel with:
- Create note button (+)
- Search button/overlay
- View toggle (2D/3D)
- Import/Export menu

### 3. NoteSidePanel.svelte

Slide-out panel for note details:
- Title and content display
- Edit/Delete actions
- Related notes section
- Metadata display

### 4. CreateNoteModal.svelte

Modal for creating new notes:
- Title input
- Content textarea (with wiki link support `[[Link]]`)
- Type selector (star, planet, moon, comet, galaxy, nebula, asteroid, satellite, blackhole, unknown)
- Save/Cancel actions

**Node Type Display Logic**:
- If node has a type → display that type
- If node has no type → display as 'unknown' (question mark in dashed circle)
- The 'unknown' type represents a conditional/indeterminate object of any shape

### 5. ConfirmModal.svelte

Reusable confirmation dialog:
- Title and message
- Confirm/Cancel buttons
- Danger mode (red styling for destructive actions)

### 6. SmartGraph.svelte

Smart component that decides between 2D and 3D:
- Detects device capabilities
- Checks WebGL support
- Respects user preference via URL param (`?force3d=1`)
- Falls back to 2D on low-end devices

### 7. Sidebar.svelte (Context Control Center) 🆕

Reserved component for the future **Context Control Center (CCC)** - a left navigation panel for advanced filtering and graph navigation.

**Status**: Reserved/Stub (width: 0, hidden)

**Planned Modules (v2.0)**:
1. **Note Groups (Projects/Folders)** - Manual grouping with drag-and-drop
2. **Dynamic Clusters** - Auto-generated groups by semantic similarity
3. **Advanced Filters** - Type, tags, keywords, custom labels with AND/OR logic
4. **Saved Searches** - Bookmarked filter combinations

**Current Implementation**:
- Skeleton component present in all pages via `+layout.svelte`
- App-shell flex layout ready for 280px panel activation
- Zero visual impact until enabled

**Activation** (future):
```css
/* In Sidebar.svelte */
.sidebar-placeholder {
  width: 280px;  /* Change from 0 */
  padding: 1rem;
}
```

### 8. Graph3D.svelte (3D Visualization)

Three.js-based 3D graph visualization with progressive loading:

**Props**:
```typescript
{
  noteId: string;              // Central note ID to build graph from
  loadDepth?: number;         // Graph traversal depth (default: 2)
  performanceMode?: boolean;  // Reduce quality for low-end devices
}
```

**Features**:
- **Progressive Rendering**: Nodes animate in from "fog" with staggered timing
- **Fog of War**: Dense fog initially obscures distant nodes, clears as graph loads
- **Auto-zoom**: Camera automatically positions to fit all nodes
- **Camera Animation**: Smooth transitions between views
- **Stats Bar**: Shows loading progress and node count

**Three.js Integration**:
```typescript
// Core modules
import { setupScene } from '$lib/three/core/sceneSetup';
import { createSimulation } from '$lib/three/simulation/forceSimulation';
import { ObjectManager } from '$lib/three/rendering/objectManager';
import { lerpCamera, autoZoomToFit } from '$lib/three/camera/cameraUtils';

// Progressive loading flow
onMount(async () => {
  const { scene, camera, renderer } = setupScene(container);
  const simulation = createSimulation();
  const objects = new ObjectManager(scene);
  
  // Load graph data
  const graphData = await fetchGraphData(noteId, loadDepth);
  
  // Progressive loading with fog animation
  await loadNodesProgressively(graphData.nodes, {
    batchSize: 5,
    delay: 200,
    onBatch: (nodes) => {
      objects.addNodes(nodes);
      simulation.addNodes(nodes);
      animateFogClearing();
    }
  });
  
  // Auto-position camera
  autoZoomToFit(camera, objects.getNodes(), { padding: 50 });
});
```

**Public Methods**:
- `resetCamera()`: Reset to initial position with animation
- `toggleAutoRotation()`: Enable/disable automatic camera rotation
- `exportView()`: Export current camera position as shareable URL

## State Management

### Graph Store (`$lib/stores/graph.ts`)

```typescript
interface GraphState {
  nodes: Node[];
  links: Link[];
  selectedNodeId: string | null;
  viewMode: '2d' | '3d';
  transform: { x: number; y: number; k: number };
}

export const graphStore = writable<GraphState>({
  nodes: [],
  links: [],
  selectedNodeId: null,
  viewMode: '2d',
  transform: { x: 0, y: 0, k: 1 }
});
```

## Routing Structure

```
/                           → Main page with note list (note-centric)
/?search=:query             → Main page with search active
/notes/:id                  → Note detail page
/notes/:id/edit             → Note edit page
/notes/create               → Create note page (alternative to modal)
/graph/:id                  → 2D graph view for note
/graph/3d/:id               → 3D graph view with progressive rendering
/search?q=:query            → Search results page (redirects to /?search=)
```

### Route Details

**Main Page (/)**
- Note list with search/filter
- Floating controls (create, search)
- Entry point for application

**Note Detail (/notes/:id)**
- Full note view with content
- Actions: edit, delete, view graph
- Back button to main page

**Graph Views**
- `/graph/:id` - 2D force-directed graph using D3
- `/graph/3d/:id` - 3D graph with Three.js progressive rendering

### Navigation Flow

```
Main Page (/) 
    ├── Click Note Card → /notes/:id
    │   ├── Click "View Graph" → /graph/3d/:id
    │   └── Click "Edit" → /notes/:id/edit
    ├── Click "Create" → CreateNoteModal or /notes/create
    └── Search → Updates URL to /?search=:query

Graph Page (/graph/3d/:id)
    ├── Click Node → Navigate to /notes/:nodeId
    └── Click "Back" → Returns to previous page
```

## SSR Error Prevention

All components follow this pattern:

```typescript
import { browser } from '$app/environment';
import { onMount } from 'svelte';

// For dynamic browser-only imports
onMount(() => {
  if (!browser) return;
  // Browser-only code here
});

// For conditional browser checks
$effect(() => {
  if (browser) {
    // Access window, document, localStorage, etc.
  }
});
```

### Common Patterns

1. **Window/Document Access**:
```typescript
$effect(() => {
  if (browser) {
    const url = new URL(window.location.href);
    // ...
  }
});
```

2. **Dynamic Library Imports**:
```typescript
const d3 = await import('d3-force');
const THREE = await import('three');
```

3. **localStorage**:
```typescript
function savePreference(key: string, value: any) {
  if (browser) {
    localStorage.setItem(key, JSON.stringify(value));
  }
}
```

4. **confirm/alert/prompt**:
```typescript
function handleDelete() {
  if (!browser) return;
  if (confirm('Delete?')) {
    // ...
  }
}
```

## BDD Testing with Cucumber

### Feature Files Location
```
tests/features/
├── graph_navigation.feature      # Graph interaction scenarios
├── import_export.feature          # Import/export functionality
├── graph_view.feature             # 2D/3D view modes
├── note_management.feature        # CRUD operations
└── search_and_discovery.feature   # Search functionality
```

### Step Definitions
```
tests/features/step_definitions/
├── graph_steps.ts                 # Graph-specific steps
├── navigation_steps.ts            # Navigation steps
└── common_steps.ts                # Shared steps
```

### Running Tests

```bash
# Run Playwright tests
npm run test

# Run Cucumber BDD tests
npm run test:cucumber

# Run all tests
npm run test:all

# Run with specific tags
CUCUMBER_TAGS="@smoke" npm run test:cucumber
```

## Performance Optimizations

1. **Lazy Loading**: 3D component is dynamically imported only when needed
2. **Canvas Rendering**: 2D graph uses Canvas API for smooth 60fps animation
3. **Device Detection**: Automatic quality adjustment based on GPU capabilities
4. **Virtual Scrolling**: For large note lists in list view
5. **Debounced Search**: 500ms debounce on search input

## Accessibility (a11y)

- Semantic HTML structure
- ARIA labels on interactive elements
- Keyboard navigation support
- Focus management in modals
- Color contrast compliance (WCAG 2.1 AA)
- Reduced motion support (`prefers-reduced-motion`)

## Build and Deployment

```bash
# Development
npm run dev

# Production build
npm run build

# Preview production build
npm run preview

# Type checking
npm run check

# Linting
npm run lint
```

## Future Enhancements

### Short Term (v1.1)
1. **Graph Export**: Export 3D graph view as image/video
2. **Node Grouping**: Group related notes visually in graph
3. **Graph Filters**: Filter by link type, weight threshold
4. **Keyboard Shortcuts**: Full keyboard navigation for graph

### Medium Term (v1.2)
5. **Offline Support**: Service Worker + IndexedDB for note caching
6. **AI-Powered Suggestions**: Inline note recommendations via NLP service
7. **Mobile App**: Capacitor packaging for iOS/Android
8. **Plugin System**: Third-party visualization plugins

### Long Term (v2.0)
9. **Context Control Center**: Left sidebar with groups, clusters, filters, saved searches
10. **Collaborative Editing**: WebRTC or WebSocket for real-time collaboration
11. **Version History**: Note versioning with diff view
12. **Advanced Analytics**: Usage statistics, most connected notes
13. **Voice Input**: Speech-to-text for note creation
