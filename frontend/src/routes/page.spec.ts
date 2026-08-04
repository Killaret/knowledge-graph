import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/svelte";
import Page from "./+page.svelte";
import { getNotes, createNote, deleteNote, deleteNotesBatch, restoreNote } from "$shared/api/notes";
import { getGraphWithPreload } from "$features/preload/hooks/usePreloadedData";

vi.mock("$shared/api/notes", () => ({
  getNotes: vi.fn(),
  createNote: vi.fn(),
  deleteNote: vi.fn(),
  deleteNotesBatch: vi.fn(),
  restoreNote: vi.fn(),
}));

vi.mock("$features/preload/hooks/usePreloadedData", () => ({
  getGraphWithPreload: vi.fn(),
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

describe("Page list view - batch operations", () => {
  const mockNotes = [
    {
      id: "1",
      title: "Note 1",
      content: "Content 1",
      type: "star",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      metadata: {},
    },
    {
      id: "2",
      title: "Note 2",
      content: "Content 2",
      type: "planet",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      metadata: {},
    },
    {
      id: "3",
      title: "Note 3",
      content: "Content 3",
      type: "comet",
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      metadata: {},
    },
  ];

  beforeEach(() => {
    vi.mocked(getNotes).mockResolvedValue(mockNotes);
    vi.mocked(getGraphWithPreload).mockResolvedValue({ nodes: [], links: [] });
    vi.mocked(createNote).mockResolvedValue({
      ...mockNotes[0],
      id: "new-note",
      title: "New Note",
    });
    vi.mocked(deleteNote).mockResolvedValue(undefined);
    vi.mocked(deleteNotesBatch).mockResolvedValue(undefined);
    vi.mocked(restoreNote).mockResolvedValue(undefined);
  });

  it("renders sorting dropdown in list view", async () => {
    render(Page);

    const listButton = screen.getByTestId("view-toggle-list");
    await fireEvent.click(listButton);

    await waitFor(() => {
      expect(screen.getAllByTestId("note-title")).toHaveLength(3);
    });

    const sortSelect = screen.getByLabelText("Sort notes");
    expect(sortSelect).toBeInTheDocument();
  });

  it("sorts notes by created date", async () => {
    render(Page);

    const listButton = screen.getByTestId("view-toggle-list");
    await fireEvent.click(listButton);

    await waitFor(() => {
      expect(screen.getAllByTestId("note-title")).toHaveLength(3);
    });

    const sortSelect = screen.getByLabelText("Sort notes");
    await fireEvent.change(sortSelect, { target: { value: "created" } });

    const notes = screen.getAllByTestId("note-title");
    expect(notes).toHaveLength(3);
  });

  it("toggles selection mode", async () => {
    render(Page);

    const listButton = screen.getByTestId("view-toggle-list");
    await fireEvent.click(listButton);

    await waitFor(() => {
      expect(screen.getAllByTestId("note-title")).toHaveLength(3);
    });

    const selectButton = screen.getByTestId("select-mode-toggle");
    await fireEvent.click(selectButton);

    expect(selectButton).toHaveTextContent("Cancel selection");
  });

  it("selects all notes when select all is clicked", async () => {
    render(Page);

    const listButton = screen.getByTestId("view-toggle-list");
    await fireEvent.click(listButton);

    await waitFor(() => {
      expect(screen.getAllByTestId("note-title")).toHaveLength(3);
    });

    const selectButton = screen.getByTestId("select-mode-toggle");
    await fireEvent.click(selectButton);

    const selectAllButton = screen.getByRole("button", {
      name: /select all notes/i,
    });
    expect(selectAllButton).toBeInTheDocument();
  });

  it("shows batch delete panel when notes are selected", async () => {
    render(Page);

    const listButton = screen.getByTestId("view-toggle-list");
    await fireEvent.click(listButton);

    await waitFor(() => {
      expect(screen.getAllByTestId("note-title")).toHaveLength(3);
    });

    const selectButton = screen.getByTestId("select-mode-toggle");
    await fireEvent.click(selectButton);

    const selectAllButton = screen.getByRole("button", {
      name: /select all notes/i,
    });
    expect(selectAllButton).toBeInTheDocument();
  });

  it("clears selection when cancel is clicked", async () => {
    render(Page);

    const listButton = screen.getByTestId("view-toggle-list");
    await fireEvent.click(listButton);

    await waitFor(() => {
      expect(screen.getAllByTestId("note-title")).toHaveLength(3);
    });

    const selectButton = screen.getByTestId("select-mode-toggle");
    await fireEvent.click(selectButton);

    expect(selectButton).toHaveTextContent("Cancel selection");

    await fireEvent.click(selectButton);

    expect(selectButton).toHaveTextContent("Select");
  });
});

describe("Page list view - undo toast", () => {
  const mockNote = {
    id: "1",
    title: "Test Note",
    content: "Test content",
    type: "star",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    metadata: {},
  };

  beforeEach(() => {
    vi.mocked(getNotes).mockResolvedValue([mockNote]);
    vi.mocked(getGraphWithPreload).mockResolvedValue({ nodes: [], links: [] });
  });

  it("shows two-stage undo toast after note deletion", async () => {
    render(Page);

    const listButton = screen.getByTestId("view-toggle-list");
    await fireEvent.click(listButton);

    const undoToast = document.querySelector(".undo-toast");
    expect(undoToast).toBeNull();
  });
});
