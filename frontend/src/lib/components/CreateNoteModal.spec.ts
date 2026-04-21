import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { tick } from 'svelte';
import CreateNoteModal from './CreateNoteModal.svelte';
import { server } from '../../../vitest-setup';
import { http, HttpResponse } from 'msw';

// Мокаем API
vi.mock('$lib/api/notes', () => ({
  createNote: vi.fn()
}));

import { createNote } from '$lib/api/notes';

describe('CreateNoteModal', () => {
  const mockNote = {
    id: '123',
    title: 'Test Note',
    content: 'Test content',
    type: 'star',
    metadata: {},
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z'
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders modal when open is true', () => {
    render(CreateNoteModal, { props: { open: true } });
    
    expect(screen.getByText('Create New Note')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter note title...')).toBeInTheDocument();
    expect(screen.getByText('Create Note')).toBeInTheDocument();
  });

  it('does not render when open is false', () => {
    render(CreateNoteModal, { props: { open: false } });
    
    expect(screen.queryByText('Create New Note')).not.toBeInTheDocument();
  });

  it('shows validation error when title is empty', async () => {
    render(CreateNoteModal, { props: { open: true } });
    
    const submitButton = screen.getByText('Create Note');
    await fireEvent.click(submitButton);
    await tick();
    
    expect(screen.getByText('Title is required')).toBeInTheDocument();
    expect(createNote).not.toHaveBeenCalled();
  });

  it('creates note successfully and calls onSuccess', async () => {
    const mockOnSuccess = vi.fn();
    vi.mocked(createNote).mockResolvedValueOnce(mockNote);
    
    render(CreateNoteModal, { 
      props: { 
        open: true, 
        onSuccess: mockOnSuccess 
      } 
    });
    
    // Заполняем форму используя паттерн для Svelte 5
    const titleInput = screen.getByPlaceholderText('Enter note title...') as HTMLInputElement;
    titleInput.value = 'Test Note';
    await fireEvent.input(titleInput);
    await tick();
    
    const contentInput = screen.getByPlaceholderText('Enter note content...') as HTMLTextAreaElement;
    contentInput.value = 'Test content';
    await fireEvent.input(contentInput);
    await tick();
    
    // Отправляем форму
    const submitButton = screen.getByText('Create Note');
    await fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(createNote).toHaveBeenCalledWith({
        title: 'Test Note',
        content: 'Test content',
        type: 'star',
        metadata: {}
      });
    });
    
    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalledWith(mockNote);
    });
  });

  it('handles API error and displays error message', async () => {
    vi.mocked(createNote).mockRejectedValueOnce(new Error('API Error'));
    
    render(CreateNoteModal, { props: { open: true } });
    
    // Заполняем форму
    const titleInput = screen.getByPlaceholderText('Enter note title...') as HTMLInputElement;
    titleInput.value = 'Test Note';
    await fireEvent.input(titleInput);
    await tick();
    
    // Отправляем форму
    const submitButton = screen.getByText('Create Note');
    await fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(screen.getByText('Failed to create note')).toBeInTheDocument();
    });
  });

  it('closes modal on Cancel button click', async () => {
    render(CreateNoteModal, { props: { open: true } });
    
    const cancelButton = screen.getByText('Cancel');
    await fireEvent.click(cancelButton);
    await tick();
    
    // Модал должен закрыться (open = false через bindable)
    expect(screen.queryByText('Create New Note')).not.toBeInTheDocument();
  });

  it('closes modal on close button click', async () => {
    render(CreateNoteModal, { props: { open: true } });
    
    const closeButton = screen.getByLabelText('Close');
    await fireEvent.click(closeButton);
    await tick();
    
    expect(screen.queryByText('Create New Note')).not.toBeInTheDocument();
  });

  it('allows selecting different note types', async () => {
    const mockOnSuccess = vi.fn();
    vi.mocked(createNote).mockResolvedValueOnce({ ...mockNote, type: 'planet' });
    
    render(CreateNoteModal, { 
      props: { 
        open: true, 
        onSuccess: mockOnSuccess 
      } 
    });
    
    // Выбираем тип Planet
    const planetButton = screen.getByText('🪐 Planet');
    await fireEvent.click(planetButton);
    await tick();
    
    // Заполняем и отправляем
    const titleInput = screen.getByPlaceholderText('Enter note title...') as HTMLInputElement;
    titleInput.value = 'Planet Note';
    await fireEvent.input(titleInput);
    await tick();
    
    const submitButton = screen.getByText('Create Note');
    await fireEvent.click(submitButton);
    
    await waitFor(() => {
      expect(createNote).toHaveBeenCalledWith(expect.objectContaining({
        type: 'planet'
      }));
    });
  });

  it('disables submit button while loading', async () => {
    vi.mocked(createNote).mockImplementation(() => new Promise(resolve => {
      setTimeout(() => resolve(mockNote), 100);
    }));
    
    render(CreateNoteModal, { props: { open: true } });
    
    // Заполняем форму
    const titleInput = screen.getByPlaceholderText('Enter note title...') as HTMLInputElement;
    titleInput.value = 'Test Note';
    await fireEvent.input(titleInput);
    await tick();
    
    // Отправляем
    const submitButton = screen.getByText('Create Note');
    await fireEvent.click(submitButton);
    
    // Кнопка должна быть disabled и показывать "Creating..."
    await waitFor(() => {
      expect(screen.getByText('Creating...')).toBeInTheDocument();
    });
    
    expect(submitButton).toBeDisabled();
  });

  it('clears form after successful creation', async () => {
    vi.mocked(createNote).mockResolvedValueOnce(mockNote);
    
    render(CreateNoteModal, { props: { open: true } });
    
    // Заполняем форму
    const titleInput = screen.getByPlaceholderText('Enter note title...') as HTMLInputElement;
    titleInput.value = 'Test Note';
    await fireEvent.input(titleInput);
    await tick();
    
    const contentInput = screen.getByPlaceholderText('Enter note content...') as HTMLTextAreaElement;
    contentInput.value = 'Test content';
    await fireEvent.input(contentInput);
    await tick();
    
    // Отправляем
    const submitButton = screen.getByText('Create Note');
    await fireEvent.click(submitButton);
    
    // После успеха форма очищается и модал закрывается
    await waitFor(() => {
      expect(screen.queryByText('Create New Note')).not.toBeInTheDocument();
    });
  });
});
