# Manual Test Feedback — ai-agents branch

> Accumulator for issues found during manual testing of the test stack.
> When enough items pile up, we will triage them into:
> - **urgent fixes** (do now)
> - **roadmap** (planned feature work)
> - **ideas/backlog** (hypotheses / nice-to-have)
>
> **Environment:** `http://127.0.0.1:3002` (test stack, SKIP_AUTH=true)
> **Last rebuild:** see git log for `fix(frontend): desynchronize 2D graph node motion and particles` and `fix(frontend): reduce 2D graph overlaps and improve readability`

---

## Test Cases Queue

Run them in order. Mark each item `[x]` when it passes, or add a **Finding** below if it fails / feels wrong.

### Case 1: 2D Graph — no more synchronized motion

**Goal:** verify the fast-fix for per-node rotation/particles actually removed the "lockstep" effect.

**Steps:**
1. Open `http://127.0.0.1:3002/` in incognito.
2. Hard refresh (`Ctrl+F5`).
3. Wait for graph to settle.
4. Watch the 2D graph for 10–15 seconds without interacting.

**What to check:**
- [x] Stars do **not** all rotate with the same phase/corona rays do not blink in unison.
- [x] Planets with rings rotate at visibly different speeds (some may reverse).
- [x] Comet tails wag with different phases, not all at once.
- [x] Small orbiting particles around nodes move in different directions / ellipses, not a uniform cloud.
- [x] New/pulsing nodes (if any are present) do not pulse in perfect sync.
- [x] Console has **no red JS errors** from the app (browser extension messages can be ignored).

*Status: passed. User found the graph still felt like a tangled mess of overlapping nodes/links, so a readability hotfix was applied (see Case 1.5).*

**If something still looks synchronized, note which element exactly:**
- Star coronas
- Planet rings
- Comet tails
- New-note pulse rings
- Particles
- Other: ___________

---

### Case 1.5: 2D Graph — readability / reduced overlaps (hotfix verification)

**Goal:** verify the hotfix that increases repulsion, link distance, collision radius, initial spread and warmup time makes the 2D graph readable instead of a tangled blob.

**Steps:**
1. Hard refresh `/`.
2. Let the graph settle (~3–5 seconds).
3. Zoom in and out if supported.
4. Switch to `List` and back to `Graph`.

**What to check:**
- [ ] Nodes are generally not stacked on top of each other.
- [ ] Link lines are distinguishable; the graph does not look like a solid "hairball".
- [ ] Node labels are readable for most nodes (not fully obscured by others).
- [ ] After switching List → Graph, the layout re-renders cleanly and remains readable.
- [ ] Performance is acceptable: no long freeze, no dropped frames.
- [ ] Console has no red JS errors.

**If still unreadable, describe what exactly:**
- Nodes still overlap too much.
- Links overlap/cross too much.
- Labels unreadable.
- Graph explodes / too sparse.
- Graph collapses / too dense.
- Other: ___________

---

### Case 2: Cosmic Cockpit Layout — panels and frame

**Goal:** verify the cockpit frame and slide-out panels render correctly in SKIP_AUTH mode.

**Steps:**
1. Stay on `/`.
2. Observe the layout.
3. Click the left edge trigger / `☰` to expand the left panel.
4. Click type filters (`Comet`, `Planet`, etc.) and then `All`.

**What to check:**
- [ ] `CockpitFrame` (decorative border/screws) is visible around the canvas.
- [ ] Top panel is pinned and contains search, filters, view switcher, create button.
- [ ] Left panel starts collapsed and expands with animation.
- [ ] Filters update both the graph and the list view.
- [ ] `graph-stats` counter updates when a filter is active.
- [ ] No panel corners overlap in an ugly way.

---

### Case 3: View switch and 3D graph

**Goal:** verify List / 2D / 3D switching works.

**Steps:**
1. From `/`, switch to `List` view (top panel or hotkey).
2. Switch back to `Graph`.
3. Switch to `3D`.

**What to check:**
- [ ] List view shows note cards with type stripe and hover glow.
- [ ] Returning to Graph restores the 2D canvas.
- [ ] 3D view loads without console errors.
- [ ] 3D scene shows nodes/links in space.

---

### Case 4: Note detail page

**Goal:** verify `/notes/{id}` shows metadata, links, similar notes.

**Steps:**
1. Click any note (or visit `/notes/{id}` directly).
2. Inspect the page.

**What to check:**
- [ ] Emoji type label, public/private chip, dates visible.
- [ ] Tags / keywords clickable and lead to `/search`.
- [ ] Connected notes section shows direction, link type icon, weight bar.
- [ ] Similar notes section shows title and score.
- [ ] Buttons: Edit, Delete, Create child, Constellation, 3D Constellation.
- [ ] Delete returns to `/` after success.
- [ ] No JS errors in console.

---

## Findings

### Urgent fixes

<!-- Things that block release or make the app unusable. We fix these immediately. -->

- **Case:** 1 / follow-up
- **What:** 2D graph looks like a tangled mess of overlapping nodes and links; hard to read which node is which.
- **Expected:** Nodes and links are spaced out enough to distinguish individual elements.
- **Actual:** Everything overlaps, looks like a solid "hairball".
- **Hotfix applied:**
  - Increased initial placement circle radius from 30% to 45% of viewport.
  - Link distance 100 → `120 * densityFactor`, link strength 0.2 → 0.35.
  - Many-body charge -100 → `-180 * densityFactor`.
  - Collision radius 25 → `35 * densityFactor`.
  - Slower alpha decay (0.1 → 0.05) and 200 → 300 warmup ticks so the layout has time to spread.
- **Status:** container rebuilt, ready for re-test under Case 1.5.
- **Screenshot / Logs:** —

### Roadmap items

<!-- Real feature work that is understood and has clear value. -->

None yet.

### Ideas / backlog

<!-- Hypotheses, nice-to-haves, or "explore later" items. -->

None yet.

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
