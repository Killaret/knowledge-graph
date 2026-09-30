import { describe, it, expect, afterEach } from "vitest";
import { render, cleanup } from "@testing-library/svelte";
import CosmicCockpitLayout from "./CosmicCockpitLayout.svelte";
import { COCKPIT_EDGE_SIZE, COCKPIT_PANEL_GAP, cockpitStore } from "$features/cosmic-cockpit";

describe("CosmicCockpitLayout — UI-PANELS-1 defaults", () => {
  it("pins the top panel open by default", () => {
    expect(cockpitStore.panels.top.pinned).toBe(true);
    expect(cockpitStore.panels.top.open).toBe(true);
  });
});

describe("CosmicCockpitLayout — UI-PANELS-1 insets", () => {
  afterEach(() => {
    cockpitStore.setPanel("right", { open: false, pinned: false, hovering: false });
    cleanup();
  });

  it("keeps the frame inset at handle size on bare hover — hovering never opens space", () => {
    const { container } = render(CosmicCockpitLayout, { props: { isAuthenticated: false } });

    cockpitStore.hoverPanel("right", true);

    const shell = container.querySelector(".cosmic-cockpit") as HTMLElement;
    expect(shell.getAttribute("style") ?? "").toContain(
      `--inset-right: ${COCKPIT_EDGE_SIZE + COCKPIT_PANEL_GAP}px`
    );
  });
});

describe("CosmicCockpitLayout — first-person Escape hotkey", () => {
  afterEach(() => {
    cockpitStore.setFirstPerson(false);
    cockpitStore.closePanel("left");
    cockpitStore.closePanel("right");
    cleanup();
  });

  it("exits first-person mode when Escape is pressed", () => {
    render(CosmicCockpitLayout, { props: { isAuthenticated: false } });

    cockpitStore.setFirstPerson(true);
    expect(cockpitStore.firstPerson).toBe(true);

    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));

    expect(cockpitStore.firstPerson).toBe(false);
  });

  it("does nothing on Escape when first-person mode is already off", () => {
    render(CosmicCockpitLayout, { props: { isAuthenticated: false } });

    cockpitStore.setFirstPerson(false);

    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));

    expect(cockpitStore.firstPerson).toBe(false);
  });

  it("ignores other keys", () => {
    render(CosmicCockpitLayout, { props: { isAuthenticated: false } });

    cockpitStore.setFirstPerson(true);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Enter" }));

    expect(cockpitStore.firstPerson).toBe(true);
  });

  it("Escape closes open unpinned panels but keeps pinned ones", () => {
    render(CosmicCockpitLayout, { props: { isAuthenticated: false } });

    cockpitStore.openPanel("left");
    cockpitStore.openPanel("right");
    expect(cockpitStore.panels.left.open).toBe(true);
    expect(cockpitStore.panels.right.open).toBe(true);

    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));

    expect(cockpitStore.panels.left.open).toBe(false);
    expect(cockpitStore.panels.right.open).toBe(false);
    expect(cockpitStore.panels.top.open).toBe(true);
  });

  it("Escape does not steal the key from a focused input", () => {
    render(CosmicCockpitLayout, { props: { isAuthenticated: false } });
    cockpitStore.openPanel("left");

    const input = document.createElement("input");
    document.body.appendChild(input);
    input.focus();
    input.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    document.body.removeChild(input);

    expect(cockpitStore.panels.left.open).toBe(true);
  });
});

describe("CosmicCockpitLayout — selectedNodeId opens/closes the right panel", () => {
  afterEach(() => {
    cockpitStore.setFirstPerson(false);
    cockpitStore.closePanel("right");
    cleanup();
  });

  it("opens the right panel when selectedNodeId becomes non-null (e.g. from a canvas click)", () => {
    const { rerender } = render(CosmicCockpitLayout, {
      props: { isAuthenticated: false, selectedNodeId: null },
    });

    expect(cockpitStore.panels.right.open).toBe(false);

    rerender({ isAuthenticated: false, selectedNodeId: "note-1" });

    expect(cockpitStore.panels.right.open).toBe(true);
  });

  it("closes the right panel when selectedNodeId goes back to null", () => {
    const { rerender } = render(CosmicCockpitLayout, {
      props: { isAuthenticated: false, selectedNodeId: "note-1" },
    });

    expect(cockpitStore.panels.right.open).toBe(true);

    rerender({ isAuthenticated: false, selectedNodeId: null });

    expect(cockpitStore.panels.right.open).toBe(false);
  });

  it("exits first-person mode when a node gets selected while in first-person", () => {
    const { rerender } = render(CosmicCockpitLayout, {
      props: { isAuthenticated: false, selectedNodeId: null },
    });

    cockpitStore.setFirstPerson(true);
    expect(cockpitStore.firstPerson).toBe(true);

    rerender({ isAuthenticated: false, selectedNodeId: "note-1" });

    expect(cockpitStore.firstPerson).toBe(false);
    expect(cockpitStore.panels.right.open).toBe(true);
  });
});
