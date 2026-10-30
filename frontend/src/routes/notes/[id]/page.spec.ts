import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/svelte";
import { writable } from "svelte/store";
import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { isAuthenticated } from "$shared/stores/auth.svelte";

vi.mock("$shared/api/notes", () => ({
  getNote: vi.fn(),
  getSuggestions: vi.fn(),
  deleteNote: vi.fn(),
  createNote: vi.fn(),
  getNotes: vi.fn(),
  updateNote: vi.fn(),
  deleteNotesBatch: vi.fn(),
  restoreNote: vi.fn(),
  publishNote: vi.fn(),
  unpublishNote: vi.fn(),
  searchNotes: vi.fn(),
}));

vi.mock("$shared/api/graph", () => ({
  getGraphData: vi.fn(),
}));

vi.mock("$shared/stores/auth.svelte", async () => {
  const actual = await vi.importActual<typeof import("$shared/stores/auth.svelte")>(
    "$shared/stores/auth.svelte"
  );
  return {
    ...actual,
    isAuthenticated: vi.fn(() => true),
    initAuth: vi.fn(),
    currentUser: vi.fn(() => null),
    isInitialized: vi.fn(() => true),
  };
});

vi.mock("$app/stores", () => ({
  page: writable({
    url: new URL("http://localhost/notes/note-1"),
    params: { id: "note-1" },
    route: { id: null },
    status: 200,
    error: null,
    data: {},
    form: undefined,
  }),
  updated: { subscribe: vi.fn(() => () => {}), check: vi.fn() },
}));

import { getNote, getSuggestions } from "$shared/api/notes";
import type { Note } from "$shared/api/notes";
import { getGraphData } from "$shared/api/graph";

function setNoteResponse(value: Note) {
  vi.mocked(getNote).mockResolvedValue(value);
}

describe("Note detail page", () => {
  beforeEach(() => {
    vi.mocked(isAuthenticated).mockReturnValue(true);
    vi.mocked(goto).mockClear();
    vi.mocked(getNote).mockReset();
    vi.mocked(getSuggestions).mockReset().mockResolvedValue([]);
    vi.mocked(getGraphData).mockReset().mockResolvedValue({ nodes: [], links: [] });
  });

  it("renders a loaded note with links and suggestions", async () => {
    setNoteResponse({
      id: "note-1",
      title: "Test note",
      content: "Note body",
      type: "star",
      is_public: true,
      created_at: "2025-01-01T00:00:00Z",
      updated_at: "2025-01-02T00:00:00Z",
      metadata: { tags: ["tag1"], keywords: ["kw1"] },
    });
    vi.mocked(getGraphData).mockResolvedValue({
      nodes: [
        { id: "note-1", title: "Test note", type: "star" },
        { id: "note-2", title: "Other note", type: "planet" },
      ],
      links: [
        { id: "l1", source: "note-1", target: "note-2", link_type: "relates_to", weight: 0.8 },
      ],
    });
    vi.mocked(getSuggestions).mockResolvedValue([
      { note_id: "note-2", title: "Other note", score: 0.95 },
    ]);

    const Page = (await import("./+page.svelte")).default;
    render(Page);

    await waitFor(() => {
      expect(screen.getByTestId("note-detail-title")).toHaveTextContent("Test note");
    });
    expect(screen.getByText(/#tag1/)).toBeInTheDocument();
    expect(screen.getAllByText(/Other note/).length).toBeGreaterThanOrEqual(1);
  });

  it("shows a 404 error and redirects", async () => {
    const error = Object.assign(new Error("Not found"), {
      response: { status: 404 },
    });
    vi.mocked(getNote).mockRejectedValue(error);

    const Page = (await import("./+page.svelte")).default;
    render(Page);

    await waitFor(() => {
      expect(screen.getByText(/not found/i)).toBeInTheDocument();
    });
  });
});
