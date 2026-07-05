# GraphCanvas Refactoring Plan

## Status: In Progress

## Completed
- ✅ Created folder structure: shared/lib/graph/renderer/{nodes,anomalies,utils}
- ✅ Created utils/helpers.ts (lightenColor, darkenColor, stringHash, hexToRgba, applyHueShiftToRGBA)
- ✅ Created utils/glow-intensity.ts
- ✅ Created utils/link-color.ts
- ✅ Created utils/line-dash.ts
- ✅ Created utils/node-color.ts
- ✅ Created utils/node-gradient.ts
- ✅ Created utils/index.ts
- ✅ Created nodes/star.ts
- ✅ Created nodes/planet.ts
- ✅ Created nodes/comet.ts
- ✅ Created nodes/galaxy.ts
- ✅ Created nodes/nebula.ts
- ✅ Created nodes/asteroid.ts
- ✅ Created nodes/debris.ts
- ✅ Created nodes/blackhole.ts
- ✅ Created nodes/technical.ts
- ✅ Created nodes/moon.ts
- ✅ Created nodes/index.ts

## Next Steps
1. Create anomaly renderer files (4 files):
   - reality-rift.ts
   - chromatic-maw.ts
   - void-whisper.ts
   - cosmic-abomination.ts

2. Create anomalies/index.ts

3. Update renderer.ts to import from new modules

4. Update GraphCanvas/index.ts

5. Run tests and fix issues
