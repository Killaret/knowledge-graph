// API-клиент для получения данных графа (узлы и связи)
import ky from 'ky';

const api = ky.create({ 
  prefixUrl: '/api',
  timeout: 30000,
  retry: {
    limit: 2,
    methods: ['get'],
    statusCodes: [408, 413, 429, 500, 502, 503, 504]
  }
});

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

// Запросить граф для заметки (возвращает все прямые связи и связанные заметки)
export async function getGraphData(noteId: string, depth: number = 2): Promise<GraphData> {
  return api.get(`notes/${noteId}/graph?depth=${depth}`).json();
}

// Запросить полный граф всех заметок и связей
export async function getFullGraphData(limit: number = 100): Promise<GraphData> {
  return api.get(`graph/all?limit=${limit}`).json();
}