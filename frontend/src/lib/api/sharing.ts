// Sharing API client
import api from './client';
import type { NoteShare, ShareLink, CreateShareRequest, CreateShareLinkRequest } from '$lib/types';

/**
 * Share a note with a user
 */
export async function shareNote(noteId: string, userId: string, permission: 'read' | 'write' = 'read'): Promise<NoteShare> {
  const response = await api.post(`v1/notes/${noteId}/share`, {
    json: { user_id: userId, permission } as CreateShareRequest
  }).json<NoteShare>();
  return response;
}

/**
 * Create a public share link for a note
 */
export async function createShareLink(noteId: string, permission: 'read' | 'write' = 'read', expiresAt?: string, maxUses?: number): Promise<ShareLink> {
  const body: CreateShareLinkRequest = { permission };
  if (expiresAt) {
    body.expires_at = expiresAt;
  }
  if (maxUses) {
    body.max_uses = maxUses;
  }
  
  const response = await api.post(`v1/notes/${noteId}/share-link`, {
    json: body
  }).json<ShareLink>();
  return response;
}

/**
 * Get all shares for a note
 */
export async function getNoteShares(noteId: string): Promise<{ user_shares: NoteShare[]; share_links: ShareLink[] }> {
  const response = await api.get(`v1/notes/${noteId}/shares`).json<{ user_shares: NoteShare[]; share_links: ShareLink[] }>();
  return response;
}

/**
 * Revoke a share (remove user's access)
 */
export async function revokeShare(noteId: string, shareId: string): Promise<void> {
  await api.delete(`v1/notes/${noteId}/shares/${shareId}`);
}

/**
 * Revoke a share link
 */
export async function revokeShareLink(linkId: string): Promise<void> {
  await api.delete(`v1/share-links/${linkId}`);
}

/**
 * Access a shared note via token
 */
export async function accessSharedNote(token: string): Promise<{ note: { id: string; title: string; content: string; type: string; metadata: Record<string, unknown>; created_at: string; updated_at: string }; permission: string }> {
  const response = await api.get(`v1/share/${token}`).json<{
    note: { id: string; title: string; content: string; type: string; metadata: Record<string, unknown>; created_at: string; updated_at: string };
    permission: string;
  }>();
  return response;
}

/**
 * Search users to share with
 */
export async function searchUsers(query: string): Promise<{ users: Array<{ id: string; login: string; email?: string }> }> {
  const response = await api.get('v1/users', {
    searchParams: { search: query }
  }).json<{ users: Array<{ id: string; login: string; email?: string }> }>();
  return response;
}
