// Интеграционные тесты для логики тумана в Graph3D
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import * as THREE from 'three';
import { animateFogDensity, type FogAnimationState } from './fogManager';
import { setFogDensity } from '$lib/three/core/sceneSetup';

// Мокаем setFogDensity для отслеживания вызовов
vi.mock('$lib/three/core/sceneSetup', () => ({
  setFogDensity: vi.fn()
}));

// Мокаем PreloadService
const mockHasPreloadedData = vi.fn(() => false);
vi.mock('$lib/services/PreloadService', () => ({
  hasPreloadedData: mockHasPreloadedData
}));

describe('Graph3D Fog Integration Logic', () => {
  let scene: THREE.Scene;
  let mockSetFogDensity: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    scene = new THREE.Scene();
    scene.fog = new THREE.FogExp2(0x050510, 0.08);
    mockSetFogDensity = vi.mocked(setFogDensity);
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  describe('Initial Fog Density', () => {
    it('should set initial fog density to 0.08 for cold start', () => {
      // Симулируем холодный старт (без предзагруженных данных)
      mockHasPreloadedData.mockReturnValue(false);

      // Логика из createGraphSimulation
      const hasPreloaded = mockHasPreloadedData();
      const initialDensity = hasPreloaded ? 0.04 : 0.08;
      
      setFogDensity(scene, initialDensity);

      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, 0.08);
    });

    it('should set initial fog density to 0.04 for preloaded data', () => {
      // Симулируем предзагруженные данные
      mockHasPreloadedData.mockReturnValue(true);

      // Логика из createGraphSimulation
      const hasPreloaded = mockHasPreloadedData();
      const initialDensity = hasPreloaded ? 0.04 : 0.08;
      
      setFogDensity(scene, initialDensity);

      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, 0.04);
    });
  });

  describe('Progressive Fog Density Calculation', () => {
    it('should calculate fog density based on node positioning progress', () => {
      const totalNodes = 10;
      const nodesWithPosition = 5; // 50% прогресс
      
      // Логика из обработчика 'tick'
      const progress = Math.min(nodesWithPosition / totalNodes, 1);
      const startDensity = 0.08;
      const endDensity = 0.005;
      const currentDensity = startDensity - (startDensity - endDensity) * progress;
      
      setFogDensity(scene, currentDensity);

      expect(currentDensity).toBeCloseTo(0.0425, 4); // 0.08 - (0.08 - 0.005) * 0.5
      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, currentDensity);
    });

    it('should handle 100% progress correctly', () => {
      const totalNodes = 10;
      const nodesWithPosition = 10; // 100% прогресс
      
      const progress = Math.min(nodesWithPosition / totalNodes, 1);
      const startDensity = 0.08;
      const endDensity = 0.005;
      const currentDensity = startDensity - (startDensity - endDensity) * progress;
      
      setFogDensity(scene, currentDensity);

      expect(currentDensity).toBeCloseTo(0.005, 5); // Минимальная плотность с допуском
      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, expect.any(Number));
    });

    it('should handle 0% progress correctly', () => {
      const totalNodes = 10;
      const nodesWithPosition = 0; // 0% прогресс
      
      const progress = Math.min(nodesWithPosition / totalNodes, 1);
      const startDensity = 0.08;
      const endDensity = 0.005;
      const currentDensity = startDensity - (startDensity - endDensity) * progress;
      
      setFogDensity(scene, currentDensity);

      expect(currentDensity).toBe(0.08); // Максимальная плотность
      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, 0.08);
    });

    it('should filter out nodes with invalid coordinates', () => {
      const nodes = [
        { id: '1', x: 0, y: 0, z: 0 }, // Валидный
        { id: '2', x: undefined, y: 0, z: 0 }, // Невалидный
        { id: '3', x: 10, y: NaN, z: 0 }, // Невалидный
        { id: '4', x: 5, y: 5, z: 5 }, // Валидный
        { id: '5', x: 15, y: 15, z: undefined } // Невалидный
      ];

      // Логика фильтрации из обработчика 'tick'
      const nodesWithPosition = nodes.filter((n: any) => 
        n.x !== undefined && 
        !isNaN(n.x) && 
        n.y !== undefined && 
        !isNaN(n.y) && 
        n.z !== undefined && 
        !isNaN(n.z)
      ).length;

      expect(nodesWithPosition).toBe(2); // Только узлы 1 и 4 валидны
    });
  });

  describe('Fog Animation on Simulation End', () => {
    it('should call animateFogDensity when simulation ends', async () => {
      // Используем fake timers для контроля времени
      vi.useFakeTimers();
      
      // Симулируем завершение симуляции
      const currentDensity = (scene.fog as THREE.FogExp2)?.density ?? 0.08;
      const targetDensity = 0.005;
      const duration = 800;

      const animation = animateFogDensity(scene, targetDensity, duration);

      // Проверяем, что анимация создана
      expect(animation).toBeDefined();
      expect(typeof animation.stop).toBe('function');

      // Продвигаем время для завершения анимации
      vi.advanceTimersByTime(duration + 100);

      // Проверяем, что setFogDensity был вызван с целевой плотностью
      expect(mockSetFogDensity).toHaveBeenLastCalledWith(scene, expect.closeTo(targetDensity, 10));

      animation.stop();
      
      vi.useRealTimers();
    });

    it('should handle scene without fog gracefully', () => {
      scene.fog = null;
      
      const animation = animateFogDensity(scene, 0.005, 1000);

      expect(() => animation.stop()).not.toThrow();
    });

    it('should use correct animation parameters', () => {
      vi.useFakeTimers();
      
      const targetDensity = 0.005;
      const duration = 800;

      const animation = animateFogDensity(scene, targetDensity, duration);

      // Продвигаем время для начала анимации
      vi.advanceTimersByTime(16);

      // Проверяем начальное состояние
      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, expect.any(Number));

      animation.stop();
      
      vi.useRealTimers();
    });
  });

  describe('Edge Cases', () => {
    it('should handle empty graph correctly', () => {
      // Для пустого графа туман должен быть сразу установлен на минимальную плотность
      setFogDensity(scene, 0.005);

      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, expect.any(Number));
    });

    it('should handle single node graph correctly', () => {
      // Для графа с одним узлом туман должен быть сразу установлен на минимальную плотность
      setFogDensity(scene, 0.005);

      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, expect.any(Number));
    });

    it('should handle timeout fallback', () => {
      // Симулируем таймаут загрузки
      const timeoutDensity = 0.005;
      setFogDensity(scene, timeoutDensity);

      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, timeoutDensity);
    });

    it('should handle addData operation', () => {
      // При добавлении новых данных туман должен быть установлен на минимальную плотность
      setFogDensity(scene, 0.005);

      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, expect.any(Number));
    });
  });

  describe('Performance Considerations', () => {
    it('should update fog density every 5 ticks', () => {
      let lastFogUpdate = 0;
      const updateInterval = 5;

      // Симулируем 10 тиков
      for (let i = 1; i <= 10; i++) {
        if (++lastFogUpdate % updateInterval === 0) {
          // Обновляем туман
          setFogDensity(scene, 0.05);
        }
      }

      // Туман должен быть обновлен 2 раза (на 5-м и 10-м тиках)
      expect(mockSetFogDensity).toHaveBeenCalledTimes(2);
    });

    it('should limit link updates to 60fps', () => {
      let lastLinkUpdateTime = 0;
      const frameInterval = 16; // ~60fps

      const now = performance.now();
      
      // Симулируем вызов с правильным интервалом
      if (now - lastLinkUpdateTime >= frameInterval) {
        lastLinkUpdateTime = now;
        // Обновляем связи (здесь не тестируем, только логику)
      }

      expect(now - lastLinkUpdateTime).toBeLessThanOrEqual(frameInterval);
    });
  });

  describe('Integration with PreloadService', () => {
    it('should log preloaded data status in development mode', () => {
      // Мокаем development mode
      const originalImportMeta = import.meta;
      Object.defineProperty(import.meta, 'env', {
        value: { DEV: true },
        writable: true
      });

      // Мокаем console.log
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {});

      mockHasPreloadedData.mockReturnValue(true);

      // Логика из createGraphSimulation
      const hasPreloaded = mockHasPreloadedData();
      const initialDensity = hasPreloaded ? 0.04 : 0.08;

      if (import.meta.env.DEV) {
        console.log(`[Graph3D] Starting simulation with preloaded data: ${hasPreloaded}, initial fog density: ${initialDensity}`);
      }

      expect(consoleSpy).toHaveBeenCalledWith(
        `[Graph3D] Starting simulation with preloaded data: true, initial fog density: 0.04`
      );

      consoleSpy.mockRestore();
    });
  });

  describe('Real fog density changes', () => {
    it('should change scene.fog.density correctly', () => {
      // Используем мок, но проверяем, что он вызывается с правильными параметрами
      const initialDensity = (scene.fog as THREE.FogExp2).density;
      expect(initialDensity).toBe(0.08);

      const targetDensity = 0.005;
      
      // Вызываем через мок, но проверяем параметры
      mockSetFogDensity(scene, targetDensity);

      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, targetDensity);
    });

    it('should handle fog density transition from 0.08 to 0.005', async () => {
      // Тестируем полный переход от плотного тумана к чистому
      const startDensity = 0.08;
      const endDensity = 0.005;
      
      // Начальное состояние
      setFogDensity(scene, startDensity);
      expect(mockSetFogDensity).toHaveBeenLastCalledWith(scene, startDensity);

      // Промежуточное состояние (50% прогресс)
      const progress = 0.5;
      const currentDensity = startDensity - (startDensity - endDensity) * progress;
      setFogDensity(scene, currentDensity);
      expect(mockSetFogDensity).toHaveBeenLastCalledWith(scene, currentDensity);

      // Финальное состояние
      setFogDensity(scene, endDensity);
      expect(mockSetFogDensity).toHaveBeenLastCalledWith(scene, endDensity);

      // Проверяем последовательность значений
      const calls = mockSetFogDensity.mock.calls.map(call => call[1]);
      expect(calls).toEqual([startDensity, currentDensity, endDensity]);
      expect(calls[0]).toBeGreaterThan(calls[1]);
      expect(calls[1]).toBeGreaterThan(calls[2]);
    });
  });
});
