import { describe, it, expect, vi, afterEach } from "vitest";
import { render, cleanup, fireEvent } from "@testing-library/svelte";
import CockpitRightPanel from "./CockpitRightPanel.svelte";

describe("CockpitRightPanel", () => {
  afterEach(() => {
    cleanup();
  });

  const notes = [{ id: "note-1", title: "Public note", type: "star" }];
  const links: { source: string; target: string; link_type?: string; weight?: number }[] = [];

  it("renders PublicNoteDetails when not authenticated", () => {
    const { getByTestId } = render(CockpitRightPanel, {
      props: {
        nodeId: "note-1",
        isAuthenticated: false,
        notes,
        links,
      },
    });

    expect(getByTestId("public-note-details")).toBeInTheDocument();
  });

  it("renders empty state when no node is selected", () => {
    const { getByTestId, queryByTestId } = render(CockpitRightPanel, {
      props: {
        nodeId: null,
        isAuthenticated: false,
        notes,
        links,
      },
    });

    expect(getByTestId("cockpit-right-panel")).toBeInTheDocument();
    expect(queryByTestId("public-note-details")).not.toBeInTheDocument();
  });

  it("calls onNodeSelect(null) when public note details close is clicked", () => {
    const onNodeSelect = vi.fn();
    const { getByTitle } = render(CockpitRightPanel, {
      props: {
        nodeId: "note-1",
        isAuthenticated: false,
        notes,
        links,
        onNodeSelect,
      },
    });

    const closeBtn = getByTitle("Close panel");
    fireEvent.click(closeBtn);
    expect(onNodeSelect).toHaveBeenCalledWith(null);
  });
});
