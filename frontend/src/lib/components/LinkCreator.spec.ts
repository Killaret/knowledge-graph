import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { server } from '../../../vitest-setup';
import LinkCreator from './LinkCreator.svelte';

describe('LinkCreator', () => {
  const sourceNoteId = 'source-123';
  const mockSearchResults = [
    { id: 'target-1', title: 'Target Note 1', content: '', metadata: {}, created_at: '', updated_at: '' },
    { id: 'target-2', title: 'Target Note 2', content: '', metadata: {}, created_at: '', updated_at: '' }
  ];

  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('renders with search input and link type selector', () => {
    render(LinkCreator, { 
      props: { sourceNoteId } 
    });

    expect(screen.getByLabelText(/search target note/i)).toBeInTheDocument();
    expect(screen.getByText(/reference/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /create link/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument();
  });

  it('searches for notes when typing in search field', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    
    server.use(
      http.get('http://localhost:8081/api/v1/notes/search*', () => {
        return HttpResponse.json({
          data: mockSearchResults,
          total: 2,
          page: 1,
          size: 20,
          totalPages: 1
        });
      })
    );

    render(LinkCreator, { props: { sourceNoteId } });

    const searchInput = screen.getByLabelText(/search target note/i);
    await userEvent.type(searchInput, 'target');

    // Ждём дебаунс (300ms)
    vi.advanceTimersByTime(350);

    await waitFor(() => {
      expect(screen.getByText('Target Note 1')).toBeInTheDocument();
      expect(screen.getByText('Target Note 2')).toBeInTheDocument();
    });

    vi.useRealTimers();
  });

  it('selects target note from search results', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    
    server.use(
      http.get('http://localhost:8081/api/v1/notes/search*', () => {
        return HttpResponse.json({
          data: mockSearchResults,
          total: 2,
          page: 1,
          size: 20,
          totalPages: 1
        });
      })
    );

    render(LinkCreator, { props: { sourceNoteId } });

    const searchInput = screen.getByLabelText(/search target note/i);
    await userEvent.type(searchInput, 'target');
    vi.advanceTimersByTime(350);

    await waitFor(() => {
      expect(screen.getByText('Target Note 1')).toBeInTheDocument();
    });

    // Выбираем первую заметку
    const resultButton = screen.getByLabelText(/select target note 1/i);
    await userEvent.click(resultButton);

    // Проверяем что выбранная заметка отображается в поле
    expect(searchInput).toHaveValue('Target Note 1');

    vi.useRealTimers();
  });

  it('allows selecting different link types', async () => {
    render(LinkCreator, { props: { sourceNoteId } });

    // Открываем dropdown типов (по умолчанию "Reference")
    const typeDropdown = screen.getByText(/reference/i);
    await userEvent.click(typeDropdown);

    // Выбираем тип "Related"
    const relatedOption = screen.getByText(/related/i);
    await userEvent.click(relatedOption);

    // Проверяем что тип изменился
    await waitFor(() => {
      expect(screen.getByText(/related/i)).toBeInTheDocument();
    });
  });

  it('successfully creates link and calls onSuccess callback', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    const onSuccess = vi.fn();
    
    server.use(
      http.get('http://localhost:8081/api/v1/notes/search*', () => {
        return HttpResponse.json({
          data: mockSearchResults,
          total: 2,
          page: 1,
          size: 20,
          totalPages: 1
        });
      }),
      http.post('http://localhost:8081/api/v1/links', () => {
        return HttpResponse.json({
          data: {
            id: 'new-link-123',
            source_note_id: sourceNoteId,
            target_note_id: 'target-1',
            link_type: 'reference',
            weight: 1.0,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString()
          }
        }, { status: 201 });
      })
    );

    render(LinkCreator, { 
      props: { sourceNoteId, onSuccess } 
    });

    // Ищем и выбираем заметку
    const searchInput = screen.getByLabelText(/search target note/i);
    await userEvent.type(searchInput, 'target');
    vi.advanceTimersByTime(350);

    await waitFor(() => {
      expect(screen.getByText('Target Note 1')).toBeInTheDocument();
    });

    await userEvent.click(screen.getByLabelText(/select target note 1/i));

    // Создаём связь
    const createButton = screen.getByRole('button', { name: /create link/i });
    await userEvent.click(createButton);

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith(expect.objectContaining({
        id: 'new-link-123',
        target_note_id: 'target-1'
      }));
    });

    vi.useRealTimers();
  });

  it('handles 409 conflict error when creating duplicate link', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    
    server.use(
      http.get('http://localhost:8081/api/v1/notes/search*', () => {
        return HttpResponse.json({
          data: mockSearchResults,
          total: 2,
          page: 1,
          size: 20,
          totalPages: 1
        });
      }),
      http.post('http://localhost:8081/api/v1/links', () => {
        return HttpResponse.json({
          error: 'Conflict',
          code: 'DUPLICATE_LINK',
          message: 'Link already exists'
        }, { status: 409 });
      })
    );

    render(LinkCreator, { props: { sourceNoteId } });

    // Ищем и выбираем заметку
    const searchInput = screen.getByLabelText(/search target note/i);
    await userEvent.type(searchInput, 'target');
    vi.advanceTimersByTime(350);

    await waitFor(() => {
      expect(screen.getByText('Target Note 1')).toBeInTheDocument();
    });

    await userEvent.click(screen.getByLabelText(/select target note 1/i));

    // Пытаемся создать связь
    const createButton = screen.getByRole('button', { name: /create link/i });
    await userEvent.click(createButton);

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(/link already exists/i);
    });

    vi.useRealTimers();
  });

  it('calls onCancel when cancel button is clicked', async () => {
    const onCancel = vi.fn();
    
    render(LinkCreator, { 
      props: { sourceNoteId, onCancel } 
    });

    const cancelButton = screen.getByRole('button', { name: /cancel/i });
    await userEvent.click(cancelButton);

    expect(onCancel).toHaveBeenCalled();
  });

  it('disables submit button when no target is selected', () => {
    render(LinkCreator, { props: { sourceNoteId } });

    const createButton = screen.getByRole('button', { name: /create link/i });
    expect(createButton).toBeDisabled();
  });

  it('excludes source note from search results', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    
    server.use(
      http.get('http://localhost:8081/api/v1/notes/search*', () => {
        return HttpResponse.json({
          data: [
            { id: sourceNoteId, title: 'Source Note', content: '', metadata: {}, created_at: '', updated_at: '' },
            { id: 'target-1', title: 'Target Note 1', content: '', metadata: {}, created_at: '', updated_at: '' }
          ],
          total: 2,
          page: 1,
          size: 20,
          totalPages: 1
        });
      })
    );

    render(LinkCreator, { props: { sourceNoteId } });

    const searchInput = screen.getByLabelText(/search target note/i);
    await userEvent.type(searchInput, 'note');
    vi.advanceTimersByTime(350);

    await waitFor(() => {
      // Исходная заметка не должна отображаться
      expect(screen.queryByText('Source Note')).not.toBeInTheDocument();
      expect(screen.getByText('Target Note 1')).toBeInTheDocument();
    });

    vi.useRealTimers();
  });
});
