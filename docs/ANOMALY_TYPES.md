# Anomaly Types in Knowledge Graph

**Updated:** July 2026  
**Status:** Active - Used for unknown node types

---

## Overview

The Knowledge Graph visualization system includes **4 anomaly types** that are automatically rendered for nodes with unknown or unrecognized types. These anomalies provide visually distinct representations for nodes that don't match standard celestial body types.

---

## Anomaly Types

### 1. Reality Rift
- **Appearance:** Dark core with jagged cracks and amoebic contour
- **Use Case:** Unknown nodes with hash % 4 === 0
- **Visual Characteristics:**
  - Dark core color
  - Multiple crack lines radiating outward
  - Deformed, organic shape

### 2. Chromatic Maw
- **Appearance:** Tentacles with gradient core
- **Use Case:** Unknown nodes with hash % 4 === 1
- **Visual Characteristics:**
  - Central gradient core
  - Multiple tentacle-like extensions
  - Color-shifting effects

### 3. Void Whisper
- **Appearance:** Particles with connections and snow effect
- **Use Case:** Unknown nodes with hash % 4 === 2
- **Visual Characteristics:**
  - Small particles around core
  - Connection lines between particles
  - Subtle crack patterns

### 4. Cosmic Abomination
- **Appearance:** Combines all three anomaly types
- **Use Case:** Unknown nodes with hash % 4 === 3
- **Visual Characteristics:**
  - Reality rift core
  - Chromatic maw tentacles
  - Void whisper particles
  - Most complex visual

---

## Implementation

### Automatic Dispatch

The system automatically selects anomaly types for nodes with `type: 'unknown'`:

```typescript
// In renderer.ts
export function drawNode(
  ctx: CanvasRenderingContext2D,
  node: SimulationNode,
  r: number,
  angle: number,
  enableShadows: boolean,
  disableVariation: boolean = false
): void {
  const type = node.type || 'unknown';
  
  switch (type) {
    case 'star':
      drawStar(ctx, x, y, r, angle, variation);
      break;
    // ... other types
    case 'unknown':
      drawUnknown(ctx, x, y, r, angle, node.id);
      break;
    default:
      drawStar(ctx, x, y, r, angle, variation);
      break;
  }
}
```

### Deterministic Selection

Anomaly type is selected based on node ID hash for consistency:

```typescript
export function drawUnknown(
  ctx: CanvasRenderingContext2D,
  x: number,
  y: number,
  r: number,
  angle: number,
  nodeId: string,
  customRenderers?: Record<number, AnomalyRenderer>
): void {
  // Select anomaly type based on hash of nodeId (deterministic)
  const hash = stringHash(nodeId);
  const anomalyType = hash % 4;
  
  const params = getAnomalyParams(nodeId);
  const renderers = customRenderers ?? {
    0: drawRealityRift,
    1: drawChromaticMaw,
    2: drawVoidWhisper,
    3: drawCosmicAbomination,
  } as Record<number, AnomalyRenderer>;

  const rendererFn = renderers[anomalyType] ?? drawRealityRift;
  rendererFn(ctx, x, y, r, params);
}
```

### Hash Function

Simple string hash for deterministic anomaly selection:

```typescript
function stringHash(str: string): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return Math.abs(hash);
}
```

---

## Configuration

Anomaly rendering is configured in `frontend/src/lib/config.ts`:

```typescript
anomaly: {
  reality_rift: {
    core_color: string;
    glow_color: string;
    crack_count_min: number;
    crack_count_max: number;
    deform_amount_min: number;
    deform_amount_max: number;
  };
  chromatic_maw: {
    tentacle_count_min: number;
    tentacle_count_max: number;
    hue_shift_base: number;
    hue_shift_range: number;
  };
  void_whisper: {
    particle_count_min: number;
    particle_count_max: number;
    hue_shift_base: number;
    hue_shift_range: number;
    connection_distance_threshold: number;
  };
  cosmic_abomination: {
    particle_count_min: number;
    particle_count_max: number;
    tentacle_count_min: number;
    tentacle_count_max: number;
    crack_count_min: number;
    crack_count_max: number;
  };
}
```

---

## Usage in Application

### Automatic Fallback

Nodes without a recognized type automatically render as anomalies:

```typescript
// In +page.svelte
nodes: allNotes.map(n => ({ 
  id: n.id, 
  title: n.title, 
  type: n.type || 'unknown'  // Fallback to unknown
}))
```

