// PreloadService - фоновая предзагрузка данных для ускорения первого взаимодействия после входа
import { browser } from '$app/environment';
import { isAuthenticated } from '$lib/stores/auth.svelte';
import { getFullGraphData, getCachedGraph, getFreshGraph, type GraphData, type GraphDelta } from '$lib/api/graph';
import { getAllAchievements } from '$lib/api/users';

// Типы для кэшированных данных
interface PreloadedGraphData {
  data: GraphData;
  timestamp: number;
  ttl: number; // Time to live в миллисекундах
  delta?: GraphDelta; // Delta for incremental updates
}

interface PreloadedAchievementsData {
  achievements: Array<{
    id: string;
    code: string;
    title: string;
    description: string;
    icon: string;
    points: number;
    earned: boolean;
    is_hidden: boolean;
  }>;
  timestamp: number;
  ttl: number;
}

// Класс для управления предзагрузкой
class PreloadServiceClass {
  private static instance: PreloadServiceClass;
  private preloadedGraph: PreloadedGraphData | null = null;
  private preloadedAchievements: PreloadedAchievementsData | null = null;
  private isPreloading: boolean = false;
  private preloadPromise: Promise<void> | null = null;

  // TTL для кэша (5 минут для графа, 10 минут для достижений)
  private readonly GRAPH_TTL = 5 * 60 * 1000;
  private readonly ACHIEVEMENTS_TTL = 10 * 60 * 1000;

  private constructor() {}

  public static getInstance(): PreloadServiceClass {
    if (!PreloadServiceClass.instance) {
      PreloadServiceClass.instance = new PreloadServiceClass();
    }
    return PreloadServiceClass.instance;
  }

  /**
   * Запускает фоновую предзагрузку данных, если пользователь не аутентифицирован
   */
  public async startPreload(): Promise<void> {
    if (!browser || this.isPreloading || isAuthenticated()) {
      return;
    }

    // Если уже есть активная предзагрузка, возвращаем её Promise
    if (this.preloadPromise) {
      return this.preloadPromise;
    }

    this.isPreloading = true;
    this.preloadPromise = this.performPreload();

    try {
      await this.preloadPromise;
    } finally {
      this.isPreloading = false;
      this.preloadPromise = null;
    }
  }

  /**
   * Выполняет фактическую предзагрузку данных
   */
  private async performPreload(): Promise<void> {
    console.log('[PreloadService] Starting background preload...');
    
    try {
      // Параллельно загружаем граф и достижения
      const [graphData, achievementsData] = await Promise.allSettled([
        this.preloadGraph(),
        this.preloadAchievements()
      ]);

      if (graphData.status === 'fulfilled') {
        console.log('[PreloadService] Graph preloaded successfully');
      } else {
        console.warn('[PreloadService] Failed to preload graph:', graphData.reason);
      }

      if (achievementsData.status === 'fulfilled') {
        console.log('[PreloadService] Achievements preloaded successfully');
      } else {
        console.warn('[PreloadService] Failed to preload achievements:', achievementsData.reason);
      }
    } catch (error) {
      console.warn('[PreloadService] Preload completed with errors:', error);
    }
  }

  /**
   * Предзагружает данные графа
   */
  private async preloadGraph(): Promise<void> {
    try {
      // Сначала пробуем получить кэшированный граф для мгновенного отображения
      const cachedGraph = await getCachedGraph();
      
      if (cachedGraph) {
        console.log('[PreloadService] Using cached graph for instant display');
        this.preloadedGraph = {
          data: cachedGraph,
          timestamp: Date.now(),
          ttl: this.GRAPH_TTL
        };
      }

      // Параллельно запрашиваем свежий граф с дельта
      const freshResponse = await getFreshGraph();
      
      this.preloadedGraph = {
        data: freshResponse.fresh,
        timestamp: Date.now(),
        ttl: this.GRAPH_TTL,
        delta: freshResponse.delta
      };

      if (freshResponse.delta) {
        console.log('[PreloadService] Delta available for incremental update:', freshResponse.delta);
      }
    } catch (error) {
      console.error('[PreloadService] Error preloading graph:', error);
      throw error;
    }
  }

