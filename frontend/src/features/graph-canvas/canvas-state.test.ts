import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  createGraphCanvasState,
  isTechnicalNode,
  pinTechnicalNodes,
  type HotkeysState,
} from "./canvas-state.svelte";

function createTestHotkeysState(): HotkeysState {
  return {
    showSearchBox: false,
    searchQuery: "",
    searchMatchIds: [],
    searchCurrentIndex: 0,
    showHelpModal: false,
    showHelpTooltip: false,
  };
}

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
  });

  afterEach(() => {
    vi.useRealTimers();
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

  it("opens and closes help modal", () => {
    const state = createGraphCanvasState();
    const hotkeys = createTestHotkeysState();

    state.openHelpModal(hotkeys);
    expect(hotkeys.showHelpModal).toBe(true);
    expect(hotkeys.showHelpTooltip).toBe(false);

    state.closeHelpModal(hotkeys);
    expect(hotkeys.showHelpModal).toBe(false);
  });

  it("opens and closes search box", () => {
    const state = createGraphCanvasState();
    const hotkeys = createTestHotkeysState();
    const redraw = vi.fn();

    state.handleOpenSearch(hotkeys);
    expect(hotkeys.showSearchBox).toBe(true);

    state.handleCloseSearch(hotkeys, redraw);
    expect(hotkeys.showSearchBox).toBe(false);
    expect(hotkeys.searchQuery).toBe("");
    expect(redraw).toHaveBeenCalled();
  });
});
