# Graph Link Visualization

## 1. Link Types

| Type | Purpose | Line color | Line style | Default weight |
|------|---------|------------|------------|----------------|
| **reference** | Direct note reference / citation | Blue `#3366ff` | Solid `[]` | 0.8 |
| **dependency** | One note depends on another | Orange `#ff6600` | Dashed `[10, 3]` | 0.7 |
| **related** | Thematic relation, weak | Grey `#999999` | Solid, dashed `[6, 4]` when weight < 0.3 | 0.5 |
| **custom** | User-defined type | Magenta `#ff66ff` | Dotted dashed `[2, 6]` | 0.5 |
| **parent** | Source is a broader topic | Teal `#2dd4bf` | Solid `[]` | 0.9 |
| **child** | Source is a subtopic | Pink `#f472b6` | Solid `[]` | 0.9 |

Each type is defined in `frontend/src/entities/link/model/link-type.ts` and exposes:

- `icon` — emoji shown in selectors, tooltips and legend.
- `label` — i18n key resolved through `formatMessage`.
- `description` / `example` — i18n keys shown in `LinkTypeSelector`.
- `color` — base hex color.
- `lineDash` — Canvas `setLineDash` pattern.
- `getColor(weight, opacity)` — color with weight-based opacity.
- `creatable` — whether the type can be created manually.

## 2. Weight Calculation

**Link weight** is a number from `0.0` to `1.0` reflecting the strength of connection between two notes.

### Manual links

- Set in `LinkCreator` / `LinkTypeSelector`.
- **Default:** `0.5`.
- Stored in the `links` table.
- Updated through `PUT /links/:id`.

### Automatic / NLP links

Weight is recalculated by `RefreshService` and `LinkWeightService`:

```text
final_weight = α × graph_weight + β × semantic_weight + γ × keyword_weight
```

| Component | Coefficient | Description |
|-----------|-------------|-------------|
| **graph_weight** | α (alpha) = 0.5 | Proximity in the explicit link graph (BFS traversal, decay=0.5) |
| **semantic_weight** | β (beta) = 0.5 | Cosine similarity of vector embeddings (pgvector) |
| **keyword_weight** | γ (gamma) = 0.2 | Keyword overlap (Jaccard Similarity) |

The `last_weight_update` column tracks when the weight was last recomputed.

## 3. Visual Weight Encoding

### Line width

```typescript
const lineWidth = Math.max(1, weight * 4);
```

| Weight | Width (px) |
|--------|------------|
| 0.1 | 1.0 |
| 0.5 | 2.0 |
| 0.8 | 3.2 |
| 1.0 | 4.0 |

### Opacity

```typescript
const baseOpacity = 0.4 + (weight ?? 0.5) * 0.4;
```

### Direct vs Recommended

| Feature | Direct link | Recommendation |
|---------|-------------|----------------|
| **Source** | `links` table | `note_recommendations` table |
| **Creation** | Manual via UI/API | Worker / NLP service |
| **Type** | Explicit (`reference`/`dependency`/`related`/`custom`/`parent`/`child`) | Usually `related` |
| **Weight** | User-defined (0.0–1.0) | Computed (α×graph + β×semantic + γ×keyword) |
| **Visual** | Type color and dash | Paler, weight-based dash |

## 4. Rendering

Main file: `frontend/src/entities/graph-canvas/lib/renderer.ts`

### `drawLink`

```typescript
export function drawLink(
  ctx: CanvasRenderingContext2D,
  link: SimulationLink,
  sourceNode: SimulationNode,
  targetNode: SimulationNode,
  opacity: number,
  hoveredNodeId: string | null,
  isDuplicateHighlighted: boolean,
  curveOffset: number
): void
```

### `drawAllLinks`

- Iterates over `simState.simLinks`.
- Uses the `linkOpacity` map for fade-in animation.
- Applies duplicate-link highlighting (yellow pulse).
- Renders bidirectional links as mirrored quadratic curves.
- Draws `dyingLinks` with `dyingLinkOpacity` for deletion fade-out.

### Bidirectional links

When both `A→B` and `B→A` exist, each link is offset perpendicularly by `BIDIRECTIONAL_LINK_OFFSET = 24` pixels to avoid overlap.

### Animations

- **Create / filter change:** New links start at `opacity = 0` and fade in via `startFadeAnimation`.
- **Delete:** Removed links are kept in `SimulationState.dyingLinks` and fade out over ~20 frames.

## 5. Interaction

### Hover

- `LinkTooltip` (in `frontend/src/features/graph-ui/LinkTooltip.svelte`) shows:
  - type icon and label,
  - color line,
  - weight and source/target titles,
  - `source_type` badge (user / auto / worker),
  - `last_weight_update` date.
- Edit / delete buttons are shown when `onLinkEdit` / `onLinkDelete` are provided.

### Legend and filtering

- `LinkTypeLegend` (`frontend/src/features/graph-ui/LinkTypeLegend.svelte`) lists all types.
- Toggling a type adds/removes it from `graphStore.hiddenLinkTypes` (empty = all visible).
- The minimum weight slider updates `graphStore.minLinkWeight`.
- `GraphCanvas` derives `visibleLinks` from the store and restarts the simulation.

## 6. Graph Data

Backend responses include `id`, enabling in-place editing:

```json
{
  "id": "uuid",
  "source": "uuid-1",
  "target": "uuid-2",
  "weight": 0.8,
  "link_type": "reference",
  "source_type": "user",
  "last_weight_update": "2026-07-29T12:00:00Z"
}
```

## 7. Related Documentation

- `docs/LINK_TYPES.md` — English link type reference.
- `docs/LINK_TYPES_RU.md` — Russian link type reference.
