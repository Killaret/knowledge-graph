# Frontend Domain Models

This directory contains lightweight domain models used by the SvelteKit frontend. They are value objects and small entities for UI state, not full backend aggregates.

## Models

| File | Export | Purpose |
|------|--------|---------|
| `celestial-body.ts` | `CelestialBody` | Visual representation of graph node types (star, planet, comet, etc.) |
| `link-type.ts` | `LinkType` | Visual/link-style metadata for graph links |
| `graph-mode.ts` | `GraphMode` | Graph visualization mode (normal/focus) |
| `graph-delta.ts` | `GraphDelta` | Incremental updates to the graph state |
| `achievement.ts` | `Achievement` | Normalized achievement data from the API |
| `filter-state.ts` | `FilterState` | Note list filtering, search, sorting state |
| `search-query.ts` | `SearchQuery` | URL-safe search query parsing/validation |
| `notification.ts` | `Notification` | Toast/overlay notification model |
| `theme.ts` | `Theme` | Standard/galactic visual theme |
| `user-points.ts` | `UserPoints` | User achievement progress |

## Example

```typescript
import { CelestialBody } from "$shared/lib/domain";

const body = CelestialBody.fromType("star");
console.log(body.label("ru")); // Russian label via i18n
```

## AI Agent Notes

- All labels and user-facing strings resolve through i18n keys; do not hardcode text.
- These are immutable value objects; use factory methods like `X.fromType(...)` or `new X(...)`.
- Each model has a corresponding `*.test.ts` file.
