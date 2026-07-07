# Manual Testing Checklists

This document contains checklists for manually verifying new Knowledge Graph functionality.

## Canvas Features

### Ghost Node Creation
- [ ] Click on empty canvas area / use hotkey `N` to activate ghost-node creation.
- [ ] Fill in the title and confirm.
- [ ] Verify a new star/planet node appears on the canvas.
- [ ] Verify the new note is persisted (refresh page and check it still exists).
- [ ] Verify the node has a turquoise "New" indicator/glow.

### Black Hole Deletion (with Undo)
- [ ] Select an existing note node.
- [ ] Drag the selected node onto the black-hole anomaly.
- [ ] Verify the node disappears from the canvas.
- [ ] Verify a toast appears with an "Undo" action.
- [ ] Click "Undo" and verify the node reappears.
- [ ] Drag again and dismiss the toast; verify the note is deleted after refresh.

### Drag-and-Drop Links
- [ ] Hover over a source node to see the link creation affordance.
- [ ] Drag the link handle to a target node.
- [ ] Verify a new link/edge is drawn between the nodes.
- [ ] Try to create the same link again; verify a "duplicate link" warning/toast appears.
- [ ] Refresh and verify the link persists.

### Hotkeys
- [ ] Press `N` — ghost-node creation should activate.
- [ ] Press `F` — search/focus input should appear.
- [ ] Press `?` — help modal with hotkeys should open.
- [ ] Press `Esc` — any open modal/tooltip should close.

### Knowledge Core (Technical Note)
- [ ] Verify the central "Knowledge Core" node is always visible.
- [ ] Click the Knowledge Core node.
- [ ] Verify it cannot be deleted via black hole.
- [ ] Verify its tooltip shows system information.

### Tooltips
- [ ] Hover over any node; verify a tooltip with title, type, and action buttons appears.
- [ ] Hover over any link; verify a tooltip with link type and weight appears.

### New Indicator
- [ ] Create a fresh note.
- [ ] Verify the node has a turquoise outline/glow labelled "New".
- [ ] Wait/perform another action; verify the indicator disappears after the note is no longer "new".

---

## Note Cards (List View)

### Visual Style
- [ ] Switch to list view.
- [ ] Verify each card has a colored accent strip matching its type.
- [ ] Verify the type icon is displayed.
- [ ] Verify a subtle glow/hover effect is present.

### Dust Style
- [ ] Create a note with type `dust`.
- [ ] Verify the card uses the special "dust" visual treatment (faded particles/dust background).

### Card Tooltip
- [ ] Hover over a card.
- [ ] Verify the tooltip shows keywords, related links count, and quick-action buttons.

### Batch Operations
- [ ] Toggle selection mode.
- [ ] Select multiple notes using checkboxes.
- [ ] Click "Delete selected".
- [ ] Verify only selected notes are removed and an Undo toast appears.

### Undo Deletion
- [ ] Delete a note from the list view.
- [ ] Verify the toast says "Done" (Galactic Lexicon) with a "Restore" button.
- [ ] Click "Restore"; verify the note reappears in the list.

### Sorting
- [ ] Open the sort dropdown.
- [ ] Sort by created date ascending and descending.
- [ ] Sort by title A-Z and Z-A.
- [ ] Sort by updated date.
- [ ] Verify the order updates correctly.

### Staggered Fade-In Animation
- [ ] Load a page with many notes.
- [ ] Verify cards appear with a staggered fade-in animation, not all at once.

### Empty State
- [ ] Clear all notes (or filter to a non-matching query).
- [ ] Verify a friendly empty-state illustration/message is shown.

---

## General UX

### Galactic Lexicon
- [ ] Trigger success actions (create note, delete note, etc.).
- [ ] Verify toast messages use cosmic-themed wording (e.g., "Star born", "Matter collapsed", "Nebula restored").
- [ ] Verify no raw/default browser messages appear.

### Browser Console
- [ ] Open browser DevTools console.
- [ ] Navigate through graph and list views.
- [ ] Verify no red errors appear in the console.
- [ ] Verify warnings are acceptable (e.g., Svelte a11y warnings).

### Language Switch
- [ ] Open profile/settings.
- [ ] Change the language toggle.
- [ ] Verify UI labels, placeholders, and toast messages switch language.
- [ ] Verify the selected language persists after reload.

---

## Reporting Defects

For each failed item:
1. Record the environment (browser, stack).
2. Attach a screenshot or short screen recording.
3. Copy relevant console/backend logs.
4. File an issue with the checklist item reference.
