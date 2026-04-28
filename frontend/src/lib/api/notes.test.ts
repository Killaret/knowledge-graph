import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../../../vitest-setup';
import {
  getNotes,
  getNote,
  createNote,
  updateNote,
  deleteNote,
  getSuggestions,
  searchNotes,
} from './notes';
import type { Note, Suggestion, SearchResponse } from './notes';

describe('notes API', () => {
  const mockNote: Note = {
    id: '1',
    title: 'Test Note',
    content: 'Test Content',
    metadata: {},
    type: 'star',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };

  describe('getNotes', () => {
    it('should return array of notes', async () => {
      const mockResponse = { notes: [mockNote], total: 1, limit: 10, offset: 0 };
      
      server.use(
        http.get('http://localhost:8081/api/notes', () => HttpResponse.json(mockResponse))
      );

      const result = await getNotes();

      expect(result).toEqual([mockNote]);
    });
  });

  describe('getNote', () => {
    it('should return a single note', async () => {
      const mockSingleNote: Note = {
        ...mockNote,
        metadata: { key: 'value' },
        type: 'planet',
      };
      
      server.use(
        http.get('http://localhost:8081/api/notes/1', () => HttpResponse.json(mockSingleNote))
      );

      const result = await getNote('1');

      expect(result).toEqual(mockSingleNote);
    });
  });

  describe('createNote', () => {
    it('should create and return new note', async () => {
      const newNote = { title: 'New Note', content: 'New Content', type: 'star' };
      const mockResponse: Note = {
        id: '2',
        ...newNote,
        metadata: {},
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };
      
      server.use(
        http.post('http://localhost:8081/api/notes', () => HttpResponse.json(mockResponse, { status: 201 }))
      );

      const result = await createNote(newNote);

      expect(result).toEqual(mockResponse);
    });
  });

  describe('updateNote', () => {
    it('should update and return note', async () => {
      const updateData = { title: 'Updated Title' };
      const mockResponse: Note = {
        id: '1',
        title: 'Updated Title',
        content: 'Original Content',
        metadata: {},
        type: 'star',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
      };
      
      server.use(
        http.put('http://localhost:8081/api/notes/1', () => HttpResponse.json(mockResponse))
      );

      const result = await updateNote('1', updateData);

      expect(result.title).toBe('Updated Title');
    });
  });

  describe('deleteNote', () => {
    it('should delete note', async () => {
      server.use(
        http.delete('http://localhost:8081/api/notes/1', () => new HttpResponse(null, { status: 204 }))
      );

      const result = await deleteNote('1');
      // deleteNote возвращает пустую строку при 204 No Content
      expect(result).toBe('');
    });
  });

  describe('getSuggestions', () => {
    it('should return suggestions array', async () => {
      const mockSuggestions: Suggestion[] = [
        { note_id: '2', title: 'Related Note', score: 0.85 },
        { note_id: '3', title: 'Another Note', score: 0.72 },
      ];
      
      server.use(
        http.get('http://localhost:8081/api/notes/1/suggestions', () => HttpResponse.json(mockSuggestions))
      );

      const result = await getSuggestions('1', 5);

      expect(result).toHaveLength(2);
      expect(result[0].score).toBe(0.85);
    });
  });

  describe('searchNotes', () => {
    it('should return search results', async () => {
      const mockResponse: SearchResponse = {
        data: [
          {
            id: '1',
            title: 'Search Result',
            content: 'Content',
            metadata: {},
            created_at: '',
            updated_at: '',
          },
        ],
        total: 1,
        page: 1,
        size: 20,
        totalPages: 1,
      };
      
      server.use(
        http.get('http://localhost:8081/api/notes/search*', () => HttpResponse.json(mockResponse))
      );

      const result = await searchNotes('search term', 1, 20);

      expect(result.data).toHaveLength(1);
      expect(result.total).toBe(1);
    });
  });

  describe('retry behavior', () => {
    // Note: Ky is configured with retry: { limit: 3, methods: ['get', 'post', 'put', 'patch'], statusCodes: [408, 413, 429, 500, 502, 503, 504] }
    // This means requests will be retried up to 3 times on transient errors

    it('should retry on 503 Service Unavailable and succeed on retry', async () => {
      let attemptCount = 0;
      server.use(
        http.get('http://localhost:8081/api/notes/1', () => {
          attemptCount++;
          if (attemptCount < 3) {
            return HttpResponse.json({ error: 'Service Unavailable' }, { status: 503 });
          }
          return HttpResponse.json(mockNote);
        })
      );

      const result = await getNote('1');
      expect(result).toEqual(mockNote);
      expect(attemptCount).toBe(3); // Initial + 2 retries
    });

    it('should retry on 504 Gateway Timeout and succeed on retry', async () => {
      let attemptCount = 0;
      server.use(
        http.get('http://localhost:8081/api/notes/1', () => {
          attemptCount++;
          if (attemptCount < 2) {
            return HttpResponse.json({ error: 'Gateway Timeout' }, { status: 504 });
          }
          return HttpResponse.json(mockNote);
        })
      );

      const result = await getNote('1');
      expect(result).toEqual(mockNote);
      expect(attemptCount).toBe(2);
    });

    it('should fail after max retries on persistent 500 errors', async () => {
      let attemptCount = 0;
      server.use(
        http.get('http://localhost:8081/api/notes/1', () => {
          attemptCount++;
          return HttpResponse.json({ error: 'Server error' }, { status: 500 });
        })
      );

      await expect(getNote('1')).rejects.toThrow();
      // limit: 3 means initial + 3 retries = 4 attempts total
      expect(attemptCount).toBe(4);
    });

    it('should retry on 502 Bad Gateway for GET requests', async () => {
      let attemptCount = 0;
      server.use(
        http.get('http://localhost:8081/api/notes/1', () => {
          attemptCount++;
          if (attemptCount < 2) {
            return HttpResponse.json({ error: 'Bad Gateway' }, { status: 502 });
          }
          return HttpResponse.json(mockNote);
        })
      );

      const result = await getNote('1');
      expect(result).toEqual(mockNote);
      expect(attemptCount).toBe(2);
    });

    it('should retry on 429 Too Many Requests for POST requests', async () => {
      let attemptCount = 0;
      const newNote = { title: 'New Note', content: 'Content', type: 'star' };
      server.use(
        http.post('http://localhost:8081/api/notes', () => {
          attemptCount++;
          if (attemptCount < 2) {
            return HttpResponse.json({ error: 'Too Many Requests' }, { status: 429 });
          }
          return HttpResponse.json({ ...mockNote, ...newNote, id: '2' }, { status: 201 });
        })
      );

      const result = await createNote(newNote);
      expect(result.title).toBe('New Note');
      expect(attemptCount).toBe(2);
    });
  });

describe('error handling', () => {
    it('should handle network errors for getNote', async () => {
      server.use(
        http.get('http://localhost:8081/api/notes/1', () => {
          return HttpResponse.error();
        })
      );

      await expect(getNote('1')).rejects.toThrow();
    });

    it('should handle HTTP 404 errors for getNote', async () => {
      server.use(
        http.get('http://localhost:8081/api/notes/999', () => HttpResponse.json({ error: 'Not found' }, { status: 404 }))
      );

      await expect(getNote('999')).rejects.toThrow();
    });

    it('should handle HTTP 500 errors for getNote', async () => {
      server.use(
        http.get('http://localhost:8081/api/notes/1', () => HttpResponse.json({ error: 'Server error' }, { status: 500 }))
      );

      await expect(getNote('1')).rejects.toThrow();
    });

    it('should handle 404 for createNote when validation fails', async () => {
      server.use(
        http.post('http://localhost:8081/api/notes', () => 
          HttpResponse.json({ error: 'Validation failed', details: ['Title required'] }, { status: 400 })
        )
      );

      await expect(createNote({ title: '', content: 'test' })).rejects.toThrow();
    });

    it('should handle 409 conflict for createNote', async () => {
      server.use(
        http.post('http://localhost:8081/api/notes', () => 
          HttpResponse.json({ error: 'Note with this title already exists' }, { status: 409 })
        )
      );

      await expect(createNote({ title: 'Duplicate', content: 'test' })).rejects.toThrow();
    });

    it('should handle 404 for updateNote when note not found', async () => {
      server.use(
        http.put('http://localhost:8081/api/notes/999', () => 
          HttpResponse.json({ error: 'Note not found' }, { status: 404 })
        )
      );

      await expect(updateNote('999', { title: 'Updated' })).rejects.toThrow();
    });

    it('should handle 500 for updateNote', async () => {
      server.use(
        http.put('http://localhost:8081/api/notes/1', () => 
          HttpResponse.json({ error: 'Database error' }, { status: 500 })
        )
      );

      await expect(updateNote('1', { title: 'Updated' })).rejects.toThrow();
    });

    it('should handle 404 for deleteNote', async () => {
      server.use(
        http.delete('http://localhost:8081/api/notes/999', () => 
          HttpResponse.json({ error: 'Note not found' }, { status: 404 })
        )
      );

      await expect(deleteNote('999')).rejects.toThrow();
    });

    it('should handle network errors for deleteNote', async () => {
      server.use(
        http.delete('http://localhost:8081/api/notes/1', () => HttpResponse.error())
      );

      await expect(deleteNote('1')).rejects.toThrow();
    });

    it('should handle 404 for getSuggestions when note not found', async () => {
      server.use(
        http.get('http://localhost:8081/api/notes/999/suggestions', () => 
          HttpResponse.json({ error: 'Note not found' }, { status: 404 })
        )
      );

      await expect(getSuggestions('999')).rejects.toThrow();
    });

    it('should handle 500 for getSuggestions', async () => {
      server.use(
        http.get('http://localhost:8081/api/notes/1/suggestions', () => 
          HttpResponse.json({ error: 'Recommendation service error' }, { status: 500 })
        )
      );

      await expect(getSuggestions('1')).rejects.toThrow();
    });

    it('should handle 400 for searchNotes with invalid query', async () => {
      server.use(
        http.get('http://localhost:8081/api/notes/search*', () => 
          HttpResponse.json({ error: 'Invalid search query' }, { status: 400 })
        )
      );

      await expect(searchNotes('')).rejects.toThrow();
    });

    it('should handle network errors for searchNotes', async () => {
      server.use(
        http.get('http://localhost:8081/api/notes/search*', () => HttpResponse.error())
      );

      await expect(searchNotes('test')).rejects.toThrow();
    });

    it('should handle 500 for getNotes', async () => {
      server.use(
        http.get('http://localhost:8081/api/notes', () => 
          HttpResponse.json({ error: 'Database error' }, { status: 500 })
        )
      );

      await expect(getNotes()).rejects.toThrow();
    });
  });
});
