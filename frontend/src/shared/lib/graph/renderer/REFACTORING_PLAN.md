# GraphCanvas Refactoring Plan

## Status: In Progress

## Completed
- ✅ Created folder structure: shared/lib/graph/renderer/{nodes,anomalies,utils}
- ✅ Created utils/helpers.ts (lightenColor, darkenColor, stringHash, hexToRgba)
- ✅ Created utils/glow-intensity.ts
- ✅ Created utils/index.ts

## Next Steps
1. Create utils files:
   - node-gradient.ts
   - link-color.ts
   - line-dash.ts
   - node-color.ts

2. Create node renderer files (10 files):
   - star.ts
   - planet.ts
   - comet.ts
   - galaxy.ts
   - nebula.ts
   - asteroid.ts
   - debris.ts
   - blackhole.ts
   - technical.ts
   - moon.ts

3. Create anomaly renderer files (4 files):
   - reality-rift.ts
   - chromatic-maw.ts
   - void-whisper.ts
   - cosmic-abomination.ts

4. Update renderer.ts to import from new modules

5. Update GraphCanvas/index.ts

6. Run tests and fix issues
