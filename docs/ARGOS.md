# Argos Visual Regression Testing

## Overview

This project uses the official [`@argos-ci/playwright`](https://www.argos-ci.com/docs/playwright) SDK for visual regression tests. The Playwright reporter uploads screenshots to [Argos](https://www.argos-ci.com/) automatically whenever `CI` or `ARGOS_UPLOAD_LOCAL` is set.

## Setup

### Repository Status

- **Repository**: `Killaret/knowledge-graph`
- **Visibility**: Public
- **Argos Plan**: Free tier (available for public repos)
- **Project Token**: Set via the `ARGOS_TOKEN` environment variable (repository / GitHub secret). For local runs you can also place the token in `frontend/argos.json` (gitignored): `{"token": "..."}`; `playwright.config.ts` and the regression script both read this file when `ARGOS_TOKEN` is not set.

### Required Environment Variables

| Variable | Purpose |
|----------|---------|
| `ARGOS_TOKEN` | Repository token for uploading screenshots. |
| `FRONTEND_URL` | Target frontend URL for tests (default: `http://localhost:3002`). |
| `ARGOS_UPLOAD_LOCAL` | Set to `true` to upload from a local run. CI uploads automatically. |
| `ARGOS_REFERENCE_BRANCH` | Baseline branch for comparisons (`main`). |
| `SKIP_AUTH` | Set to `true` for the test stack so the backend accepts the test user. |

### GitHub CI Workflow

See `.github/workflows/main.yml` for the `visual-regression` job. It performs:

1. Starts the isolated test stack (`docker compose -f docker-compose.test.yml up -d --build --wait`) with `SKIP_AUTH=false`.
2. Seeds a small deterministic fixture (`NOTE_COUNT=20 LINK_COUNT=10 SEED=42 PUBLIC_PERCENT=50`).
3. Runs `npx playwright test --project=visual --project=visual-real-auth`.
4. The Argos Playwright reporter uploads screenshots automatically.

```yaml
- name: Run visual tests with Argos
  env:
    ARGOS_TOKEN: ${{ secrets.ARGOS_TOKEN }}
  run: npx playwright test --project=visual --project=visual-real-auth
```

The manual `npx argos upload` step is no longer needed.

## Visual Tests

### Test Location

The suite is split into two files by audience ([`VIS-1`](tasks/VIS-1-split-visual-baselines.md)):

- `frontend/tests/visual/visual-anonymous.spec.ts` — what a logged-out visitor sees: login and register pages, the public 2D graph, search page, empty state, responsive home. Runs in the `visual` project with no `storageState` and no `__SKIP_AUTH__` injection.
- `frontend/tests/visual/visual-authenticated.spec.ts` — scenarios that need a session: list view, star filter, ghost node form, help modal, selected NoteCard, 3D view. Runs in the `visual-real-auth` project; `tests/setup/auth.setup.ts` logs in as the seeded `testuser` and persists `storageState`.

### Test Scenarios

The suite captures stable, deterministic views only:

1. **Anonymous (`visual`)**
   - Login and register pages
   - Public 2D graph
   - Search page and empty state
   - Home default and responsive viewports

2. **Authenticated (`visual-real-auth`)**
   - Home: default view, list view, filtered by star node type
   - 2D Graph: full view, ghost node creation form (press `N`), help modal (press `?`)
   - NoteCard selected state in list view
   - Search: page, query (`star`), empty state
   - 3D Graph view
   - Home responsive viewports

### Running Tests Locally

**Windows:**
```powershell
$env:SKIP_AUTH = "false"
./scripts/testing/start-test.ps1
./scripts/testing/seed-test-data.ps1 -NoteCount 20 -LinkCount 10 -Seed 42 -PublicPercent 50

cd frontend
$env:FRONTEND_URL = "http://localhost:3002"
$env:SKIP_AUTH = "false"
npx playwright test --project=visual --project=visual-real-auth
```

**Linux / Mac:**
```bash
SKIP_AUTH=false ./scripts/testing/start-test.sh
NOTE_COUNT=20 LINK_COUNT=10 SEED=42 PUBLIC_PERCENT=50 ./scripts/testing/seed-test-data.sh

cd frontend
FRONTEND_URL=http://localhost:3002 SKIP_AUTH=false ARGOS_UPLOAD_LOCAL=true npm run test:visual
```

### Baselines

Argos uses the `ARGOS_REFERENCE_BRANCH` (`main`) as the baseline for new PRs. The first upload on a branch creates the baseline; subsequent uploads are compared against it.

Both projects run with `SKIP_AUTH=false`: the anonymous baseline records the real logged-out state (public notes only, `applyNoteScope` filters by owner), the authorized baseline records the seeded `testuser` session. The two baselines are expected to differ — the public graph shows only the published subset.

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
- The `visual` project runs `tests/visual/visual-anonymous.spec.ts` with no storage state.
- The `visual-real-auth` project runs `tests/visual/visual-authenticated.spec.ts` with `storageState` from `tests/setup/.auth/testuser.json` (produced by the `setup-auth` dependency).
- The default `chromium` project uses `grepInvert: /@3d|@visual/`.

## Troubleshooting

### Screenshots Not Showing Data

1. **Verify the test stack is running:**
   ```bash
   docker compose -f docker-compose.test.yml ps
   curl http://localhost:18083/health
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
