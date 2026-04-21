import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  getNotes,
  getNote,
  createNote,
  updateNote,
  deleteNote,
  getSuggestions,
  searchNotes,
  type Note,
  type Suggestion,
  type SearchResponse
} from './notes';
import { resetKyMocks, mockKyResponse } from './__mocks__/ky';

describe('notes API', () => {
  beforeEach(() => {
    resetMocks();
  });

  it('getNotes should return array of notes', async () => {
    const mockNotes: Note[] = [
      {
        id: '1',
        title: 'Test Note',
        content: 'Test Content',
        metadata: {},
        type: 'star',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z'
      }
    ];

    // Настраиваем мок для этого теста
    const ky = await import('ky');
    ky.default.get.mockResolvedValueOnce({
      json: () => Promise.resolve({ notes: mockNotes, total: 1, limit: 10, offset: 0 })
    });

    const result = await getNotes();
    expect(result).toEqual(mockNotes);
  });

  it('getNote should return a single note', async () => {
    const mockNote: Note = {
      id: '1',
      title: 'Test Note',
      content: 'Test Content',
      metadata: { key: 'value' },
      type: 'planet',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z'
    };

    const ky = await import('ky');
    ky.default.get.mockResolvedValueOnce({
      json: () => Promise.resolve(mockNote)
    });

    const result = await getNote('1');
    expect(result).toEqual(mockNote);
  });

  it('createNote should create and return new note', async () => {
    const newNote = { title: 'New Note', content: 'New Content', type: 'star' };
    const mockResponse: Note = {
      id: '2',
      ...newNote,
      metadata: {},
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z'
    };

    const ky = await import('ky');
    ky.default.post.mockResolvedValueOnce({
      json: () => Promise.resolve(mockResponse)
    });

    const result = await createNote(newNote);
    expect(result).toEqual(mockResponse);
  });

  it('updateNote should update and return note', async () => {
    const updateData = { title: 'Updated Title' };
    const mockResponse: Note = {
      id: '1',
      title: 'Updated Title',
      content: 'Original Content',
      metadata: {},
      type: 'star',
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-02T00:00:00Z'
    };

    const ky = await import('ky');
    ky.default.put.mockResolvedValueOnce({
      json: () => Promise.resolve(mockResponse)
    });

    const result = await updateNote('1', updateData);
    expect(result.title).toBe('Updated Title');
  });

  it('deleteNote should delete note', async () => {
    const ky = await import('ky');
    ky.default.delete.mockResolvedValueOnce(Promise.resolve());

    await deleteNote('1');
    expect(ky.default.delete).toHaveBeenCalled();
  });

  it('getSuggestions should return suggestions array', async () => {
    const mockSuggestions: Suggestion[] = [
      { note_id: '2', title: 'Related Note', score: 0.85 },
      { note_id: '3', title: 'Another Note', score: 0.72 }
    ];

    const ky = await import('ky');
    ky.default.get.mockResolvedValueOnce({
      json: () => Promise.resolve(mockSuggestions)
    });

    const result = await getSuggestions('1', 5);
    expect(result).toHaveLength(2);
    expect(result[0].score).toBe(0.85);
  });

  it('searchNotes should return search results', async () => {
    const mockResponse: SearchResponse = {
      data: [{ id: '1', title: 'Search Result', content: 'Content', metadata: {}, created_at: '', updated_at: '' }],
      total: 1,
      page: 1,
      size: 20,
      totalPages: 1
    };

    const ky = await import('ky');
    ky.default.get.mockResolvedValueOnce({
      json: () => Promise.resolve(mockResponse)
    });

    const result = await searchNotes('search term', 1, 20);
    expect(result.data).toHaveLength(1);
    expect(result.total).toBe(1);
  });
});
