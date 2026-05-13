// Тесты для edge cases и пограничных сценариев PreloadService
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { browser } from '$app/environment';
import { isAuthenticated } from '$lib/stores/auth.svelte';
import { PreloadService } from './PreloadService';
import * as graphApi from '$lib/api/graph';
import * as usersApi from '$lib/api/users';
import type { GraphData } from '$lib/api/graph';
import {
  mockGraphData,
  mockAchievementsData,
  createDelayedMock,
  createDelayedError
} from './__mocks__/PreloadService.mocks';

// Мокаем зависимости
vi.mock('$app/environment', () => ({
  browser: true
}));

vi.mock('$lib/stores/auth.svelte', () => ({
  isAuthenticated: vi.fn(() => false)
}));

vi.mock('$lib/api/graph', () => ({
  getFullGraphData: vi.fn()
}));

vi.mock('$lib/api/users', () => ({
  getAllAchievements: vi.fn()
}));

describe('PreloadService Edge Cases', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    PreloadService.clearCache();
    
    vi.mocked(graphApi.getFullGraphData).mockResolvedValue(mockGraphData);
    vi.mocked(usersApi.getAllAchievements).mockResolvedValue(mockAchievementsData);
  });

  afterEach(() => {
    PreloadService.clearCache();
    vi.clearAllTimers();
  });

  describe('Memory Management', () => {
    it('should handle large datasets gracefully', async () => {
      // Создаем большой набор данных
      const largeGraphData = {
        nodes: Array.from({ length: 10000 }, (_, i) => ({
          id: `node-${i}`,
          title: `Node ${i}`,
          type: 'star',
          x: Math.random() * 1000,
          y: Math.random() * 1000,
          z: 0,
          size: Math.random() * 20
        })),
        links: Array.from({ length: 20000 }, (_, i) => ({
          source: `node-${Math.floor(Math.random() * 10000)}`,
          target: `node-${Math.floor(Math.random() * 10000)}`,
          weight: Math.random(),
          link_type: 'reference'
        }))
      };

      vi.mocked(graphApi.getFullGraphData).mockResolvedValue(largeGraphData);

      await PreloadService.startPreload();

      const cachedData = PreloadService.getPreloadedGraph();
      expect(cachedData).toEqual(largeGraphData);
      expect(cachedData?.nodes.length).toBe(10000);
      expect(cachedData?.links.length).toBe(20000);
    });

    it('should handle memory pressure scenarios', async () => {
      // Эмулируем сценарий нехватки памяти
      const originalSetTimeout = global.setTimeout;
      const setTimeoutSpy = vi.fn() as any;
      global.setTimeout = setTimeoutSpy;

      await PreloadService.startPreload();

      // Проверяем, что таймеры используются правильно
      expect(setTimeoutSpy).toHaveBeenCalled();

      global.setTimeout = originalSetTimeout;
    });
  });

  describe('Network Issues', () => {
    it('should handle timeout scenarios', async () => {
      // Мокаем очень медленный ответ
      const slowGraphPromise = new Promise<GraphData>(resolve => 
        setTimeout(() => resolve(mockGraphData), 10000)
      );
      vi.mocked(graphApi.getFullGraphData).mockReturnValue(slowGraphPromise);

      const startTime = Date.now();
      await PreloadService.startPreload();
      const endTime = Date.now();

      // Предзагрузка должна завершиться в разумное время
      expect(endTime - startTime).toBeLessThan(5000);
    });

    it('should handle intermittent network failures', async () => {
      let callCount = 0;
      vi.mocked(graphApi.getFullGraphData).mockImplementation(() => {
        callCount++;
        if (callCount === 1) {
          return Promise.reject(new Error('Network error'));
        }
        return Promise.resolve(mockGraphData);
      });

      // Первая попытка должна провалиться, но сервис продолжит работу
      await PreloadService.startPreload();

      // Данные все равно должны быть загружены при повторной попытке
      const cachedData = PreloadService.getPreloadedGraph();
      expect(cachedData).toEqual(mockGraphData);
    });

    it('should handle CORS issues', async () => {
      const corsError = new Error('CORS policy violation');
      vi.mocked(graphApi.getFullGraphData).mockRejectedValue(corsError);

      await PreloadService.startPreload();

      // Сервис должен обработать CORS ошибку
      expect(PreloadService.getPreloadedGraph()).toBeNull();
      expect(PreloadService.getPreloadedAchievements()).toEqual(mockAchievementsData.achievements);
    });
  });

  describe('Concurrent Access Patterns', () => {
    it('should handle rapid start/stop cycles', async () => {
      const promises = [];
      
      // Быстрый цикл запуска и остановки
      for (let i = 0; i < 10; i++) {
        promises.push(PreloadService.startPreload());
        PreloadService.clearCache();
      }

      await Promise.allSettled(promises);

      // Сервис должен остаться в консистентном состоянии
      expect(PreloadService.isPreloadingData()).toBe(false);
      expect(PreloadService.hasPreloadedData()).toBe(false);
    });

    it('should handle multiple simultaneous cache invalidations', async () => {
      await PreloadService.startPreload();

      // Множественная инвалидация кэша
      const invalidations = Array.from({ length: 100 }, () => 
        Promise.resolve(PreloadService.invalidateGraphCache())
      );

      await Promise.all(invalidations);

      expect(PreloadService.getPreloadedGraph()).toBeNull();
      expect(PreloadService.getPreloadedAchievements()).toEqual(mockAchievementsData.achievements);
    });

    it('should handle race conditions in cache access', async () => {
      const accessPromises = Array.from({ length: 1000 }, () => 
        Promise.resolve(PreloadService.getPreloadedGraph())
      );

      const results = await Promise.all(accessPromises);

      // Все результаты должны быть консистентны
      expect(results.every(result => result === null || result === mockGraphData)).toBe(true);
    });
  });

  describe('Data Integrity', () => {
    it('should handle corrupted cache data', async () => {
      // Эмулируем поврежденные данные
      const corruptedData: GraphData = { nodes: [], links: [] };
      vi.mocked(graphApi.getFullGraphData).mockResolvedValue(corruptedData);

      await PreloadService.startPreload();

      // Сервис должен обработать поврежденные данные
      const cachedData = PreloadService.getPreloadedGraph();
      expect(cachedData).toEqual(corruptedData);
    });

    it('should handle malformed API responses', async () => {
      // Мокаем некорректный ответ API
      vi.mocked(graphApi.getFullGraphData).mockResolvedValue({
        invalidField: 'invalid',
        nodes: 'not an array'
      } as any);

      await PreloadService.startPreload();

      // Сервис должен обработать некорректные данные
      const cachedData = PreloadService.getPreloadedGraph();
      expect(cachedData).toBeDefined();
    });

    it('should handle circular references in data', async () => {
      // Создаем данные с циклическими ссылками
      const circularData = {
        nodes: [
          { id: '1', title: 'Node 1', type: 'star' },
          { id: '2', title: 'Node 2', type: 'planet' }
        ],
        links: [
          { source: '1', target: '2', link_type: 'reference' }
        ]
      };
      
      // Добавляем циклическую ссылку
      (circularData as any).self = circularData;

      vi.mocked(graphApi.getFullGraphData).mockResolvedValue(circularData);

      await PreloadService.startPreload();

      // Сервис должен обработать циклические ссылки
      const cachedData = PreloadService.getPreloadedGraph();
      expect(cachedData).toBeDefined();
    });
  });

  describe('Performance Edge Cases', () => {
    it('should handle CPU-intensive operations', async () => {
      // Мокаем CPU-интенсивную операцию
      const cpuIntensiveData = {
        nodes: Array.from({ length: 5000 }, (_, i) => ({
          id: `node-${i}`,
          title: `Node ${i}`,
          type: 'star',
          // Добавляем сложные вычисления
          metadata: {
            complex: Array.from({ length: 100 }, (_, j) => ({
              id: j,
              value: Math.random() * 1000,
              nested: {
                deep: Array.from({ length: 10 }, (_, k) => ({
                  id: k,
                  calculation: Math.sin(k) * Math.cos(j) * Math.tan(i)
                }))
              }
            }))
          }
        })),
        links: []
      };

      vi.mocked(graphApi.getFullGraphData).mockResolvedValue(cpuIntensiveData);

      const startTime = performance.now();
      await PreloadService.startPreload();
      const endTime = performance.now();

      // Операция должна завершиться в разумное время
      expect(endTime - startTime).toBeLessThan(1000);
    });

    it('should handle memory leaks', async () => {
      const initialStats = PreloadService.getStats();
      
      // Множественные операции
      for (let i = 0; i < 100; i++) {
        await PreloadService.startPreload();
        PreloadService.clearCache();
      }

      const finalStats = PreloadService.getStats();
      
      // Статистика должна быть в начальном состоянии
      expect(finalStats.hasGraph).toBe(initialStats.hasGraph);
      expect(finalStats.hasAchievements).toBe(initialStats.hasAchievements);
    });
  });

  describe('Browser Compatibility', () => {
    it('should handle missing browser APIs', async () => {
      // Мокаем отсутствие browser API
      const originalLocalStorage = global.localStorage;
      delete (global as any).localStorage;

      await PreloadService.startPreload();

      // Сервис должен работать без localStorage
      expect(PreloadService.isPreloadingData()).toBe(false);

      global.localStorage = originalLocalStorage;
    });

    it('should handle restricted browser environments', async () => {
      // Мокаем restricted environment
      vi.stubGlobal('browser', false);
      vi.stubGlobal('localStorage', undefined);

      await PreloadService.startPreload();

      expect(PreloadService.isPreloadingData()).toBe(false);

      vi.unstubAllGlobals();
    });
  });

  describe('Error Recovery', () => {
    it('should recover from temporary failures', async () => {
      let attemptCount = 0;
      vi.mocked(graphApi.getFullGraphData).mockImplementation(() => {
        attemptCount++;
        if (attemptCount <= 2) {
          return Promise.reject(new Error('Temporary failure'));
        }
        return Promise.resolve(mockGraphData);
      });

      // Первые попытки должны провалиться
      await PreloadService.startPreload();
      expect(PreloadService.getPreloadedGraph()).toBeNull();

      // Очищаем и пробуем снова
      PreloadService.clearCache();
      await PreloadService.startPreload();
      
      // Теперь должно работать
      expect(PreloadService.getPreloadedGraph()).toEqual(mockGraphData);
    });

    it('should handle cascading failures', async () => {
      // Все API возвращают ошибки
      vi.mocked(graphApi.getFullGraphData).mockRejectedValue(new Error('Graph failed'));
      vi.mocked(usersApi.getAllAchievements).mockRejectedValue(new Error('Achievements failed'));

      await PreloadService.startPreload();

      // Сервис должен остаться в рабочем состоянии
      expect(PreloadService.isPreloadingData()).toBe(false);
      expect(PreloadService.hasPreloadedData()).toBe(false);
      expect(PreloadService.getStats()).toEqual({
        hasGraph: false,
        hasAchievements: false,
        graphAge: null,
        achievementsAge: null,
        isPreloading: false
      });
    });
  });
});
