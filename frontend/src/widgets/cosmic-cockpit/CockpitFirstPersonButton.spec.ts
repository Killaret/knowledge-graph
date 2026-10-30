import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import CockpitFirstPersonButton from "./CockpitFirstPersonButton.svelte";
import { cockpitStore } from "$features/cosmic-cockpit";

describe("CockpitFirstPersonButton", () => {
  beforeEach(() => {
    cockpitStore.setFirstPerson(true);
  });

  afterEach(() => {
    cockpitStore.setFirstPerson(false);
    cleanup();
  });

  it("is always rendered and visible (not hover-only)", () => {
    render(CockpitFirstPersonButton);

    const button = screen.getByTestId("first-person-exit");
    expect(button).toBeVisible();
  });

  it("shows a hint that Esc exits first-person mode", () => {
    render(CockpitFirstPersonButton);

    expect(screen.getByTestId("first-person-exit")).toHaveTextContent(/esc/i);
  });

  it("calls cockpitStore.exitFirstPerson on click", async () => {
    render(CockpitFirstPersonButton);

    await fireEvent.click(screen.getByTestId("first-person-exit"));

    expect(cockpitStore.firstPerson).toBe(false);
  });

  it("applies the animated gradient-text class to its label with a phase offset", () => {
    render(CockpitFirstPersonButton);

    const label = screen.getByTestId("first-person-exit").querySelector(".cockpit-gradient-text");
    expect(label).not.toBeNull();
    expect(label?.getAttribute("style")).toContain("--cockpit-text-delay");
  });
});
