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
> **Last rebuild:** see git log for `fix(frontend): desynchronize 2D graph node motion and particles`, `fix(frontend): reduce 2D graph overlaps and improve readability`, `fix(frontend): improve first-person exit and panel handles`, and `feat(frontend): add Altered Carbon cockpit gradient text theme`

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

- **Case:** Cockpit §3 / §8 (3D view only shows labels)
- **What:** 3D graph loads but only HTML labels are visible; node spheres and links are not seen, so the scene looks like floating text on a black background.
- **Expected:** Nodes appear as colored geometry and links are visible; fog is light enough that the graph is not hidden.
- **Actual:** WebGL spheres were too dim (MeshStandardMaterial under distance lighting) and fog density was heavy enough to blend nodes into the background.
- **Hotfix applied:**
  - Switched node material to `MeshBasicMaterial` with the celestial body glow color so nodes are bright regardless of scene lighting.
  - Added a `fitScale` in `NodeManager` that scales node size with the graph bounding radius so nodes stay visible at the camera distance used by auto-zoom.
  - Reduced label background opacity and blur so spheres are not completely covered by label boxes.
  - Tuned fog presets in `knowledge-graph.config.json` and the default config (birth: 0.01 → 0.0006, nebula: 0.005 → 0.0003) so fog adds atmosphere without hiding geometry.
  - The engine now sets fog to the final light density immediately after the synchronous simulation warm-up.
- **Status:** container rebuilt, ready for re-test. See Cockpit §8 checklist.
- **Screenshot / Logs:** —

- **Case:** Cockpit §4 (isolated 2D nodes fly far away)
- **What:** Some disconnected or weakly connected nodes drift far from the main cluster, making the graph zoom out to a tiny center and forcing the user to zoom back in.
- **Expected:** All nodes remain within a comfortable viewport without extreme outliers.
- **Actual:** The 2D force simulation had no strong inward pull; the existing bounding force only kicked in at a large radius.
- **Hotfix applied:**
  - Tightened the 2D bounding radius from `0.55` to `0.4` of the smaller viewport dimension.
  - Increased the bounding-force strength to 1.0 so outliers are pulled back more quickly.
- **Status:** container rebuilt, ready for re-test. Dense center clusters are still a readability issue until semantic clustering is implemented.
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

- **Case:** Cockpit §1 (panel handles, left/top — found while implementing gradient text)
- **What:** The right/bottom handle fix above only covered the right/bottom handle block. `CockpitPanel.svelte` renders a **second**, separate handle block for left/top panels that still lacked the `!isOpen` guard — same bug, different code path.
- **Expected:** Left/top pull-handles also disappear once their panel is open.
- **Actual:** Left/top arrows stayed visible even when expanded.
- **Fix applied:** `{#if position === "left" || position === "top"}` → `{#if (position === "left" || position === "top") && !isOpen}`, and dropped the now-dead `class:open={isOpen}` binding to match the right/bottom block.
- **Regression test:** `CockpitPanel.spec.ts` — `shows the pull-handle for %s when collapsed` / `hides the pull-handle for %s once the panel is open`, parametrized over all four positions (`right`, `bottom`, `left`, `top`).
- **Status:** fixed, covered by unit tests, container rebuilt.
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

- **Case:** Cockpit §6 (Altered Carbon theming — implemented)
- **What:** Original proposal wanted a distinct "Altered Carbon" neon look via new `--cockpit-*` CSS variables, plus static two-color gradient text.
- **Review notes (before implementing):**
  - The project already has an "Allotropic Carbon" theme (`--carbon-*`, `--color-info`, `--color-glow-purple` in `shared/styles/global.css`) applied site-wide — a parallel `--cockpit-*` set would have duplicated it.
  - `CockpitFrame.svelte` already implements grid + stars + corner bolts + cyan/magenta border gradient, just with hardcoded rgba instead of variables.
  - Proposed target components `CosmicHUD`/`CosmicNotificationCenter`/`CosmicToast` do not exist under those names; actual files are `CockpitHUD.svelte` and the app-wide `widgets/notification/ToastNotification.svelte` (not cockpit-scoped, left untouched).
