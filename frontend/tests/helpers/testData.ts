import type { APIRequestContext } from '@playwright/test';

export interface NoteData {
  id?: string;
  title: string;
  content?: string;
  type?: string;
  metadata?: Record<string, unknown>;
}

export interface LinkData {
  id?: string;
  sourceNoteId: string;
  targetNoteId: string;
  weight?: number;
  link_type?: string;
  metadata?: Record<string, unknown>;
}

/**
 * Get backend base URL from environment or default
 */
export function getBackendUrl(): string {
  return process.env.BACKEND_URL || 'http://localhost:8080';
}

/**
 * Create a note via API
 */
export async function createNote(
  request: APIRequestContext,
  data: Partial<NoteData>
): Promise<{ id: string; [key: string]: unknown }> {
  const payload = {
    title: data.title || 'Test Note',
    content: data.content || 'Test content',
    type: data.type,
    metadata: data.metadata || {},
  };

  const response = await request.post(`${getBackendUrl()}/notes`, {
    data: payload,
  });

  if (!response.ok()) {
    const errorText = await response.text();
    throw new Error(`Failed to create note: ${response.status()} - ${errorText}`);
  }

  return await response.json();
}

/**
 * Create multiple notes in batch
 */
export async function createNotes(
  request: APIRequestContext,
  notes: Partial<NoteData>[]
): Promise<Array<{ id: string; [key: string]: unknown }>> {
  const created = [];
  for (const noteData of notes) {
    const note = await createNote(request, noteData);
    created.push(note);
  }
  return created;
}

/**
 * Create a link between two notes
 */
export async function createLink(
  request: APIRequestContext,
  sourceId: string,
  targetId: string,
  weight = 0.5,
  linkType = 'related'
): Promise<{ id: string; [key: string]: unknown }> {
  // Debug logging
  console.log('[createLink] Creating link:', { sourceId, targetId, weight, linkType });
  
  // Validate inputs
  if (!sourceId || !targetId) {
    throw new Error(`Invalid parameters: sourceId=${sourceId}, targetId=${targetId}`);
  }
  
  // Go backend expects PascalCase field names
  const payload = {
    SourceNoteID: sourceId,
    TargetNoteID: targetId,
    Weight: weight,
    LinkType: linkType,
    Metadata: {},
  };
  
  console.log('[createLink] Payload:', JSON.stringify(payload));

  const response = await request.post(`${getBackendUrl()}/links`, {
    data: payload,
  });

  if (!response.ok()) {
    const errorText = await response.text();
    throw new Error(`Failed to create link: ${response.status()} - ${errorText}`);
  }

  return await response.json();
}

/**
 * Create a star topology - center note with surrounding notes
 */
export async function createStarTopology(
  request: APIRequestContext,
  centerNote: Partial<NoteData>,
  surroundingNotes: Partial<NoteData>[],
  linkWeight = 0.7
): Promise<{
  center: { id: string };
  surrounding: Array<{ id: string }>;
}> {
  const center = await createNote(request, centerNote);
  const surrounding = [];

  for (const noteData of surroundingNotes) {
    const note = await createNote(request, noteData);
    surrounding.push(note);
    await createLink(request, center.id, note.id, linkWeight);
  }

  return { center, surrounding };
}

/**
 * Create a chain topology - notes linked in sequence
 */
export async function createChainTopology(
  request: APIRequestContext,
  notes: Partial<NoteData>[],
  linkWeight = 0.8
): Promise<Array<{ id: string }>> {
  const created = [];

  for (let i = 0; i < notes.length; i++) {
    const note = await createNote(request, notes[i]);
    created.push(note);

    // Link to previous note
    if (i > 0) {
      await createLink(request, created[i - 1].id, note.id, linkWeight);
    }
  }

  return created;
}

/**
 * Delete a note via API
 */
export async function deleteNote(request: APIRequestContext, noteId: string): Promise<void> {
  const response = await request.delete(`${getBackendUrl()}/notes/${noteId}`);

  if (!response.ok() && response.status() !== 404) {
    const errorText = await response.text();
    throw new Error(`Failed to delete note: ${response.status()} - ${errorText}`);
  }
}

/**
 * Clean up all test data
 */
export async function cleanupTestData(
  request: APIRequestContext,
  noteIds: string[]
): Promise<void> {
  for (const id of noteIds) {
    try {
      await deleteNote(request, id);
    } catch (e) {
      // Ignore errors during cleanup
      console.warn(`Failed to delete note ${id}:`, e);
    }
  }
}

/**
 * Check if backend is available
 */
export async function isBackendAvailable(request: APIRequestContext): Promise<boolean> {
  try {
    const response = await request.get(`${getBackendUrl()}/notes`, {
      timeout: 5000,
    });
    return response.status() < 500;
  } catch {
    return false;
  }
}
