import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/svelte";
import { writable } from "svelte/store";
import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { getNote, updateNote } from "$shared/api/notes";
import type { Note } from "$shared/api/notes";

vi.mock("$shared/api/notes", () => ({
  getNote: vi.fn(),
  updateNote: vi.fn(),
  createNote: vi.fn(),
  getNotes: vi.fn(),
  deleteNote: vi.fn(),
  deleteNotesBatch: vi.fn(),
  restoreNote: vi.fn(),
  publishNote: vi.fn(),
  unpublishNote: vi.fn(),
  getSuggestions: vi.fn(),
  searchNotes: vi.fn(),
}));

vi.mock("$app/stores", () => ({
  page: writable({
    url: new URL("http://localhost/notes/note-1/edit"),
    params: { id: "note-1" },
    route: { id: null },
    status: 200,
    error: null,
    data: {},
    form: undefined,
  }),
  updated: { subscribe: vi.fn(() => () => {}), check: vi.fn() },
}));

describe("Edit note page", () => {
  beforeEach(() => {
    vi.mocked(goto).mockClear();
    vi.mocked(getNote).mockReset();
    vi.mocked(updateNote).mockReset();
  });

  it("shows loading then note form", async () => {
    vi.mocked(getNote).mockResolvedValue({
      id: "note-1",
      title: "Old title",
      content: "Old content",
      metadata: {},
      created_at: "2025-01-01T00:00:00Z",
      updated_at: "2025-01-01T00:00:00Z",
    } as Note);
    const Page = (await import("./+page.svelte")).default;
    render(Page);
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
    await waitFor(() => {
      expect(screen.getByDisplayValue("Old title")).toBeInTheDocument();
    });
  });

  it("shows an error when the note fails to load", async () => {
    vi.mocked(getNote).mockRejectedValue(new Error("not found"));
    const Page = (await import("./+page.svelte")).default;
    render(Page);
    await waitFor(() => {
      expect(screen.getByText(/not found/i)).toBeInTheDocument();
    });
  });

  it("shows validation error when title is empty", async () => {
    vi.mocked(getNote).mockResolvedValue({
      id: "note-1",
      title: "Old title",
      content: "Old content",
      metadata: {},
      created_at: "2025-01-01T00:00:00Z",
      updated_at: "2025-01-01T00:00:00Z",
    } as Note);
    const Page = (await import("./+page.svelte")).default;
    const { container } = render(Page);
    await waitFor(() => {
      expect(screen.getByDisplayValue("Old title")).toBeInTheDocument();
    });
    const titleInput = screen.getByDisplayValue("Old title") as HTMLInputElement;
    await fireEvent.input(titleInput, { target: { value: "" } });

    const form = container.querySelector("form") as HTMLFormElement;
    await fireEvent.submit(form);
    await waitFor(() => {
      expect(screen.getByText(/Title is required/i)).toBeInTheDocument();
    });
  });

  it("updates the note and navigates", async () => {
    vi.mocked(getNote).mockResolvedValue({
      id: "note-1",
      title: "Old title",
      content: "Old content",
      metadata: {},
      created_at: "2025-01-01T00:00:00Z",
      updated_at: "2025-01-01T00:00:00Z",
    } as Note);
    vi.mocked(updateNote).mockResolvedValue({
      id: "note-1",
      title: "New title",
      content: "New content",
      metadata: {},
      created_at: "2025-01-01T00:00:00Z",
      updated_at: "2025-01-01T00:00:00Z",
    } as Note);
    const Page = (await import("./+page.svelte")).default;
    const { container } = render(Page);
    await waitFor(() => {
      expect(screen.getByDisplayValue("Old title")).toBeInTheDocument();
    });

    const titleInput = screen.getByDisplayValue("Old title") as HTMLInputElement;
    await fireEvent.input(titleInput, { target: { value: "New title" } });

    const form = container.querySelector("form") as HTMLFormElement;
    await fireEvent.submit(form);

    await waitFor(() => {
      expect(goto).toHaveBeenCalledWith("/notes/note-1");
    });
    expect(updateNote).toHaveBeenCalledWith("note-1", { title: "New title", content: "Old content" });
  });

  it("shows an error when update fails", async () => {
    vi.mocked(getNote).mockResolvedValue({
      id: "note-1",
      title: "Old title",
      content: "Old content",
      metadata: {},
      created_at: "2025-01-01T00:00:00Z",
      updated_at: "2025-01-01T00:00:00Z",
    } as Note);
    vi.mocked(updateNote).mockRejectedValue(new Error("update failed"));
    const Page = (await import("./+page.svelte")).default;
    const { container } = render(Page);
    await waitFor(() => {
      expect(screen.getByDisplayValue("Old title")).toBeInTheDocument();
    });

    const form = container.querySelector("form") as HTMLFormElement;
    await fireEvent.submit(form);

    await waitFor(() => {
      expect(screen.getByText(/Failed to update note/i)).toBeInTheDocument();
    });
  });
});
