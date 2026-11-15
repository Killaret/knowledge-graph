# Link Types in Knowledge Graph

**Updated:** September 2026 (LINK-TYPES-1)

## Overview

Link types define the nature of relationships between notes in the knowledge graph. Choosing the right type helps structure knowledge, improves navigation, and makes the graph visually meaningful.

Current user-visible types:

| Type | Icon | Default weight | Line style | Color |
|------|------|----------------|------------|-------|
| Dependency | 🔗 | 0.7 | dashed `[10, 3]` | `#ff6600` |
| Related | 🔀 | 0.5 | solid, weak → dashed `[6, 4]` under 0.3 | `#999999` |
| Parent | ⬆️ | 0.9 | solid | `#2dd4bf` |
| Child | ⬇️ | 0.9 | solid | `#f472b6` |
| Auto link (model) | ✨ | — | thin solid | `#4ade80` |

> **Legacy types.** `reference` and `custom` are deprecated (LINK-TYPES-1). The
> API still accepts them, but they are persisted as `related`; migration 037
> rewrote existing rows and merged duplicate pairs (the survivor keeps the
> pair's maximum weight, absorbed rows are soft-deleted). Both types are gone
> from pickers, filters and the legend. `parent`/`child` remain system-managed
> origin links until ORIGIN-1.

## Available types

### Related — default

- **Icon:** 🔀
- **Meaning:** Notes are thematically connected, but neither strictly depends on the other. The default type everywhere: database column, API payloads that omit the type, and the creation form.
- **When to use:** Similar topics, alternative approaches, "see also" mentions — everything that used to be `reference`/`custom`.
- **Example:** "JavaScript" → Related → "TypeScript".
- **Visual:** Grey solid line; becomes dashed when the weight drops below 0.3.

### Dependency — a directed chain

- **Icon:** 🔗
- **Meaning:** A→B means A is the prerequisite and B follows it — direction runs "from what is done first to what comes later".
- **When to use:** Task chains, prerequisites, ordered steps, learning paths.
- **Example:** "Docker Basics" → Dependency → "Docker Compose".
- **Visual:** Orange dashed line; the dash animation runs along the direction.
- **Chain highlight (2D and 3D):** hovering a note that has dependency links highlights the whole chain through it — prerequisites and dependents alike — and dims everything else. Brightness decays with the distance from the hovered node; the walk is bounded by `frontend.graph.dependency_highlight_depth` (default 10). Re-hovering another chain node recomputes the chain from it.
- **Cycles:** links that lie on a dependency cycle render red, and the note panel shows a warning plus **"Requires:"** (incoming dependencies) and **"Needed for:"** (outgoing) rows.

### Parent / Child

- **Icons:** ⬆️ / ⬇️
- **Meaning:** origin links — the source note is the broader topic / a subtopic. Managed by the system; their migration to a dedicated origin concept is planned under ORIGIN-1.
- **Visual:** Teal / pink solid lines.

### Auto link (model)

- **Icon:** ✨
- **Meaning:** `source_type = "gamma"` — a link proposed by the similarity model, not a manual type. Shown in the legend as its own row; not toggleable through the type filter.
- **Visual:** Green (`#4ade80`) — a hue no manual type uses — thinner and more transparent than manual links; brightness follows the similarity weight. Confirming the link in the tooltip promotes it to `source_type = "user"` and it then takes its type's colour.

## Choosing a type

```
Is one note a prerequisite for the other?
  Yes → Dependency
  No →
    Is one note a broader / narrower topic?
      Yes → Parent / Child (system-managed)
      No → Related
```

## Visual encoding

- **Color** — manual links take `LinkType.color`; auto links use `AUTO_LINK_COLOR` (`#4ade80`).
- **Line dash pattern** is taken from `LinkType.lineDash`. `related` switches to `[6, 4]` when `weight < 0.3`; `dependency` dashes animate to show direction.
- **Line width** is `Math.max(1, weight * 4)`, multiplied by `0.55` for auto links and by `1.5` while a duplicate link warning is active.
- **Opacity** uses `0.4 + weight * 0.4` as the base; auto links are additionally scaled by `0.45`.
- **Dependency chain:** chain links fade with BFS depth (`max(0.35, 1 - depth * 0.12)`), non-chain elements drop to ~0.15 opacity; cycle links are red (`#ef4444`).
- **Bidirectional links** (A→B and B→A) are rendered as two mirrored quadratic curves so they do not overlap.
- **Deleted links** fade out; **new links** fade in through the opacity map animation.

## UI integration

- `LinkTypeSelector` (in `frontend/src/components/molecules/LinkTypeSelector.svelte`) shows creatable types with icon, label, color and a short description.
- `CockpitNoteDetails` lists links with type icon, weight bar, `source_type` badge and `last_weight_update`, plus dependency "Requires / Needed for" rows and a cycle warning.
- `LinkTooltip` on the graph shows the type icon, color, weight, source/target, `source_type` and `last_weight_update`. For `source_type = "gamma"` links it offers **Confirm link** (promotes the row to `source_type='user'`, origin kept in `metadata.gamma`) and **Not related** (deletes the link and records a rejection in `link_suppressions`, so the pair is not proposed again).
- `LinkTypeLegend` on the graph shows manual types (filter + minimum weight) and a static "Auto link (model)" row — in both 2D and 3D.

## API

### Create a link

```bash
POST /api/v1/links
{
  "source_note_id": "uuid-1",
  "target_note_id": "uuid-2",
  "link_type": "related",
  "weight": 0.8
}
```

`link_type` accepts `reference`, `dependency`, `related`, `custom`, `parent`, `child`. The legacy values `reference`/`custom` are normalized to `related` before persisting — the response reports `"link_type": "related"`.

### Update a link

```bash
PUT /api/v1/links/:id
{
  "link_type": "dependency",
  "weight": 0.9
}
```

### Graph data

Graph responses (`/api/v1/graph/public`, `/api/v1/notes/:id/graph`) include:

```json
{
  "id": "uuid",
  "source": "uuid-1",
  "target": "uuid-2",
  "weight": 0.8,
  "link_type": "related",
  "source_type": "user",
  "last_weight_update": "2026-07-29T12:00:00Z"
}
```

## Related files

- `frontend/src/entities/link/model/link-type.ts` — type definitions + `AUTO_LINK_COLOR`.
- `frontend/src/entities/graph-canvas/lib/dependency-chain.ts` — chain walk, cycle detection.
- `frontend/src/components/molecules/LinkTypeSelector.svelte` — type selector.
- `frontend/src/features/graph-ui/LinkTooltip.svelte` — graph hover tooltip.
- `frontend/src/features/graph-ui/LinkTypeLegend.svelte` — graph legend and filters.
- `frontend/src/widgets/cosmic-cockpit/CockpitNoteDetails.svelte` — note links panel.
- `frontend/src/entities/graph-canvas/lib/link-renderers.ts` — 2D link rendering.
- `frontend/src/features/graph-3d/lib/links.ts` — 3D link rendering.
- `docs/product/LINK_TYPES_RU.md` — Russian translation.
