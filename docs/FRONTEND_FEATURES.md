# Frontend Features

## Double-Tap Zoom on Graph Canvas

### Overview

The graph canvas supports double-tap zoom functionality for mobile and touch devices, providing intuitive zoom interactions similar to map applications.

### How It Works

1. **First Double-Tap**: Zoom in 2x at the touch point
   - Scale factor: `transform.k *= 2`
   - Centers zoom at touch coordinates
   - Maintains relative positioning

2. **Second Double-Tap**: Reset to default view
   - Calls `resetView()` function
   - Returns to optimal zoom level for all nodes
   - Resets transform to default state

### Technical Implementation

**Location:** `frontend/src/components/organisms/GraphCanvas.svelte`

**State Variables:**
```typescript
let lastTouchTime = 0;
let lastTouchPos = { x: 0, y: 0 };
let tapCount = 0;
```

**Touch Detection Logic:**
```typescript
function handleTouchStart(e: TouchEvent) {
  if (!browser) return; // Skip in SSR

  if (e.touches.length === 1) {
    const now = Date.now();
    const touch = e.touches[0];
    const dx = Math.abs(touch.clientX - lastTouchPos.x);
    const dy = Math.abs(touch.clientY - lastTouchPos.y);

    // Double-tap detection: <300ms interval, <30px distance
    if (now - lastTouchTime < 300 && dx < 30 && dy < 30) {
      tapCount++;
      handleDoubleTap(touch.clientX, touch.clientY);
      e.preventDefault();
    } else {
      tapCount = 0;
    }

    lastTouchTime = now;
    lastTouchPos = { x: touch.clientX, y: touch.clientY };
  }
}
```

**Zoom Logic:**
```typescript
function handleDoubleTap(clientX: number, clientY: number) {
  const rect = canvas.getBoundingClientRect();
  const x = (clientX - rect.left - transform.x) / transform.k;
  const y = (clientY - rect.top - transform.y) / transform.k;

  if (tapCount === 1) {
    // Zoom in 2x
    const newScale = transform.k * 2;
    const centerX = x * newScale;
    const centerY = y * newScale;

    transform.x = clientX - rect.left - centerX;
    transform.y = clientY - rect.top - centerY;
    transform.k = newScale;
  } else if (tapCount === 2) {
    // Reset zoom
    const simNodes = getSimulationNodes(simState);
    if (ctx && simNodes.length > 0) {
      resetView(ctx, width, height, simNodes, transform);
      tapCount = 0;
    }
  }
}
```

### Canvas Integration

```html
<canvas
  bind:this={canvas}
  onmousedown={onPanStart}
  onmousemove={onPanMove}
  onmouseup={onPanEnd}
  onclick={onClick}
  ontouchstart={handleTouchStart}
  onwheel={onZoom}
/>
```

### Testing

**Unit Tests:** `frontend/src/components/organisms/GraphCanvas.interactions.spec.ts`
- Verifies touch handler is attached to canvas
- Confirms canvas element is rendered correctly

**Manual Testing:**
1. Open graph page on mobile device or browser DevTools touch emulation
2. Tap twice quickly (<300ms) on the same area (<30px apart)
3. Verify zoom in 2x at touch point
4. Tap twice again quickly
5. Verify zoom resets to default view

### Responsive Design Improvements

### Global CSS Changes

**Location:** `frontend/src/shared/styles/global.css`

Added global box-sizing for better responsive behavior:
```css
*, *::before, *::after {
  box-sizing: border-box;
}
```

### Graph Page Layout

**Location:** `frontend/src/routes/graph/+page.svelte`

**Changes:**
- Changed `height: 100vh` to `min-height: 100%` for `.graph-page`
- Changed `height: 100%` to `min-height: 400px` for `.graph-container`
- Added media queries for tablet breakpoints

**Media Queries:**
```css
@media (min-width: 768px) {
  .graph-page {
    padding: 1.5rem;
  }
}

@media (max-width: 768px) {
  .graph-container {
    min-height: 400px;
  }
}
```

### Note Side Panel

**Location:** `frontend/src/components/organisms/NoteSidePanel.svelte`

**Changes:**
- Changed `height: 100vh` to `max-height: 100vh`
- Prevents overflow issues on smaller screens

### Benefits

