import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, cleanup } from "@testing-library/svelte";
import { tick } from "svelte";

vi.mock("./GraphTooltip.svelte", () => ({
  default: vi.fn().mockImplementation(() => {
    return {
      showNodeTooltip: vi.fn(),
      hide: vi.fn(),
    };
  }),
}));

vi.mock("./LinkTooltip.svelte", () => ({
  default: vi.fn().mockImplementation(() => {
    return {};
  }),
}));

import Overlay from "./overlay.svelte";

function createCanvas() {
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

function createHotkeysState(overrides: Record<string, unknown> = {}) {
  return {
    showSearchBox: false,
    searchQuery: "",
    searchMatchIds: [] as string[],
    searchCurrentIndex: 0,
    showHelpModal: false,
    showHelpTooltip: false,
    helpTooltipMessage: "",
    helpTooltipPosition: { x: -1, y: -1 },
    inactivityTimeout: null,
    lastActivityTime: 0,
    ...overrides,
  };
}

describe("GraphOverlay Component", () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    cleanup();
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it("renders empty when no active overlays", async () => {
    const canvas = createCanvas();
    const hotkeys = createHotkeysState();
    const { container } = render(Overlay, {
      props: {
        canvas,
        nodes: [],
        hotkeysState: hotkeys,
      },
    });
    await tick();
    expect(container.textContent?.trim()).toBe("");
  });

  it("shows duplicate warning", async () => {
    const canvas = createCanvas();
    const hotkeys = createHotkeysState();
    const { getByText } = render(Overlay, {
      props: {
        canvas,
        nodes: [],
        hotkeysState: hotkeys,
        duplicateWarning: { message: "Already linked", x: 10, y: 10, linkId: "l1" },
      },
    });
    await tick();
    expect(getByText("Already linked")).toBeTruthy();
  });

  it("shows focus mode indicator", async () => {
    const canvas = createCanvas();
    const hotkeys = createHotkeysState();
    const { getByText } = render(Overlay, {
      props: {
        canvas,
        nodes: [],
        hotkeysState: hotkeys,
        focusMode: true,
      },
    });
    await tick();
    expect(getByText(/Focus/i)).toBeTruthy();
  });

  it("shows fog recovery warning", async () => {
    const canvas = createCanvas();
    const hotkeys = createHotkeysState();
    const { getByText } = render(Overlay, {
      props: {
        canvas,
        nodes: [],
        hotkeysState: hotkeys,
        fogWarning: "recovery",
      },
    });
    await tick();
    expect(getByText(/fog/i)).toBeTruthy();
  });

  it("shows fog danger warning", async () => {
    const canvas = createCanvas();
    const hotkeys = createHotkeysState();
    const { getByText } = render(Overlay, {
      props: {
        canvas,
        nodes: [],
        hotkeysState: hotkeys,
        fogWarning: "danger",
      },
    });
    await tick();
    expect(getByText(/fog/i)).toBeTruthy();
  });

  it("shows undo toast with restore option", async () => {
    const canvas = createCanvas();
    const hotkeys = createHotkeysState();
    const onRestore = vi.fn();
    const { getByText } = render(Overlay, {
      props: {
        canvas,
        nodes: [],
        hotkeysState: hotkeys,
        showUndoToast: true,
        undoToastStage: "restore",
        onRestoreDeletedNode: onRestore,
      },
    });
    await tick();
    expect(getByText(/Restore/i)).toBeTruthy();
  });

  it("shows search box with match counter", async () => {
    const canvas = createCanvas();
    const hotkeys = createHotkeysState({
      showSearchBox: true,
      searchMatchIds: ["n1", "n2"],
      searchCurrentIndex: 0,
    });
    const { getByText, getByPlaceholderText } = render(Overlay, {
      props: {
        canvas,
        nodes: [],
        hotkeysState: hotkeys,
      },
    });
    await tick();
    expect(getByPlaceholderText(/Search/i)).toBeTruthy();
    expect(getByText("1/2")).toBeTruthy();
  });

  it("shows help tooltip at default position", async () => {
    const canvas = createCanvas();
    const hotkeys = createHotkeysState({
      showHelpTooltip: true,
      helpTooltipMessage: "Press / to search",
      helpTooltipPosition: { x: -1, y: -1 },
    });
    const { getByText } = render(Overlay, {
      props: {
        canvas,
        nodes: [],
        hotkeysState: hotkeys,
      },
    });
    await tick();
    expect(getByText("Press / to search")).toBeTruthy();
  });

  it("shows help tooltip at custom position", async () => {
    const canvas = createCanvas();
    const hotkeys = createHotkeysState({
      showHelpTooltip: true,
      helpTooltipMessage: "Hover a node",
      helpTooltipPosition: { x: 100, y: 100 },
    });
    const { getByText } = render(Overlay, {
      props: {
        canvas,
        nodes: [],
        hotkeysState: hotkeys,
      },
    });
    await tick();
    expect(getByText("Hover a node")).toBeTruthy();
  });
});
