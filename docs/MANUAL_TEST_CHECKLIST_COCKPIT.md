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
- [ ] Pull-handle (arrow) on **every** edge — left, right, top, bottom — is visible **only while collapsed** and disappears once that panel is open (covers all four positions, not just right/bottom).
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
4. Switch back to `Graph`.
5. Switch to `3D` again.

**Checks:**
- [ ] List view shows note cards with type stripe and hover glow.
- [ ] Returning to Graph restores the 2D canvas cleanly (no leftover DOM/canvas artifacts).
- [ ] 3D view loads without console errors and shows visible nodes/colored spheres, not only labels.
- [ ] Links are still present after each 2D ↔ 3D round-trip.
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
- [ ] Isolated or sparsely connected nodes do not fly far away from the rest of the graph.
- [ ] List → Graph switch re-renders cleanly and stays readable.
- [ ] The fog button (cloud icon) toggles the vignette visibly and the canvas redraws immediately.
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

## 6. Cockpit visual theme — animated gradient text (Allotropic Carbon aliases)

> Implemented in `feat(frontend): add Altered Carbon cockpit gradient text
> theme` — reuses `--carbon-*`/`--color-info`/`--color-comet` through
> cockpit-scoped aliases in `shared/styles/global.css`; no new palette was
> introduced. Covered by unit tests in `CockpitPanel.spec.ts`,
> `CockpitFirstPersonButton.spec.ts`, `CockpitHUD.spec.ts`. See
> `docs/MANUAL_TEST_FEEDBACK.md` (Cockpit §6) for the implementation summary
> and review notes on the original proposal.

**Steps:**
1. Open `/` in SKIP_AUTH mode, hard refresh.
2. Watch panel titles (top/bottom/left/right) and the left panel's section
   headings ("Navigation", "Graph Controls", "Filters", etc.) for ~6-10s.
3. Enter first-person mode and look at the exit button label.
4. Check the HUD's cluster name and numeric metrics (notes/links/FPS/health).
5. Enable "Reduce motion" at the OS level (or `prefers-reduced-motion` via
   DevTools rendering emulation) and reload.

**Checks — decorative headings (should shimmer, desynchronized):**
- [ ] Panel titles (`CockpitPanel`) show a moving cyan→violet gradient text, not a static two-color label.
- [ ] Different panels' titles are visibly **out of phase** with each other (not all brightening/shifting at the same time).
- [ ] Left panel's section headings shimmer with their own offsets, not in lockstep with each other or with the panel titles.
- [ ] First-person exit button label ("Exit first-person (Esc)") also has the moving gradient.

**Checks — data-critical HUD text (should stay legible, not distractingly animated):**
- [ ] Cluster name in the HUD has a gradient look but does **not** move/shimmer (static gradient).
- [ ] Numeric metrics (notes, links, FPS, health%, sync status) render as plain solid-color text — easy to read at a glance, no gradient/animation.

**Checks — accessibility & consistency:**
- [ ] With reduced-motion enabled, all `cockpit-gradient-text` elements stop animating (frozen at one gradient position) instead of disappearing or breaking layout.
- [ ] Text remains readable (sufficient contrast against the panel background) at every point during the animation, not just at the gradient's end stops.
- [ ] No new hue was introduced — accents visually match the teal/violet already used by `NoteCard`/`Button` elsewhere in the app (Allotropic Carbon theme).
- [ ] Public/unauthenticated view is untouched — no gradient text or `.cosmic-cockpit` variables leak outside the authenticated cockpit layout.
- [ ] `npm run check` and `npm run test:unit` pass.

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

## 8. 3D view — node visibility, links, and fog

> This section covers the recent fix where 3D showed only labels and no geometry.

**Steps:**
1. From the 2D graph view, click **3D** in the top bar.
2. Wait for the spinner to disappear (should be <2s).
3. Rotate the scene by dragging; zoom with the wheel.
4. Hover a node and observe its label.
5. Click the fog button (cloud icon) in the top bar twice — off, then on.
6. Click **2D** to return to the graph.
7. Click **3D** again.

**Checks:**
- [ ] 3D loads without a long blank/black screen and without console errors.
- [ ] Colored spheres (nodes) are visible, not just text labels floating in space.
- [ ] Lines (links) are visible between related nodes, at least when zoomed in.
- [ ] The graph can be rotated, panned, and zoomed with mouse/touch.
- [ ] Hovering a node brightens the node and its neighbors; distant nodes are subdued.
- [ ] Toggling fog off/on changes the canvas/background visibly.
- [ ] Switching 2D → 3D → 2D → 3D does not lose links or crash the view.
- [ ] The top-bar controls (reset, search, focus, fog, language) do not overlap each other.

**If still unreadable:**
- Note whether the problem is (a) nodes too small to see, (b) spheres too big and overlapping, (c) labels overlapping, or (d) the center cluster is too dense. Layout readability for 50+ random notes is only solvable by semantic clustering; the force layout alone cannot guarantee a clean cluster.

---

## 9. Graph top bar control spacing and fog toggle

> This section covers the recent fix where the fog button overlapped neighboring controls.

**Steps:**
1. Open the graph at the default window size (≈1280–1920 px wide).
2. Look at the top bar to the right of the view switcher (2D/3D/List).
3. Click the fog button and the language switcher.
4. Resize the window to 768 px wide and repeat.

**Checks:**
- [ ] Reset, search, focus, and fog buttons are separate, non-overlapping 34×34 targets.
- [ ] The fog button does not touch or overlap the language switcher or the right edge.
- [ ] Fog toggle is reachable and its pressed/unpressed state is visible.
- [ ] At 768 px, the top bar still fits without wrapping or clipping.

---

## 10. Responsive / mobile (pending)

> Mobile touch handling is on the todo list; use this section to record early observations.

**Steps:**
1. Use browser DevTools to emulate a phone (e.g. iPhone SE, 375×667 px).
2. Open the graph and rotate the device to landscape.
3. Try to drag the canvas and pinch to zoom.

**Checks:**
- [ ] The graph canvas fills the viewport and does not scroll the page.
- [ ] Top-bar controls remain usable and do not overlap.
- [ ] Touch drag pans the graph; pinch zooms.
- [ ] Cockpit panels do not cover the entire screen when collapsed.
- [ ] No console errors on touch events.

---

## Known limitations (carry over from main checklist)

- Free node dragging on canvas is not supported (drag creates links, not moves nodes).
- `Ctrl+Z` is a placeholder — no real undo yet.
- Black hole glows on hover/drag but does not delete nodes on drop (not implemented).
- Semantic clustering of 50+ random notes is a planned feature; the current 2D/3D force layout can only keep the graph within bounds, not guarantee a clean readable cluster.

---

## How findings flow from here

1. Run the relevant section above.
2. If something fails, add an entry to `docs/MANUAL_TEST_FEEDBACK.md` under
   **Urgent fixes / Roadmap items / Ideas** (see that file for the entry
   template), referencing the section number from this checklist (e.g.
   "Case: Cockpit §5").
3. Once fixed, come back and re-check the box here.
