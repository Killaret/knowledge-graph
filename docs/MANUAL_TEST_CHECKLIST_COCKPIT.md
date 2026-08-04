# Cosmic Cockpit — Unified Manual Test Checklist

> Consolidates what used to be spread across:
> - `MANUAL_TEST_CHECKLISTS_RU.md` § 0 "Cosmic Cockpit Layout"
> - `MANUAL_TEST_FEEDBACK.md` Cases 1–4
>
> This is the **single source of truth** for Cosmic Cockpit manual testing.
> `MANUAL_TEST_FEEDBACK.md` keeps only the running **Findings log** (urgent
> fixes / roadmap / ideas); it links back here for the test steps themselves.
>
> **Environment:** `http://127.0.0.1:3002` (test stack, SKIP_AUTH=true unless noted)
> Use incognito + `Ctrl+F5` hard refresh before each run.

---

## 1. Layout & panels

**Steps:**
1. Open `/` in SKIP_AUTH mode.
2. Observe the layout without interacting.
3. Hover the left/right/bottom edges to auto-expand panels.
4. Click the pin icon on a panel, then unpin.

**Checks:**
- [ ] `CockpitFrame` (decorative border, corner bolts, grid+stars background) is visible around the canvas.
- [ ] Top panel (`CockpitTopPanel`) is pinned by default; contains search, filters, view switcher, create button.
- [ ] Left panel (`CockpitLeftPanel`) starts collapsed; hover/click expands it with animation.
- [ ] Right/bottom panel pull-handle (arrow) is visible **only while collapsed** — disappears once the panel is open.
- [ ] Pinning a panel keeps it open after mouse leaves; unpinning restores auto-collapse behavior.
- [ ] No panel corners overlap or clip each other at default window size.

---

## 2. Type filters & list/graph sync

**Steps:**
1. Create a note of type `comet` via API/UI, reload.
2. Expand left panel, click `Comet` filter.
3. Click `All`.

**Checks:**
- [ ] Filter click updates both graph and list; `graph-stats` counter reflects the filtered count (not the total).
- [ ] Switching back to `All` restores all notes.
- [ ] List view: cards render, sorting and search work.

---

## 3. View switching (List / 2D / 3D)

**Steps:**
1. From `/`, switch to `List`.
2. Switch back to `Graph`.
3. Switch to `3D`.

**Checks:**
- [ ] List view shows note cards with type stripe and hover glow.
- [ ] Returning to Graph restores the 2D canvas cleanly (no leftover DOM/canvas artifacts).
- [ ] 3D view loads without console errors.
- [ ] Selected node stays in sync across view switches.

---

## 4. 2D Graph — motion & readability

**Steps:**
1. Hard refresh `/`, let the graph settle (~3-5s).
2. Watch for 10-15s without interacting.
3. Zoom in/out.

**Checks — desync (per-node motion):**
- [ ] Star coronas do not blink in unison; different phase per node.
- [ ] Planet rings rotate at different speeds; some may reverse.
- [ ] Comet tails wag with different phases.
- [ ] Particles orbit in different directions/ellipses, not a uniform cloud.
- [ ] New/pulsing nodes do not pulse in sync.

**Checks — readability (post-hotfix, see `simulation.ts` force tuning):**
- [ ] Nodes are not stacked on top of each other.
- [ ] Links are distinguishable; graph does not look like a solid "hairball".
- [ ] Node labels are readable for most nodes.
- [ ] List → Graph switch re-renders cleanly and stays readable.
- [ ] No long freeze / dropped frames during warmup.
- [ ] No red console errors.

**If still not OK, note precisely what:** which element still syncs, or
whether nodes/links/labels are still unreadable (too dense / too sparse).

---

## 5. First-person mode

**Steps:**
1. Open bottom panel, click "First Person" toggle.
2. Observe UI.
3. Press `Esc`.
4. Re-enter first-person, click the visible exit button instead of `Esc`.

**Checks:**
- [ ] Entering first-person hides the cockpit panels/frame.
- [ ] An exit control is **always visible** (not only on hover), labeled with the `Esc` hint.
- [ ] `Esc` exits first-person mode; panels return.
- [ ] Clicking the exit button also exits first-person mode.
- [ ] HUD toggle button label switches between enter/exit text correctly.

---

## 6. Cockpit visual theme (Allotropic Carbon baseline)

> Today the cockpit re-implements a cyan/magenta neon look with **hardcoded
> rgba values** in several components instead of reusing the site-wide
> Allotropic Carbon theme (`--carbon-*`, `--color-info`, `--color-glow-purple`
> in `shared/styles/global.css`). Use this section to verify consistency
> before/after any theming refactor.

**Checks (current state):**
- [ ] `CockpitFrame` border gradient and corner bolts render (cyan → dark → magenta).
- [ ] `CockpitHUD` metrics text and sync indicator use the teal accent consistent with `NoteCard`/`Button` elsewhere in the app.
- [ ] No visually conflicting accent colors between cockpit components and the rest of the app (e.g. a different cyan hue, different glow intensity).

**Checks (after a theming refactor, if/when done):**
- [ ] Cockpit components consume shared CSS variables (aliases into `--carbon-*`/`--color-info`) rather than new hardcoded hex/rgba.
- [ ] Public/unauthenticated view is untouched — no `.cosmic-cockpit` class or cockpit variables leak outside the authenticated layout.
- [ ] `npm run check` and `npm run test:unit` pass after the refactor.
- [ ] Visual regression: toggling panels, first-person mode, and 2D/3D switching does not break the new styles.

---

## 7. Note detail page (reached from cockpit right panel)

**Steps:**
1. Click any note (or open `/notes/{id}` directly).

**Checks:**
- [ ] Emoji type label, public/private chip, dates visible.
- [ ] Tags/keywords are clickable, lead to `/search`.
- [ ] Connected notes section shows direction, link type icon, weight bar.
- [ ] Similar notes section shows title and score.
- [ ] Buttons: Edit, Delete, Create child, Constellation, 3D Constellation all work.
- [ ] Delete returns to `/` after success.
- [ ] No console errors.

---

## Known limitations (carry over from main checklist)

- Free node dragging on canvas is not supported (drag creates links, not moves nodes).
- `Ctrl+Z` is a placeholder — no real undo yet.
- Black hole glows on hover/drag but does not delete nodes on drop (not implemented).

---

## How findings flow from here

1. Run the relevant section above.
2. If something fails, add an entry to `docs/MANUAL_TEST_FEEDBACK.md` under
   **Urgent fixes / Roadmap items / Ideas** (see that file for the entry
   template), referencing the section number from this checklist (e.g.
   "Case: Cockpit §5").
3. Once fixed, come back and re-check the box here.
