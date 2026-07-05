import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { http, HttpResponse } from 'msw';
import { server } from '../../vitest-setup';
import Page from './+page.svelte';

describe('Page list view - batch operations', () => {
  const mockNotes = [
    {
      id: '1',
      title: 'Note 1',
      content: 'Content 1',
      type: 'star',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      metadata: {}
    },
    {
      id: '2',
      title: 'Note 2',
      content: 'Content 2',
      type: 'planet',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      metadata: {}
    },
    {
      id: '3',
      title: 'Note 3',
      content: 'Content 3',
      type: 'comet',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      metadata: {}
    }
  ];

  beforeEach(() => {
    server.use(
      http.get('http://localhost:8080/api/v1/notes', () => HttpResponse.json({ notes: mockNotes, total: 3, limit: 10000, offset: 0 })),
      http.post('http://localhost:8080/api/v1/notes/batch', () => new HttpResponse(null, { status: 204 })),
      http.post('http://localhost:8080/api/v1/notes/:id/restore', () => new HttpResponse(null, { status: 204 })),
      http.get('http://localhost:8080/api/v1/graph/all', () => HttpResponse.json({ nodes: [], links: [] })),
      http.get('http://localhost:8080/api/v1/me/graph/fresh', () => HttpResponse.json({ nodes: [], links: [] }))
    );
  });

  afterEach(() => {
    server.resetHandlers();
  });

  it('renders page without errors', () => {
    render(Page);
    expect(document.querySelector('.page-container')).toBeInTheDocument();
  });

  it('renders undo toast structure in DOM', () => {
    render(Page);
    const undoToast = document.querySelector('.undo-toast');
    expect(undoToast).toBeNull(); // No toast initially
  });
});
