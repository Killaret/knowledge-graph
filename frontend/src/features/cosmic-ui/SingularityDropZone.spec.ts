import { describe, it, expect, afterEach } from "vitest";
import { render, screen, cleanup } from "@testing-library/svelte";
import SingularityDropZone from "./SingularityDropZone.svelte";

describe("SingularityDropZone", () => {
  afterEach(() => {
    cleanup();
  });

  it("does not render when not visible", () => {
    render(SingularityDropZone, { props: { visible: false, hovered: false } });

    expect(screen.queryByTestId("singularity-drop-zone")).not.toBeInTheDocument();
  });

  it("renders when visible", () => {
    render(SingularityDropZone, { props: { visible: true, hovered: false } });

    const zone = screen.getByTestId("singularity-drop-zone");
    expect(zone).toBeInTheDocument();
    expect(zone).not.toHaveClass("hovered");
  });

  it("applies hovered class and aria-label", () => {
    render(SingularityDropZone, { props: { visible: true, hovered: true } });

    const zone = screen.getByTestId("singularity-drop-zone");
    expect(zone).toHaveClass("hovered");
    expect(zone).toHaveAttribute("role", "region");
    expect(zone).toHaveAttribute("aria-label");
  });
});
