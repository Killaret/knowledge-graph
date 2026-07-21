import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  createHotkeysState,
  handleKeyDownEvent,
  updateSearch,
  focusNextSearchMatch,
  updateActivity,
  showRandomTip,
} from "./hotkeys";

function createCanvas(): HTMLCanvasElement {
  const canvas = document.createElement("canvas");
  canvas.getBoundingClientRect = () => ({
    left: 0,
    top: 0,
    width: 800,
    height: 600,
    right: 800,
    bottom: 600,
    x: 0,
    y: 0,
    toJSON: () => "",
  });
  return canvas;
}

describe("hotkeys", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal("requestAnimationFrame", (cb: FrameRequestCallback) => {
      cb(0);
      return 0;
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it("creates hotkeys state", () => {
    const state = createHotkeysState();
    expect(state.showSearchBox).toBe(false);
    expect(state.searchMatchIds).toEqual([]);
  });

  it("opens search on 'f' key", () => {
    const state = createHotkeysState();
    const searchInput = document.createElement("input");
    document.body.appendChild(searchInput);
    const callbacks = {
      onSearchOpen: vi.fn(),
    };

    handleKeyDownEvent(
      new KeyboardEvent("keydown", { key: "f" }),
      state,
      createCanvas(),
      { x: 0, y: 0, k: 1 },
      [],
      { x: 0, y: 0, radius: 15, hovered: false, pulsePhase: 0, active: true },
      null,
      false,
      false,
      searchInput,
      callbacks,
    );

    expect(callbacks.onSearchOpen).toHaveBeenCalled();
    expect(document.activeElement).toBe(searchInput);
    document.body.removeChild(searchInput);
  });

  it("toggles focus on Escape", () => {
    const state = createHotkeysState();
    const callbacks = { onFocusModeToggle: vi.fn() };

    handleKeyDownEvent(
      new KeyboardEvent("keydown", { key: "Escape" }),
      state,
      createCanvas(),
      { x: 0, y: 0, k: 1 },
      [],
      { x: 0, y: 0, radius: 15, hovered: false, pulsePhase: 0, active: true },
      null,
      false,
      false,
      null,
      callbacks,
    );

    expect(callbacks.onFocusModeToggle).toHaveBeenCalled();
  });

  it("ignores hotkeys while typing in input", () => {
    const input = document.createElement("input");
    document.body.appendChild(input);
    input.focus();

    const state = createHotkeysState();
    const callbacks = { onFocusModeToggle: vi.fn() };

    handleKeyDownEvent(
      new KeyboardEvent("keydown", { key: "Delete" }),
      state,
      createCanvas(),
      { x: 0, y: 0, k: 1 },
      [],
      { x: 0, y: 0, radius: 15, hovered: false, pulsePhase: 0, active: true },
      null,
      false,
      false,
      null,
      callbacks,
    );

    expect(callbacks.onFocusModeToggle).not.toHaveBeenCalled();
    document.body.removeChild(input);
  });

  it("focuses next search match on Enter", () => {
    const state = createHotkeysState();
    state.showSearchBox = true;
    state.searchQuery = "node";
    state.searchMatchIds = ["n1", "n2"];
    const transform = { x: 0, y: 0, k: 1 };
    const simNodes = [
      { id: "n1", x: 100, y: 100, title: "Node 1" },
      { id: "n2", x: 200, y: 200, title: "Node 2" },
    ];

    focusNextSearchMatch(state, transform, simNodes as any, createCanvas());
    expect(state.searchCurrentIndex).toBe(1);
    expect(transform.k).toBeGreaterThanOrEqual(1.2);
  });

  it("updates search matches", () => {
    const state = createHotkeysState();
    state.searchQuery = "one";
    const simNodes = [
      { id: "n1", title: "One", x: 0, y: 0 },
      { id: "n2", title: "Two", x: 0, y: 0 },
    ];

    updateSearch(state, simNodes as any);

    expect(state.searchMatchIds).toEqual(["n1"]);
  });

  it("creates ghost node on 'n' key", () => {
    const state = createHotkeysState();
    const callbacks = { onGhostNodeCreate: vi.fn() };

    handleKeyDownEvent(
      new KeyboardEvent("keydown", { key: "n" }),
      state,
      createCanvas(),
      { x: 0, y: 0, k: 1 },
      [],
      { x: 0, y: 0, radius: 15, hovered: false, pulsePhase: 0, active: true },
      null,
      false,
      false,
      null,
      callbacks,
    );

    expect(callbacks.onGhostNodeCreate).toHaveBeenCalled();
  });

  it("deletes selected node on Delete key", () => {
    const state = createHotkeysState();
    const callbacks = { onNodeDelete: vi.fn() };

    handleKeyDownEvent(
      new KeyboardEvent("keydown", { key: "Delete" }),
      state,
      createCanvas(),
      { x: 0, y: 0, k: 1 },
      [],
      { x: 0, y: 0, radius: 15, hovered: false, pulsePhase: 0, active: true },
      "node-1",
      false,
      false,
      null,
      callbacks,
    );

    expect(callbacks.onNodeDelete).toHaveBeenCalledWith("node-1");
  });

  it("triggers undo on Ctrl+Z", () => {
    const state = createHotkeysState();
    const callbacks = { onUndo: vi.fn() };

    handleKeyDownEvent(
      new KeyboardEvent("keydown", { key: "z", ctrlKey: true }),
      state,
      createCanvas(),
      { x: 0, y: 0, k: 1 },
      [],
      { x: 0, y: 0, radius: 15, hovered: false, pulsePhase: 0, active: true },
      null,
      false,
      false,
      null,
      callbacks,
    );

    expect(callbacks.onUndo).toHaveBeenCalled();
  });

  it("shows help modal on '?' key", () => {
    const state = createHotkeysState();
    const callbacks = { onHelpToggle: vi.fn() };

    handleKeyDownEvent(
      new KeyboardEvent("keydown", { key: "?" }),
      state,
      createCanvas(),
      { x: 0, y: 0, k: 1 },
      [],
      { x: 0, y: 0, radius: 15, hovered: false, pulsePhase: 0, active: true },
      null,
      false,
      false,
      null,
      callbacks,
    );

    expect(callbacks.onHelpToggle).toHaveBeenCalled();
  });

  it("shows and hides random tip", () => {
    const state = createHotkeysState();
    const tips = ["tip1", "tip2"];

    showRandomTip(state, tips);

    expect(state.showHelpTooltip).toBe(true);
    expect(tips).toContain(state.helpTooltipMessage);

    vi.advanceTimersByTime(4000);
    expect(state.showHelpTooltip).toBe(false);
  });

  it("inactivity timer triggers tip", () => {
    const state = createHotkeysState();
    const onTip = vi.fn();

    updateActivity(state, onTip);
    expect(state.lastActivityTime).toBeGreaterThan(0);

    vi.advanceTimersByTime(10000);
    expect(onTip).toHaveBeenCalled();
  });
});
