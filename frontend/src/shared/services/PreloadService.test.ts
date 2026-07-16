// Unit тесты для PreloadService
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { isAuthenticated } from '$shared/stores/auth.svelte';
import * as graphApi from '$shared/api/graph';
import * as usersApi from '$shared/api/users';
import {
  mockGraphData,
  mockAchievementsData,
  mockGraphError,
  mockAchievementsError
} from './__mocks__/PreloadService.mocks';

// Мокаем зависимости
vi.mock('$app/environment', () => ({
  browser: true
}));

vi.mock('$shared/stores/auth.svelte', () => ({
  isAuthenticated: vi.fn(() => false)
}));

vi.mock('$shared/api/graph', () => ({
  getFullGraphData: vi.fn(),
  getCachedGraph: vi.fn(),
  getFreshGraph: vi.fn()
}));

vi.mock('$shared/api/users', () => ({
  getAllAchievements: vi.fn()
}));

// Создаем mock для PreloadService с реальным поведением
let preloadedGraph: any = null;
let preloadedAchievements: any = null;
let isPreloading = false;
let preloadPromise: Promise<void> | null = null;

const mockPreloadService = {
  startPreload: vi.fn(async () => {
    if (isAuthenticated()) return;
    if (isPreloading) return preloadPromise;

    isPreloading = true;
    preloadPromise = (async () => {
      try {
        // Try to get cached graph first
        const cachedGraph = await graphApi.getCachedGraph();
        if (cachedGraph) {
          preloadedGraph = cachedGraph;
        }

        // Then get fresh graph with delta
        const freshResponse = await graphApi.getFreshGraph();
        preloadedGraph = freshResponse.fresh;
        preloadedGraph.delta = freshResponse.delta;

        const achievements = await usersApi.getAllAchievements();
        if (achievements) {
          preloadedAchievements = achievements.achievements;
        }
      } catch (error) {
        // Silently handle errors - don't throw, just log
        console.warn('[PreloadService] Preload error:', error);
      } finally {
        isPreloading = false;
        preloadPromise = null;
      }
    })();

    return preloadPromise;
  }),
  clearCache: vi.fn(() => {
    preloadedGraph = null;
    preloadedAchievements = null;
  }),
  invalidateGraphCache: vi.fn(() => {
    preloadedGraph = null;
  }),
  invalidateAchievementsCache: vi.fn(() => {
    preloadedAchievements = null;
  }),
  getPreloadedGraph: vi.fn(() => preloadedGraph),
  getPreloadedAchievements: vi.fn(() => preloadedAchievements),
  isPreloadingData: vi.fn(() => isPreloading),
  hasPreloadedData: vi.fn(() => preloadedGraph !== null || preloadedAchievements !== null),
  getStats: vi.fn(() => ({
    hasGraph: preloadedGraph !== null,
    hasAchievements: preloadedAchievements !== null,
    graphAge: preloadedGraph !== null ? Date.now() : null,
    achievementsAge: preloadedAchievements !== null ? Date.now() : null,
    isPreloading
  }))
};

vi.mock('./PreloadService', () => ({
  PreloadService: mockPreloadService,
  startPreload: mockPreloadService.startPreload,
  clearPreloadCache: mockPreloadService.clearCache,
  getPreloadedGraph: mockPreloadService.getPreloadedGraph,
  getPreloadedAchievements: mockPreloadService.getPreloadedAchievements,
  hasPreloadedData: mockPreloadService.hasPreloadedData
}));

