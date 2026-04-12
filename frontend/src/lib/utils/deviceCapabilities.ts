// Utility to detect device capabilities and optimize 3D rendering accordingly
// This module caches results to avoid repeated WebGL context creation

export interface DeviceCapabilities {
  /** Whether the device is a mobile device */
  isMobile: boolean;
  /** Whether the device should use low-power mode */
  isLowPower: boolean;
  /** GPU performance tier */
  gpuTier: 'high' | 'medium' | 'low';
  /** Maximum recommended nodes for this device */
  maxNodes: number;
  /** Number of stars in background */
  starCount: number;
  /** Whether to enable particle effects */
  enableParticles: boolean;
  /** Whether to enable glow effects */
  enableGlow: boolean;
  /** Recommended pixel ratio for performance */
  pixelRatio: number;
}

/** Cached capabilities to avoid repeated detection */
let cachedCapabilities: DeviceCapabilities | null = null;

/** Tier configuration with documented reasoning */
const TIER_CONFIG = {
  low: {
    maxNodes: 30,
    starCount: 100,
    enableParticles: false,
    enableGlow: false,
    reason: 'Mobile devices, integrated GPUs, or battery saving'
  },
  medium: {
    maxNodes: 50,
    starCount: 500,
    enableParticles: true,
    enableGlow: true,
    reason: 'Modern integrated GPUs (Intel Xe, Apple M1+)'
  },
  high: {
    maxNodes: 100,
    starCount: 1000,
    enableParticles: true,
    enableGlow: true,
    reason: 'Discrete GPUs (NVIDIA, AMD)'
  }
} as const;

export function detectDeviceCapabilities(): DeviceCapabilities {
  // Return cached result if available
  if (cachedCapabilities) {
    return cachedCapabilities;
  }

  if (typeof window === 'undefined') {
    // SSR fallback - return conservative settings
    return {
      isMobile: false,
      isLowPower: true,
      gpuTier: 'low',
      maxNodes: 50,
      starCount: 200,
      enableParticles: false,
      enableGlow: false,
      pixelRatio: 1
    };
  }

  // Check for mobile device
  const isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
    navigator.userAgent
  ) || window.innerWidth < 768;

  // Check for touch device
  const _isTouch = 'ontouchstart' in window || navigator.maxTouchPoints > 0;

  // Check hardware concurrency (CPU cores)
  const cpuCores = navigator.hardwareConcurrency || 2;

  // Check device memory (if available)
  const deviceMemory = (navigator as any).deviceMemory || 4;

  // Try to detect GPU tier
  let gpuTier: 'high' | 'medium' | 'low' = 'medium';

  const canvas = document.createElement('canvas');
  const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');

  if (gl) {
    const debugInfo = (gl as any).getExtension('WEBGL_debug_renderer_info');
    if (debugInfo) {
      const renderer = (gl as any).getParameter(debugInfo.UNMASKED_RENDERER_WEBGL);
      const _vendor = (gl as any).getParameter(debugInfo.UNMASKED_VENDOR_WEBGL);

      // If vendor indicates generic or Microsoft software renderer, mark as low
      if (_vendor && /microsoft/i.test(_vendor)) {
        gpuTier = 'low';
      }

      // Check for software renderer or low-end GPUs
      const lowEndGPUs = ['swiftshader', 'llvmpipe', 'software', 'microsoft basic render'];
      const isSoftwareRenderer = lowEndGPUs.some(gpu =>
        renderer.toLowerCase().includes(gpu)
      );

      if (isSoftwareRenderer) {
        gpuTier = 'low';
      } else if (renderer.toLowerCase().includes('intel')) {
        // Check for modern Intel GPUs (Xe, Arc)
        const modernIntel = ['intel(r) uhd graphics 7', 'intel(r) iris', 'intel(r) xe', 'intel(r) arc'];
        const isModernIntel = modernIntel.some(g => renderer.toLowerCase().includes(g.toLowerCase()));
        gpuTier = isModernIntel ? 'medium' : 'low';
      } else if (isMobile) {
        gpuTier = 'medium';
      } else {
        gpuTier = 'high';
      }
    }

    // Check max texture size as GPU capability indicator
    const maxTextureSize = (gl as any).getParameter((gl as any).MAX_TEXTURE_SIZE);
    if (maxTextureSize < 4096) {
      gpuTier = 'low';
    }
  }

  // Determine tier config and low power mode
  const tierConfig = TIER_CONFIG[gpuTier];
  const isLowPower = isMobile || _isTouch || cpuCores <= 4 || deviceMemory <= 4 || gpuTier === 'low';

  // Build result
  const result: DeviceCapabilities = {
    isMobile,
    isLowPower,
    gpuTier,
    maxNodes: tierConfig.maxNodes,
    starCount: tierConfig.starCount,
    enableParticles: tierConfig.enableParticles,
    enableGlow: tierConfig.enableGlow,
    pixelRatio: Math.min(window.devicePixelRatio, gpuTier === 'high' ? 2 : gpuTier === 'medium' ? 1.5 : 1)
  };

  // Cache and return
  cachedCapabilities = result;
  return result;
}

/** Clear the cached capabilities (useful for testing) */
export function clearDeviceCapabilitiesCache(): void {
  cachedCapabilities = null;
}

export function shouldUse3D(capabilities: DeviceCapabilities): boolean {
  // Use 2D fallback for very low-end devices
  if (capabilities.gpuTier === 'low' && capabilities.isMobile) {
    return false;
  }
  return true;
}
