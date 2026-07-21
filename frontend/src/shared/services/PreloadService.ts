// PreloadService - фоновая предзагрузка данных для ускорения первого взаимодействия после входа
import { browser } from "$app/environment";
import { isAuthenticated } from "$shared/stores/auth-session.svelte";
import {
  getFullGraphData,
  getFreshGraph,
  type GraphData,
  type GraphDeltaData,
} from "$shared/api/graph";

// Типы для кэшированных данных
interface PreloadedGraphData {
  data: GraphData;
  timestamp: number;
  ttl: number; // Time to live в миллисекундах
  delta?: GraphDeltaData; // Delta for incremental updates
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

  // TTL для кэша (5 минут для графа)
  private readonly GRAPH_TTL = 5 * 60 * 1000;

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
    if (this.preloadPromise) {
      return this.preloadPromise;
    }

    if (!browser || this.isPreloading || isAuthenticated()) {
      return;
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

  public async preloadAuthenticatedGraph(): Promise<void> {
    if (this.preloadPromise) {
      return this.preloadPromise;
    }

    if (!browser || this.isPreloading || !isAuthenticated()) {
      return;
    }

    this.isPreloading = true;
    this.preloadPromise = this.performAuthenticatedPreload();

    try {
      await this.preloadPromise;
    } finally {
      this.isPreloading = false;
      this.preloadPromise = null;
    }
  }

  /**
   * Выполняет фактическую предзагрузку данных для гостя
   */
  private async performPreload(): Promise<void> {
    if (import.meta.env.DEV) {
      console.log("[PreloadService] Starting background preload...");
    }

    try {
      await this.preloadGraph();

      if (import.meta.env.DEV) {
        console.log("[PreloadService] Graph preloaded successfully");
      }
    } catch (error) {
      if (import.meta.env.DEV) {
        console.warn("[PreloadService] Preload completed with errors:", error);
      }
    }
  }

  /**
   * Выполняет предзагрузку данных графа для аутентифицированного пользователя
   */
  private async performAuthenticatedPreload(): Promise<void> {
    if (import.meta.env.DEV) {
      console.log("[PreloadService] Starting authenticated graph preload...");
    }

    try {
      // Получаем только свежий граф (cached не нужен, так как fresh перезапишет)
      const freshResult = await getFreshGraph();

      this.preloadedGraph = {
        data: freshResult.fresh,
        timestamp: Date.now(),
        ttl: this.GRAPH_TTL,
        delta: freshResult.delta,
      };

      if (freshResult.delta) {
        if (import.meta.env.DEV) {
          console.log(
            "[PreloadService] Delta available for authenticated update",
          );
        }
      }
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error(
          "[PreloadService] Error during authenticated graph preload:",
          error,
        );
      }
      throw error;
    }
  }

  /**
   * Предзагружает данные графа
   */
  private async preloadGraph(): Promise<void> {
    try {
      const graphData = await getFullGraphData();
      this.preloadedGraph = {
        data: graphData,
        timestamp: Date.now(),
        ttl: this.GRAPH_TTL,
      };
      if (import.meta.env.DEV) {
        console.log("[PreloadService] Public graph preloaded successfully");
      }
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error("[PreloadService] Error preloading public graph:", error);
      }
      throw error;
    }
  }

  /**
   * Получает предзагруженные данные графа, если они актуальны
   */
  public getPreloadedGraphData(): PreloadedGraphData | null {
    if (!this.preloadedGraph) {
      return null;
    }

    const now = Date.now();
    if (now - this.preloadedGraph.timestamp > this.preloadedGraph.ttl) {
      this.preloadedGraph = null;
      return null;
    }

    return this.preloadedGraph;
  }

  public getPreloadedGraph(): GraphData | null {
    return this.getPreloadedGraphData()?.data ?? null;
  }

  public getPreloadedGraphDelta(): GraphDeltaData | null {
    return this.getPreloadedGraphData()?.delta ?? null;
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
    if (
      now - this.preloadedAchievements.timestamp >
      this.preloadedAchievements.ttl
    ) {
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
    if (import.meta.env.DEV) {
      console.log("[PreloadService] Clearing preload cache...");
    }
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
      graphAge: this.preloadedGraph
        ? now - this.preloadedGraph.timestamp
        : null,
      achievementsAge: this.preloadedAchievements
        ? now - this.preloadedAchievements.timestamp
        : null,
      isPreloading: this.isPreloading,
    };
  }
}

// Экспортируем синглтон
export const PreloadService = PreloadServiceClass.getInstance();

// Экспортием удобные функции для использования в компонентах
export function startPreload(): Promise<void> {
  return PreloadService.startPreload();
}

export function preloadAuthenticatedGraph(): Promise<void> {
  return PreloadService.preloadAuthenticatedGraph();
}

export function getPreloadedGraph(): GraphData | null {
  return PreloadService.getPreloadedGraph();
}

export function getPreloadedGraphData(): PreloadedGraphData | null {
  return PreloadService.getPreloadedGraphData();
}

export function getPreloadedGraphDelta(): GraphDeltaData | null {
  return PreloadService.getPreloadedGraphDelta();
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

export function isPreloadingData(): boolean {
  return PreloadService.isPreloadingData();
}

export function getStats(): {
  hasGraph: boolean;
  hasAchievements: boolean;
  graphAge: number | null;
  achievementsAge: number | null;
  isPreloading: boolean;
} {
  return PreloadService.getStats();
}
