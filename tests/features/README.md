# BDD Feature Tests

This directory contains Gherkin feature files for Behavior-Driven Development (BDD) testing of the Knowledge Graph application.

## Structure

```
tests/features/
├── README.md                      # This file
├── graph_navigation.feature       # Core graph interaction scenarios
├── import_export.feature          # Import/export functionality
├── graph_view.feature             # 2D/3D view mode switching
├── note_management.feature        # Note CRUD operations
├── search_and_discovery.feature   # Search and discovery features
├── step_definitions/              # TypeScript step implementations
│   ├── graph_steps.ts
│   ├── navigation_steps.ts
│   └── common_steps.ts
└── support/                       # Test support files
    ├── hooks.ts                   # Before/After hooks
    └── world.ts                 # Custom World type
```

## Running Tests

### Prerequisites

Install Cucumber and dependencies:

```bash
npm install --save-dev @cucumber/cucumber
```

### Run All Features

```bash
npx cucumber-js
```

### Run Specific Feature

```bash
npx cucumber-js tests/features/graph_navigation.feature
```

### Run with Tags

```bash
# Run only smoke tests
npx cucumber-js --tags "@smoke"

# Run all except slow tests
npx cucumber-js --tags "not @slow"

# Run smoke OR api tests
npx cucumber-js --tags "@smoke or @api"
```

### Environment Variables

```bash
# Set base URL
BASE_URL=http://localhost:5173 npx cucumber-js

# Run in headed mode (visible browser)
HEADLESS=false npx cucumber-js

# Slow down execution for debugging
SLOW_MO=1000 npx cucumber-js

# Set default timeout
DEFAULT_TIMEOUT=60000 npx cucumber-js
```

## Feature Tags

Available tags for organizing tests:

- `@smoke` - Critical path tests, run on every build
- `@regression` - Full regression suite
- `@slow` - Long-running tests (excluded from quick runs)
- `@visual` - Tests requiring visual verification
- `@api` - API/integration tests
- `@e2e` - End-to-end tests

## Writing New Scenarios

### Feature File Template

```gherkin
Feature: Feature Name
  As a [role]
  I want [feature]
  So that [benefit]

  Background:
    Given the application is open
    And [common setup]

  @smoke
  Scenario: Basic scenario
    Given [initial state]
    When [action]
    Then [expected result]

  Scenario Outline: Parametrized scenario
    Given I have <count> notes
    When I delete one
    Then I should have <remaining> notes

    Examples:
      | count | remaining |
      | 5     | 4         |
      | 3     | 2         |
```

### Step Definition Pattern

```typescript
import { Given, When, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import type { ITestWorld } from '../support/world';

Given('the application is open', async function(this: ITestWorld) {
  await this.page.goto('/');
  await expect(this.page.locator('.graph-canvas')).toBeVisible();
});

When('I click {string}', async function(this: ITestWorld, text: string) {
  await this.page.click(`text="${text}"`);
});

Then('I should see {string}', async function(this: ITestWorld, text: string) {
  await expect(this.page.locator(`text="${text}"`)).toBeVisible();
});
```

## Best Practices

1. **Use Background** for common setup across scenarios
2. **Tag appropriately** with @smoke, @slow, etc.
3. **Keep scenarios independent** - no shared state
4. **Use data tables** for multiple related inputs
5. **Scenario Outlines** for data-driven tests
6. **Descriptive names** that explain the business value

## Troubleshooting

### Timeout Issues

Increase timeout in cucumber.mjs or via env var:
```bash
DEFAULT_TIMEOUT=60000 npx cucumber-js
```

### Screenshot on Failure

Screenshots are automatically captured on failure and saved to `test-results/screenshots/`.

### Debug Mode

Run with headed browser and slow motion:
```bash
HEADLESS=false SLOW_MO=1000 npx cucumber-js
```

## CI/CD Integration

Example GitHub Actions step:

```yaml
- name: Run BDD Tests
  run: npm run test:cucumber
  env:
    BASE_URL: http://localhost:4173
    HEADLESS: true
```
