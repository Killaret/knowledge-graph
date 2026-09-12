import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/svelte";
import CockpitTypeFilter from "./CockpitTypeFilter.svelte";

describe("CockpitTypeFilter", () => {
  const filters = [
    { id: "all", label: "All", emoji: "🔎", description: "Show all notes", example: "note, bookmark" },
    { id: "note", label: "Notes", emoji: "📝" },
    { id: "bookmark", label: "Bookmarks", emoji: "📑", example: "https://example.com" },
  ];

  beforeEach(() => {
    cleanup();
  });

  it("renders all filter chips", () => {
    render(CockpitTypeFilter, { props: { filters, selected: "all", onSelect: vi.fn() } });

    expect(screen.getByText("All")).toBeInTheDocument();
    expect(screen.getByText("Notes")).toBeInTheDocument();
    expect(screen.getByText("Bookmarks")).toBeInTheDocument();
  });

  it("marks selected chip as active", () => {
    render(CockpitTypeFilter, { props: { filters, selected: "note", onSelect: vi.fn() } });

    const noteChip = screen.getByTestId("filter-chip-note");
    const allChip = screen.getByTestId("filter-chip-all");

    expect(noteChip).toHaveAttribute("aria-pressed", "true");
    expect(noteChip).toHaveClass("active");
    expect(allChip).toHaveAttribute("aria-pressed", "false");
    expect(allChip).not.toHaveClass("active");
  });

  it("calls onSelect when a chip is clicked", async () => {
    const onSelect = vi.fn();
    render(CockpitTypeFilter, { props: { filters, selected: "all", onSelect } });

    const notesChip = screen.getByTestId("filter-chip-note");
    await fireEvent.click(notesChip);

    expect(onSelect).toHaveBeenCalledWith("note");
  });

  it("renders count badge when type count is provided", () => {
    render(CockpitTypeFilter, {
      props: { filters, selected: "all", onSelect: vi.fn(), typeCounts: { all: 12, note: 5 } },
    });

    expect(screen.getByTestId("filter-count-all")).toHaveTextContent("12");
    expect(screen.getByTestId("filter-count-note")).toHaveTextContent("5");
    expect(screen.queryByTestId("filter-count-bookmark")).not.toBeInTheDocument();
  });

  it("shows description and example in title when available", () => {
    render(CockpitTypeFilter, {
      props: { filters, selected: "all", onSelect: vi.fn() },
    });

    const allChip = screen.getByTestId("filter-chip-all");
    const title = allChip.getAttribute("title");
    expect(title).toMatch(/example/i);
  });

  it("falls back to generic filter label when description is missing", () => {
    render(CockpitTypeFilter, {
      props: { filters, selected: "all", onSelect: vi.fn() },
    });

    const notesChip = screen.getByTestId("filter-chip-note");
    const title = notesChip.getAttribute("title");
    expect(title).toMatch(/Notes/i);
  });
});
