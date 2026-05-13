// Unit тесты для PreloadService
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { browser } from '$app/environment';
import { isAuthenticated } from '$lib/stores/auth.svelte';
import { PreloadService } from './PreloadService';
import * as graphApi from '$lib/api/graph';
import * as usersApi from '$lib/api/users';
import {
  mockGraphData,
  mockAchievementsData,
  mockGraphError,
  mockAchievementsError,
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

describe('PreloadService', () => {
  let service: typeof PreloadService;

  beforeEach(() => {
    vi.clearAllMocks();
    
    // Полный сброс PreloadService через удаление статического свойства
    delete (PreloadService as any).instance;
    
    // Получаем новый инстанс сервиса
    service = PreloadService;
    
    // Принудительно очищаем все внутренние состояния
    (service as any).preloadedGraph = null;
    (service as any).preloadedAchievements = null;
    (service as any).isPreloading = false;
    (service as any).preloadPromise = null;
    
    // Мокаем значения по умолчанию - разрешаем предзагрузку
    vi.mocked(isAuthenticated).mockImplementation(() => false);
    vi.mocked(graphApi.getFullGraphData).mockResolvedValue(mockGraphData);
    vi.mocked(usersApi.getAllAchievements).mockResolvedValue(mockAchievementsData);
  });

  afterEach(() => {
    service.clearCache();
    vi.clearAllTimers();
  });

  describe('Basic Preloading', () => {
    it('should preload data when not authenticated', async () => {
      const result = await service.startPreload();
      
      expect(result).toBeUndefined();
      expect(graphApi.getFullGraphData).toHaveBeenCalled();
      expect(usersApi.getAllAchievements).toHaveBeenCalled();
    });

    it('should not preload when authenticated', async () => {
      vi.mocked(isAuthenticated).mockImplementation(() => true);
      
      const result = await service.startPreload();
      
      expect(result).toBeUndefined();
      expect(graphApi.getFullGraphData).not.toHaveBeenCalled();
      expect(usersApi.getAllAchievements).not.toHaveBeenCalled();
    });

    it('should handle preload errors gracefully', async () => {
      vi.mocked(graphApi.getFullGraphData).mockRejectedValue(mockGraphError);
      vi.mocked(usersApi.getAllAchievements).mockRejectedValue(mockAchievementsError);
      
      const result = await service.startPreload();
      
      expect(result).toBeUndefined();
      expect(service.getPreloadedGraph()).toBeNull();
      expect(service.getPreloadedAchievements()).toBeNull();
    });
  });

  describe('Cache Management', () => {
    it('should cache preloaded data', async () => {
      await service.startPreload();
      
      expect(service.getPreloadedGraph()).toEqual(mockGraphData);
      expect(service.getPreloadedAchievements()).toEqual(mockAchievementsData.achievements);
    });

    it('should clear cache', async () => {
      await service.startPreload();
      
      service.clearCache();
      
      expect(service.getPreloadedGraph()).toBeNull();
      expect(service.getPreloadedAchievements()).toBeNull();
    });

    it('should invalidate graph cache', async () => {
      await service.startPreload();
      
      service.invalidateGraphCache();
      
      expect(service.getPreloadedGraph()).toBeNull();
      expect(service.getPreloadedAchievements()).toEqual(mockAchievementsData.achievements);
    });

    it('should invalidate achievements cache', async () => {
      await service.startPreload();
      
      service.invalidateAchievementsCache();
      
      expect(service.getPreloadedGraph()).toEqual(mockGraphData);
      expect(service.getPreloadedAchievements()).toBeNull();
    });
  });

  describe('Status Methods', () => {
    it('should return correct preloading status', async () => {
      expect(service.isPreloadingData()).toBe(false);
      
      const performPreloadSpy = vi.spyOn(service as any, 'performPreload');
      performPreloadSpy.mockImplementation(() => new Promise(resolve => setTimeout(resolve, 100)));
      
      const promise = service.startPreload();
      expect(service.isPreloadingData()).toBe(true);
      
      await promise;
      expect(service.isPreloadingData()).toBe(false);
    });

    it('should return correct data availability status', async () => {
      vi.useFakeTimers();
      
      expect(service.hasPreloadedData()).toBe(false);
      
      await service.startPreload();
      
      // Запускаем все таймеры для завершения асинхронных операций
      vi.advanceTimersByTime(1000);
      
      expect(service.hasPreloadedData()).toBe(true);
      
      vi.useRealTimers();
    });

    it('should return correct stats', async () => {
      vi.useFakeTimers();
      
      const stats = service.getStats();
      
      expect(stats).toEqual({
        hasGraph: false,
        hasAchievements: false,
        graphAge: null,
        achievementsAge: null,
        isPreloading: false
      });
      
      await service.startPreload();
      
      // Запускаем все таймеры для завершения асинхронных операций
      vi.advanceTimersByTime(1000);
      
      const newStats = service.getStats();
      
      expect(newStats.hasGraph).toBe(true);
      expect(newStats.hasAchievements).toBe(true);
      expect(newStats.graphAge).toBeGreaterThanOrEqual(0);
      expect(newStats.achievementsAge).toBeGreaterThanOrEqual(0);
      expect(newStats.isPreloading).toBe(false);
      
      vi.useRealTimers();
    });
  });

  describe('Concurrent Access', () => {
    it('should handle concurrent preload requests', async () => {
      vi.useFakeTimers();
      
      const promises = [
        service.startPreload(),
        service.startPreload(),
        service.startPreload()
      ];
      
      await Promise.all(promises);
      
      // Запускаем все таймеры для завершения асинхронных операций
      vi.advanceTimersByTime(1000);
      
      // API должен вызваться только один раз
      expect(graphApi.getFullGraphData).toHaveBeenCalledTimes(1);
      expect(usersApi.getAllAchievements).toHaveBeenCalledTimes(1);
      
      vi.useRealTimers();
    });

    it('should handle concurrent cache access', async () => {
      vi.useFakeTimers();
      
      await service.startPreload();
      
      // Запускаем все таймеры для завершения асинхронных операций
      vi.advanceTimersByTime(1000);
      
      const promises = [
        Promise.resolve(service.getPreloadedGraph()),
        Promise.resolve(service.getPreloadedAchievements()),
        Promise.resolve(service.getPreloadedGraph())
      ];
      
      const results = await Promise.all(promises);
      
      expect(results[0]).toEqual(mockGraphData);
      expect(results[1]).toEqual(mockAchievementsData.achievements);
      expect(results[2]).toEqual(mockGraphData);
      
      vi.useRealTimers();
    });
  });
});
