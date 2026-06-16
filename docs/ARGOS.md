# Argos Visual Regression Testing

## Overview

This project uses [Argos](https://www.argos-ci.com/) for automated visual regression testing. Argos provides a free tier for public repositories, which is available for the `Killaret/knowledge-graph` repository.

## Setup

### Repository Status

- **Repository**: `Killaret/knowledge-graph`
- **Visibility**: Public ✅
- **Argos Plan**: Free tier (available for public repos)
- **Token**: `arp_grf4tj2faxgijlzrdpqzh6b6x9r3msqbq8uh`

### Configuration

**Frontend argos.json:**
```json
{
  "token": "ARGOS_TOKEN",
  "bucket": "Killaret/knowledge-graph",
  "branch": "origin/main",
  "baseBranch": "origin/main"
}
```

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

```bash
cd frontend
npm run test:visual
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

Visual tests use the test helper `tests/helpers/testData.ts` which:

- Uses backend URL: `http://127.0.0.1:9000`
- Creates test notes via API before screenshot capture
- Implements SKIP_AUTH mode for test authentication
- Generates consistent test data for reliable screenshots

## Troubleshooting

### Screenshots Not Showing Data

If screenshots appear stuck on login page or show no data:

1. **Check Backend Status:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Verify SKIP_AUTH Mode:**
   - Check `tests/setup/global-setup.ts` sets `process.env.SKIP_AUTH = 'true'`
   - Verify backend SkipAuth middleware is configured

3. **Test Data Creation:**
   - Ensure test notes are created successfully before screenshots
   - Check API connectivity: `http://127.0.0.1:9000`

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
