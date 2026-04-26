import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { tick } from 'svelte';
import NoteEditor from './NoteEditor.svelte';

// Мокируем API
const mockCreateNote = vi.fn();
const mockUpdateNote = vi.fn();
const mockGetNote = vi.fn();

vi.mock('$lib/api/notes', () => ({
	createNote: (...args: any[]) => mockCreateNote(...args),
	updateNote: (...args: any[]) => mockUpdateNote(...args),
	getNote: (...args: any[]) => mockGetNote(...args)
}));

// Мокируем навигацию
const mockGoto = vi.fn();

vi.mock('$app/navigation', () => ({
	goto: (...args: any[]) => mockGoto(...args)
}));

// Мокируем stores
vi.mock('$app/stores', () => ({
	page: {
		subscribe: vi.fn()
	}
}));

describe('NoteEditor', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.resetAllMocks();
	});

	describe('Create mode', () => {
		it('should render create form with empty fields', () => {
			render(NoteEditor, { props: {} });

			expect(screen.getByTestId('note-editor')).toBeInTheDocument();
			expect(screen.getByTestId('title-input')).toHaveValue('');
			expect(screen.getByTestId('content-input')).toHaveValue('');
			expect(screen.getByTestId('type-select')).toHaveValue('star');
			expect(screen.getByTestId('save-button')).toHaveTextContent('Create');
		});

		it('should validate empty title and block submission', async () => {
			render(NoteEditor, { props: {} });

			const saveButton = screen.getByTestId('save-button');
			await fireEvent.click(saveButton);
			await tick();

			expect(screen.getByTestId('title-error')).toHaveTextContent('Title is required');
			expect(mockCreateNote).not.toHaveBeenCalled();
		});

		it('should create note and redirect on success', async () => {
			const newNote = { id: '123', title: 'Test Note', content: 'Content', type: 'star' };
			mockCreateNote.mockResolvedValueOnce(newNote);

			render(NoteEditor, { props: {} });

			// Заполняем форму
			const titleInput = screen.getByTestId('title-input') as HTMLInputElement;
			const contentInput = screen.getByTestId('content-input') as HTMLTextAreaElement;

			titleInput.value = 'Test Note';
			await fireEvent.input(titleInput);
			await tick();

			contentInput.value = 'Note content';
			await fireEvent.input(contentInput);
			await tick();

			// Отправляем форму
			const saveButton = screen.getByTestId('save-button');
			await fireEvent.click(saveButton);

			await waitFor(() => {
				expect(mockCreateNote).toHaveBeenCalledWith({
					title: 'Test Note',
					content: 'Note content',
					type: 'star'
				});
			});

			await waitFor(() => {
				expect(mockGoto).toHaveBeenCalledWith('/notes/123');
			});
		});

		it('should display error message when creation fails', async () => {
			mockCreateNote.mockRejectedValueOnce(new Error('Server error'));

			render(NoteEditor, { props: {} });

			const titleInput = screen.getByTestId('title-input') as HTMLInputElement;
			titleInput.value = 'Test Note';
			await fireEvent.input(titleInput);
			await tick();

			const saveButton = screen.getByTestId('save-button');
			await fireEvent.click(saveButton);

			await waitFor(() => {
				expect(screen.getByTestId('error-message')).toHaveTextContent('Failed to save note');
			});

			expect(mockGoto).not.toHaveBeenCalled();
		});

		it('should call onCancel when cancel button is clicked', async () => {
			const onCancel = vi.fn();
			render(NoteEditor, { props: { onCancel } });

			const cancelButton = screen.getByTestId('cancel-button');
			await fireEvent.click(cancelButton);

			expect(onCancel).toHaveBeenCalled();
			expect(mockGoto).not.toHaveBeenCalled();
		});

		it('should navigate home when cancel clicked without onCancel', async () => {
			render(NoteEditor, { props: {} });

			const cancelButton = screen.getByTestId('cancel-button');
			await fireEvent.click(cancelButton);

			expect(mockGoto).toHaveBeenCalledWith('/');
		});
	});

	describe('Edit mode', () => {
		const existingNote = {
			id: '456',
			title: 'Existing Note',
			content: 'Existing content',
			type: 'planet'
		};

		it('should load and display note data', async () => {
			mockGetNote.mockResolvedValueOnce(existingNote);

			render(NoteEditor, { props: { noteId: '456' } });

			expect(screen.getByTestId('loading')).toBeInTheDocument();

			await waitFor(() => {
				expect(screen.getByTestId('title-input')).toHaveValue('Existing Note');
			});

			expect(screen.getByTestId('content-input')).toHaveValue('Existing content');
			expect(screen.getByTestId('type-select')).toHaveValue('planet');
			expect(screen.getByTestId('save-button')).toHaveTextContent('Update');
			expect(mockGetNote).toHaveBeenCalledWith('456');
		});

		it('should display error when loading fails', async () => {
			mockGetNote.mockRejectedValueOnce(new Error('Load failed'));

			render(NoteEditor, { props: { noteId: '456' } });

			await waitFor(() => {
				expect(screen.getByTestId('error-message')).toHaveTextContent('Failed to load note');
			});
		});

		it('should update note on save', async () => {
			mockGetNote.mockResolvedValueOnce(existingNote);
			mockUpdateNote.mockResolvedValueOnce({ ...existingNote, title: 'Updated Title' });

			render(NoteEditor, { props: { noteId: '456' } });

			await waitFor(() => {
				expect(screen.getByTestId('title-input')).toBeInTheDocument();
			});

			const titleInput = screen.getByTestId('title-input') as HTMLInputElement;
			titleInput.value = 'Updated Title';
			await fireEvent.input(titleInput);
			await tick();

			const saveButton = screen.getByTestId('save-button');
			await fireEvent.click(saveButton);

			await waitFor(() => {
				expect(mockUpdateNote).toHaveBeenCalledWith('456', {
					title: 'Updated Title',
					content: 'Existing content',
					type: 'planet'
				});
			});

			// При редактировании не делаем redirect
			expect(mockGoto).not.toHaveBeenCalled();
		});

		it('should show saving state during update', async () => {
			mockGetNote.mockResolvedValueOnce(existingNote);
			let resolveUpdate: (value: any) => void;
			const updatePromise = new Promise((resolve) => {
				resolveUpdate = resolve;
			});
			mockUpdateNote.mockReturnValueOnce(updatePromise);

			render(NoteEditor, { props: { noteId: '456' } });

			await waitFor(() => {
				expect(screen.getByTestId('title-input')).toBeInTheDocument();
			});

			const saveButton = screen.getByTestId('save-button');
			await fireEvent.click(saveButton);

			expect(saveButton).toHaveTextContent('Saving...');
			expect(saveButton).toBeDisabled();
			expect(screen.getByTestId('title-input')).toBeDisabled();

			// Разрешаем промис
			resolveUpdate!({ ...existingNote });

			await waitFor(() => {
				expect(saveButton).toHaveTextContent('Update');
			});
		});
	});

	describe('Form interactions', () => {
		it('should update note type selection', async () => {
			render(NoteEditor, { props: {} });

			const typeSelect = screen.getByTestId('type-select') as HTMLSelectElement;
			typeSelect.value = 'comet';
			await fireEvent.change(typeSelect);
			await tick();

			expect(screen.getByTestId('type-select')).toHaveValue('comet');
		});

		it('should trim whitespace from title and content', async () => {
			mockCreateNote.mockResolvedValueOnce({ id: '789', title: 'Trimmed', content: 'Content', type: 'star' });

			render(NoteEditor, { props: {} });

			const titleInput = screen.getByTestId('title-input') as HTMLInputElement;
			const contentInput = screen.getByTestId('content-input') as HTMLTextAreaElement;

			titleInput.value = '  Trimmed Title  ';
			await fireEvent.input(titleInput);
			await tick();

			contentInput.value = '  Content with spaces  ';
			await fireEvent.input(contentInput);
			await tick();

			const saveButton = screen.getByTestId('save-button');
			await fireEvent.click(saveButton);

			await waitFor(() => {
				expect(mockCreateNote).toHaveBeenCalledWith({
					title: 'Trimmed Title',
					content: 'Content with spaces',
					type: 'star'
				});
			});
		});
	});
});
