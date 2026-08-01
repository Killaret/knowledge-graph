export interface NoteFormState {
  showNoteForm: boolean;
  noteFormPosition: { x: number; y: number };
  newNoteTitle: string;
  newNoteContent: string;
  newNoteType: string;
}

export interface NoteFormCallbacks {
  onNoteCreate?: (data: { title: string; content: string; type: string }) => void;
  onFormClose?: () => void;
}

export const NOTE_DEFAULT_TYPE = "planet";

export function createNoteFormState(): NoteFormState {
  return {
    showNoteForm: false,
    noteFormPosition: { x: 0, y: 0 },
    newNoteTitle: "",
    newNoteContent: "",
    newNoteType: NOTE_DEFAULT_TYPE,
  };
}

export function openNoteForm(state: NoteFormState, x: number, y: number): void {
  state.showNoteForm = true;
  state.noteFormPosition = { x, y };
  state.newNoteTitle = "";
  state.newNoteContent = "";
  state.newNoteType = NOTE_DEFAULT_TYPE;
}

export function closeNoteForm(state: NoteFormState): void {
  state.showNoteForm = false;
  state.newNoteTitle = "";
  state.newNoteContent = "";
  state.newNoteType = NOTE_DEFAULT_TYPE;
}

export function createNote(state: NoteFormState, callbacks: NoteFormCallbacks): void {
  if (state.newNoteTitle.trim() && callbacks.onNoteCreate) {
    callbacks.onNoteCreate({
      title: state.newNoteTitle.trim(),
      content: state.newNoteContent.trim(),
      type: state.newNoteType,
    });
  }
  closeNoteForm(state);
  callbacks.onFormClose?.();
}

export function isNoteFormValid(state: NoteFormState): boolean {
  return state.newNoteTitle.trim().length > 0;
}
