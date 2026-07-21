import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  createGraphCanvasState,
  isTechnicalNode,
  pinTechnicalNodes,
} from "./canvas-state.svelte";
import { createNoteFormState } from "$features/graph-forms/note-form";
import { createLinkFormState } from "$features/graph-forms/link-form";
import { createHotkeysState } from "$features/graph-interaction/hotkeys";

vi.mock("$components/organisms/GraphCanvas", () => ({
  getSimulationNodes: vi.fn(() => []),
  resetView: vi.fn(),
}));

const { getSimulationNodes, resetView } =
  await import("$components/organisms/GraphCanvas");

describe("canvas-state helpers", () => {
  it("isTechnicalNode detects technical node types", () => {
    const nodes = [
      { id: "n1", type: "star" },
      { id: "n2", type: "technical" },
    ];
    expect(isTechnicalNode(nodes, "n1")).toBe(false);
    expect(isTechnicalNode(nodes, "n2")).toBe(true);
    expect(isTechnicalNode(nodes, "missing")).toBe(false);
  });

  it("pinTechnicalNodes fixes position for technical nodes", () => {
    const nodes = [
      { id: "n1", title: "Star", type: "star" },
      { id: "n2", title: "Tech", type: "technical" },
    ];
    const result = pinTechnicalNodes(nodes);
    expect(result[1].x).toBe(60);
    expect(result[1].y).toBe(60);
    expect(result[1].fx).toBe(60);
    expect(result[1].fy).toBe(60);

    expect(result[1].id).toBe("n2");
    expect(result[0].x).toBeUndefined();
  });
});

