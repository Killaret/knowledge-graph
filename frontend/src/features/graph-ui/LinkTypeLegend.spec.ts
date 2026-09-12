import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, cleanup, waitFor } from "@testing-library/svelte";
import LinkTypeLegend from "./LinkTypeLegend.svelte";
import { LinkType } from "$entities";

describe("LinkTypeLegend", () => {
  const allTypes = LinkType.ALL_TYPES;

  beforeEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("renders all link types", () => {
    render(LinkTypeLegend, { props: { hiddenTypes: [] } });

    for (const type of allTypes) {
      expect(screen.getByText(type.label)).toBeInTheDocument();
    }
  });

  it("collapses and expands content", async () => {
    render(LinkTypeLegend, { props: { hiddenTypes: [], collapsible: true } });

    const header = screen.getByRole("button", { name: /link types/i });
    await fireEvent.click(header);

    await waitFor(() => {
      expect(screen.queryByText(allTypes[0].label)).not.toBeInTheDocument();
    });

    await fireEvent.click(header);

    await waitFor(() => {
      expect(screen.getByText(allTypes[0].label)).toBeInTheDocument();
    });
  });

  it("toggles a link type and calls onToggle", async () => {
    const onToggle = vi.fn();
    render(LinkTypeLegend, { props: { hiddenTypes: [], onToggle } });

    const firstType = allTypes[0];
    const chip = screen.getByText(firstType.label).closest("button") as HTMLButtonElement;
    expect(chip).not.toBeDisabled();
    await fireEvent.click(chip);

    expect(onToggle).toHaveBeenCalledWith(firstType.type);
  });

  it("disables toggle when not interactive", () => {
    render(LinkTypeLegend, { props: { hiddenTypes: [] } });

    const firstType = allTypes[0];
    const chip = screen.getByText(firstType.label).closest("button") as HTMLButtonElement;
    expect(chip).toBeDisabled();
  });

  it("show all and hide all trigger onToggle for visible state", async () => {
    const onToggle = vi.fn();
    const hiddenTypes = [allTypes[0].type, allTypes[1].type];
    render(LinkTypeLegend, {
      props: { hiddenTypes, onToggle },
    });

    const showAll = screen.getByRole("button", { name: /^Show all$/i });
    const hideAll = screen.getByRole("button", { name: /^Hide all$/i });

    expect(showAll).not.toBeDisabled();
    expect(hideAll).not.toBeDisabled();

    await fireEvent.click(showAll);
    for (const type of allTypes) {
      if (hiddenTypes.includes(type.type)) {
        expect(onToggle).toHaveBeenCalledWith(type.type);
      }
    }

    onToggle.mockClear();
    await fireEvent.click(hideAll);
    for (const type of allTypes) {
      if (!hiddenTypes.includes(type.type)) {
        expect(onToggle).toHaveBeenCalledWith(type.type);
      }
    }
  });

  it("disables show all when all types are visible", () => {
    render(LinkTypeLegend, { props: { hiddenTypes: [], onToggle: vi.fn() } });

    const showAll = screen.getByRole("button", { name: /^Show all$/i });
    const hideAll = screen.getByRole("button", { name: /^Hide all$/i });

    expect(showAll).toBeDisabled();
    expect(hideAll).not.toBeDisabled();
  });

  it("disables hide all when all types are hidden", () => {
    const hiddenTypes = allTypes.map((t) => t.type);
    render(LinkTypeLegend, { props: { hiddenTypes, onToggle: vi.fn() } });

    const showAll = screen.getByRole("button", { name: /^Show all$/i });
    const hideAll = screen.getByRole("button", { name: /^Hide all$/i });

    expect(showAll).not.toBeDisabled();
    expect(hideAll).toBeDisabled();
  });

  it("calls onMinWeightChange when range input changes", async () => {
    const onMinWeightChange = vi.fn();
    render(LinkTypeLegend, {
      props: { hiddenTypes: [], showMinWeight: true, minWeight: 0.3, onMinWeightChange },
    });

    const input = screen.getByRole("slider");
    await fireEvent.input(input, { target: { value: "0.7" } });

    expect(onMinWeightChange).toHaveBeenCalledWith(0.7);
  });
});
