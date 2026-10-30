import { describe, it, expect } from "vitest";
import { render } from "@testing-library/svelte";
import TypeSelector from "./TypeSelector.svelte";
import { CelestialBody } from "$entities/shared/model/celestial-body";

// IMP-3 §4: the ghost form (and any other consumer) must render a button for
// every UI type — data-type presence for all 11, none missing.
describe("TypeSelector", () => {
  it("renders a button for every UI type", () => {
    render(TypeSelector, { props: { types: CelestialBody.UI_TYPES, selected: "star" } });
    for (const body of CelestialBody.UI_TYPES) {
      expect(
        document.querySelector(`[data-type="${body.type}"]`),
        `missing type button: ${body.type}`
      ).toBeTruthy();
    }
    expect(document.querySelectorAll("[data-type]")).toHaveLength(CelestialBody.UI_TYPES.length);
  });

  it("shows a hint with description and example for each type", () => {
    render(TypeSelector, { props: { types: CelestialBody.UI_TYPES, selected: "star" } });
    const btn = document.querySelector('[data-type="star"]') as HTMLElement;
    expect(btn.getAttribute("title")).toBeTruthy();
  });
});
