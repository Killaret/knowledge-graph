import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import {
  detectDeviceCapabilities,
  shouldUse3D,
  type DeviceCapabilities,
} from "./deviceCapabilities";

describe("deviceCapabilities", () => {
  // Тест: SSR fallback (typeof window === 'undefined')
  it("should return SSR fallback when window is undefined", () => {
    const originalWindow = global.window;
    // @ts-expect-error - удаляем window для теста SSR
    global.window = undefined;

    const capabilities = detectDeviceCapabilities();

    expect(capabilities.isMobile).toBe(false);
    expect(capabilities.isLowPower).toBe(true);
    expect(capabilities.gpuTier).toBe("low");
    expect(capabilities.maxNodes).toBe(50);
    expect(capabilities.starCount).toBe(200);
    expect(capabilities.enableParticles).toBe(false);
    expect(capabilities.enableGlow).toBe(false);
    expect(capabilities.pixelRatio).toBe(1);

    // Восстанавливаем window
    global.window = originalWindow;
  });

  // Тест: shouldUse3D для low-end мобильных устройств
  it("shouldUse3D should return false for low-end mobile devices", () => {
    const capabilities: DeviceCapabilities = {
      isMobile: true,
      isLowPower: true,
      gpuTier: "low",
      maxNodes: 30,
      starCount: 100,
      enableParticles: false,
      enableGlow: false,
      pixelRatio: 1,
    };

    expect(shouldUse3D(capabilities)).toBe(false);
  });

  // Тест: shouldUse3D для high-end устройств
  it("shouldUse3D should return true for high-end devices", () => {
    const capabilities: DeviceCapabilities = {
      isMobile: false,
      isLowPower: false,
      gpuTier: "high",
      maxNodes: 100,
      starCount: 1000,
      enableParticles: true,
      enableGlow: true,
      pixelRatio: 2,
    };

    expect(shouldUse3D(capabilities)).toBe(true);
  });

  // Тест: shouldUse3D для medium-tier мобильных устройств
  it("shouldUse3D should return true for medium-tier mobile devices", () => {
    const capabilities: DeviceCapabilities = {
      isMobile: true,
      isLowPower: true,
      gpuTier: "medium",
      maxNodes: 50,
      starCount: 500,
      enableParticles: true,
      enableGlow: true,
      pixelRatio: 1.5,
    };

    expect(shouldUse3D(capabilities)).toBe(true);
  });

  // Тест: shouldUse3D для десктопа с low GPU
  it("shouldUse3D should return true for low-end desktop", () => {
    const capabilities: DeviceCapabilities = {
      isMobile: false,
      isLowPower: true,
      gpuTier: "low",
      maxNodes: 30,
      starCount: 100,
      enableParticles: false,
      enableGlow: false,
      pixelRatio: 1,
    };

    // Desktop с low GPU все еще может использовать 3D
    expect(shouldUse3D(capabilities)).toBe(true);
  });

  // Тест: проверка типов интерфейса DeviceCapabilities
  it("DeviceCapabilities interface should have all required fields", () => {
    const capabilities: DeviceCapabilities = {
      isMobile: false,
      isLowPower: false,
      gpuTier: "high",
      maxNodes: 100,
      starCount: 1000,
      enableParticles: true,
      enableGlow: true,
      pixelRatio: 2,
    };

    expect(capabilities.isMobile).toBe(false);
    expect(capabilities.gpuTier).toBe("high");
    expect(capabilities.maxNodes).toBe(100);
    expect(capabilities.starCount).toBe(1000);
  });
});

