import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/svelte";
import Page from "./+page.svelte";
import { goto } from "$app/navigation";
import { getFullGraphData, getGraphData } from "$shared/api/graph";
import { getNotes, getNote, createNote, updateNote } from "$shared/api/notes";
import { createLink, getNoteLinks } from "$shared/api/links";

vi.mock("$shared/api/notes", () => ({
  getNotes: vi.fn(),
  getNote: vi.fn(),
  createNote: vi.fn(),
  updateNote: vi.fn(),
  deleteNote: vi.fn(),
  deleteNotesBatch: vi.fn(),
  restoreNote: vi.fn(),
  publishNote: vi.fn(),
  unpublishNote: vi.fn(),
  getSuggestions: vi.fn(),
  searchNotes: vi.fn(),
}));

vi.mock("$shared/api/graph", () => ({
  getGraphData: vi.fn(),
  getFullGraphData: vi.fn(),
}));

vi.mock("$shared/api/links", () => ({
  getLinks: vi.fn(),
  getLink: vi.fn(),
  createLink: vi.fn(),
  deleteLink: vi.fn(),
  getNoteLinks: vi.fn(),
  deleteAllNoteLinks: vi.fn(),
}));

vi.mock("$shared/stores/auth.svelte", async () => {
  const actual = await vi.importActual<typeof import("$shared/stores/auth.svelte")>(
    "$shared/stores/auth.svelte"
  );
  return {
    ...actual,
    isAuthenticated: vi.fn(() => true),
    initAuth: vi.fn(),
  };
});

const KNOWLEDGE_CORE_ID = "00000000-0000-0000-0000-000000000001";

const mockKnowledgeCore = {
  id: KNOWLEDGE_CORE_ID,
  title: "Knowledge Core",
  content: "Help content",
  type: "technical",
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  metadata: {},
};

const mockGraphNode = {
  id: "node-1",
  title: "Test Note",
  type: "star",
  created_at: new Date().toISOString(),
};

const mockGraph = {
  nodes: [mockGraphNode],
  links: [],
};

const mockNote = {
  id: "note-1",
  title: "Local Note",
  content: "Local content",
  type: "planet",
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
  metadata: {},
};

describe("Graph page - Cosmic Cockpit integration", () => {
  beforeEach(() => {
    vi.mocked(getFullGraphData).mockResolvedValue(mockGraph);
    vi.mocked(getGraphData).mockResolvedValue({
      nodes: [mockNote],
      links: [],
    });
    vi.mocked(getNotes).mockResolvedValue([mockNote]);
    vi.mocked(getNote).mockImplementation(async (id: string) => {
      if (id === KNOWLEDGE_CORE_ID) return mockKnowledgeCore;
      return { ...mockNote, id };
    });
    vi.mocked(createNote).mockResolvedValue({ ...mockNote, id: "new-note" });
    vi.mocked(updateNote).mockResolvedValue({ ...mockNote, title: "Updated" });
    vi.mocked(createLink).mockResolvedValue({
      id: "link-1",
      source_note_id: "node-1",
      target_note_id: "node-1",
      link_type: "related",
      weight: 1,
      metadata: {},
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    });
    vi.mocked(getNoteLinks).mockResolvedValue([]);
    vi.mocked(goto).mockClear();
  });

  it("loads and renders the full graph by default", async () => {
    render(Page);

    await waitFor(
      () => {
        expect(screen.getByTestId("graph-canvas")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    expect(getFullGraphData).toHaveBeenCalledWith(0, undefined, false);
    expect(getNote).toHaveBeenCalledWith(KNOWLEDGE_CORE_ID);
  });

  it("toggles to local graph view and loads centered graph", async () => {
    render(Page);

    await waitFor(
      () => {
        expect(screen.getByTestId("graph-canvas")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    const toggle = screen.getByTestId("full-graph-toggle");
    await fireEvent.change(toggle, { target: { checked: false } });

    await waitFor(
      () => {
        expect(getNotes).toHaveBeenCalled();
      },
      { timeout: 2000 }
    );

    expect(getGraphData).toHaveBeenCalledWith("note-1", 3);
  });

  it("navigates to list view when the list view toggle is clicked", async () => {
    render(Page);

    await waitFor(
      () => {
        expect(screen.getByTestId("view-toggle-list")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    const listButton = screen.getByTestId("view-toggle-list");
    await fireEvent.click(listButton);

    expect(goto).toHaveBeenCalledWith("/");
  });

  it("opens note details in the right panel when a note tree item is selected", async () => {
    render(Page);

    await waitFor(
      () => {
        expect(screen.getAllByTestId("cockpit-note-tree-item").length).toBeGreaterThan(0);
      },
      { timeout: 2000 }
    );

    const [firstTreeItem] = screen.getAllByTestId("cockpit-note-tree-item");
    await fireEvent.click(firstTreeItem);

    await waitFor(
      () => {
        expect(screen.getByTestId("cockpit-right-panel")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    expect(getNote).toHaveBeenCalledWith("node-1");
    expect(getNoteLinks).toHaveBeenCalledWith("node-1");
  });

  it("opens the create note modal from the cockpit top panel", async () => {
    render(Page);

    await waitFor(
      () => {
        expect(screen.getByTestId("graph-canvas")).toBeInTheDocument();
      },
      { timeout: 2000 }
    );

    const createButton = screen.getByTestId("create-note-button");
    await fireEvent.click(createButton);

    expect(screen.getByRole("dialog")).toBeInTheDocument();
  });
});
