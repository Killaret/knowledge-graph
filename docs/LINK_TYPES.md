# Link Types in Knowledge Graph

**Updated:** July 2026

## Overview

Link types define the nature of relationships between notes in the knowledge graph. Choosing the right type helps structure knowledge, improves navigation, and makes the graph visually meaningful.

The following types are available:

| Type | Icon | Default weight | Line style | Color |
|------|------|----------------|------------|-------|
| Reference | 📖 | 0.8 | solid | `#3366ff` |
| Dependency | 🔗 | 0.7 | dashed `[10, 3]` | `#ff6600` |
| Related | 🔀 | 0.5 | solid, weak → dashed `[6, 4]` under 0.3 | `#999999` |
| Custom | ✨ | 0.5 | dashed `[2, 6]` | `#ff66ff` |
| Parent | ⬆️ | 0.9 | solid | `#2dd4bf` |
| Child | ⬇️ | 0.9 | solid | `#f472b6` |

## Available types

### Reference

- **Icon:** 📖
- **Meaning:** One note simply mentions another.
- **When to use:** Cross-references, citations, "see also" links.
- **Example:** "Quantum physics" → Reference → "Uncertainty principle".
- **Visual:** Blue solid line.

### Dependency

- **Icon:** 🔗
- **Meaning:** The target note depends on the source note for understanding or execution.
- **When to use:** Prerequisites, ordered steps, learning paths.
- **Example:** "Docker Compose" → Dependency → "Docker Basics".
- **Visual:** Orange dashed line.

### Related

- **Icon:** 🔀
- **Meaning:** Notes are thematically connected, but neither strictly depends on the other.
- **When to use:** Similar topics, alternative approaches, related concepts.
- **Example:** "JavaScript" → Related → "TypeScript".
- **Visual:** Grey solid line; becomes dashed when the weight drops below 0.3.

### Custom

- **Icon:** ✨
- **Meaning:** A user-defined type for relationships that do not fit the standard set.
- **When to use:** Project-specific relations, experimental categories.
- **Example:** "Follow-up", "contradicts", "inspired by".
- **Visual:** Magenta dashed line.

### Parent

- **Icon:** ⬆️
- **Meaning:** The source note is the broader topic that contains the target.
- **When to use:** Hierarchies, categories, parent–child topic trees.
- **Example:** "Programming" → Parent → "Functional programming".
- **Visual:** Teal solid line.

### Child

- **Icon:** ⬇️
- **Meaning:** The source note is a subtopic of the target.
- **When to use:** Subtopics, subcategories, drill-down relationships.
- **Example:** "Functional programming" → Child → "Programming".
- **Visual:** Pink solid line.

## Choosing a type

```
Is one note a prerequisite for the other?
  Yes → Dependency
  No →
    Is one note a broader / narrower topic?
      Yes → Parent / Child
      No →
        Is one note a simple mention or source?
          Yes → Reference
          No →
            Are the notes thematically close?
              Yes → Related
              No → Custom
```

## Visual encoding

- **Color** is taken from the type definition (`LinkType.color`).
- **Line dash pattern** is taken from `LinkType.lineDash`. `related` switches to `[6, 4]` when `weight < 0.3`.
- **Line width** is `Math.max(1, weight * 4)` and multiplied by `1.5` while a duplicate link warning is active.
- **Opacity** uses `LinkType.getColor(weight, fadeOpacity)` with a base opacity of `0.4 + weight * 0.4`.
- **Bidirectional links** (A→B and B→A) are rendered as two mirrored quadratic curves so they do not overlap.
- **Deleted links** fade out; **new links** fade in through the opacity map animation.

## UI integration

- `LinkTypeSelector` (in `frontend/src/components/molecules/LinkTypeSelector.svelte`) shows all creatable types with icon, label, color and a short description.
- `CockpitNoteDetails` lists links with type icon, weight bar, `source_type` badge and `last_weight_update`.
- `LinkTooltip` on the graph shows the type icon, color, weight, source/target, `source_type` and `last_weight_update`.
- `LinkTypeLegend` on the graph shows all types and allows filtering by type and minimum weight.

## API

### Create a link

```bash
POST /api/v1/links
{
  "source_note_id": "uuid-1",
  "target_note_id": "uuid-2",
  "link_type": "reference",
  "weight": 0.8
}
```

### Update a link

```bash
PUT /api/v1/links/:id
{
  "link_type": "dependency",
  "weight": 0.9
}
```

### Graph data

Graph responses (`/api/v1/graph/all`, `/api/v1/notes/:id/graph`) include:

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

## Related files

- `frontend/src/entities/link/model/link-type.ts` — type definitions.
- `frontend/src/components/molecules/LinkTypeSelector.svelte` — type selector.
- `frontend/src/features/graph-ui/LinkTooltip.svelte` — graph hover tooltip.
- `frontend/src/features/graph-ui/LinkTypeLegend.svelte` — graph legend and filters.
- `frontend/src/widgets/cosmic-cockpit/CockpitNoteDetails.svelte` — note links panel.
- `frontend/src/entities/graph-canvas/lib/renderer.ts` — graph rendering.
- `docs/LINK_TYPES_RU.md` — Russian translation.
