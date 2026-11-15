import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/svelte";
import CockpitPanel from "./CockpitPanel.svelte";
import { COCKPIT_CLOSE_DELAY, COCKPIT_EDGE_SIZE, cockpitStore } from "$features/cosmic-cockpit";

vi.mock("$app/environment", () => ({
  browser: true,
}));

function resetPanels() {
  (["top", "bottom", "left", "right"] as const).forEach((position) => {
    cockpitStore.setPanel(position, { open: false, pinned: false, hovering: false });
  });
}

function resetCockpit() {
  resetPanels();
  cockpitStore.setFirstPerson(false);
  cockpitStore.hoverDelay = 350;
  cockpitStore.edgeSensitivity = 40;
  cockpitStore.autoCollapse = true;
  cockpitStore.reducedMotion = false;
}

describe("CockpitPanel", () => {
  beforeEach(() => {
    resetCockpit();
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    vi.useRealTimers();
    resetCockpit();
    cleanup();
  });

  it.each(["right", "bottom", "left", "top"] as const)(
    "shows the pull-handle for %s when collapsed",
    (position) => {
      render(CockpitPanel, { props: { position, size: 300, title: "Test" } });

      expect(screen.getByTestId(`cockpit-handle-${position}`)).toBeInTheDocument();
    }
  );

  it.each(["right", "bottom", "left", "top"] as const)(
    "hides the pull-handle for %s once the panel is open",
    (position) => {
      cockpitStore.setPanel(position, { open: true });

      render(CockpitPanel, { props: { position, size: 300, title: "Test" } });

      expect(screen.queryByTestId(`cockpit-handle-${position}`)).not.toBeInTheDocument();
    }
  );

  it("renders the panel title with the animated gradient-text class", () => {
    render(CockpitPanel, { props: { position: "top", size: 64, title: "Navigation" } });

    const title = screen.getByText("Navigation");
    expect(title).toHaveClass("cockpit-gradient-text");
    expect(title.getAttribute("style")).toContain("--cockpit-text-delay");
  });

  it("uses a different gradient-text delay for different panel positions", () => {
    const { unmount } = render(CockpitPanel, {
      props: { position: "top", size: 64, title: "Top Title" },
    });
    const topStyle = screen.getByText("Top Title").getAttribute("style");
    unmount();
    cleanup();

    render(CockpitPanel, { props: { position: "left", size: 300, title: "Left Title" } });
    const leftStyle = screen.getByText("Left Title").getAttribute("style");

    expect(topStyle).not.toEqual(leftStyle);
  });

  it("reports a clamped, content-based size via onSizeChange when minSize/maxSize are provided", () => {
    const onSizeChange = vi.fn();

    render(CockpitPanel, {
      props: { position: "top", size: 64, minSize: 56, maxSize: 120, onSizeChange, title: "Nav" },
    });

    // jsdom reports 0 for scrollHeight/scrollWidth by default, so the
    // reported size should be clamped up to the configured minimum rather
    // than collapsing the panel to nothing.
    expect(onSizeChange).toHaveBeenCalledWith(56);
  });

  it("does not attempt content-based measurement when minSize/maxSize are not provided (left/right panels)", () => {
    const onSizeChange = vi.fn();

    render(CockpitPanel, {
      props: { position: "left", size: 320, onSizeChange, title: "Operations" },
    });

    expect(onSizeChange).not.toHaveBeenCalled();
  });

  it("toggles pin state when the pin button is clicked", () => {
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const pin = screen.getByTestId("cockpit-panel-pin-right");
    fireEvent.click(pin);
    expect(cockpitStore.panels.right.pinned).toBe(true);
    expect(cockpitStore.panels.right.open).toBe(true);

    fireEvent.click(pin);
    expect(cockpitStore.panels.right.pinned).toBe(false);
  });

  it("calls onClose and resets the panel when the close button is clicked", () => {
    cockpitStore.setPanel("right", { open: true, pinned: true });
    const onClose = vi.fn();

    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right", onClose } });

    const close = screen.getByTestId("cockpit-panel-close-right");
    fireEvent.click(close);

    expect(cockpitStore.panels.right.open).toBe(false);
    expect(cockpitStore.panels.right.pinned).toBe(false);
    expect(onClose).toHaveBeenCalled();
  });

  // UI-PANELS-1: visibility must not depend on the bare `hovering` flag —
  // a cursor resting on the edge never makes the panel start sliding.
  it("does not slide out on a bare hover (flicker guard)", () => {
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const panel = screen.getByTestId("cockpit-panel-right");
    fireEvent.mouseEnter(panel);

    // No time passed yet — the handle is still the only visible part.
    expect(cockpitStore.panels.right.open).toBe(false);
    expect(screen.getByTestId("cockpit-handle-right")).toBeInTheDocument();
    expect(panel.style.getPropertyValue("--panel-visible")).toBe(`${COCKPIT_EDGE_SIZE}px`);
  });

  it("cancels the delayed open when the cursor leaves before the delay", () => {
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const panel = screen.getByTestId("cockpit-panel-right");
    fireEvent.mouseEnter(panel);
    vi.advanceTimersByTime(100);
    fireEvent.mouseLeave(panel);

    vi.advanceTimersByTime(1000);
    expect(cockpitStore.panels.right.open).toBe(false);
  });

  it("ignores hover-open entirely when auto-collapse is off", () => {
    cockpitStore.autoCollapse = false;
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const panel = screen.getByTestId("cockpit-panel-right");
    fireEvent.mouseEnter(panel);
    vi.advanceTimersByTime(1000);

    expect(cockpitStore.panels.right.open).toBe(false);
    expect(screen.getByTestId("cockpit-handle-right")).toBeInTheDocument();
  });

  it("stays open on mouse leave when auto-collapse is off", () => {
    cockpitStore.autoCollapse = false;
    cockpitStore.setPanel("right", { open: true });
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const panel = screen.getByTestId("cockpit-panel-right");
    fireEvent.mouseEnter(panel);
    fireEvent.mouseLeave(panel);
    vi.advanceTimersByTime(2000);

    expect(cockpitStore.panels.right.open).toBe(true);
  });

  it("does not hide while a dropdown inside the panel is open", () => {
    cockpitStore.setPanel("right", { open: true });
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const panel = screen.getByTestId("cockpit-panel-right");
    const dropdown = document.createElement("div");
    dropdown.setAttribute("aria-expanded", "true");
    panel.appendChild(dropdown);

    fireEvent.mouseLeave(panel);
    vi.advanceTimersByTime(2000);

    expect(cockpitStore.panels.right.open).toBe(true);
  });

  it("does not hide while focus is inside the panel", () => {
    cockpitStore.setPanel("right", { open: true });
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const panel = screen.getByTestId("cockpit-panel-right");
    const input = document.createElement("input");
    panel.appendChild(input);
    input.focus();

    fireEvent.mouseLeave(panel);
    vi.advanceTimersByTime(2000);

    expect(cockpitStore.panels.right.open).toBe(true);
  });

  it("closes after the dwell delay once the cursor leaves and nothing inside holds it", () => {
    cockpitStore.setPanel("right", { open: true });
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const panel = screen.getByTestId("cockpit-panel-right");
    fireEvent.mouseEnter(panel);
    fireEvent.mouseLeave(panel);

    // Not hidden at the old 350ms cadence — the dwell must be ~600ms.
    vi.advanceTimersByTime(COCKPIT_CLOSE_DELAY - 250);
    expect(cockpitStore.panels.right.open).toBe(true);

    vi.advanceTimersByTime(400);
    expect(cockpitStore.panels.right.open).toBe(false);
  });

  it("opens the panel after the configured hover delay", () => {
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const panel = screen.getByTestId("cockpit-panel-right");
    fireEvent.mouseEnter(panel);

    expect(cockpitStore.panels.right.hovering).toBe(true);
    expect(cockpitStore.panels.right.open).toBe(false);

    vi.advanceTimersByTime(360);
    expect(cockpitStore.panels.right.open).toBe(true);
  });

  it("cancels a pending close on re-enter", () => {
    cockpitStore.setPanel("right", { open: true });
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const panel = screen.getByTestId("cockpit-panel-right");
    fireEvent.mouseLeave(panel);

    vi.advanceTimersByTime(180);
    fireEvent.mouseEnter(panel);

    vi.advanceTimersByTime(360);
    expect(cockpitStore.panels.right.open).toBe(true);
  });

  it("does not open from hover when first-person mode is active", () => {
    cockpitStore.setFirstPerson(true);
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const panel = screen.getByTestId("cockpit-panel-right");
    fireEvent.mouseEnter(panel);

    vi.advanceTimersByTime(360);
    expect(cockpitStore.panels.right.hovering).toBe(false);
    expect(cockpitStore.panels.right.open).toBe(false);
  });

  it("opens the panel with the Enter key on the handle", () => {
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const handle = screen.getByTestId("cockpit-handle-right");
    fireEvent.keyDown(handle, { key: "Enter" });

    expect(cockpitStore.panels.right.open).toBe(true);
  });

  it.each([
    { position: "left", start: 100, end: 150 },
    { position: "right", start: 150, end: 100 },
    { position: "top", start: 100, end: 150 },
    { position: "bottom", start: 150, end: 100 },
  ] as const)(
    "opens the $position panel when pulled past the drag threshold",
    ({ position, start, end }) => {
      render(CockpitPanel, { props: { position, size: 300, title: position } });

      const handle = screen.getByTestId(`cockpit-handle-${position}`);
      fireEvent.pointerDown(handle, { clientX: start, clientY: start, pointerId: 1 });
      fireEvent.pointerMove(handle, { clientX: end, clientY: end, pointerId: 1 });

      expect(cockpitStore.panels[position].open).toBe(true);
    }
  );

  it("does not pull open in first-person mode", () => {
    cockpitStore.setFirstPerson(true);
    render(CockpitPanel, { props: { position: "right", size: 300, title: "Right" } });

    const handle = screen.getByTestId("cockpit-handle-right");
    fireEvent.pointerDown(handle, { clientX: 150, clientY: 0, pointerId: 1 });
    fireEvent.pointerMove(handle, { clientX: 100, clientY: 0, pointerId: 1 });

    expect(cockpitStore.panels.right.open).toBe(false);
  });

  it("removes panel transitions when reduced motion is enabled", () => {
    cockpitStore.reducedMotion = true;
    const { container } = render(CockpitPanel, {
      props: { position: "right", size: 300, title: "Right" },
    });

    const panel = container.querySelector('[data-testid="cockpit-panel-right"]') as HTMLElement;
    expect(panel?.getAttribute("style") ?? "").toContain("transition: none");
  });

  it("removes panel transitions when the OS prefers reduced motion", () => {
    cockpitStore.setSystemReducedMotion(true);
    const { container } = render(CockpitPanel, {
      props: { position: "right", size: 300, title: "Right" },
    });

    const panel = container.querySelector('[data-testid="cockpit-panel-right"]') as HTMLElement;
    expect(panel?.getAttribute("style") ?? "").toContain("transition: none");
    cockpitStore.setSystemReducedMotion(false);
  });
});
