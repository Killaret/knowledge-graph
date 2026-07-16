import { describe, it, expect } from 'vitest';
import { http, HttpResponse } from 'msw';
import { server } from '../../../vitest-setup';
import { getLinks, getLink, createLink, deleteLink, getNoteLinks, type Link, type CreateLinkData } from './links';

describe('links API', () => {
  const mockLink: Link = {
    id: 'link-1',
    source_note_id: 'note-1',
    target_note_id: 'note-2',
    link_type: 'reference',
    weight: 1.0,
    metadata: {},
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  };

  describe('getLinks', () => {
    it('should return array of links', async () => {
      const mockResponse = { data: [mockLink] };
      
      server.use(
        http.get('http://localhost:8080/api/v1/links', () => HttpResponse.json(mockResponse))
      );

      const result = await getLinks();

      expect(result).toEqual([mockLink]);
    });

    it('should handle empty links array', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/links', () => HttpResponse.json({ data: [] }))
      );

      const result = await getLinks();

      expect(result).toEqual([]);
    });

    it('should throw on 404 error', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/links', () => 
          HttpResponse.json({ error: 'Not found' }, { status: 404 })
        )
      );

      await expect(getLinks()).rejects.toThrow();
    });

    it('should throw on 500 error', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/links', () => 
          HttpResponse.json({ error: 'Server error' }, { status: 500 })
        )
      );

      await expect(getLinks()).rejects.toThrow();
    });

    it('should handle network errors', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/links', () => HttpResponse.error())
      );

      await expect(getLinks()).rejects.toThrow();
    });
  });

  describe('getLink', () => {
    it('should return a single link', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/links/link-1', () => 
          HttpResponse.json({ data: mockLink })
        )
      );

      const result = await getLink('link-1');

      expect(result).toEqual(mockLink);
    });

    it('should throw on 404 when link not found', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/links/nonexistent', () => 
          HttpResponse.json({ error: 'Link not found' }, { status: 404 })
        )
      );

      await expect(getLink('nonexistent')).rejects.toThrow();
    });

    it('should throw on 500 error', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/links/link-1', () => 
          HttpResponse.json({ error: 'Server error' }, { status: 500 })
        )
      );

      await expect(getLink('link-1')).rejects.toThrow();
    });
  });

  describe('createLink', () => {
    it('should create and return new link', async () => {
      const createData: CreateLinkData = {
        source_note_id: 'note-1',
        target_note_id: 'note-2',
        link_type: 'reference',
        weight: 0.8,
      };

      server.use(
        http.post('http://localhost:8080/api/v1/links', () => 
          HttpResponse.json({ data: { ...mockLink, ...createData } }, { status: 201 })
        )
      );

      const result = await createLink(createData);

      expect(result.source_note_id).toBe('note-1');
      expect(result.target_note_id).toBe('note-2');
      expect(result.link_type).toBe('reference');
    });

    it('should throw on 409 conflict (duplicate link)', async () => {
      const createData: CreateLinkData = {
        source_note_id: 'note-1',
        target_note_id: 'note-2',
        link_type: 'reference',
      };

      server.use(
        http.post('http://localhost:8080/api/v1/links', () => 
          HttpResponse.json(
            { error: 'Conflict', code: 'DUPLICATE_LINK', message: 'Link already exists' },
            { status: 409 }
          )
        )
      );

      await expect(createLink(createData)).rejects.toThrow();
    });

    it('should throw on 400 bad request (invalid data)', async () => {
      const createData: CreateLinkData = {
        source_note_id: '',
        target_note_id: 'note-2',
        link_type: 'reference',
      };

      server.use(
        http.post('http://localhost:8080/api/v1/links', () => 
          HttpResponse.json(
            { error: 'Bad Request', message: 'Invalid source_note_id' },
            { status: 400 }
          )
        )
      );

      await expect(createLink(createData)).rejects.toThrow();
    });

    it('should throw on 500 error', async () => {
      const createData: CreateLinkData = {
        source_note_id: 'note-1',
        target_note_id: 'note-2',
        link_type: 'reference',
      };

      server.use(
        http.post('http://localhost:8080/api/v1/links', () => 
          HttpResponse.json({ error: 'Server error' }, { status: 500 })
        )
      );

      await expect(createLink(createData)).rejects.toThrow();
    });

    it('should handle network errors', async () => {
      const createData: CreateLinkData = {
        source_note_id: 'note-1',
        target_note_id: 'note-2',
        link_type: 'reference',
      };

      server.use(
        http.post('http://localhost:8080/api/v1/links', () => HttpResponse.error())
      );

      await expect(createLink(createData)).rejects.toThrow();
    });
  });

  describe('deleteLink', () => {
    it('should delete link successfully', async () => {
      server.use(
        http.delete('http://localhost:8080/api/v1/links/link-1', () => new HttpResponse(null, { status: 204 }))
      );

      // Should not throw
      await expect(deleteLink('link-1')).resolves.toBeUndefined();
    });

    it('should throw on 404 when link not found', async () => {
      server.use(
        http.delete('http://localhost:8080/api/v1/links/nonexistent', () => 
          HttpResponse.json({ error: 'Link not found' }, { status: 404 })
        )
      );

      await expect(deleteLink('nonexistent')).rejects.toThrow();
    });

    it('should throw on 500 error', async () => {
      server.use(
        http.delete('http://localhost:8080/api/v1/links/link-1', () => 
          HttpResponse.json({ error: 'Server error' }, { status: 500 })
        )
      );

      await expect(deleteLink('link-1')).rejects.toThrow();
    });

    it('should handle network errors', async () => {
      server.use(
        http.delete('http://localhost:8080/api/v1/links/link-1', () => HttpResponse.error())
      );

      await expect(deleteLink('link-1')).rejects.toThrow();
    });
  });

  describe('getNoteLinks', () => {
    it('should return links for a note', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/notes/note-1/links', () => 
          HttpResponse.json({ data: [mockLink] })
        )
      );

      const result = await getNoteLinks('note-1');

      expect(result).toHaveLength(1);
      expect(result[0].id).toBe('link-1');
    });

    it('should return empty array when note has no links', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/notes/note-1/links', () => 
          HttpResponse.json({ data: [] })
        )
      );

      const result = await getNoteLinks('note-1');

      expect(result).toEqual([]);
    });

    it('should throw on 404 when note not found', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/notes/nonexistent/links', () => 
          HttpResponse.json({ error: 'Note not found' }, { status: 404 })
        )
      );

      await expect(getNoteLinks('nonexistent')).rejects.toThrow();
    });

    it('should throw on 500 error', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/notes/note-1/links', () => 
          HttpResponse.json({ error: 'Server error' }, { status: 500 })
        )
      );

      await expect(getNoteLinks('note-1')).rejects.toThrow();
    });

    it('should handle network errors', async () => {
      server.use(
        http.get('http://localhost:8080/api/v1/notes/note-1/links', () => HttpResponse.error())
      );

      await expect(getNoteLinks('note-1')).rejects.toThrow();
    });
  });
});
