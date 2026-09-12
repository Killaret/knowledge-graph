import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, fireEvent } from "@testing-library/svelte";
import { goto } from "$app/navigation";
import { createNote } from "$shared/api/notes";
import type { Note } from "$shared/api/notes";

vi.mock("$shared/api/notes", () => ({
  createNote: vi.fn(),
  getNotes: vi.fn(),
  getNote: vi.fn(),
  updateNote: vi.fn(),
  deleteNote: vi.fn(),
  deleteNotesBatch: vi.fn(),
  restoreNote: vi.fn(),
  publishNote: vi.fn(),
  unpublishNote: vi.fn(),
  getSuggestions: vi.fn(),
  searchNotes: vi.fn(),
}));

describe("New note page", () => {
  beforeEach(() => {
    vi.mocked(goto).mockClear();
    vi.mocked(createNote).mockReset();
  });

  it("shows a validation error when title is empty", async () => {
    const Page = (await import("./+page.svelte")).default;
    const { container } = render(Page);
    const form = container.querySelector("form") as HTMLFormElement;
    await fireEvent.submit(form);
    await waitFor(() => {
      expect(screen.getByText(/Title is required/i)).toBeInTheDocument();
    });
  });

  it("creates a note and navigates to it", async () => {
    vi.mocked(createNote).mockResolvedValue({
      id: "note-1",
      title: "Test note",
      content: "Content",
      metadata: {},
      created_at: "2025-01-01T00:00:00Z",
      updated_at: "2025-01-01T00:00:00Z",
    } as Note);
    const Page = (await import("./+page.svelte")).default;
    render(Page);

    const titleInput = screen.getByPlaceholderText(/title/i);
    const contentInput = screen.getByPlaceholderText(/content/i);
    await fireEvent.input(titleInput, { target: { value: "Test note" } });
    await fireEvent.input(contentInput, { target: { value: "Content" } });

    const submitButton = screen.getByRole("button", { name: /create/i });
    await fireEvent.click(submitButton);

    await waitFor(() => {
      expect(goto).toHaveBeenCalledWith("/notes/note-1");
    });
    expect(createNote).toHaveBeenCalledWith({ title: "Test note", content: "Content", metadata: {} });
  });

  it("shows an error when note creation fails", async () => {
    vi.mocked(createNote).mockRejectedValue(new Error("create failed"));
    const Page = (await import("./+page.svelte")).default;
    render(Page);

    const titleInput = screen.getByPlaceholderText(/title/i);
    await fireEvent.input(titleInput, { target: { value: "Test note" } });

    const submitButton = screen.getByRole("button", { name: /create/i });
    await fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText(/Failed to create note/i)).toBeInTheDocument();
    });
  });
});
