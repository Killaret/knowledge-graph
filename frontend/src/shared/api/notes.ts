// API-клиент для работы с заметками и рекомендациями
import { api } from "./client";

// Тип данных заметки (соответствует ответу бэкенда)
export interface Note {
  id: string;
  title: string;
  content: string;
  metadata: Record<string, unknown>;
  type?: string;
  is_public?: boolean;
  created_at: string;
  updated_at: string;
}

// Тип рекомендации (похожая заметка)
export interface Suggestion {
  note_id: string;
  title: string;
  score: number; // вес от 0 до 1
}

// Получить все заметки (GET /notes)
// API возвращает { notes: Note[], total, limit, offset }
export async function getNotes(): Promise<Note[]> {
  const response = await api
    .get("v1/notes", {
      searchParams: { limit: 10000 },
      cache: "no-store",
    })
    .json<{ notes: Note[]; total: number; limit: number; offset: number }>();
  return response.notes;
}

// Получить одну заметку по ID
export async function getNote(id: string): Promise<Note> {
  const response = await api.get(`v1/notes/${id}`).json<{ data: Note }>();
  return response.data;
}

// Создать новую заметку
export async function createNote(data: {
  title: string;
  content: string;
  type?: string;
  metadata?: Record<string, unknown>;
}): Promise<Note> {
  return api.post("v1/notes", { json: data }).json();
}

// Обновить существующую заметку
export async function updateNote(id: string, data: Partial<Note>): Promise<Note> {
  return api.put(`v1/notes/${id}`, { json: data }).json();
}

// Delete a single note
export async function deleteNote(id: string): Promise<void> {
  await api.delete(`v1/notes/${id}`);
}

// Delete multiple notes in a single batch request
export async function deleteNotesBatch(ids: string[]): Promise<void> {
  await api.post("v1/notes/batch", { json: { ids } });
}

// Restore a deleted note
export async function restoreNote(id: string): Promise<void> {
  await api.post(`v1/notes/${id}/restore`);
}

// Publish a note to the public graph
export async function publishNote(id: string): Promise<Note> {
  return api.post(`v1/notes/${id}/publish`).json<Note>();
}

// Unpublish a note from the public graph
export async function unpublishNote(id: string): Promise<Note> {
  return api.post(`v1/notes/${id}/unpublish`).json<Note>();
}

// Получить рекомендации для заметки (похожие по явным связям и эмбеддингам)
export async function getSuggestions(id: string, limit = 10): Promise<Suggestion[]> {
  return api.get(`v1/notes/${id}/suggestions`, { searchParams: { limit } }).json();
}

// Search response type
export interface SearchResponse {
  data: Note[];
  total: number;
  page: number;
  size: number;
  totalPages: number;
}

// Search notes with full-text search
export async function searchNotes(query: string, page = 1, size = 20): Promise<SearchResponse> {
  const searchParams = new URLSearchParams({
    q: query,
    page: page.toString(),
    size: size.toString(),
  });
  return api.get(`v1/notes/search?${searchParams}`).json();
}
