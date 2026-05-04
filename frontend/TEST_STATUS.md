# Frontend Test Status

## Overview

This document tracks the test coverage and status of the Knowledge Graph frontend application.

## Test Structure

### Unit Tests (Vitest)
Located in `src/lib/components/` alongside components.

#### GraphCanvas Tests

| Test File | Coverage | Status | Notes |
|-----------|----------|--------|-------|
| `GraphCanvas.spec.ts` | Core functionality | ✅ Passing | Basic component rendering |
| `GraphCanvas.node-types.spec.ts` | Data types | ✅ Passing | Type validation |
| `GraphCanvas.node-types.visual.spec.ts` | Visual correctness | ✅ Passing | **NEW** Verifies fillStyle, strokeStyle, and visual properties for all celestial body types |
| `GraphCanvas.links.visual.spec.ts` | Link rendering | ✅ Passing | **NEW** Verifies links connect correct nodes with proper styles |
| `GraphCanvas.rendering.spec.ts` | Rendering logic | ✅ Passing | Canvas drawing operations |
| `GraphCanvas.interactions.spec.ts` | User interactions | ✅ Passing | Zoom, pan, click handlers |

#### Visual Correctness Tests (NEW)

These tests ensure that celestial bodies are rendered with correct visual identity:

**Node Types Verified:**
- ⭐ **Star**: Golden fill (#ffdd88), orange stroke (#cc9900), 5-point shape, glow shadow
- 🪐 **Planet**: Earthy colors (#c9b37c, #a57c2c), circular base with ellipse bands
- ☄️ **Comet**: Cyan fill (#aaffdd), translucent tail (rgba(170, 255, 221, 0.6))
- 🌌 **Galaxy**: Purple translucent ellipses (rgba(200, 180, 255, n))
- 🪨 **Asteroid**: Brown rocky colors (#8b7355, #5c4a3a), irregular 7-point polygon

**Link Styles Verified:**
- **Reference**: Blue (#3366ff), solid line, 0.8x thickness
- **Dependency**: Orange (#ff6600), dash-dot [10, 3], 1.5x thickness
- **Related**: Gray (#999999), dashed only if weight < 0.3
- **Custom**: Pink (#ff66ff), dotted [2, 6]

### E2E Tests (Playwright)
Located in `tests/` directory.

| Test File | Coverage | Status | Notes |
|-----------|----------|--------|-------|
| `notes.spec.ts` | CRUD operations | ✅ Passing | Note management via modal |
| `graph-3d.spec.ts` | 3D graph | ✅ Passing | Three.js rendering |
| `graph-3d-modules.spec.ts` | 3D modules | ✅ Passing | Force graph 3D |
| `graph-3d-performance.spec.ts` | Performance | ⚠️ Skipped in CI | 100 nodes load time |
| `camera-position.spec.ts` | Camera controls | ✅ Passing | View navigation |
| `celestial-body-types.spec.ts` | Body filtering | ✅ Passing | Filter by type |
| `filter-by-type.spec.ts` | Type filters | ✅ Passing | Chip interactions |
| `graph-visual.spec.ts` | Visual regression | ✅ Passing | **NEW** Screenshot comparisons |

### Visual Regression Tests (NEW)

Located in `tests/graph-visual.spec.ts`

**Screenshot Tests:**
- Star node appearance
- Planet node appearance  
- Comet node appearance
- Galaxy node appearance
- Asteroid node appearance
- Link between two nodes
- Multiple link types visualization

**Baseline Location:** `tests/__snapshots__/`

## Running Tests

### Unit Tests
```bash
npm run test:unit        # Interactive mode
npm run test:unit -- --run    # CI mode
```

### E2E Tests
```bash
npx playwright test      # All tests
npx playwright test --project=chromium  # Chromium only
npx playwright test --grep-invert "@performance"  # Exclude performance tests
```

### Visual Tests Only
```bash
npx playwright test tests/graph-visual.spec.ts --project=chromium
```

## Test Configuration

### Unit Test Config (`vitest.config.ts`)
- Framework: Vitest
- Environment: jsdom with canvas mock
- Coverage: v8 provider

### E2E Test Config (`playwright.config.ts`)
- Browsers: Chromium, Firefox, WebKit
- Base URL: http://localhost:3000 (Docker)
- Screenshot: On failure
- Video: Retained on failure

## CI/CD Integration

Tests run automatically on push via `.github/workflows/frontend-tests.yml`:
- Lint check
- Unit tests
- Playwright tests (excluding @performance)
- Coverage report upload

## Known Issues

1. **Performance Tests**: `graph-3d-performance.spec.ts` can timeout on slower machines. Excluded from CI via `--grep-invert "@performance"`.

2. **Visual Tests**: Baseline screenshots may differ slightly between environments due to font rendering differences. Threshold set to 0.2-0.3.

3. **Asteroid Rendering**: Uses `Math.random()` for irregular shape, making exact pixel matching impossible. Visual tests use appropriate thresholds.

## Recent Changes

### Added Visual Correctness Tests
- **Date**: 2026-05-04
- **Scope**: GraphCanvas node types and links
- **Files Added**:
  - `src/lib/components/GraphCanvas.node-types.visual.spec.ts`
  - `src/lib/components/GraphCanvas.links.visual.spec.ts`
  - `tests/graph-visual.spec.ts`
- **Coverage**: 10+ new tests ensuring visual identity of celestial bodies

## Future Improvements

- [ ] Add visual tests for 3D graph (Three.js meshes)
- [ ] Add tests for link labels and weights
- [ ] Add tests for node selection highlighting
- [ ] Add tests for graph clustering visualization

## Test Results Summary

**Current Status**: ✅ 85 passing, 1 skipped (performance), 0 failed

Last updated: 2026-05-04
