# Manual Testing Checklists

This document contains checklists for manually verifying new Knowledge Graph functionality.

## Pre-Testing Setup

### Check Stacks Health
- [ ] Run `scripts/check-stacks-health.ps1` (Windows) or `scripts/check-stacks-health.sh` (Linux/Mac) → Health check script executes and reports status
- [ ] Verify dev stack is healthy (http://localhost:8080/health) → Dev stack returns healthy status
- [ ] Verify personal stack is healthy (http://localhost:8082/health) → Personal stack returns healthy status
- [ ] Verify dev API is accessible (http://localhost:8080/api/v1/notes?limit=1) → Dev API returns valid response
- [ ] Verify personal API is accessible (http://localhost:8082/api/v1/notes?limit=1) → Personal API returns valid response
- [ ] If any checks fail, start the affected stack before proceeding → Affected stack becomes healthy

### Start Test Stack
- [ ] Run `scripts/start-test.ps1` (Windows) or `scripts/start-test.sh` (Linux/Mac) → Test stack containers start and become healthy
- [ ] Wait for all test containers to be healthy → All containers show "healthy" status
- [ ] Verify test stack is ready (http://localhost:3002) → Frontend loads successfully
- [ ] Verify test backend API is accessible (http://localhost:8083) → API returns valid responses

### Seed Test Data
- [ ] Run `scripts/seed-test-data.ps1` (Windows) or `scripts/seed-test-data.sh` (Linux/Mac) → Test data creation script executes successfully
- [ ] Verify test user is registered (login: testuser, password: TestPassword123!) → User can login with these credentials
- [ ] Verify 5 test notes are created (star, planet, comet, galaxy, asteroid) → Notes appear in the system with correct types
- [ ] Verify 2 test links are created between notes → Links are visible on the graph connecting the notes

### Quick Start (Full Cycle)
- [ ] Run `scripts/run-full-test-cycle.ps1` (Windows) or `scripts/run-full-test-cycle.sh` (Linux/Mac) → Script executes full test cycle automatically
- [ ] This will automatically: check stacks health, start test stack, seed data, wait for manual testing, destroy test stack, verify stacks health again → All steps complete without errors

### Regression Test Plan (Optional)
- [ ] Review REGRESSION_TEST_PLAN.md for comprehensive testing procedures
- [ ] Run full regression cycle if deploying to production or after major changes
- [ ] Verify all regression test checkpoints pass before proceeding to manual testing

---

## Smoke Tests

### Public Access
- [ ] Open http://localhost:3002 in browser without authentication → Canvas loads successfully with public graph visible
- [ ] Verify no authentication errors appear → Page loads cleanly, no 401/403 errors

### Authentication Flow
- [ ] Click "Login" button → Login form appears with username/password fields
- [ ] Enter testuser / TestPassword123! → Successfully authenticated and redirected to main view
- [ ] Verify user profile is accessible → Profile displays user information correctly

### Graph Display
- [ ] Navigate to graph view → Graph canvas displays with nodes and edges
- [ ] Verify Knowledge Core node is visible → Central node appears and cannot be deleted
- [ ] Verify basic graph interactions work → Zoom, pan, and node selection function correctly

### Note Creation
- [ ] Press `N` or click empty canvas area → Ghost node creation form appears
- [ ] Enter title "Test Note" and confirm → New note appears on canvas with "New" indicator
- [ ] Refresh page → Note persists and indicator remains

### Logout
- [ ] Click logout button → Successfully logged out and redirected to login page
- [ ] Verify session is cleared → Cannot access protected endpoints without re-authentication

---

## Public Graph Connection Verification

### Create Public Notes (Without Auth)
- [ ] Open http://localhost:3002 in browser → Canvas loads without authentication
- [ ] Use API to create public note: `POST http://localhost:8083/api/v1/notes` with `{"title": "Public Test", "content": "Test content", "is_public": true}` → Note created successfully (201)
- [ ] Verify note appears on public graph → New node visible on canvas without login

### Create Public Links (Without Auth)
- [ ] Create two public notes via API → Both notes created successfully
- [ ] Create link between them: `POST http://localhost:8083/api/v1/links` with `{"source_id": "...", "target_id": "...", "type": "related"}` → Link created successfully (201)
- [ ] Verify link appears on public graph → Edge visible between nodes on canvas

### Verify API Access Without Auth
- [ ] Call `GET http://localhost:8083/api/v1/graph` without Authorization header → Returns public graph data (200)
- [ ] Call `GET http://localhost:8083/api/v1/notes?is_public=true` without auth → Returns only public notes (200)
- [ ] Verify private notes are not accessible → Private notes excluded from response

### Verify Frontend Public Access
- [ ] Load http://localhost:3002 in incognito/private window → Canvas loads with public graph
- [ ] Verify only public nodes/edges are visible → Private content not displayed
- [ ] Try to access protected routes (e.g., profile) → Redirected to login or shows 401 error

---

## Canvas Features

### Ghost Node Creation
- [ ] Click on empty canvas area / use hotkey `N` to activate ghost-node creation → Ghost node form appears with backdrop blur and gradient styling
- [ ] Fill in the title and confirm → New star/planet node appears on the canvas
- [ ] Refresh the page → New note persists and still exists
- [ ] Verify the node has a turquoise "New" indicator/glow → Turquoise outline/glow labelled "New" is visible on the node

### Black Hole Deletion (with Undo)
- [ ] Select an existing note node → Node becomes highlighted/selected
- [ ] Drag the selected node onto the black-hole anomaly → Node disappears from the canvas
- [ ] Verify a toast appears with an "Undo" action → Toast shows "Done" then "Restore" button after 1.5s
- [ ] Click "Undo" and verify the node reappears → Node returns to its previous position on canvas
- [ ] Drag again and dismiss the toast; verify the note is deleted after refresh → Note is permanently deleted after toast dismissal
- [ ] Open note side panel and verify "Delete all links" button exists → Button is visible in link management section
- [ ] Click "Delete all links" and confirm via modal → Confirmation modal appears
- [ ] Confirm deletion → All links are removed from the note

### Drag-and-Drop Links
- [ ] Hover over a source node to see the link creation affordance → Link handle becomes visible on the node
- [ ] Drag the link handle to a target node → Visual preview link appears while dragging
- [ ] Release on target node → Link form appears with enhanced styling
- [ ] Fill in the link type and weight → Form accepts input and validates
- [ ] Confirm link creation → New link/edge is drawn between the nodes
- [ ] Try to create the same link again → "Duplicate link" warning/toast appears
- [ ] Refresh the page → Link persists and remains visible

### Hotkeys
- [ ] Press `N` → Ghost-node creation activates
- [ ] Press `F` → Search/focus input appears
- [ ] Press `?` → Help modal with hotkeys opens
- [ ] Press `Esc` → Any open modal/tooltip closes
- [ ] Press `Ctrl+Z` → Undo last action (if implemented)
- [ ] Press `Delete`/`Backspace` → Delete selected node (if implemented)

### Knowledge Core (Technical Note)
- [ ] Verify the central "Knowledge Core" node is always visible → Central node remains visible at all times
- [ ] Click the Knowledge Core node → Node becomes selected
- [ ] Try to drag it to black hole → Node cannot be deleted via black hole
- [ ] Hover over the node → Tooltip shows system information

### Tooltips
- [ ] Hover over any node → Tooltip with title, type, and action buttons appears
- [ ] Hover over any link → Tooltip with link type and weight appears

### New Indicator
- [ ] Create a fresh note → Node has a turquoise outline/glow labelled "New"
- [ ] Wait/perform another action → Indicator disappears after the note is no longer "new"

---

## Note Cards (List View)

### Visual Style
- [ ] Switch to list view → List view displays with note cards
- [ ] Verify each card has a colored accent strip matching its type → Color strip corresponds to note type (star, planet, comet, galaxy, asteroid)
- [ ] Verify the type icon is displayed → Type emoji/icon appears in top-left corner
- [ ] Verify a subtle glow/hover effect is present → Card lifts and glows on hover

### Dust Style
- [ ] Create a note with type `dust` → Card uses the special "dust" visual treatment (faded particles/dust background)

### Card Tooltip
- [ ] Hover over a card → Tooltip shows keywords, related links count, and quick-action buttons

### Batch Operations
- [ ] Toggle selection mode → Selection mode activates with checkboxes visible
- [ ] Select multiple notes using checkboxes → Selected notes are highlighted
- [ ] Click "Delete selected" → Only selected notes are removed and Undo toast appears
- [ ] Verify the Undo toast shows "Done" first, then "Restore" after 1.5s → Toast transitions from "Done" to "Restore" button
- [ ] Click "Restore" within 5s → Notes reappear in the list

### Undo Deletion
- [ ] Delete a note from the list view → Toast says "Done" (Galactic Lexicon) with a "Restore" button
- [ ] Click "Restore" → Note reappears in the list

### Sorting
- [ ] Open the sort dropdown → Sort options are displayed
- [ ] Sort by created date ascending → Notes ordered by oldest first
- [ ] Sort by created date descending → Notes ordered by newest first
- [ ] Sort by title A-Z → Notes ordered alphabetically
- [ ] Sort by title Z-A → Notes ordered reverse alphabetically
- [ ] Sort by updated date → Notes ordered by most recently updated
- [ ] Verify the order updates correctly → List reorders immediately after selection

### Staggered Fade-In Animation
- [ ] Load a page with many notes → Cards appear with a staggered fade-in animation, not all at once

### Empty State
- [ ] Clear all notes (or filter to a non-matching query) → Friendly empty-state illustration/message is shown

---

## General UX

### Galactic Lexicon
- [ ] Trigger success actions (create note, delete note, etc.) → Toast messages use cosmic-themed wording (e.g., "Star born", "Matter collapsed", "Nebula restored")
- [ ] Verify no raw/default browser messages appear → All messages use custom Galactic Lexicon styling

### Browser Console
- [ ] Open browser DevTools console → Console panel opens
- [ ] Navigate through graph and list views → No red errors appear in the console
- [ ] Verify warnings are acceptable → Only acceptable warnings appear (e.g., Svelte a11y warnings)

### Language Switch
- [ ] Open profile/settings → Profile/settings panel opens
- [ ] Change the language toggle → UI labels, placeholders, and toast messages switch language
- [ ] Reload the page → Selected language persists

---

## Post-Testing Cleanup

### Stop Test Stack
- [ ] Run `scripts/stop-test.ps1` (Windows) or `scripts/stop-test.sh` (Linux/Mac) → Test stack is destroyed (containers stopped and volumes removed)
- [ ] Verify no test containers are running (`docker ps --filter "name=kg-test"`) → No test containers appear in the list

### Verify Stacks Health
- [ ] Run `scripts/check-stacks-health.ps1` (Windows) or `scripts/check-stacks-health.sh` (Linux/Mac) → Dev and personal stacks health checks execute
- [ ] Verify dev stack is still healthy after testing → Dev stack returns healthy status
- [ ] Verify personal stack is still healthy after testing → Personal stack returns healthy status
- [ ] Verify no data leakage from test stack to dev/personal stacks → No test data appears in dev/personal stacks

### Cleanup Test Data (if needed)
- [ ] If test data was created in dev/personal stacks, manually delete test notes → Test notes are removed from dev/personal stacks
- [ ] Delete test user if created in dev/personal stacks → Test user is removed from dev/personal stacks
- [ ] Verify no test artifacts remain in production stacks → No test data or users remain in production stacks

---

## Reporting Defects

For each failed item:
- [ ] Record the environment (browser, stack) → Document browser version, OS, and which stack was tested
- [ ] Attach a screenshot or short screen recording → Capture visual evidence of the issue
- [ ] Copy relevant console/backend logs → Save error messages and stack traces
- [ ] File an issue with the checklist item reference → Create GitHub issue referencing the specific checklist item that failed

### Issue Template

When filing issues, include:
- **Checklist Item:** Reference the specific section and item (e.g., "Canvas → Ghost Node Creation → Item 3")
- **Environment:** Browser, OS, stack (dev/personal/test)
- **Steps to Reproduce:** Exact actions taken
- **Expected Result:** What should have happened
- **Actual Result:** What actually happened
- **Screenshots/Logs:** Attach visual evidence and error logs