1. **Better Mobile Experience**
   - Touch-friendly zoom interaction
   - Responsive layout adaptation
   - Improved readability on small screens

2. **Consistent Sizing**
   - Box-sizing prevents layout shifts
   - Minimum heights ensure usability
   - Flexible container behavior

3. **Cross-Device Compatibility**
   - Works on touch devices
   - Maintains desktop functionality
   - Graceful degradation

### Browser Compatibility

- **Touch Events:** Supported in modern browsers (Chrome, Firefox, Safari, Edge)
- **Fallback:** Standard mouse wheel zoom continues to work
- **SSR Safe:** Touch handler skips during server-side rendering

### Performance Considerations

- Touch detection uses simple time/distance thresholds
- No expensive calculations during detection
- Zoom transformation uses existing canvas transform system
- Minimal performance impact on rendering

### Future Enhancements

- Pinch-to-zoom gesture support
- Zoom limits (min/max scale)
- Zoom animation smoothing
- Haptic feedback on zoom
- Customizable zoom sensitivity

---

## Carbon (Allotropic) Theme

### Overview

The application uses a dark "allotropic carbon" visual language: deep carbon surfaces, subtle graphite gradients, and neon glows that map to specific UI semantics. The theme is implemented with CSS custom properties in `frontend/src/shared/styles/global.css` and applied to shared components.

### Design Tokens

**Location:** `frontend/src/shared/styles/global.css`

Key token groups:
- `--carbon-*` color tokens (backgrounds, surfaces, text, borders)
- `--carbon-gradient-*` for primary (cyan → violet), amber, danger, success, surface, and card gradients
- `--carbon-glow-*` for neon glows: cyan, amber, red, green, purple

### Restyled Components

- `Button` (`frontend/src/components/atoms/Button.svelte`) — primary, secondary, danger, ghost variants with glow.
- `Modal` (`frontend/src/components/atoms/Modal.svelte`) — carbon surface with gradient border and blur.
- `NoteForm` + `TypeSelector` (`frontend/src/components/molecules/NoteForm.svelte`, `TypeSelector.svelte`) — carbon inputs, cyan focus glow, amber validation glow.
- `NoteCard` (`frontend/src/widgets/notes/NoteCard.svelte`) — nebula gradient, type-colored left stripe, hover lift and glow, new/updated indicators.
- `BackButton`, `SearchBar`, `AuthCard`, `EditNoteModal`, `CreateNoteModal` — aligned with carbon surfaces and gradients.

### Color Semantics

- **Primary:** cyan → violet gradient for main CTAs.
- **Warning / quick capture:** amber glow.
- **Danger:** red glow for destructive actions.
- **Success:** green for positive states.
- **Type colors:** star, planet, comet, galaxy, asteroid, blackhole, etc. remain unique and are used for type chips, NoteCard stripes, and graph nodes.

---

## Note Detail Page

### Overview

The note detail page (`frontend/src/routes/notes/[id]/+page.svelte`) was rebuilt to show richer metadata, related note links, and graph navigation in the carbon theme.

### Features

- **Hero section:** title, type emoji, type label, description, visibility chip (public/private).
- **Metadata chips:** creation and update datetimes, public/private badge.
- **Tags and keywords:** rendered from `note.metadata.tags` and `note.metadata.keywords` as links to `/search?q=...`.
- **Connected notes:** list of related notes derived from `getGraphData` for the current note, showing:
  - Link type icon and label
  - Link direction (incoming / outgoing)
  - Celestial body type emoji
  - Link weight visual bar
- **Similar notes:** powered by `getSuggestions` with similarity score.
- **Actions:** Edit, Delete, Create child note, Open 2D graph, Open 3D constellation.
- **Transitions:** page and section fly/fade animations using Svelte transitions.
- **i18n:** all labels use `frontend/src/shared/utils/i18n.ts` keys in Russian and English.

### Manual Testing Notes

1. Open any note at `/notes/{id}`.
2. Verify the type chip, visibility chip, and datetime chips render.
3. If the note has `metadata.tags` or `metadata.keywords`, they appear as clickable search links.
4. If the note has graph links, the related notes section renders with direction arrows and weight bars.
5. Edit, delete, and child-note actions open the corresponding modals.
6. Graph/constellation buttons navigate to `/graph/{id}` and `/graph/3d/{id}`.
7. No console errors and no layout shifts on load.
