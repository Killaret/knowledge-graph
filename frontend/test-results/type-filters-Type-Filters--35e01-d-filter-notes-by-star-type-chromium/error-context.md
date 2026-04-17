# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: type-filters.spec.ts >> Type Filters - Home Page Filtering >> should filter notes by star type
- Location: tests\type-filters.spec.ts:18:3

# Error details

```
Error: locator.click: Target page, context or browser has been closed
Call log:
  - waiting for locator('button:has-text("⭐"), button:has-text("Stars"), button[data-filter="star"]').first()
    - locator resolved to <button title="Stars" class="filter-chip  s-B7JUFFnK4lnr">…</button>
  - attempting click action
    - waiting for element to be visible, enabled and stable

```

```
Error: browserContext.close: Target page, context or browser has been closed
```