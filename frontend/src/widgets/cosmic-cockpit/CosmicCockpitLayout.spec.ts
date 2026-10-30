import { describe, it, expect, afterEach } from "vitest";
import { render, cleanup } from "@testing-library/svelte";
import CosmicCockpitLayout from "./CosmicCockpitLayout.svelte";
import { cockpitStore } from "$features/cosmic-cockpit";

describe("CosmicCockpitLayout — first-person Escape hotkey", () => {
  afterEach(() => {
    cockpitStore.setFirstPerson(false);
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
