import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/svelte";
import StateIllustration from "./StateIllustration.svelte";

describe("StateIllustration", () => {
  it.each([
    ["empty", "Empty state illustration"],
    ["error", "Error state illustration"],
    ["404", "404 illustration"],
    ["offline", "Offline illustration"],
    ["no-links", "No links illustration"],
    ["no-results", "No results illustration"],
  ] as const)("renders %s illustration with accessible label", (type, label) => {
    render(StateIllustration, { props: { type } });
    expect(screen.getByRole("img", { name: label })).toBeInTheDocument();
  });

  it("renders an svg element inside the wrapper", () => {
    const { container } = render(StateIllustration, {
      props: { type: "empty" },
    });
    expect(container.querySelector("svg")).toBeInTheDocument();
  });
});