### Type Selector

The TypeSelector component includes standard types but not anomalies:

```typescript
type CelestialType = 'star' | 'planet' | 'comet' | 'galaxy' | 'asteroid';
```

Anomalies are reserved for system-generated unknown types.

---

## Testing

### Unit Tests

Anomaly rendering is tested in `frontend/src/lib/components/GraphCanvas.node-types.spec.ts`:

```typescript
describe('Anomaly Rendering (Unknown Node Types)', () => {
  it('drawRealityRift renders without errors');
  it('drawChromaticMaw renders tentacles with gradient core');
  it('drawVoidWhisper renders particles with connections');
  it('drawCosmicAbomination combines all anomaly types');
  it('drawUnknown dispatches to one anomaly renderer based on nodeId');
  it('drawUnknown is deterministic for the same nodeId');
  it('getAnomalyParams returns stable and different values for different nodeIds');
  it('drawNode dispatches to drawUnknown for unknown type');
});
```

### Integration Tests

Anomaly functions are also tested in `frontend/src/lib/components/GraphCanvas/renderer.test.ts`:

```typescript
describe('renderer anomaly functions', () => {
  describe('drawRealityRift', () => { /* ... */ });
  describe('drawChromaticMaw', () => { /* ... */ });
  describe('drawVoidWhisper', () => { /* ... */ });
  describe('drawCosmicAbomination', () => { /* ... */ });
});
```

---

## Visual Examples

### Reality Rift
- Dark, cracked appearance
- Organic, deformed shape
- Subtle glow effects

### Chromatic Maw
- Central gradient core
- Radiating tentacles
- Color-shifting animations

### Void Whisper
- Particle-based rendering
- Connection lines
- Snow-like effect

### Cosmic Abomination
- Most complex visual
- Combines all anomaly types
- High visual impact

---

## Performance Considerations

### Rendering Cost

Anomaly rendering is more expensive than standard types:
- **Reality Rift:** Medium cost (cracks + deformations)
- **Chromatic Maw:** High cost (tentacles + gradients)
- **Void Whisper:** Medium cost (particles + connections)
- **Cosmic Abomination:** Very high cost (all combined)

### Optimization

For large graphs with many unknown nodes:
1. Consider limiting unknown node count
2. Use `disableVariation` flag for stable rendering
3. Implement LOD (Level of Detail) for distant nodes

---

## Future Enhancements

### Potential Improvements

1. **Custom Anomaly Types:** Allow users to define custom anomaly renderers
2. **Anomaly Severity:** Different visual styles based on "how unknown" a node is
3. **Anomaly Metadata:** Store additional context about why a node is unknown
4. **Anomaly Interaction:** Special interactions for anomaly nodes (e.g., "investigate unknown")

### Configuration Options

```typescript
// Potential future configuration
anomaly: {
  enabled: true;
  maxUnknownNodes: 50;
  fallbackToStar: false;
  customRenderers: Record<string, AnomalyRenderer>;
}
```

---

## Troubleshooting

### All Nodes Render as Anomalies

**Problem:** All nodes show as anomalies instead of standard types

**Solution:** Check that node types are correctly set:
```typescript
// Ensure types are set correctly
type: n.type || 'unknown'  // Should be 'star', 'planet', etc.
```

### Anomaly Type Changes on Re-render

**Problem:** Same node shows different anomaly types

**Solution:** Ensure node IDs are consistent:
```typescript
// Hash is based on nodeId - must be stable
const hash = stringHash(nodeId);
```

### Performance Issues with Many Unknown Nodes

**Problem:** Graph rendering slows down with many anomalies

**Solution:** 
1. Reduce number of unknown nodes
2. Use simpler anomaly types
3. Implement LOD for distant nodes

---

## Related Documentation

- [GraphCanvas Renderer](../frontend/src/lib/components/GraphCanvas/renderer.ts)
- [Configuration](../frontend/src/lib/config.ts)
- [Node Types](./NODE_TYPES.md)
- [Visual Testing](./ARGOS.md)

---

## Summary

Anomaly types provide a visually distinct and deterministic way to render unknown nodes in the Knowledge Graph. The system automatically selects one of four anomaly types based on node ID hash, ensuring consistent rendering while providing visual interest for unrecognized content.