// API-клиент для работа со связями между заметками
import { api } from "./client";

// Тип связи между заметками
export interface Link {
  id: string;
  source_note_id: string;
  target_note_id: string;
  link_type: string;
  weight: number;
  source_type?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  last_weight_update?: string;
}

// Данные для создания связи
export interface CreateLinkData {
  source_note_id: string;
  target_note_id: string;
  link_type: string;
  weight?: number;
  metadata?: Record<string, unknown>;
}

// Данные для обновления связи
export interface UpdateLinkData {
  link_type?: string;
  weight?: number;
  metadata?: Record<string, unknown>;
}

// Получить все связи
export async function getLinks(): Promise<Link[]> {
  const response = await api.get("v1/links").json<{ data: Link[] }>();
  return response.data;
}

// Получить связь по ID
export async function getLink(id: string): Promise<Link> {
  const response = await api.get(`v1/links/${id}`).json<{ data: Link }>();
  return response.data;
}

// Создать новую связь
export async function createLink(data: CreateLinkData): Promise<Link> {
  const response = await api.post("v1/links", { json: data }).json<{ data: Link }>();
  return response.data;
}

// Обновить связь
export async function updateLink(id: string, data: UpdateLinkData): Promise<Link> {
  const response = await api.put(`v1/links/${id}`, { json: data }).json<{ data: Link }>();
  return response.data;
}

// Удалить связь
export async function deleteLink(id: string): Promise<void> {
  await api.delete(`v1/links/${id}`);
}

// Получить связи для заметки (исходящие и входящие)
// Backend returns {data: {incoming, outgoing}} — merge them; a self-loop link
// lands in both lists, so dedupe by id. Flat array tolerated for compatibility.
export async function getNoteLinks(noteId: string): Promise<Link[]> {
  const response = await api
    .get(`v1/notes/${noteId}/links`)
    .json<{ data: Link[] | { incoming?: Link[]; outgoing?: Link[] } }>();
  const d = response.data;
  if (Array.isArray(d)) return d;
  const merged = [...(d?.outgoing ?? []), ...(d?.incoming ?? [])];
  return merged.filter((l, i) => merged.findIndex((x) => x.id === l.id) === i);
}

// Delete all links for a note
export async function deleteAllNoteLinks(noteId: string): Promise<void> {
  await api.delete(`v1/notes/${noteId}/links`);
}
