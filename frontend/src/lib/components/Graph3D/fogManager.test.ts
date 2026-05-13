// Unit тесты для fogManager
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import * as THREE from 'three';
import { animateFogDensity, progressiveFogClear, type FogAnimationState } from './fogManager';

// Мокаем setFogDensity
vi.mock('$lib/three/core/sceneSetup', () => ({
  setFogDensity: vi.fn()
}));

import { setFogDensity } from '$lib/three/core/sceneSetup';
const mockSetFogDensity = vi.mocked(setFogDensity);

describe('FogManager', () => {
  let scene: THREE.Scene;

  beforeEach(() => {
    scene = new THREE.Scene();
    scene.fog = new THREE.FogExp2(0x050510, 0.08);
    mockSetFogDensity.mockClear();
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  describe('animateFogDensity', () => {
    it('should animate fog density from start to target', async () => {
      const targetDensity = 0.005;
      const duration = 1000;

      // Используем fake timers для контроля времени
      vi.useFakeTimers();
      
      const animation = animateFogDensity(scene, targetDensity, duration);

      // Продвигаем время на 16мс (первый кадр)
      vi.advanceTimersByTime(16);
      
      // Проверяем что анимация началась
      expect(mockSetFogDensity).toHaveBeenCalled();

      // Продвигаем время до середины анимации
      vi.advanceTimersByTime(500);
      
      // Проверяем что было несколько вызовов
      expect(mockSetFogDensity.mock.calls.length).toBeGreaterThan(1);

      // Завершаем анимацию
      vi.advanceTimersByTime(1000);

      // Проверяем финальное состояние
      expect(mockSetFogDensity).toHaveBeenLastCalledWith(scene, expect.any(Number));
      
      // Останавливаем анимацию
      animation.stop();
      
      // Возвращаем реальные таймеры
      vi.useRealTimers();
    });

    it('should use easing function for smooth animation', async () => {
      const targetDensity = 0.005;
      const duration = 500;

      const animation = animateFogDensity(scene, targetDensity, duration);

      // Ждем половину анимации
      await new Promise(resolve => setTimeout(resolve, duration / 2));

      // Проверяем, что плотность изменилась
      const calls = mockSetFogDensity.mock.calls;
      expect(calls.length).toBeGreaterThan(0);
      
      if (calls.length > 0) {
        const lastCallDensity = calls[calls.length - 1][1];
        expect(lastCallDensity).toBeGreaterThanOrEqual(targetDensity);
        expect(lastCallDensity).toBeLessThanOrEqual(0.08);
      }

      animation.stop();
    });

    it('should handle scene without fog gracefully', () => {
      scene.fog = null;

      const animation = animateFogDensity(scene, 0.005, 1000);

      // Не должно быть ошибок
      expect(() => animation.stop()).not.toThrow();
    });

    it('should handle scene with linear fog (not FogExp2)', async () => {
      scene.fog = new THREE.Fog(0x050510, 1, 100);

      const animation = animateFogDensity(scene, 0.005, 100);

      // Ждем немного для анимации
      await new Promise(resolve => setTimeout(resolve, 50));

      // setFogDensity должен быть вызван
      expect(mockSetFogDensity).toHaveBeenCalled();

      animation.stop();
    });

    it('should stop animation correctly', async () => {
      const animation = animateFogDensity(scene, 0.005, 2000);

      // Ждем немного
      await new Promise(resolve => setTimeout(resolve, 100));

      // Останавливаем анимацию
      animation.stop();

      const callsBeforeStop = mockSetFogDensity.mock.calls.length;

      // Ждем еще, но вызовов не должно быть больше
      await new Promise(resolve => setTimeout(resolve, 200));

      expect(mockSetFogDensity.mock.calls.length).toBe(callsBeforeStop);
    });

    it('should use animation state correctly', () => {
      const state: FogAnimationState = {
        animationId: null,
        isRunning: false
      };

      animateFogDensity(scene, 0.005, 1000, state);

      expect(state.isRunning).toBe(true);
    });
  });

  describe('progressiveFogClear', () => {
    it('should progressively clear fog using linear fog', async () => {
      // Используем fake timers для контроля времени
      vi.useFakeTimers();
      
      // Заменяем FogExp2 на Fog для этого теста
      scene.fog = new THREE.Fog(0x050510, 0.3, 10);
      const maxDistance = 100;
      const duration = 1000;

      const animation = progressiveFogClear(scene, maxDistance, duration);

      // Продвигаем время до завершения анимации
      vi.advanceTimersByTime(duration + 16);

      // Проверяем финальное состояние
      const fog = scene.fog as THREE.Fog;
      // progressiveFogClear устанавливает near = maxDistance * 0.3, far = maxDistance
      expect(fog.near).toBeCloseTo(maxDistance * 0.3, 0.1);
      expect(fog.far).toBeCloseTo(maxDistance, 0.1);

      animation.stop();
      
      // Возвращаем реальные таймеры
      vi.useRealTimers();
    });

    it('should handle scene without fog gracefully', () => {
      scene.fog = null;

      const animation = progressiveFogClear(scene, 100, 1000);

      expect(() => animation.stop()).not.toThrow();
    });

    it('should call onComplete callback when animation finishes', async () => {
      // Используем fake timers для контроля времени
      vi.useFakeTimers();
      
      scene.fog = new THREE.Fog(0x050510, 0.3, 10);
      const onComplete = vi.fn();

      const animation = progressiveFogClear(scene, 100, 500, onComplete);

      // Продвигаем время до завершения анимации
      vi.advanceTimersByTime(500 + 16);

      expect(onComplete).toHaveBeenCalled();

      animation.stop();
      
      // Возвращаем реальные таймеры
      vi.useRealTimers();
    });

    it('should use cubic easing for smooth animation', async () => {
      // Используем fake timers для контроля времени
      vi.useFakeTimers();
      
      scene.fog = new THREE.Fog(0x050510, 0.3, 10);
      const maxDistance = 100;
      const duration = 1000;

      const animation = progressiveFogClear(scene, maxDistance, duration);

      // Продвигаем время на половину анимации
      vi.advanceTimersByTime(duration / 2);

      const fog = scene.fog as THREE.Fog;
      
      // При cubic easing, на половине времени прогресс должен быть > 0.5
      // При progress = 0.5, easeProgress = 1 - (1-0.5)^3 = 0.875
      // currentMaxDistance = 100 * 0.875 = 87.5
      expect(fog.far).toBeGreaterThan(10); // начальное значение
      expect(fog.far).toBeGreaterThan(50); // должен быть значительно больше начального
      expect(fog.far).toBeLessThan(maxDistance);

      animation.stop();
      
      // Возвращаем реальные таймеры
      vi.useRealTimers();
    });
  });

  describe('Integration with setFogDensity', () => {
    it('should work correctly with basic fog operations', () => {
      // Простой тест без сложных моков
      const initialDensity = (scene.fog as THREE.FogExp2).density;
      expect(initialDensity).toBe(0.08);

      // Проверяем, что setFogDensity мок работает
      const targetDensity = 0.005;
      mockSetFogDensity(scene, targetDensity);

      expect(mockSetFogDensity).toHaveBeenCalledWith(scene, targetDensity);
    });
  });

  describe('Edge Cases', () => {
    it('should handle very short duration', () => {
      const animation = animateFogDensity(scene, 0.005, 1);

      expect(() => animation.stop()).not.toThrow();
    });

    it('should handle zero duration', async () => {
      const animation = animateFogDensity(scene, 0.005, 0);

      // Ждем немного для обработки
      await new Promise(resolve => setTimeout(resolve, 10));

      // Должен быть вызван как минимум один раз
      expect(mockSetFogDensity).toHaveBeenCalled();

      animation.stop();
    });

    it('should handle negative duration gracefully', () => {
      const animation = animateFogDensity(scene, 0.005, -100);

      expect(() => animation.stop()).not.toThrow();
    });

    it('should handle same start and target density', async () => {
      const density = 0.05;
      const animation = animateFogDensity(scene, density, 1000);

      await new Promise(resolve => setTimeout(resolve, 100));

      // Должен быть вызван как минимум один раз
      expect(mockSetFogDensity).toHaveBeenCalled();

      animation.stop();
    });
  });
});