- **What was implemented** (revised prompt, `docs/PROMPT_COCKPIT_ALTERED_CARBON_THEME.md`, now removed after landing):
  - `shared/styles/global.css`: cockpit-scoped aliases (`--cockpit-accent` → `--color-info`, `--cockpit-accent-2` → `--color-comet`, `--cockpit-gradient-text` → `--carbon-gradient-primary`, etc.) — no new hues.
  - `.cockpit-gradient-text` utility class: animated moving gradient text via `background-position` + `@keyframes cockpit-gradient-shift`, phase offset via `--cockpit-text-delay` (desynchronized per instance, not lockstep). `.cockpit-gradient-text--static` modifier for data-critical text. `prefers-reduced-motion` disables the animation.
  - Applied animated gradient text to decorative headings: `CockpitPanel` panel titles (delay per position), `CockpitFirstPersonButton` exit label, `CockpitLeftPanel` section headings (delay per section).
  - Applied **static** gradient text only to `CockpitHUD`'s cluster name; numeric metrics (notes/links/FPS/health/sync) stay plain solid-color text for readability.
  - Replaced hardcoded `#2dd4bf`/`rgba(45, 212, 191, ...)` with `var(--cockpit-accent, ...)` in `CockpitPanel`, `CockpitFirstPersonButton`, `CockpitHUD`, and `CockpitFrame`'s corner bolts.
- **Regression tests:** `CockpitPanel.spec.ts`, `CockpitFirstPersonButton.spec.ts`, `CockpitHUD.spec.ts` (20 new tests, all passing). `npm run check` — 0 errors. Full `npm run test:unit` — 858/858 passing.
- **Status:** implemented, tested, checklist updated (Cockpit §6). Ready for manual re-test.
- **Screenshot / Logs:** —

---

## Verification

- **Case:** A-1 / 3D visual regression sensitivity
- **What:** Verified that the 3D visual test now produces a stable, deterministic element screenshot and that changing the fog `density_final` produces a visibly different image.
- **Expected:** Two consecutive runs with the same config produce identical 1280×720 crops; raising `birth.density_final` from 0.0006 to 0.02 produces a clearly different frame.
- **Actual:** Baseline and second-run crops are 1240×680 and pixel-identical across runs. The dense-fog crop differs by 114 528 / 843 200 pixels (13.58 %). Differences are concentrated in the background fog and label dimming; node positions remain identical.
- **Screenshot / Logs:** `docs/assets/a1-3d-visual-regression/3d-baseline.png`, `docs/assets/a1-3d-visual-regression/3d-fog-dense.png`, `frontend/test-results/visual-visual-regression-V-df9ff--3D-Graph---renders-3D-view-visual/argos/visual/3d-graph-view.png`; `npm run test:unit` 987/987; `npm run check` clean.

---

## Verification

- **Case:** AUD-4 / Yandex OAuth login route contract
- **What:** Verified that `/api/v1/auth/yandex/login` returns JSON `{"url": "..."}` with all required OAuth parameters, that the old `/api/v1/auth/yandex` path is no longer registered, and that `YANDEX_CLIENT_ID` is wired through to the test backend container.
- **Expected:** Old path (`/api/v1/auth/yandex`) returns an error or 404/401; new path returns `200` with a URL pointing to `https://oauth.yandex.com/authorize` and containing `client_id`, `response_type=code`, `state`, `scope`, `code_challenge` and `code_challenge_method=S256`.
- **Actual:** Before fix: `GET /api/v1/auth/yandex/login` returned `404 Not Found`; `GET /api/v1/auth/yandex` returned `501 Not Implemented` (no client id configured). After fix and rebuild with `YANDEX_CLIENT_ID=test-yandex-client-id`: `GET /api/v1/auth/yandex/login` returns `200 OK` and JSON `{"url":"https://oauth.yandex.com/authorize?client_id=test-yandex-client-id&code_challenge=...&code_challenge_method=S256&response_type=code&scope=login%3Aemail+login%3Ainfo&state=..."}`. `GET /api/v1/auth/yandex` returns `401` (route removed, falls through to JWT middleware).
- **Screenshot / Logs:** `curl -s -D - http://127.0.0.1:18083/api/v1/auth/yandex/login` (before: `404`, after: `200`); `go test ./...` green; `go vet ./...` clean; `npm run test:unit` 987/987; `npm run check` 0 errors; `npm run lint` no new warnings.

## Verification

- **Case:** Playwright real-auth setup for visual regression
- **What:** Verified that `tests/setup/auth.setup.ts` can log in as the seeded `testuser` and persist a `storageState` file, and that the new `visual-real-auth` project runs `@visual` tests against the real-auth test stack.
- **Expected:** `setup-auth` project completes; `visual-real-auth` loads the authenticated page and a sample `@visual` test passes without `__SKIP_AUTH__` injection.
- **Actual:**
  - `npx playwright test --project=setup-auth` → `1 passed` in 6.3s.
  - `npx playwright test --project=visual-real-auth --grep "Home page - default view"` → `2 passed` (setup + test) in 9.8s.
  - `frontend/tests/setup/.auth/testuser.json` was created and used by the `visual-real-auth` project.
- **Screenshot / Logs:** Playwright `line` reporter output for `setup-auth` and `visual-real-auth` projects.

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