describe("createGraphCanvasState", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("creates a note via handleCreateNote", () => {
    const state = createGraphCanvasState();
    const noteForm = createNoteFormState();
    noteForm.newNoteTitle = "New note";
    noteForm.newNoteContent = "Content";
    noteForm.newNoteType = "star";
    const onNoteCreate = vi.fn();
    const redraw = vi.fn();

    state.handleCreateNote(noteForm, onNoteCreate, redraw);

    expect(onNoteCreate).toHaveBeenCalledWith({
      title: "New note",
      content: "Content",
      type: "star",
    });
    expect(redraw).toHaveBeenCalled();
    expect(noteForm.showNoteForm).toBe(false);
  });

  it("closes note form via handleNoteFormClose", () => {
    const state = createGraphCanvasState();
    const noteForm = createNoteFormState();
    noteForm.newNoteTitle = "draft";
    const redraw = vi.fn();

    state.handleNoteFormClose(noteForm, redraw);

    expect(noteForm.showNoteForm).toBe(false);
    expect(noteForm.newNoteTitle).toBe("");
    expect(redraw).toHaveBeenCalled();
  });

  it("creates a link via handleCreateLink", () => {
    const state = createGraphCanvasState();
    const linkForm = createLinkFormState();
    linkForm.linkSourceNodeId = "a";
    linkForm.linkTargetNodeId = "b";
    linkForm.newLinkType = "related";
    linkForm.newLinkWeight = 1;
    const onLinkCreate = vi.fn();
    const redraw = vi.fn();

    state.handleCreateLink(linkForm, [], onLinkCreate, redraw);

    expect(onLinkCreate).toHaveBeenCalledWith({
      source: "a",
      target: "b",
      link_type: "related",
      weight: 1,
    });
    expect(redraw).toHaveBeenCalled();
  });

  it("shows duplicate warning for existing link", () => {
    const state = createGraphCanvasState();
    const linkForm = createLinkFormState();
    linkForm.linkSourceNodeId = "a";
    linkForm.linkTargetNodeId = "b";
    linkForm.linkFormPosition = { x: 100, y: 200 };
    const onLinkCreate = vi.fn();
    const redraw = vi.fn();
    const links = [{ source: "a", target: "b", link_type: "related" }];

    state.handleCreateLink(linkForm, links, onLinkCreate, redraw);

    expect(onLinkCreate).not.toHaveBeenCalled();
    expect(state.duplicateWarning).not.toBeNull();
    expect(state.highlightedLinkId).toBe("a-b-related");
  });

  it("clears duplicate warning after timeout", () => {
    const state = createGraphCanvasState();
    state.showDuplicateWarning("a", "b", "related", 100, 200);

    expect(state.duplicateWarning).not.toBeNull();

    vi.advanceTimersByTime(2001);
    expect(state.duplicateWarning).toBeNull();
  });

  it("edits and deletes hovered links", () => {
    const state = createGraphCanvasState();
    const hoveredLink = {
      source: "a",
      target: "b",
      link_type: "related",
      weight: 1,
      source_type: "star",
    };
    state.hoveredLink = hoveredLink;

    const onLinkEdit = vi.fn();
    state.handleLinkEdit(onLinkEdit);
    expect(onLinkEdit).toHaveBeenCalledWith(hoveredLink);
    expect(state.hoveredLink).toBeNull();

    state.hoveredLink = hoveredLink;
    const onLinkDelete = vi.fn();
    state.handleLinkDelete(onLinkDelete);
    expect(onLinkDelete).toHaveBeenCalledWith({
      source: "a",
      target: "b",
      link_type: "related",
    });
    expect(state.hoveredLink).toBeNull();
  });

  it("shows undo toast and restores deleted node", () => {
    const state = createGraphCanvasState();
    const onNoteRestore = vi.fn();

    state.showUndoToastFor("node-1");
    expect(state.showUndoToast).toBe(true);
    expect(state.lastDeletedNodeId).toBe("node-1");

    state.restoreDeletedNode(onNoteRestore);
    expect(onNoteRestore).toHaveBeenCalledWith("node-1");
    expect(state.showUndoToast).toBe(false);
    expect(state.lastDeletedNodeId).toBeNull();
  });

  it("cancels undo toast", () => {
    const state = createGraphCanvasState();
    state.showUndoToastFor("node-1");
    state.cancelUndo();

    expect(state.showUndoToast).toBe(false);
    expect(state.lastDeletedNodeId).toBeNull();
  });

  it("toggles focus mode", () => {
    const state = createGraphCanvasState();
    const redraw = vi.fn();

    expect(state.focusMode).toBe(false);
    state.handleToggleFocus(redraw);
    expect(state.focusMode).toBe(true);
    expect(redraw).toHaveBeenCalled();
  });

  it("handles search open/close/update", () => {
    const state = createGraphCanvasState();
    const hotkeys = createHotkeysState();
    const redraw = vi.fn();

    state.handleOpenSearch(hotkeys);
    expect(hotkeys.showSearchBox).toBe(true);

    state.handleCloseSearch(hotkeys, redraw);
    expect(hotkeys.showSearchBox).toBe(false);
    expect(redraw).toHaveBeenCalled();

    const simState = {
      simulation: { nodes: () => [{ id: "n1", title: "Test", x: 0, y: 0 }] },
    } as any;
    vi.mocked(getSimulationNodes).mockReturnValue([
      { id: "n1", title: "Test", x: 0, y: 0 },
    ]);
    hotkeys.searchQuery = "Test";
    state.handleUpdateSearch(hotkeys, simState, redraw);
    expect(hotkeys.searchMatchIds).toEqual(["n1"]);
  });

  it("opens and closes help modal", () => {
    const state = createGraphCanvasState();
    const hotkeys = createHotkeysState();

    state.openHelpModal(hotkeys);
    expect(hotkeys.showHelpModal).toBe(true);
    expect(hotkeys.showHelpTooltip).toBe(false);

    state.closeHelpModal(hotkeys);
    expect(hotkeys.showHelpModal).toBe(false);
  });

  it("resets view when canvas and nodes exist", () => {
    const state = createGraphCanvasState();
    const simState = {
      simulation: { nodes: () => [{ id: "n1", title: "Test", x: 0, y: 0 }] },
    } as any;
    const transform = { x: 0, y: 0, k: 1 };
    const ctx = {} as CanvasRenderingContext2D;

    vi.mocked(getSimulationNodes).mockReturnValue([
      { id: "n1", title: "Test", x: 0, y: 0 },
    ]);
    state.handleResetView(ctx, 800, 600, simState, transform);

    expect(getSimulationNodes).toHaveBeenCalledWith(simState);
    expect(resetView).toHaveBeenCalledWith(
      ctx,
      800,
      600,
      expect.any(Array),
      transform,
    );
  });
});
