# Frontend Architecture Documentation

## Overview

The Knowledge Graph frontend is a **graph-centric Svelte 5 application** with a focus on visual interaction and intuitive note management. The architecture prioritizes **2D visualization as the primary interface**, with 3D as an optional enhancement for users with capable devices.

## Core Principles

1. **Graph-First Design**: The graph is the primary interface, not a separate page
2. **SSR-Safe**: All browser APIs are properly guarded to prevent server-side errors
3. **Progressive Enhancement**: 2D works everywhere, 3D is optional
4. **Accessibility**: Keyboard navigation, screen reader support, high contrast modes

## Technology Stack

- **Framework**: Svelte 5 (Runes mode)
- **Language**: TypeScript
- **Styling**: CSS with CSS variables for theming
- **Visualization**: D3-force (2D), Three.js (3D optional)
- **Testing**: Playwright + Cucumber (BDD)
- **Build**: Vite

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Main Page (/)                           │
│  ┌─────────────────────────────────────────────────────┐  │
│  │              GraphCanvas (2D Primary)                │  │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │  │
│  │  │   Nodes     │  │    Links    │  │  Animation  │   │  │
│  │  │  (D3-force) │  │ (D3-force)  │  │ (Canvas 2D) │   │  │
│  │  └─────────────┘  └─────────────┘  └─────────────┘   │  │
│  └─────────────────────────────────────────────────────┘  │
│                           │                                  │
│         ┌─────────────────┼─────────────────┐               │
│         │                 │                 │               │
│    ┌────▼────┐      ┌────▼────┐      ┌────▼────┐          │
│    │Floating │      │  Note   │      │ Create  │          │
│    │Controls │      │SidePanel│      │  Modal  │          │
│    └─────────┘      └─────────┘      └─────────┘          │
└─────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────▼─────────┐
                    │   Optional 3D     │
                    │  (lazy-loaded)    │
                    │   Graph3D.svelte  │
                    └───────────────────┘
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
- Type selector (Star, Planet, Comet, Galaxy)
- Save/Cancel actions

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
/                          → Main graph view (graph-centric)
/?note=:id                 → Main view with specific note selected
/graph/:id                 → Redirect to /?note=:id
/graph/3d/:id              → Optional 3D view page
/notes/:id                 → Redirect to /?note=:id (legacy support)
/notes/:id/edit            → Edit note page (for complex editing)
/search?q=:query          → Search results overlay
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

1. **Offline Support**: Service Worker + IndexedDB
2. **Collaborative Editing**: WebRTC or WebSocket integration
3. **AI-Powered Suggestions**: Integration with NLP service
4. **Mobile App**: Capacitor or PWA packaging
5. **Plugin System**: Third-party visualization plugins
