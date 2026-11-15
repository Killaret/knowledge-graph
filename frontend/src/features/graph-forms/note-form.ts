import { localInputToIso, remindChoiceToSeconds, type RemindChoice } from "$shared/utils/comet";

export interface NoteFormState {
  showNoteForm: boolean;
  noteFormPosition: { x: number; y: number };
  newNoteTitle: string;
  newNoteContent: string;
  newNoteType: string;
  // COMET-1: поля кометы на быстрой форме графа
  newDueAtLocal: string;
  newRemindChoice: RemindChoice;
  newCustomMinutes: number;
}

export interface NoteFormCallbacks {
  onNoteCreate?: (data: {
    title: string;
    content: string;
    type: string;
    due_at?: string | null;
    remind_before_seconds?: number | null;
  }) => void;
  onFormClose?: () => void;
}

export const NOTE_DEFAULT_TYPE = "star";

export function createNoteFormState(): NoteFormState {
  return {
    showNoteForm: false,
    noteFormPosition: { x: 0, y: 0 },
    newNoteTitle: "",
    newNoteContent: "",
    newNoteType: NOTE_DEFAULT_TYPE,
    newDueAtLocal: "",
    newRemindChoice: "none",
    newCustomMinutes: 60,
  };
}

export function openNoteForm(state: NoteFormState, x: number, y: number): void {
  state.showNoteForm = true;
  state.noteFormPosition = { x, y };
  state.newNoteTitle = "";
  state.newNoteContent = "";
  state.newNoteType = NOTE_DEFAULT_TYPE;
  state.newDueAtLocal = "";
  state.newRemindChoice = "none";
  state.newCustomMinutes = 60;
}

export function closeNoteForm(state: NoteFormState): void {
  state.showNoteForm = false;
  state.newNoteTitle = "";
  state.newNoteContent = "";
  state.newNoteType = NOTE_DEFAULT_TYPE;
  state.newDueAtLocal = "";
  state.newRemindChoice = "none";
  state.newCustomMinutes = 60;
}

export function createNote(state: NoteFormState, callbacks: NoteFormCallbacks): void {
  if (state.newNoteTitle.trim() && callbacks.onNoteCreate) {
    const isComet = state.newNoteType === "comet";
    callbacks.onNoteCreate({
      title: state.newNoteTitle.trim(),
      content: state.newNoteContent.trim(),
      type: state.newNoteType,
      ...(isComet
        ? {
            due_at: localInputToIso(state.newDueAtLocal),
            remind_before_seconds: remindChoiceToSeconds(
              state.newRemindChoice,
              state.newCustomMinutes
            ),
          }
        : {}),
    });
  }
  closeNoteForm(state);
  callbacks.onFormClose?.();
}

export function isNoteFormValid(state: NoteFormState): boolean {
  return state.newNoteTitle.trim().length > 0;
}
