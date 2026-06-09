// Хуки для использования предзагруженных данных в компонентах
import { getPreloadedGraph, getPreloadedGraphDelta, getPreloadedAchievements, hasPreloadedData } from '$lib/services/PreloadService';
import { getFullGraphData } from '$lib/api/graph';
import { getAllAchievements, getMyAchievements } from '$lib/api/users';
import type { GraphData } from '$lib/api/graph';

/**
 * Хук для получения данных графа с использованием предзагруженных данных
 * Возвращает предзагруженные данные если они есть, иначе загружает с сервера
 */
export async function getGraphWithPreload(limit: number = 1000): Promise<GraphData> {
  // Сначала пробуем получить предзагруженные данные
  const preloadedData = getPreloadedGraph();
  
  if (preloadedData) {
    console.log('[usePreloadedData] Using preloaded graph data');
    return preloadedData;
  }
  
  // Если нет предзагруженных данных, загружаем с сервера
  console.log('[usePreloadedData] Loading graph data from server');
  return await getFullGraphData(limit);
}

/**
 * Хук для получения достижений с использованием предзагруженных данных
 * Возвращает предзагруженные данные если они есть, иначе загружает с сервера
 */
export async function getAchievementsWithPreload(usePersonal: boolean = false): Promise<{
  achievements: Array<{
    id: string;
    code: string;
    title: string;
    description: string;
    icon: string;
    points: number;
    earned?: boolean;
    is_hidden?: boolean;
    obtained_at?: string;
  }>;
  total_points?: number;
}> {
  // Сначала пробуем получить предзагруженные данные
  const preloadedData = getPreloadedAchievements();
  
  if (preloadedData && !usePersonal) {
    console.log('[usePreloadedData] Using preloaded achievements data');
    return { achievements: preloadedData };
  }
  
  // Если нужны персональные данные или нет предзагруженных, загружаем с сервера
  console.log(`[usePreloadedData] Loading ${usePersonal ? 'personal' : 'all'} achievements from server`);
  
  if (usePersonal) {
    return await getMyAchievements();
  } else {
    return await getAllAchievements();
  }
}

/**
 * Реактивный хук для отслеживания наличия предзагруженных данных
 */
export function usePreloadedDataStatus() {
  return {
    hasData: hasPreloadedData(),
    hasGraph: !!getPreloadedGraph(),
    hasAchievements: !!getPreloadedAchievements()
  };
}

/**
 * Хук для получения мгновенного доступа к предзагруженным данным
 * Используется для немедленного отображения UI после входа
 */
export function useInstantData() {
  const graph = getPreloadedGraph();
  const delta = getPreloadedGraphDelta();
  const achievements = getPreloadedAchievements();
  
  return {
    graph: graph || { nodes: [], links: [] },
    delta,
    achievements: achievements || [],
    hasInstantData: !!(graph || achievements),
    isDataReady: hasPreloadedData()
  };
}

/**
 * Комбинированный хук для загрузки данных с приоритетом на предзагруженные
 */
export async function loadAppData(options: {
  limit?: number;
  usePersonalAchievements?: boolean;
  fallbackToServer?: boolean;
} = {}): Promise<{
  graph: GraphData;
  achievements: Array<{
    id: string;
    code: string;
    title: string;
    description: string;
    icon: string;
    points: number;
    earned?: boolean;
    is_hidden?: boolean;
    obtained_at?: string;
  }>;
  totalPoints?: number;
  usedPreloaded: {
    graph: boolean;
    achievements: boolean;
  };
}> {
  const {
    limit = 1000,
    usePersonalAchievements = false
  } = options;
  
  // Загружаем граф и достижения параллельно
  const [graphPromise, achievementsPromise] = [
    getGraphWithPreload(limit),
    getAchievementsWithPreload(usePersonalAchievements)
  ];
  
  const [graph, achievementsResult] = await Promise.all([
    graphPromise,
    achievementsPromise
  ]);
  
  const usedPreloaded = {
    graph: !!getPreloadedGraph(),
    achievements: !!getPreloadedAchievements() && !usePersonalAchievements
  };
  
  return {
    graph,
    achievements: achievementsResult.achievements,
    totalPoints: achievementsResult.total_points,
    usedPreloaded
  };
}
