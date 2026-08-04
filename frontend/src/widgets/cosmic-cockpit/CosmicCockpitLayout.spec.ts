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
