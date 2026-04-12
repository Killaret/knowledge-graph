// API-клиент для получения данных графа (узлы и связи)
import ky from 'ky';

const api = ky.create({ prefixUrl: '/api' });

// Узел графа – заметка (звезда)
export interface GraphNode {
  id: string;
  title: string;
  type: string;
}

// Ребро графа – связь между заметками
export interface GraphLink {
  source: string;   // ID исходной заметки
  target: string;   // ID целевой заметки
  weight: number;    // вес связи (толщина линии)
}

// Данные графа: список узлов и рёбер
export interface GraphData {
  nodes: GraphNode[];
  links: GraphLink[];
}

// Запросить граф для заметки (возвращает все прямые связи и связанные заметки)
export async function getGraphData(noteId: string): Promise<GraphData> {
  return api.get(`notes/${noteId}/graph`).json();
}

// Запросить глобальный граф всех заметок
export async function getGlobalGraphData(): Promise<GraphData> {
  try {
    return api.get('graph').json();
  } catch {
    // Fallback: если эндпоинта нет, возвращаем пустой граф
    console.warn('Global graph endpoint not available, returning empty graph');
    return { nodes: [], links: [] };
  }
}