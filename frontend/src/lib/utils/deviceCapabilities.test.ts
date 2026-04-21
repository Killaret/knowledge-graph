import { describe, it, expect } from 'vitest';
import {
  detectDeviceCapabilities,
  shouldUse3D,
  type DeviceCapabilities
} from './deviceCapabilities';

describe('deviceCapabilities', () => {
  // Тест: SSR fallback (typeof window === 'undefined')
  it('should return SSR fallback when window is undefined', () => {
    const originalWindow = global.window;
    // @ts-expect-error - удаляем window для теста SSR
    global.window = undefined;

    const capabilities = detectDeviceCapabilities();

    expect(capabilities.isMobile).toBe(false);
    expect(capabilities.isLowPower).toBe(true);
    expect(capabilities.gpuTier).toBe('low');
    expect(capabilities.maxNodes).toBe(50);
    expect(capabilities.starCount).toBe(200);
    expect(capabilities.enableParticles).toBe(false);
    expect(capabilities.enableGlow).toBe(false);
    expect(capabilities.pixelRatio).toBe(1);

    // Восстанавливаем window
    global.window = originalWindow;
  });

  // Тест: shouldUse3D для low-end мобильных устройств
  it('shouldUse3D should return false for low-end mobile devices', () => {
    const capabilities: DeviceCapabilities = {
      isMobile: true,
      isLowPower: true,
      gpuTier: 'low',
      maxNodes: 30,
      starCount: 100,
      enableParticles: false,
      enableGlow: false,
      pixelRatio: 1
    };

    expect(shouldUse3D(capabilities)).toBe(false);
  });

  // Тест: shouldUse3D для high-end устройств
  it('shouldUse3D should return true for high-end devices', () => {
    const capabilities: DeviceCapabilities = {
      isMobile: false,
      isLowPower: false,
      gpuTier: 'high',
      maxNodes: 100,
      starCount: 1000,
      enableParticles: true,
      enableGlow: true,
      pixelRatio: 2
    };

    expect(shouldUse3D(capabilities)).toBe(true);
  });

  // Тест: shouldUse3D для medium-tier мобильных устройств
  it('shouldUse3D should return true for medium-tier mobile devices', () => {
    const capabilities: DeviceCapabilities = {
      isMobile: true,
      isLowPower: true,
      gpuTier: 'medium',
      maxNodes: 50,
      starCount: 500,
      enableParticles: true,
      enableGlow: true,
      pixelRatio: 1.5
    };

    expect(shouldUse3D(capabilities)).toBe(true);
  });

  // Тест: shouldUse3D для десктопа с low GPU
  it('shouldUse3D should return true for low-end desktop', () => {
    const capabilities: DeviceCapabilities = {
      isMobile: false,
      isLowPower: true,
      gpuTier: 'low',
      maxNodes: 30,
      starCount: 100,
      enableParticles: false,
      enableGlow: false,
      pixelRatio: 1
    };

    // Desktop с low GPU все еще может использовать 3D
    expect(shouldUse3D(capabilities)).toBe(true);
  });

  // Тест: проверка типов интерфейса DeviceCapabilities
  it('DeviceCapabilities interface should have all required fields', () => {
    const capabilities: DeviceCapabilities = {
      isMobile: false,
      isLowPower: false,
      gpuTier: 'high',
      maxNodes: 100,
      starCount: 1000,
      enableParticles: true,
      enableGlow: true,
      pixelRatio: 2
    };

    expect(capabilities.isMobile).toBe(false);
    expect(capabilities.gpuTier).toBe('high');
    expect(capabilities.maxNodes).toBe(100);
    expect(capabilities.starCount).toBe(1000);
  });
});
