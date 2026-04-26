// API-клиент для получения данных графа (узлы и связи)
import { api } from './client';
import { apiConfig } from '$lib/config';

// Узел графа – заметка (звезда)
export interface GraphNode {
  id: string;
  title: string;
  type?: string;
  x?: number;
  y?: number;
  z?: number;
  size?: number;
}

// Ребро графа – связь между заметками
export interface GraphLink {
  source: string;   // ID исходной заметки
  target: string;   // ID целевой заметки
  weight?: number;    // вес связи (толщина линии)
  link_type?: string; // тип связи: reference, dependency, related, custom
}

// Данные графа: список узлов и рёбер
export interface GraphData {
  nodes: GraphNode[];
  links: GraphLink[];
}

/**
 * Обрабатывает ошибки API графа и возвращает понятное сообщение
 */
function handleGraphError(error: unknown, context: string): never {
  console.error(`[Graph API] ${context}:`, error);
  
  if (error instanceof Error) {
    if (error.message.includes('404') || error.message.includes('Not Found')) {
      throw new Error('Граф не найден. Возможно, заметка была удалена.');
    }
    if (error.message.includes('500')) {
      throw new Error('Ошибка сервера при загрузке графа. Попробуйте позже.');
    }
    if (error.message.includes('Failed to fetch') || error.message.includes('NetworkError')) {
      throw new Error('Не удалось подключиться к серверу. Проверьте интернет-соединение.');
    }
    throw new Error(`Ошибка загрузки графа: ${error.message}`);
  }
  
  throw new Error('Неизвестная ошибка при загрузке графа');
}

// Запросить граф для заметки (возвращает все прямые связи и связанные заметки)
export async function getGraphData(noteId: string, depth: number = 2): Promise<GraphData> {
  try {
    const response = await api.get(`notes/${noteId}/graph?depth=${depth}`).json<GraphData>();
    return response;
  } catch (error) {
    return handleGraphError(error, `Failed to load graph for note ${noteId}`);
  }
}

// Запросить полный граф всех заметок и связей
export async function getFullGraphData(limit: number = apiConfig.default_limit): Promise<GraphData> {
  try {
    // link_limit=0 означает загрузить все связи (без ограничения)
    const response = await api.get(`graph/all?limit=${limit}&link_limit=0`).json<GraphData>();
    return response;
  } catch (error) {
    return handleGraphError(error, 'Failed to load full graph');
  }
}