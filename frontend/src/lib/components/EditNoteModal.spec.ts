import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { tick } from 'svelte';
import EditNoteModal from './EditNoteModal.svelte';

// Мокируем API
const mockGetNote = vi.fn();
const mockUpdateNote = vi.fn();

vi.mock('$lib/api/notes', () => ({
  getNote: (...args: any[]) => mockGetNote(...args),
  updateNote: (...args: any[]) => mockUpdateNote(...args)
}));

describe('EditNoteModal', () => {
  const mockNote = {
    id: '456',
    title: 'Existing Note',
    content: 'Existing content',
    type: 'planet',
    metadata: {},
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-02T00:00:00Z'
  };

  const updatedNote = {
    ...mockNote,
    title: 'Updated Title',
    content: 'Updated content'
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it('renders modal when open is true', async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: '456' } });

    expect(screen.getByText('Edit Note')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'Save Changes' })).toBeInTheDocument();
    });
  });

  it('does not render when open is false', () => {
    render(EditNoteModal, { props: { open: false, noteId: '456' } });

    expect(screen.queryByText('Edit Note')).not.toBeInTheDocument();
  });

  it('loads note data when opened', async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: '456' } });

    await waitFor(() => {
      expect(mockGetNote).toHaveBeenCalledWith('456');
    });

    await waitFor(() => {
      expect(screen.getByDisplayValue('Existing Note')).toBeInTheDocument();
    });

    expect(screen.getByDisplayValue('Existing content')).toBeInTheDocument();
    // TypeSelector показывает выбранный тип через aria-pressed
    const planetButton = screen.getByRole('button', { name: /Planet/i });
    expect(planetButton).toHaveAttribute('aria-pressed', 'true');
  });

  it('shows loading state while loading', async () => {
    let resolveGetNote: (value: any) => void;
    const getNotePromise = new Promise((resolve) => {
      resolveGetNote = resolve;
    });
    mockGetNote.mockReturnValueOnce(getNotePromise);

    render(EditNoteModal, { props: { open: true, noteId: '456' } });

    expect(screen.getByText('Loading...')).toBeInTheDocument();

    // Разрешаем промис
    resolveGetNote!(mockNote);

    await waitFor(() => {
      expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });
  });

  it('shows error when loading fails', async () => {
    mockGetNote.mockRejectedValueOnce(new Error('Failed to load'));

    render(EditNoteModal, { props: { open: true, noteId: '456' } });

    await waitFor(() => {
      expect(screen.getByText('Failed to load note')).toBeInTheDocument();
    });
  });

  it('shows validation error when title is empty', async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: '456' } });

    await waitFor(() => {
      expect(screen.getByDisplayValue('Existing Note')).toBeInTheDocument();
    });

    // Очищаем заголовок
    const titleInput = screen.getByTestId('edit-title-input') as HTMLInputElement;
    titleInput.value = '';
    await fireEvent.input(titleInput);
    await tick();

    const saveButton = screen.getByRole('button', { name: 'Save Changes' });
    await fireEvent.click(saveButton);
    await tick();

    expect(screen.getByText('Title is required')).toBeInTheDocument();
    expect(mockUpdateNote).not.toHaveBeenCalled();
  });

  it('updates note successfully and calls onSuccess', async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);
    mockUpdateNote.mockResolvedValueOnce(updatedNote);
    const mockOnSuccess = vi.fn();

    render(EditNoteModal, {
      props: {
        open: true,
        noteId: '456',
        onSuccess: mockOnSuccess
      }
    });

    await waitFor(() => {
      expect(screen.getByDisplayValue('Existing Note')).toBeInTheDocument();
    });

    // Изменяем данные
    const titleInput = screen.getByTestId('edit-title-input') as HTMLInputElement;
    titleInput.value = 'Updated Title';
    await fireEvent.input(titleInput);
    await tick();

    const contentInput = screen.getByTestId('edit-content-input') as HTMLTextAreaElement;
    contentInput.value = 'Updated content';
    await fireEvent.input(contentInput);
    await tick();

    // Сохраняем
    const saveButton = screen.getByRole('button', { name: 'Save Changes' });
    await fireEvent.click(saveButton);

    await waitFor(() => {
      expect(mockUpdateNote).toHaveBeenCalledWith('456', {
        title: 'Updated Title',
        content: 'Updated content',
        type: 'planet',
        metadata: {}
      });
    });

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalledWith(updatedNote);
    });

    // Модаль закрывается
    expect(screen.queryByText('Edit Note')).not.toBeInTheDocument();
  });

  it('shows error when update fails', async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);
    mockUpdateNote.mockRejectedValueOnce(new Error('Failed to update'));

    render(EditNoteModal, { props: { open: true, noteId: '456' } });

    await waitFor(() => {
      expect(screen.getByDisplayValue('Existing Note')).toBeInTheDocument();
    });

    const saveButton = screen.getByRole('button', { name: 'Save Changes' });
    await fireEvent.click(saveButton);

    await waitFor(() => {
      expect(screen.getByText('Failed to update note')).toBeInTheDocument();
    });

    // Модаль не закрывается
    expect(screen.getByText('Edit Note')).toBeInTheDocument();
  });

  it('shows saving state during update', async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);
    let resolveUpdate: (value: any) => void;
    const updatePromise = new Promise((resolve) => {
      resolveUpdate = resolve;
    });
    mockUpdateNote.mockReturnValueOnce(updatePromise);

    render(EditNoteModal, { props: { open: true, noteId: '456' } });

    await waitFor(() => {
      expect(screen.getByDisplayValue('Existing Note')).toBeInTheDocument();
    });

    const saveButton = screen.getByRole('button', { name: 'Save Changes' });
    await fireEvent.click(saveButton);

    expect(saveButton).toHaveTextContent('Saving...');
    expect(saveButton).toBeDisabled();
    expect(screen.getByTestId('edit-title-input')).toBeDisabled();

    // Разрешаем промис
    resolveUpdate!(updatedNote);

    await waitFor(() => {
      expect(screen.queryByText('Edit Note')).not.toBeInTheDocument();
    });
  });

  it('closes modal when cancel is clicked', async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: '456' } });

    await waitFor(() => {
      expect(screen.getByText('Edit Note')).toBeInTheDocument();
    });

    const cancelButton = screen.getByRole('button', { name: 'Cancel' });
    await fireEvent.click(cancelButton);

    expect(screen.queryByText('Edit Note')).not.toBeInTheDocument();
  });

  it('closes modal when close button is clicked', async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: '456' } });

    await waitFor(() => {
      expect(screen.getByText('Edit Note')).toBeInTheDocument();
    });

    const closeButton = screen.getByLabelText('Close');
    await fireEvent.click(closeButton);

    expect(screen.queryByText('Edit Note')).not.toBeInTheDocument();
  });

  it('updates note type selection', async () => {
    mockGetNote.mockResolvedValueOnce(mockNote);

    render(EditNoteModal, { props: { open: true, noteId: '456' } });

    await waitFor(() => {
      expect(screen.getByDisplayValue('Existing Note')).toBeInTheDocument();
    });

    // Выбираем Comet через TypeSelector (вместо старого select)
    const cometButton = screen.getByRole('button', { name: /Comet/i });
    await fireEvent.click(cometButton);
    await tick();

    // Проверяем что Comet выбран
    expect(cometButton).toHaveAttribute('aria-pressed', 'true');
  });
});
