// API-клиент для работы со связями (links) между заметками
import ky from 'ky';

const api = ky.create({ prefixUrl: '/api' });

// Тип связи
export interface Link {
  id: string;
  source_note_id: string;
  target_note_id: string;
  link_type: string;
  weight: number;
  created_at: string;
  updated_at: string;
}

// Создать связь между заметками
export async function createLink(
  sourceNoteId: string,
  targetNoteId: string,
  linkType: string = 'reference',
  weight: number = 1.0
): Promise<Link> {
  return api.post('links', {
    json: {
      source_note_id: sourceNoteId,
      target_note_id: targetNoteId,
      link_type: linkType,
      weight: weight
    }
  }).json();
}

// Получить все связи заметки
export async function getNoteLinks(noteId: string): Promise<Link[]> {
  return api.get(`notes/${noteId}/links`).json();
}

// Удалить связь
export async function deleteLink(linkId: string): Promise<void> {
  return api.delete(`links/${linkId}`).json();
}

// Обновить связь
export async function updateLink(
  linkId: string,
  updates: Partial<Pick<Link, 'link_type' | 'weight'>>
): Promise<Link> {
  return api.put(`links/${linkId}`, { json: updates }).json();
}

// Batch создание связей (если бэкенд поддерживает)
export async function createLinksBatch(
  links: Array<{
    source_note_id: string;
    target_note_id: string;
    link_type: string;
    weight: number;
  }>
): Promise<Link[]> {
  try {
    // Пробуем batch endpoint если есть
    return api.post('links/batch', { json: { links } }).json();
  } catch (e) {
    // Fallback: создаём последовательно
    const created: Link[] = [];
    for (const link of links) {
      const newLink = await createLink(
        link.source_note_id,
        link.target_note_id,
        link.link_type,
        link.weight
      );
      created.push(newLink);
    }
    return created;
  }
}
