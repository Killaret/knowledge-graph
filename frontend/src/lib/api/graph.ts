// API-клиент для получения данных графа (узлы и связи)
import { api } from './client';
import { apiConfig } from '$lib/config';

function getGraphApi() {
  // Используем graph service через прокси в SvelteKit hooks
  // Прокси перенаправляет /graph-service/api/* -> http://localhost:9091/api/*
  return api.extend({ prefixUrl: '/graph-service/api' });
}

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

// API response wrapper structure
interface GraphApiResponse {
  data: GraphData;
  meta?: {
    total_nodes?: number;
    total_links?: number;
    limit?: number;
  };
}

function buildQuery(params: Record<string, string | number | undefined>): string {
  return Object.entries(params)
    .filter(([, value]) => value !== undefined && value !== '')
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)
    .join('&');
}

// Запросить граф для заметки (возвращает все прямые связи и связанные заметки)
export async function getGraphData(noteId: string, depth: number = 2, userId?: string): Promise<GraphData> {
  try {
    const query = buildQuery({ depth, user_id: userId });
    const response = await getGraphApi().get(`v1/graph/note/${encodeURIComponent(noteId)}${query ? `?${query}` : ''}`).json<GraphApiResponse>();
    return response.data || { nodes: [], links: [] };
  } catch (error) {
    return handleGraphError(error, `Failed to load graph for note ${noteId}`);
  }
}

// Запросить полный граф всех заметок и связей
export async function getFullGraphData(limit: number = apiConfig.default_limit, userId?: string): Promise<GraphData> {
  try {
    const query = buildQuery({ limit, user_id: userId });
    const response = await getGraphApi().get(`v1/graph/full${query ? `?${query}` : ''}`).json<GraphApiResponse>();
    return response.data || { nodes: [], links: [] };
  } catch (error) {
    return handleGraphError(error, 'Failed to load full graph');
  }
}

// Graph delta structure for incremental updates
export interface GraphDelta {
  added_nodes?: GraphNode[];
  removed_nodes?: string[];
  updated_nodes?: GraphNode[];
  added_links?: GraphLink[];
  removed_links?: GraphLink[];
}

// Fresh graph response with optional delta
export interface FreshGraphResponse {
  fresh: GraphData;
  delta?: GraphDelta;
}

// Запросить кэшированный граф пользователя
export async function getCachedGraph(): Promise<GraphData | null> {
  try {
    const response = await api.get('v1/me/graph/cached');
    if (response.status === 204) {
      return null; // No cached data
    }
    return await response.json<GraphData>();
  } catch (error) {
    console.warn('[Graph API] Failed to get cached graph:', error);
    return null; // Return null on error, fallback to fresh
  }
}

// Запросить свежий граф с опциональным дельта-обновлением
export async function getFreshGraph(userId?: string): Promise<FreshGraphResponse> {
  try {
    if (userId) {
      const response = await getGraphApi().get(`v1/graph/full?${buildQuery({ user_id: userId })}`);
      const fullGraph = await response.json<GraphData>();
      return { fresh: fullGraph };
    }

    const response = await api.get('v1/me/graph/fresh').json<FreshGraphResponse>();
    return response;
  } catch (error) {
    return handleGraphError(error, 'Failed to load fresh graph');
  }
}