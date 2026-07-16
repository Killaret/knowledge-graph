// Моки для API функций, используемых в PreloadService
import { vi } from 'vitest';
import type { GraphData } from '$shared/api/graph';

// Моковые данные для графа
export const mockGraphData: GraphData = {
  nodes: [
    {
      id: '1',
      title: 'Test Note 1',
      type: 'star',
      x: 100,
      y: 100,
      z: 0,
      size: 10
    },
    {
      id: '2',
      title: 'Test Note 2',
      type: 'planet',
      x: 200,
      y: 200,
      z: 0,
      size: 15
    },
    {
      id: '3',
      title: 'Test Note 3',
      type: 'moon',
      x: 300,
      y: 150,
      z: 0,
      size: 8
    }
  ],
  links: [
    {
      source: '1',
      target: '2',
      weight: 2,
      link_type: 'reference'
    },
    {
      source: '2',
      target: '3',
      weight: 1,
      link_type: 'related'
    }
  ]
};

// Моковые данные для достижений
export const mockAchievementsData = {
  achievements: [
    {
      id: '1',
      code: 'first_note',
      title: 'First Note',
      description: 'Create your first note',
      icon: '📝',
      points: 10,
      earned: false,
      is_hidden: false
    },
    {
      id: '2',
      code: 'graph_explorer',
      title: 'Graph Explorer',
      description: 'Explore the knowledge graph',
      icon: '🌐',
      points: 25,
      earned: true,
      is_hidden: false
    },
    {
      id: '3',
      code: 'secret_achievement',
      title: 'Secret Achievement',
      description: 'Hidden achievement',
      icon: '🎯',
      points: 50,
      earned: false,
      is_hidden: true
    }
  ]
};

// Моковые персональные достижения
export const mockPersonalAchievementsData = {
  achievements: [
    {
      id: '1',
      code: 'first_note',
      title: 'First Note',
      description: 'Create your first note',
      icon: '📝',
      points: 10,
      earned: true,
      obtained_at: '2024-01-01T00:00:00Z'
    },
    {
      id: '2',
      code: 'graph_explorer',
      title: 'Graph Explorer',
      description: 'Explore the knowledge graph',
      icon: '🌐',
      points: 25,
      earned: true,
      obtained_at: '2024-01-02T00:00:00Z'
    }
  ],
  total_points: 35
};

// Моковые функции API
export const mockGetFullGraphData = vi.fn().mockResolvedValue(mockGraphData);
export const mockGetAllAchievements = vi.fn().mockResolvedValue(mockAchievementsData);
export const mockGetMyAchievements = vi.fn().mockResolvedValue(mockPersonalAchievementsData);

// Моки для ошибок
export const mockGraphError = new Error('Failed to load graph');
export const mockAchievementsError = new Error('Failed to load achievements');

export const mockGetFullGraphDataError = vi.fn().mockRejectedValue(mockGraphError);
export const mockGetAllAchievementsError = vi.fn().mockRejectedValue(mockAchievementsError);

// Мок для задержки (имитация сетевого запроса)
export const createDelayedMock = <T>(data: T, delay: number = 100) => {
  return vi.fn().mockImplementation(() => 
    new Promise<T>((resolve) => setTimeout(() => resolve(data), delay))
  );
};

// Мок для ошибки с задержкой
export const createDelayedError = (error: Error, delay: number = 100) => {
  return vi.fn().mockImplementation(() => 
    new Promise<never>((_, reject) => setTimeout(() => reject(error), delay))
  );
};