  /**
   * Предзагружает список достижений
   */
  private async preloadAchievements(): Promise<void> {
    try {
      const achievementsData = await getAllAchievements();
      
      this.preloadedAchievements = {
        achievements: achievementsData.achievements,
        timestamp: Date.now(),
        ttl: this.ACHIEVEMENTS_TTL
      };
    } catch (error) {
      console.error('[PreloadService] Error preloading achievements:', error);
      throw error;
    }
  }

  /**
   * Получает предзагруженные данные графа, если они актуальны
   */
  public getPreloadedGraph(): GraphData | null {
    if (!this.preloadedGraph) {
      return null;
    }

    const now = Date.now();
    if (now - this.preloadedGraph.timestamp > this.preloadedGraph.ttl) {
      // Кэш устарел
      this.preloadedGraph = null;
      return null;
    }

    return this.preloadedGraph.data;
  }

  /**
   * Получает предзагруженные данные достижений, если они актуальны
   */
  public getPreloadedAchievements(): Array<{
    id: string;
    code: string;
    title: string;
    description: string;
    icon: string;
    points: number;
    earned: boolean;
    is_hidden: boolean;
  }> | null {
    if (!this.preloadedAchievements) {
      return null;
    }

    const now = Date.now();
    if (now - this.preloadedAchievements.timestamp > this.preloadedAchievements.ttl) {
      // Кэш устарел
      this.preloadedAchievements = null;
      return null;
    }

    return this.preloadedAchievements.achievements;
  }

  /**
   * Проверяет,正在进行 ли предзагрузка
   */
  public isPreloadingData(): boolean {
    return this.isPreloading;
  }

  /**
   * Проверяет, есть ли предзагруженные данные
   */
  public hasPreloadedData(): boolean {
    return !!(this.preloadedGraph || this.preloadedAchievements);
  }

  /**
   * Очищает весь кэш (вызывается при выходе)
   */
  public clearCache(): void {
    console.log('[PreloadService] Clearing preload cache...');
    this.preloadedGraph = null;
    this.preloadedAchievements = null;
  }

  /**
   * Инвалидирует только кэш графа
   */
  public invalidateGraphCache(): void {
    this.preloadedGraph = null;
  }

  /**
   * Инвалидирует только кэш достижений
   */
  public invalidateAchievementsCache(): void {
    this.preloadedAchievements = null;
  }

  /**
   * Получает статистику предзагрузки (для отладки)
   */
  public getStats(): {
    hasGraph: boolean;
    hasAchievements: boolean;
    graphAge: number | null;
    achievementsAge: number | null;
    isPreloading: boolean;
  } {
    const now = Date.now();
    
    return {
      hasGraph: !!this.preloadedGraph,
      hasAchievements: !!this.preloadedAchievements,
      graphAge: this.preloadedGraph ? now - this.preloadedGraph.timestamp : null,
      achievementsAge: this.preloadedAchievements ? now - this.preloadedAchievements.timestamp : null,
      isPreloading: this.isPreloading
    };
  }
}

// Экспортируем синглтон
export const PreloadService = PreloadServiceClass.getInstance();

// Экспортием удобные функции для использования в компонентах
export function startPreload(): Promise<void> {
  return PreloadService.startPreload();
}

export function getPreloadedGraph(): GraphData | null {
  return PreloadService.getPreloadedGraph();
}

export function getPreloadedAchievements(): Array<{
  id: string;
  code: string;
  title: string;
  description: string;
  icon: string;
  points: number;
  earned: boolean;
  is_hidden: boolean;
}> | null {
  return PreloadService.getPreloadedAchievements();
}

export function clearPreloadCache(): void {
  PreloadService.clearCache();
}

export function hasPreloadedData(): boolean {
  return PreloadService.hasPreloadedData();
}