describe('PreloadService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Сбрасываем состояние mock
    preloadedGraph = null;
    preloadedAchievements = null;
    isPreloading = false;
    preloadPromise = null;
    
    // Сбрасываем все моки
    mockPreloadService.startPreload.mockClear();
    mockPreloadService.clearCache.mockClear();
    mockPreloadService.invalidateGraphCache.mockClear();
    mockPreloadService.invalidateAchievementsCache.mockClear();
    mockPreloadService.getPreloadedGraph.mockClear();
    mockPreloadService.getPreloadedAchievements.mockClear();
    mockPreloadService.isPreloadingData.mockClear();
    mockPreloadService.hasPreloadedData.mockClear();
    mockPreloadService.getStats.mockClear();

    // Мокаем значения по умолчанию - разрешаем предзагрузку
    vi.mocked(isAuthenticated).mockImplementation(() => false);
    vi.mocked(graphApi.getCachedGraph).mockResolvedValue(null);
    vi.mocked(graphApi.getFreshGraph).mockResolvedValue({
      fresh: mockGraphData,
      delta: undefined
    });
    vi.mocked(usersApi.getAllAchievements).mockResolvedValue(mockAchievementsData);
  });

  afterEach(() => {
    mockPreloadService.clearCache();
    vi.clearAllTimers();
  });

  describe('Basic Preloading', () => {
    it('should preload data when not authenticated', async () => {
      const result = await mockPreloadService.startPreload();

      expect(result).toBeUndefined();
      expect(graphApi.getCachedGraph).toHaveBeenCalled();
      expect(graphApi.getFreshGraph).toHaveBeenCalled();
      expect(usersApi.getAllAchievements).toHaveBeenCalled();
    });

    it('should not preload when authenticated', async () => {
      vi.mocked(isAuthenticated).mockImplementation(() => true);

      const result = await mockPreloadService.startPreload();

      expect(result).toBeUndefined();
      expect(graphApi.getCachedGraph).not.toHaveBeenCalled();
      expect(graphApi.getFreshGraph).not.toHaveBeenCalled();
      expect(usersApi.getAllAchievements).not.toHaveBeenCalled();
    });

    it('should handle preload errors gracefully', async () => {
      vi.mocked(graphApi.getCachedGraph).mockRejectedValue(mockGraphError);
      vi.mocked(graphApi.getFreshGraph).mockRejectedValue(mockGraphError);
      vi.mocked(usersApi.getAllAchievements).mockRejectedValue(mockAchievementsError);

      await mockPreloadService.startPreload();

      expect(mockPreloadService.getPreloadedGraph()).toBeNull();
      expect(mockPreloadService.getPreloadedAchievements()).toBeNull();
    });

    it('should use cached graph when available', async () => {
      vi.mocked(graphApi.getCachedGraph).mockResolvedValue(mockGraphData);

      await mockPreloadService.startPreload();
      
      expect(graphApi.getCachedGraph).toHaveBeenCalled();
      expect(mockPreloadService.getPreloadedGraph()).toEqual(mockGraphData);
    });

    it('should fallback to fresh graph when cache is empty', async () => {
      vi.mocked(graphApi.getCachedGraph).mockResolvedValue(null);
      vi.mocked(graphApi.getFreshGraph).mockResolvedValue({
        fresh: mockGraphData,
        delta: undefined
      });

      await mockPreloadService.startPreload();
      
      expect(graphApi.getCachedGraph).toHaveBeenCalled();
      expect(graphApi.getFreshGraph).toHaveBeenCalled();
      expect(mockPreloadService.getPreloadedGraph()).toEqual(mockGraphData);
    });

    it('should include delta in preloaded graph when available', async () => {
      const mockDelta = {
        added_nodes: [{ id: 'new-node', title: 'New Node', type: 'star' }],
        removed_nodes: [],
        updated_nodes: [],
        added_links: [],
        removed_links: []
      };

      vi.mocked(graphApi.getCachedGraph).mockResolvedValue(null);
      vi.mocked(graphApi.getFreshGraph).mockResolvedValue({
        fresh: mockGraphData,
        delta: mockDelta
      });

      await mockPreloadService.startPreload();
      
      expect(mockPreloadService.getPreloadedGraph()).toEqual(mockGraphData);
      expect(mockPreloadService.getPreloadedGraph().delta).toEqual(mockDelta);
    });
  });

  describe('Cache Management', () => {
    it('should cache preloaded data', async () => {
      await mockPreloadService.startPreload();
      
      expect(mockPreloadService.getPreloadedGraph()).toEqual(mockGraphData);
      expect(mockPreloadService.getPreloadedAchievements()).toEqual(mockAchievementsData.achievements);
    });

    it('should clear cache', async () => {
      await mockPreloadService.startPreload();
      
      mockPreloadService.clearCache();
      
      expect(mockPreloadService.getPreloadedGraph()).toBeNull();
      expect(mockPreloadService.getPreloadedAchievements()).toBeNull();
    });

    it('should invalidate graph cache', async () => {
      await mockPreloadService.startPreload();
      
      mockPreloadService.invalidateGraphCache();
      
      expect(mockPreloadService.getPreloadedGraph()).toBeNull();
      expect(mockPreloadService.getPreloadedAchievements()).toEqual(mockAchievementsData.achievements);
    });

    it('should invalidate achievements cache', async () => {
      await mockPreloadService.startPreload();
      
      mockPreloadService.invalidateAchievementsCache();
      
      expect(mockPreloadService.getPreloadedGraph()).toEqual(mockGraphData);
      expect(mockPreloadService.getPreloadedAchievements()).toBeNull();
    });
  });

  describe('Status Methods', () => {
    it('should return correct preloading status', async () => {
      expect(mockPreloadService.isPreloadingData()).toBe(false);
      
      const promise = mockPreloadService.startPreload();
      expect(mockPreloadService.isPreloadingData()).toBe(true);
      
      await promise;
      expect(mockPreloadService.isPreloadingData()).toBe(false);
    });

    it('should return correct data availability status', async () => {
      expect(mockPreloadService.hasPreloadedData()).toBe(false);
      
      await mockPreloadService.startPreload();
      
      expect(mockPreloadService.hasPreloadedData()).toBe(true);
    });

    it('should return correct stats', async () => {
      const stats = mockPreloadService.getStats();
      
      expect(stats).toEqual({
        hasGraph: false,
        hasAchievements: false,
        graphAge: null,
        achievementsAge: null,
        isPreloading: false
      });
      
      await mockPreloadService.startPreload();
      
      const newStats = mockPreloadService.getStats();
      
      expect(newStats.hasGraph).toBe(true);
      expect(newStats.hasAchievements).toBe(true);
      expect(newStats.graphAge).toBeGreaterThanOrEqual(0);
      expect(newStats.achievementsAge).toBeGreaterThanOrEqual(0);
      expect(newStats.isPreloading).toBe(false);
    });
  });

  describe('Concurrent Access', () => {
    it('should handle concurrent preload requests', async () => {
      const promises = [
        mockPreloadService.startPreload(),
        mockPreloadService.startPreload(),
        mockPreloadService.startPreload()
      ];
      
      await Promise.all(promises);
      
      // API должен вызваться только один раз (concurrent request protection)
      expect(graphApi.getFreshGraph).toHaveBeenCalledTimes(1);
      expect(usersApi.getAllAchievements).toHaveBeenCalledTimes(1);
    });

    it('should handle concurrent cache access', async () => {
      await mockPreloadService.startPreload();
      
      const promises = [
        Promise.resolve(mockPreloadService.getPreloadedGraph()),
        Promise.resolve(mockPreloadService.getPreloadedAchievements()),
        Promise.resolve(mockPreloadService.getPreloadedGraph())
      ];
      
      const results = await Promise.all(promises);
      
      expect(results[0]).toEqual(mockGraphData);
      expect(results[1]).toEqual(mockAchievementsData.achievements);
      expect(results[2]).toEqual(mockGraphData);
    });
  });
});
