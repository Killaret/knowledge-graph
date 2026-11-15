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
  // COMET-1: поля планирования — null, когда не заданы
  due_at?: string | null;
  remind_before_seconds?: number | null;
  done_at?: string | null;
  created_at: string;
  updated_at: string;
}

// COMET-1: поля, которые можно передать при создании/обновлении.
// Семантика обновления — merge по присутствию ключа: null очищает.
export interface CometFields {
  due_at?: string | null;
  remind_before_seconds?: number | null;
  done_at?: string | null;
}

// «Ближайшие дела» — ответ GET /api/v1/notes/comets
export interface CometsResponse {
  overdue: Note[];
  upcoming: Note[];
  undated: Note[];
}

// Тип рекомендации (похожая заметка)
export interface Suggestion {
  note_id: string;
  title: string;
  score: number; // вес от 0 до 1
}

// Получить все заметки (GET /notes)
// API возвращает { notes: Note[], total, limit, offset }.
// Backend clamps `limit` to pagination.max_limit — read all pages so a corpus
// larger than the cap is not silently truncated (NOTES-LIMIT-1).
export async function getNotes(): Promise<Note[]> {
  const pageSize = 10000;
  const all: Note[] = [];
  let total = Infinity;
  while (all.length < total) {
    const response = await api
      .get("v1/notes", {
        searchParams: { limit: pageSize, offset: all.length },
        cache: "no-store",
      })
      .json<{ notes: Note[]; total: number; limit: number; offset: number }>();
    all.push(...response.notes);
    total = response.total;
    if (response.notes.length === 0) break;
  }
  return all;
}

// Получить одну заметку по ID
export async function getNote(id: string): Promise<Note> {
  const response = await api.get(`v1/notes/${id}`).json<{ data: Note }>();
  return response.data;
}

// Создать новую заметку
export async function createNote(
  data: {
    title: string;
    content: string;
    type?: string;
    metadata?: Record<string, unknown>;
  } & CometFields
): Promise<Note> {
  return api.post("v1/notes", { json: data }).json();
}

// COMET-1: «Ближайшие дела» — незакрытые кометы пользователя по группам.
export async function listComets(): Promise<CometsResponse> {
  const body = await api.get("v1/notes/comets", { cache: "no-store" }).json<{
    data: CometsResponse;
  }>();
  return body.data;
}

// COMET-1 stage D: «Добавить в календарь» — скачивает .ics датированной кометы.
export async function downloadCometIcs(id: string): Promise<Blob> {
  return api.get(`v1/notes/${id}/calendar.ics`).blob();
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
  await api.post("v1/notes/batch/delete", { json: { ids } });
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

// Получить рекомендации для заметки (похожие по явным связям и эмбеддингам).
// The API answers { suggestions, generated_at } (note_handler.go
// SuggestionsResponse); a bare array is accepted too.
export async function getSuggestions(id: string, limit = 10): Promise<Suggestion[]> {
  const body = await api
    .get(`v1/notes/${id}/suggestions`, { searchParams: { limit } })
    .json<{ suggestions?: Suggestion[] | null } | Suggestion[]>();
  if (Array.isArray(body)) return body;
  return body?.suggestions ?? [];
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
