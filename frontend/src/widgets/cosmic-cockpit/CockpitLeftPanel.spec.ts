import { describe, it, expect, afterEach, vi } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import CockpitLeftPanel from "./CockpitLeftPanel.svelte";
import { graphStore } from "$shared/stores/graph.svelte";
import { LinkType } from "$entities";

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

describe("CockpitLeftPanel — link types filter", () => {
  afterEach(() => {
    graphStore.hiddenLinkTypes = [];
    graphStore.minLinkWeight = 0;
    cleanup();
  });

  it("renders a chip for every link type", () => {
    render(CockpitLeftPanel);

    for (const linkType of LinkType.ALL_TYPES) {
      expect(screen.getByTestId(`link-type-chip-${linkType.type}`)).toBeInTheDocument();
    }
  });

  it("toggles a link type as hidden in graphStore when its chip is clicked", async () => {
    render(CockpitLeftPanel);

    await fireEvent.click(screen.getByTestId("link-type-chip-parent"));
    expect(graphStore.hiddenLinkTypes).toContain("parent");

    await fireEvent.click(screen.getByTestId("link-type-chip-parent"));
    expect(graphStore.hiddenLinkTypes).not.toContain("parent");
  });

  it("'Show all' is disabled from the default (all visible) state and becomes actionable once one type is hidden", async () => {
    render(CockpitLeftPanel);

    expect(screen.getByTestId("link-types-show-all")).toBeDisabled();

    await fireEvent.click(screen.getByTestId("link-type-chip-parent"));
    expect(screen.getByTestId("link-types-show-all")).not.toBeDisabled();

    await fireEvent.click(screen.getByTestId("link-types-show-all"));
    expect(graphStore.hiddenLinkTypes).toEqual([]);
  });

  it("'Hide all' hides every type and is disabled once everything is hidden", async () => {
    render(CockpitLeftPanel);

    expect(screen.getByTestId("link-types-hide-all")).not.toBeDisabled();

    await fireEvent.click(screen.getByTestId("link-types-hide-all"));
    expect(graphStore.hiddenLinkTypes.length).toBe(LinkType.ALL_TYPES.length);
    expect(screen.getByTestId("link-types-hide-all")).toBeDisabled();

    await fireEvent.click(screen.getByTestId("link-types-show-all"));
    expect(graphStore.hiddenLinkTypes).toEqual([]);
  });

  it("renders a min-weight slider bound to graphStore.minLinkWeight", async () => {
    render(CockpitLeftPanel);

    const slider = screen.getByTestId("cockpit-min-weight") as HTMLInputElement;
    expect(slider.value).toBe("0");

    await fireEvent.input(slider, { target: { value: "0.5" } });
    expect(graphStore.minLinkWeight).toBe(0.5);
  });
});
