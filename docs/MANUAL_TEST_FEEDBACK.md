# Manual Test Feedback — ai-agents branch

> Accumulator for issues found during manual testing of the test stack.
> When enough items pile up, we will triage them into:
> - **urgent fixes** (do now)
> - **roadmap** (planned feature work)
> - **ideas/backlog** (hypotheses / nice-to-have)
>
> **Test steps live in [`MANUAL_TEST_CHECKLIST_COCKPIT.md`](MANUAL_TEST_CHECKLIST_COCKPIT.md)**
> (layout/panels, filters, view switching, 2D motion/readability, first-person,
> visual theme, note detail page). This file only tracks the **findings log**
> below — add an entry here when a checklist item fails, referencing the
> section number (e.g. "Cockpit §4").
>
> **Environment:** `http://127.0.0.1:3002` (test stack, SKIP_AUTH=true)
> **Last rebuild:** see git log for `fix(frontend): desynchronize 2D graph node motion and particles`, `fix(frontend): reduce 2D graph overlaps and improve readability`, and `fix(frontend): improve first-person exit and panel handles`

---

## Findings

### Urgent fixes

<!-- Things that block release or make the app unusable. We fix these immediately. -->

- **Case:** Cockpit §4 (2D Graph readability)
- **What:** 2D graph looks like a tangled mess of overlapping nodes and links; hard to read which node is which.
- **Expected:** Nodes and links are spaced out enough to distinguish individual elements.
- **Actual:** Everything overlaps, looks like a solid "hairball".
- **Hotfix applied:**
  - Increased initial placement circle radius from 30% to 45% of viewport.
  - Link distance 100 → `120 * densityFactor`, link strength 0.2 → 0.35.
  - Many-body charge -100 → `-180 * densityFactor`.
  - Collision radius 25 → `35 * densityFactor`.
  - Slower alpha decay (0.1 → 0.05) and 200 → 300 warmup ticks so the layout has time to spread.
- **Status:** container rebuilt, ready for re-test under Cockpit §4.
- **Screenshot / Logs:** —

- **Case:** Cockpit §5 (first-person mode)
- **What:** There is no obvious way to exit first-person mode: the floating exit button is invisible until you hover it, there is no Esc hotkey, and no hint about how to get out.
- **Expected:** User always sees how to exit first-person mode, can use Esc, and gets a clear label on the exit control.
- **Actual:** Once first-person is active, UI panels disappear and the exit button is hidden; users feel trapped.
- **Hotfix applied:**
  - `CockpitFirstPersonButton` is now always visible and shows "Exit first-person (Esc)".
  - `CosmicCockpitLayout` listens to the `Escape` key and exits first-person mode.
  - HUD button still toggles first-person; its own label already changes between enter/exit.
- **Status:** container rebuilt, ready for re-test.
- **Screenshot / Logs:** —

- **Case:** Cockpit §1 (panel handles)
- **What:** After auto-expanding the right/bottom panel, the arrow handle stays visible even though the panel is already open.
- **Expected:** The pull-handle (arrow) is only visible when the panel is collapsed.
- **Actual:** Arrows remain on screen when the panel is expanded, creating visual noise.
- **Hotfix applied:** `CockpitPanel` now renders the handle only when `!isOpen`.
- **Status:** container rebuilt, ready for re-test.
- **Screenshot / Logs:** —

### Roadmap items

<!-- Real feature work that is understood and has clear value. -->

None yet.

### Ideas / backlog

<!-- Hypotheses, nice-to-haves, or "explore later" items. -->

- **Case:** Cockpit §1 (visual depth)
- **What:** Cockpit panels look flat 2D. User wants a 2.5D / pseudo-3D effect where panel ends look voluminous, as if they really go into the screen depth, increasing immersion in the canvas.
- **Expected:** Panels have subtle 3D extrusion / bevel / perspective so they feel like physical surfaces receding into space.
- **Actual:** Panels are flat rectangles.
- **Proposed idea:**
  - Add CSS `perspective` to the cockpit container and small `rotateX`/`rotateY` transforms on panel bodies.
  - Use layered borders with gradient shadows on panel edges to simulate thickness.
  - Animate the depth subtly on hover/focus.
  - Keep it optional and ensure it does not hurt click targets or accessibility.
- **Triage:** backlog / roadmap candidate (Phase 11 or UI-polish phase). Not a bug.
- **Screenshot / Logs:** —

- **Case:** Cockpit §6 (Altered Carbon theming)
- **What:** Proposal to give cockpit components a distinct "Altered Carbon" neon look via new `--cockpit-*` CSS variables.
- **Review notes:**
  - The project already has an "Allotropic Carbon" theme (`--carbon-*`, `--color-info`, `--color-glow-purple` in `shared/styles/global.css`) applied site-wide — a parallel `--cockpit-*` set would duplicate it.
  - `CockpitFrame.svelte` already implements grid + stars + corner bolts + cyan/magenta border gradient, just with hardcoded rgba instead of variables.
  - Proposed target components `CosmicHUD`/`CosmicNotificationCenter`/`CosmicToast` do not exist under those names; actual files are `CockpitHUD.svelte` and the app-wide `widgets/notification/ToastNotification.svelte` (not cockpit-scoped).
- **Proposed approach:** refactor cockpit components to consume existing shared variables (or aliases into them) instead of hardcoded rgba, rather than introducing a second color system. See Cockpit §6 checklist for what to verify once this is attempted.
- **Triage:** roadmap candidate, needs a scoped-down follow-up prompt before implementation.
- **Screenshot / Logs:** —

---

## How to add a finding

Create a new bullet under the right section with:

```markdown
- **Case:** 1 / 2 / 3 / 4 / other
- **What:** concise description
- **Expected:** what you expected
- **Actual:** what happened
- **Screenshot / Logs:** attach if possible
- **Proposed fix / idea:** optional
```
