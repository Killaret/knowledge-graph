// API-клиент для работа со связями между заметками
import { api } from './client';

// Тип связи между заметками
export interface Link {
  id: string;
  source_note_id: string;
  target_note_id: string;
  link_type: string;
  weight: number;
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

// Данные для создания связи
export interface CreateLinkData {
  source_note_id: string;
  target_note_id: string;
  link_type: string;
  weight?: number;
  metadata?: Record<string, any>;
}

// Получить все связи
export async function getLinks(): Promise<Link[]> {
  const response = await api.get('v1/links').json<{ data: Link[] }>();
  return response.data;
}

// Получить связь по ID
export async function getLink(id: string): Promise<Link> {
  const response = await api.get(`v1/links/${id}`).json<{ data: Link }>();
  return response.data;
}

// Создать новую связь
export async function createLink(data: CreateLinkData): Promise<Link> {
  const response = await api.post('v1/links', { json: data }).json<{ data: Link }>();
  return response.data;
}

// Удалить связь
export async function deleteLink(id: string): Promise<void> {
  await api.delete(`v1/links/${id}`);
}

// Получить связи для заметки (исходящие и входящие)
export async function getNoteLinks(noteId: string): Promise<Link[]> {
  const response = await api.get(`v1/notes/${noteId}/links`).json<{ data: Link[] }>();
  return response.data;
}
