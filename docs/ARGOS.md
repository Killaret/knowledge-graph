# Argos Visual Regression Testing

## Overview

This project uses the official [`@argos-ci/playwright`](https://www.argos-ci.com/docs/playwright) SDK for visual regression tests. The Playwright reporter uploads screenshots to [Argos](https://www.argos-ci.com/) automatically whenever `CI` or `ARGOS_UPLOAD_LOCAL` is set.

## Setup

### Repository Status

- **Repository**: `Killaret/knowledge-graph`
- **Visibility**: Public
- **Argos Plan**: Free tier (available for public repos)
- **Project Token**: Set via the `ARGOS_TOKEN` environment variable (repository / GitHub secret)

### Required Environment Variables

| Variable | Purpose |
|----------|---------|
| `ARGOS_TOKEN` | Repository token for uploading screenshots. |
| `FRONTEND_URL` | Target frontend URL for tests (default: `http://localhost:3002`). |
| `ARGOS_UPLOAD_LOCAL` | Set to `true` to upload from a local run. CI uploads automatically. |
| `ARGOS_REFERENCE_BRANCH` | Baseline branch for comparisons (`ai-agents` for this work). |
| `SKIP_AUTH` | Set to `true` for the test stack so the backend accepts the test user. |

### GitHub CI Workflow

See `.github/workflows/main.yml` for the `visual-regression` job. It performs:

1. Starts the isolated test stack (`docker compose -f docker-compose.test.yml up -d --build --wait`) with `SKIP_AUTH=true`.
2. Seeds a small deterministic fixture (`NOTE_COUNT=20 LINK_COUNT=10 SEED=42`).
3. Runs `npx playwright test --project=visual`.
4. The Argos Playwright reporter uploads screenshots automatically.

```yaml
- name: Run visual tests with Argos
  env:
    ARGOS_TOKEN: ${{ secrets.ARGOS_TOKEN }}
  run: npx playwright test --project=visual
```

The manual `npx argos upload` step is no longer needed.

## Visual Tests

### Test Location

`frontend/tests/visual/visual-regression.spec.ts`

### Test Scenarios

The suite captures stable, deterministic views only:

1. **Home**
   - Default view
   - List view
   - Filtered by star node type

2. **2D Graph**
   - Full graph view
   - Ghost node creation form (press `N`)
   - Help hotkeys modal (press `?`)

3. **NoteCard**
   - Selected state in list view

4. **Search**
   - Search page
   - Search with query (`star`)
   - Empty state

5. **3D Graph**
   - Frozen notice redirect to 2D graph

### Running Tests Locally

**Windows:**
```powershell
./scripts/testing/start-test.ps1
./scripts/testing/seed-test-data.ps1 -NoteCount 20 -LinkCount 10 -Seed 42

cd frontend
$env:FRONTEND_URL = "http://localhost:3002"
$env:SKIP_AUTH = "true"
npm run test:visual:upload
```

**Linux / Mac:**
```bash
SKIP_AUTH=true ./scripts/testing/start-test.sh
NOTE_COUNT=20 LINK_COUNT=10 SEED=42 ./scripts/testing/seed-test-data.sh

cd frontend
FRONTEND_URL=http://localhost:3002 ARGOS_UPLOAD_LOCAL=true npm run test:visual
```

### Baselines

Argos uses the `ARGOS_REFERENCE_BRANCH` (`ai-agents`) as the baseline for new PRs. The first upload on a branch creates the baseline; subsequent uploads are compared against it.

## Test Data

Visual tests run against the isolated test stack seeded with a small fixture. To keep screenshots deterministic:

- `SEED=42` makes link creation reproducible in `seed-test-data.sh` / `-Seed 42` in `seed-test-data.ps1`.
- The browser gets a seeded `Math.random` linear congruential generator to stabilize canvas particle positions and d3-force jitter.
- `?stableRender=true` disables animations on the graph canvas.
- `data-visual-test="transparent"` hides dynamic content such as timestamps, new/updated indicators, and tooltips.
- `data-testid="graph-canvas"` exposes `data-test-stable="true"` once the force simulation reaches a steady state.

## Configuration

Playwright configuration is in `frontend/playwright.config.ts`:

- `reporter` includes `createArgosReporterOptions` from `@argos-ci/playwright/reporter`.
- `use.bypassCSP: true` allows Argos stabilization script injection.
- The `visual` project is filtered by `grep: /@visual/` and depends on `setup`.
- The default `chromium` project uses `grepInvert: /@3d|@visual/`.

## Troubleshooting

### Screenshots Not Showing Data

1. **Verify the test stack is running:**
   ```bash
   docker compose -f docker-compose.test.yml ps
   curl http://localhost:8083/health
   ```

2. **Seed data with the deterministic fixture:**
   ```bash
   NOTE_COUNT=20 LINK_COUNT=10 SEED=42 ./scripts/testing/seed-test-data.sh
   ```

3. **Confirm `SKIP_AUTH=true`:**
   - The test stack backend needs `SKIP_AUTH=true` so the frontend can load notes without a real login.
   - The Playwright test injects `window.__SKIP_AUTH__ = true` on the client.

4. **Check the graph service:**
   ```bash
   curl http://localhost:9091/health
   ```

### Argos Upload Failures

1. **Token Issues:**
   - Verify `ARGOS_TOKEN` is set in GitHub secrets or your local environment.
   - Check that the token has proper upload permissions.

2. **Network Issues:**
   - Ensure the CI runner / local machine can reach `api.argos-ci.com`.

## Best Practices

1. **Consistent Test Data:** Use the seeded fixture and `stableRender=true` for every visual test.
2. **Timing and States:** Wait for `[data-testid="graph-canvas"][data-test-stable="true"]` before graph screenshots.
3. **Mask Dynamic Content:** Add `data-visual-test="transparent"` to any element with unstable rendering (timestamps, live indicators, animated tooltips).
4. **Viewport Consistency:** Use `argosScreenshot(..., { viewports: [...] })` for responsive scenarios, but keep the standard suite on a fixed desktop viewport.

## Resources

- [Argos Playwright Documentation](https://argos-ci.com/docs/playwright)
- [Playwright Screenshot Testing](https://playwright.dev/docs/screenshot-testing)
- [Project CI/CD Pipeline](../.github/workflows/main.yml)
