import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/svelte";
import { tick } from "svelte";
import EditNoteModal from "./EditNoteModal.svelte";

// Мокируем API
const mockGetNote = vi.fn();
const mockUpdateNote = vi.fn();

vi.mock("$shared/api/notes", () => ({
  getNote: (...args: any[]) => mockGetNote(...args),
  updateNote: (...args: any[]) => mockUpdateNote(...args),
}));

// Мокаем lexicon-settings для получения текстов валидации
vi.mock("$shared/stores/lexicon-settings", () => ({
  getMessage: vi.fn().mockResolvedValue("Title is required"),
  mode: {
    subscribe: vi.fn((cb) => {
      cb("standard");
      return () => {};
    }),
  },
}));

describe("EditNoteModal", () => {
  const mockNote = {
    id: "456",
    title: "Existing Note",
    content: "Existing content",
    type: "planet",
    metadata: {},
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-02T00:00:00Z",
  };

  const updatedNote = {
    ...mockNote,
    title: "Updated Title",
    content: "Updated content",
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it("renders modal when open is true", async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    expect(screen.getByText("Edit Note")).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Save Changes" })).toBeInTheDocument();
    });
  });

  it("does not render when open is false", () => {
    render(EditNoteModal, { props: { open: false, noteId: "456" } });

    expect(screen.queryByText("Edit Note")).not.toBeInTheDocument();
  });

  it("loads note data when opened", async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(mockGetNote).toHaveBeenCalledWith("456");
    });

    await waitFor(() => {
      expect(screen.getByDisplayValue("Existing Note")).toBeInTheDocument();
    });

    expect(screen.getByDisplayValue("Existing content")).toBeInTheDocument();
    // TypeSelector показывает выбранный тип через aria-pressed
    const planetButton = screen.getByTestId("type-btn-planet");
    expect(planetButton).toHaveAttribute("aria-pressed", "true");
  });

  it("shows loading state while loading", async () => {
    let resolveGetNote: (value: any) => void;
    const getNotePromise = new Promise((resolve) => {
      resolveGetNote = resolve;
    });
    mockGetNote.mockReturnValueOnce(getNotePromise);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    expect(screen.getByText("Loading...")).toBeInTheDocument();

    // Разрешаем промис
    resolveGetNote!(mockNote);

    await waitFor(() => {
      expect(screen.queryByText("Loading...")).not.toBeInTheDocument();
    });
  });

  it("shows error when loading fails", async () => {
    mockGetNote.mockRejectedValueOnce(new Error("Failed to load"));

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(screen.getByText("Failed to load note")).toBeInTheDocument();
    });
  });

  it("shows validation error when title is empty", async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(screen.getByDisplayValue("Existing Note")).toBeInTheDocument();
    });

    // Очищаем заголовок
    const titleInput = screen.getByTestId("edit-title-input") as HTMLInputElement;
    titleInput.value = "";
    await fireEvent.input(titleInput);
    await tick();

    const saveButton = screen.getByRole("button", { name: "Save Changes" });
    await fireEvent.click(saveButton);
    await tick();

    // Проверяем что отображается ошибка валидации (код ошибки вместо сообщения)
    await waitFor(
      () => {
        expect(screen.getByText("Error: VALIDATION_ERROR")).toBeInTheDocument();
      },
      { timeout: 1000 }
    );

    expect(mockUpdateNote).not.toHaveBeenCalled();
  });

  it("updates note successfully and calls onSuccess", async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);
    mockUpdateNote.mockResolvedValueOnce(updatedNote);
    const mockOnSuccess = vi.fn();

    render(EditNoteModal, {
      props: {
        open: true,
        noteId: "456",
        onSuccess: mockOnSuccess,
      },
    });

    await waitFor(() => {
      expect(screen.getByDisplayValue("Existing Note")).toBeInTheDocument();
    });

    // Изменяем данные
    const titleInput = screen.getByTestId("edit-title-input") as HTMLInputElement;
    titleInput.value = "Updated Title";
    await fireEvent.input(titleInput);
    await tick();

    const contentInput = screen.getByTestId("edit-content-input") as HTMLTextAreaElement;
    contentInput.value = "Updated content";
    await fireEvent.input(contentInput);
    await tick();

    // Сохраняем
    const saveButton = screen.getByRole("button", { name: "Save Changes" });
    await fireEvent.click(saveButton);

    await waitFor(() => {
      expect(mockUpdateNote).toHaveBeenCalledWith("456", {
        title: "Updated Title",
        content: "Updated content",
        type: "planet",
        metadata: {},
      });
    });

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalledWith(updatedNote);
    });

    // Модаль закрывается
    expect(screen.queryByText("Edit Note")).not.toBeInTheDocument();
  });

  it("shows error when update fails", async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);
    mockUpdateNote.mockRejectedValueOnce(new Error("Failed to update"));

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(screen.getByDisplayValue("Existing Note")).toBeInTheDocument();
    });

    const saveButton = screen.getByRole("button", { name: "Save Changes" });
    await fireEvent.click(saveButton);

    await waitFor(() => {
      expect(screen.getByText("Failed to update note")).toBeInTheDocument();
    });

    // Модаль не закрывается
    expect(screen.getByText("Edit Note")).toBeInTheDocument();
  });

  it("shows saving state during update", async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);
    let resolveUpdate: (value: any) => void;
    const updatePromise = new Promise((resolve) => {
      resolveUpdate = resolve;
    });
    mockUpdateNote.mockReturnValueOnce(updatePromise);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(screen.getByDisplayValue("Existing Note")).toBeInTheDocument();
    });

    const saveButton = screen.getByRole("button", { name: "Save Changes" });
    await fireEvent.click(saveButton);

    expect(saveButton).toHaveTextContent("Saving...");
    expect(saveButton).toBeDisabled();
    expect(screen.getByTestId("edit-title-input")).toBeDisabled();

    // Разрешаем промис
    resolveUpdate!(updatedNote);

    await waitFor(() => {
      expect(screen.queryByText("Edit Note")).not.toBeInTheDocument();
    });
  });

  it("closes modal when cancel is clicked", async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(screen.getByText("Edit Note")).toBeInTheDocument();
    });

    const cancelButton = screen.getByRole("button", { name: "Cancel" });
    await fireEvent.click(cancelButton);

    expect(screen.queryByText("Edit Note")).not.toBeInTheDocument();
  });

  it("closes modal when close button is clicked", async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(screen.getByText("Edit Note")).toBeInTheDocument();
    });

    const closeButton = screen.getByLabelText("Close");
    await fireEvent.click(closeButton);

    expect(screen.queryByText("Edit Note")).not.toBeInTheDocument();
  });

  it("updates note type selection", async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(screen.getByDisplayValue("Existing Note")).toBeInTheDocument();
    });

    // Выбираем Comet через TypeSelector (вместо старого select)
    const cometButton = screen.getByTestId("type-btn-comet");
    await fireEvent.click(cometButton);
    await tick();

    // Проверяем что Comet выбран
    expect(cometButton).toHaveAttribute("aria-pressed", "true");
  });

  // COMET-1 этап B: в редактировании у кометы есть дата, напоминание и «Сделано».
  it("populates comet scheduling fields and marks the note done", async () => {
    const dueIso = new Date("2026-03-01T12:00").toISOString();
    mockGetNote.mockResolvedValueOnce({
      ...mockNote,
      type: "comet",
      due_at: dueIso,
      remind_before_seconds: 3600,
      done_at: null,
    });
    mockUpdateNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(screen.getByDisplayValue("Existing Note")).toBeInTheDocument();
    });

    const due = document.querySelector('[data-testid="edit-comet-due"]') as HTMLInputElement;
    expect(due).toBeTruthy();
    const pad = (n: number) => String(n).padStart(2, "0");
    const d = new Date(dueIso);
    const expectedLocal = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
    expect(due.value).toBe(expectedLocal);

    const remind = document.querySelector('[data-testid="edit-comet-remind"]') as HTMLSelectElement;
    expect(remind.value).toBe("1h");

    const doneCheckbox = document.querySelector(
      '[data-testid="edit-comet-done"]'
    ) as HTMLInputElement;
    expect(doneCheckbox.checked).toBe(false);
    await fireEvent.click(doneCheckbox);
    await tick();

    const submitButton = screen.getByRole("button", { name: "Save Changes" });
    await fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockUpdateNote).toHaveBeenCalled();
    });
    const payload = mockUpdateNote.mock.calls[0][1];
    expect(payload.due_at).toBe(dueIso);
    expect(payload.remind_before_seconds).toBe(3600);
    // done=true без сохранённого done_at ставит «сейчас»
    expect(typeof payload.done_at).toBe("string");
  });

  it("clears done_at when the done checkbox is unchecked", async () => {
    mockGetNote.mockResolvedValueOnce({
      ...mockNote,
      type: "comet",
      due_at: null,
      remind_before_seconds: null,
      done_at: "2026-01-10T09:00:00.000Z",
    });
    mockUpdateNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(screen.getByDisplayValue("Existing Note")).toBeInTheDocument();
    });

    const doneCheckbox = document.querySelector(
      '[data-testid="edit-comet-done"]'
    ) as HTMLInputElement;
    expect(doneCheckbox.checked).toBe(true);
    await fireEvent.click(doneCheckbox);
    await tick();

    const submitButton = screen.getByRole("button", { name: "Save Changes" });
    await fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockUpdateNote).toHaveBeenCalled();
    });
    const payload = mockUpdateNote.mock.calls[0][1];
    expect(payload.done_at).toBeNull();
  });

  it("does not render comet fields for a non-comet note", async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: "456" } });

    await waitFor(() => {
      expect(screen.getByDisplayValue("Existing Note")).toBeInTheDocument();
    });

    expect(document.querySelector('[data-testid="edit-comet-fields"]')).toBeNull();
  });
});