describe("detectDeviceCapabilities browser path", () => {
  const originalNavigator = global.navigator;

  beforeEach(() => {
    vi.stubGlobal("navigator", {
      ...originalNavigator,
      userAgent: "desktop",
      hardwareConcurrency: 8,
      deviceMemory: 8,
    });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  function mockWebGL(renderer: string, maxTextureSize: number) {
    const debugInfo = { UNMASKED_RENDERER_WEBGL: 0x9245 } as unknown as WEBGL_debug_renderer_info;
    const gl = {
      MAX_TEXTURE_SIZE: 0x0d33,
      getExtension: vi.fn().mockReturnValue(debugInfo),
      getParameter: vi.fn((p: number) => {
        if (p === 0x9245) return renderer;
        if (p === 0x0d33) return maxTextureSize; // MAX_TEXTURE_SIZE
        return 0;
      }),
    } as unknown as WebGLRenderingContext;

    vi.stubGlobal(
      "document",
      {
        createElement: vi.fn().mockReturnValue({
          getContext: vi.fn((type: string) => (type === "webgl" ? gl : null)),
        }),
      }
    );
  }

  it("detects high-end desktop GPU", () => {
    mockWebGL("NVIDIA GeForce RTX 4090", 16384);
    vi.stubGlobal("window", { ...window, innerWidth: 1920, devicePixelRatio: 2 });

    const caps = detectDeviceCapabilities();
    expect(caps.gpuTier).toBe("high");
    expect(caps.isLowPower).toBe(false);
    expect(caps.maxNodes).toBe(100);
    expect(caps.pixelRatio).toBe(2);
  });

  it("downgrades to low for a software renderer", () => {
    mockWebGL("llvmpipe", 16384);
    vi.stubGlobal("window", { ...window, innerWidth: 1920, devicePixelRatio: 1 });

    const caps = detectDeviceCapabilities();
    expect(caps.gpuTier).toBe("low");
    expect(caps.isLowPower).toBe(true);
    expect(caps.maxNodes).toBe(30);
  });

  it("downgrades to low for Intel GPU", () => {
    mockWebGL("Intel Iris Xe", 16384);
    vi.stubGlobal("window", { ...window, innerWidth: 1920, devicePixelRatio: 1 });

    const caps = detectDeviceCapabilities();
    expect(caps.gpuTier).toBe("low");
    expect(caps.isLowPower).toBe(true);
    expect(caps.maxNodes).toBe(30);
  });

  it("downgrades to low for small max texture size", () => {
    mockWebGL("NVIDIA GeForce", 2048);
    vi.stubGlobal("window", { ...window, innerWidth: 1920, devicePixelRatio: 1 });

    const caps = detectDeviceCapabilities();
    expect(caps.gpuTier).toBe("low");
    expect(caps.isLowPower).toBe(true);
  });

  it("uses medium tier on mobile", () => {
    mockWebGL("Mali-G78", 8192);
    vi.stubGlobal("navigator", {
      ...originalNavigator,
      userAgent: "Android 14",
      hardwareConcurrency: 8,
      deviceMemory: 8,
    });
    vi.stubGlobal("window", { ...window, innerWidth: 400, devicePixelRatio: 2.5 });

    const caps = detectDeviceCapabilities();
    expect(caps.isMobile).toBe(true);
    expect(caps.gpuTier).toBe("medium");
    expect(caps.isLowPower).toBe(true);
    expect(caps.maxNodes).toBe(30);
    expect(caps.pixelRatio).toBe(1);
  });

  it("falls back to medium when WebGL is unavailable", () => {
    vi.stubGlobal(
      "document",
      {
        createElement: vi.fn().mockReturnValue({
          getContext: vi.fn().mockReturnValue(null),
        }),
      }
    );
    vi.stubGlobal("window", { ...window, innerWidth: 1920, devicePixelRatio: 1 });

    const caps = detectDeviceCapabilities();
    expect(caps.gpuTier).toBe("medium");
  });

  it("detects mobile by viewport width", () => {
    vi.stubGlobal("navigator", {
      ...originalNavigator,
      userAgent: "Mozilla/5.0",
      hardwareConcurrency: 8,
      deviceMemory: 8,
    });
    vi.stubGlobal("window", { ...window, innerWidth: 400, devicePixelRatio: 1 });
    mockWebGL("NVIDIA GeForce RTX 4090", 16384);

    const caps = detectDeviceCapabilities();
    expect(caps.isMobile).toBe(true);
    expect(caps.gpuTier).toBe("medium");
    expect(caps.isLowPower).toBe(true);
    expect(caps.maxNodes).toBe(30);
    expect(caps.pixelRatio).toBe(1);
  });
});
