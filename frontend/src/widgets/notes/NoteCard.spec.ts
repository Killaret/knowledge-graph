import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/svelte";
import NoteCard from "./NoteCard.svelte";
import type { Note } from "$shared/api/notes";
import tippy from "tippy.js";

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

vi.mock("tippy.js", () => ({
  default: vi.fn(() => ({
    show: vi.fn(),
    hide: vi.fn(),
    destroy: vi.fn(),
  })),
}));

function createNote(overrides: Partial<Note> = {}): Note {
  return {
    id: "123e4567-e89b-12d3-a456-426614174000",
    title: "Test Note Title",
    content: "Test note content",
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    metadata: { type: "star" },
    type: "star",
    ...overrides,
  };
}

beforeEach(() => {
  vi.mocked(tippy).mockClear();
});

describe("NoteCard", () => {
  it("renders note title and content", () => {
    render(NoteCard, { props: { note: createNote() } });

    expect(screen.getByText("Test Note Title")).toBeInTheDocument();
    expect(screen.getByText("Test note content")).toBeInTheDocument();
  });

  it("displays formatted date", () => {
    render(NoteCard, { props: { note: createNote() } });

    const dateElement = screen.getByTestId("note-date");
    expect(dateElement).toBeInTheDocument();
    expect(dateElement.textContent).toMatch(/Star lit:/);
  });

  it("renders different note types with correct styling", () => {
    const starNote = createNote({ type: "star" });
    const { container } = render(NoteCard, { props: { note: starNote } });

    const card = container.querySelector('[data-note-type="star"]');
    expect(card).toBeInTheDocument();
  });

  it("handles click events", async () => {
    const onClick = vi.fn();
    render(NoteCard, { props: { note: createNote(), onClick } });

    const card = document.querySelector(".note-card");
    expect(card).toBeTruthy();
    expect(onClick).not.toHaveBeenCalled();

    await fireEvent.click(card!);
    expect(onClick).toHaveBeenCalledWith(expect.objectContaining({ title: "Test Note Title" }));
  });

  it("renders with minimal note data", () => {
    const minimalNote = createNote({
      title: "Minimal",
      content: "",
      type: undefined,
    });

    render(NoteCard, { props: { note: minimalNote } });
    expect(screen.getByText("Minimal")).toBeInTheDocument();
  });

  it("shows new indicator for notes created within the last 24 hours", () => {
    const newNote = createNote({ created_at: new Date().toISOString() });
    const { container } = render(NoteCard, { props: { note: newNote } });

    const indicator = container.querySelector(".note-card__indicator--new");
    expect(indicator).toBeInTheDocument();
  });

  it("does not show new indicator for notes older than 24 hours", () => {
    const oldDate = new Date(Date.now() - 25 * 60 * 60 * 1000).toISOString();
    const oldNote = createNote({ created_at: oldDate, updated_at: oldDate });
    const { container } = render(NoteCard, { props: { note: oldNote } });

    const indicator = container.querySelector(".note-card__indicator--new");
    expect(indicator).not.toBeInTheDocument();
  });

  it("shows updated indicator for notes updated within the last hour", () => {
    const now = new Date().toISOString();
    const past = new Date(Date.now() - 25 * 60 * 60 * 1000).toISOString();
    const updatedNote = createNote({ created_at: past, updated_at: now });
    const { container } = render(NoteCard, { props: { note: updatedNote } });

    const indicator = container.querySelector(".note-card__indicator--updated");
    expect(indicator).toBeInTheDocument();
    expect(screen.getByTestId("note-updated-date")).toBeInTheDocument();
  });

  it("does not show updated indicator when created_at equals updated_at", () => {
    const now = new Date().toISOString();
    const note = createNote({ created_at: now, updated_at: now });
    const { container } = render(NoteCard, { props: { note } });

    const indicator = container.querySelector(".note-card__indicator--updated");
    expect(indicator).not.toBeInTheDocument();
  });

  it("renders dust style for quick-capture notes", () => {
    const dustNote = createNote({ type: "dust" });
    const { container } = render(NoteCard, { props: { note: dustNote } });

    const card = container.querySelector(".note-card.dust");
    expect(card).toBeInTheDocument();
  });

  it("renders selection checkbox and calls onSelect when toggled", async () => {
    const onSelect = vi.fn();
    render(NoteCard, {
      props: { note: createNote(), selectMode: true, onSelect },
    });

    const checkbox = screen.getByRole("checkbox") as HTMLInputElement;
    expect(checkbox).toBeInTheDocument();
    expect(checkbox.checked).toBe(false);

    await fireEvent.click(checkbox);
    expect(onSelect).toHaveBeenCalledWith(
      expect.objectContaining({ title: "Test Note Title" }),
      true
    );
  });

  it("renders selected state when selected prop is true", () => {
    const { container } = render(NoteCard, {
      props: { note: createNote(), selected: true },
    });

    const card = container.querySelector(".note-card.selected");
    expect(card).toBeInTheDocument();
  });

  it("renders view, edit and delete actions in the tooltip by default", async () => {
    render(NoteCard, { props: { note: createNote() } });

    await waitFor(() => expect(tippy).toHaveBeenCalled());

    const [, options] = vi.mocked(tippy).mock.calls[0] as [unknown, { content: string }];
    expect(options.content).toContain('data-action="view"');
    expect(options.content).toContain('data-action="edit"');
    expect(options.content).toContain('data-action="delete"');
  });

  it("hides edit and delete actions in the tooltip when readonly", async () => {
    render(NoteCard, { props: { note: createNote(), readonly: true } });

    await waitFor(() => expect(tippy).toHaveBeenCalled());

    const [, options] = vi.mocked(tippy).mock.calls[0] as [unknown, { content: string }];
    expect(options.content).toContain('data-action="view"');
    expect(options.content).not.toContain('data-action="edit"');
    expect(options.content).not.toContain('data-action="delete"');
  });
});
