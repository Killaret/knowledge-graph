import { describe, it, expect, vi } from "vitest";
import {
  createNoteFormState,
  openNoteForm,
  closeNoteForm,
  createNote,
  isNoteFormValid,
  type NoteFormCallbacks,
} from "./note-form";

describe("note-form", () => {
  it("creates default form state", () => {
    const state = createNoteFormState();
    expect(state).toEqual({
      showNoteForm: false,
      noteFormPosition: { x: 0, y: 0 },
      newNoteTitle: "",
      newNoteContent: "",
      newNoteType: "planet",
    });
  });

  it("opens the note form at a position", () => {
    const state = createNoteFormState();
    openNoteForm(state, 100, 200);
    expect(state.showNoteForm).toBe(true);
    expect(state.noteFormPosition).toEqual({ x: 100, y: 200 });
    expect(state.newNoteTitle).toBe("");
    expect(state.newNoteType).toBe("planet");
  });

  it("closes and resets the form", () => {
    const state = createNoteFormState();
    openNoteForm(state, 10, 20);
    state.newNoteTitle = "Keep";
    closeNoteForm(state);
    expect(state.showNoteForm).toBe(false);
    expect(state.newNoteTitle).toBe("");
    expect(state.newNoteContent).toBe("");
  });

  it("creates a note when valid and invokes callbacks", () => {
    const state = createNoteFormState();
    state.newNoteTitle = "  My Note  ";
    state.newNoteContent = "Content";
    state.newNoteType = "star";

    const callbacks: NoteFormCallbacks = {
      onNoteCreate: vi.fn(),
      onFormClose: vi.fn(),
    };

    createNote(state, callbacks);
    expect(callbacks.onNoteCreate).toHaveBeenCalledWith({
      title: "My Note",
      content: "Content",
      type: "star",
    });
    expect(callbacks.onFormClose).toHaveBeenCalled();
    expect(state.newNoteTitle).toBe("");
  });

  it("does not create a note with an empty title", () => {
    const state = createNoteFormState();
    const callbacks: NoteFormCallbacks = {
      onNoteCreate: vi.fn(),
      onFormClose: vi.fn(),
    };

    createNote(state, callbacks);
    expect(callbacks.onNoteCreate).not.toHaveBeenCalled();
    expect(callbacks.onFormClose).toHaveBeenCalled();
  });

  it("reports form validity", () => {
    const valid = createNoteFormState();
    valid.newNoteTitle = "Title";
    expect(isNoteFormValid(valid)).toBe(true);

    const invalid = createNoteFormState();
    expect(isNoteFormValid(invalid)).toBe(false);
  });
});
