# Argos Visual Regression Testing

## Overview

This project uses [Argos](https://www.argos-ci.com/) for automated visual regression testing. Argos provides a free tier for public repositories, which is available for the `Killaret/knowledge-graph` repository.

## Setup

### Repository Status

- **Repository**: `Killaret/knowledge-graph`
- **Visibility**: Public ✅
- **Argos Plan**: Free tier (available for public repos)
- **Project Token**: Set via the `ARGOS_TOKEN` environment variable (configured in repository settings)

### Configuration

Argos is configured through environment variables.

- `ARGOS_TOKEN` — repository token (required for upload).
- `ARGOS_BRANCH` — branch name (optional, defaults to the current Git branch).
- `ARGOS_COMMIT` — commit SHA (optional, defaults to the current Git HEAD).
- `ARGOS_REFERENCE_BRANCH` — branch used as the baseline for comparison.
- `ARGOS_REFERENCE_COMMIT` — commit used as the baseline for comparison.

For the local CLI, the token can also be passed with `--token`.

**GitHub CI Workflow:**
```yaml
- name: Upload screenshots to Argos
  if: always()
  env:
    ARGOS_TOKEN: ${{ secrets.ARGOS_TOKEN }}
    ARGOS_COMMIT: ${{ github.sha }}
    ARGOS_BRANCH: ${{ github.ref_name }}
  run: |
    cd frontend
    npx argos upload ./argos-screenshots --token $ARGOS_TOKEN
```

## Visual Tests

### Test Location
`frontend/tests/visual/visual-regression.spec.ts`

### Test Scenarios

1. **3D Graph Views**
   - Full graph with all node types
   - Loading state
   - Planet node focus
   - Star node focus
   - With link connections
   - Without links

2. **2D Graph Views**
   - Empty state
   - Default view
   - Filtered by star nodes
   - List view
   - Modal confirm dialog
   - Search with results

### Running Tests Locally

**Dev Mode (Local):**
```bash
cd frontend
npm run test:visual
```

**Docker Mode:**
```bash
# Start dev stack
docker-compose up -d postgres redis backend graph-service frontend nginx

# Run tests (they use Docker frontend at http://localhost:5173)
cd frontend
npm run test:visual

# Upload screenshots to Argos
cd frontend
npx argos upload argos-screenshots/ --token $ARGOS_TOKEN
```

### CI/CD Pipeline

The Argos integration is configured in `.github/workflows/main.yml`:

1. **On Push to Main:**
   - Runs Playwright visual tests
   - Uploads screenshots to Argos
   - Creates/updates review in Argos dashboard

2. **Review Process:**
   - Argos compares screenshots against baseline
   - Creates review with detected differences
   - Team can approve/reject changes
   - Approved changes become new baseline

## Test Data

Visual tests use the test helper `tests/setup/skip-auth.setup.ts` which:

- **Authentication**: Uses `__SKIP_AUTH__` flag injected via beforeEach hook to bypass authentication
- **Data Loading**: Loads graph data through nginx proxy (http://localhost:8080/graph-service/api/v1/graph/full)
- **Backend URL**: Dev mode uses `http://localhost:9000`, Docker uses nginx proxy `http://localhost:8080`
- **Graph Mode**: Tests enable "Full Graph" toggle to ensure links are rendered
- **Timing**: Uses increased timeouts (8000ms) to ensure links are fully rendered before screenshot capture

## Troubleshooting

### Screenshots Not Showing Data

If screenshots appear stuck on login page or show no data:

1. **Check Docker Services:**
   ```bash
   docker-compose ps
   # Verify backend, graph-service, nginx are healthy
   ```

2. **Check Nginx Proxy:**
   ```bash
   curl http://localhost:8080/health
   curl http://localhost:8080/graph-service/api/v1/graph/full
   ```

3. **Verify SKIP_AUTH Mode:**
   - Check `tests/setup/skip-auth.setup.ts` injects `__SKIP_AUTH__` flag
   - Verify backend SkipAuth middleware is configured
   - Check frontend .env or backend .env for `SKIP_AUTH=true`

4. **Check Graph Service:**
   ```bash
   docker logs kg-graph-service
   curl http://localhost:9091/health
   curl http://localhost:9091/api/v1/graph/full
   ```

5. **Test Data Loading:**
   - Ensure "Full Graph" toggle is enabled in tests
   - Check increased timeouts are sufficient for link rendering
   - Verify graph service returns nodes with links data

### Argos Upload Failures

1. **Token Issues:**
   - Verify `ARGOS_TOKEN` is set in GitHub secrets
   - Check token has proper permissions

2. **Network Issues:**
   - Ensure CI runner can access Argos API
   - Check firewall/proxy settings

## Best Practices

1. **Consistent Test Data:**
   - Use the same test data generation for all visual tests
   - Avoid random data that causes false positives

2. **Timing and States:**
   - Ensure screenshots are captured after UI is fully loaded
   - Use `waitFor` for dynamic content

3. **Viewport Consistency:**
   - Use consistent viewport sizes across tests
   - Test responsive behavior in separate scenarios

## Resources

- [Argos Documentation](https://argos-ci.com/docs)
- [Playwright Visual Testing](https://playwright.dev/docs/screenshot-testing)
- [Project CI/CD Pipeline](../.github/workflows/main.yml)
