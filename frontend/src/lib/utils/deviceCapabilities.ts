// Utility to detect device capabilities and optimize 3D rendering accordingly

export interface DeviceCapabilities {
  isMobile: boolean;
  isLowPower: boolean;
  gpuTier: 'high' | 'medium' | 'low';
  maxNodes: number;
  starCount: number;
  enableParticles: boolean;
  enableGlow: boolean;
  pixelRatio: number;
}

export function detectDeviceCapabilities(): DeviceCapabilities {
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

      // Check for software renderer or low-end GPUs
      const lowEndGPUs = ['swiftshader', 'llvmpipe', 'software', 'microsoft basic render'];
      const isSoftwareRenderer = lowEndGPUs.some(gpu =>
        renderer.toLowerCase().includes(gpu)
      );

      if (isSoftwareRenderer) {
        gpuTier = 'low';
      } else if (renderer.toLowerCase().includes('intel')) {
        gpuTier = 'low';
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

  // Determine if low power mode
  const isLowPower = isMobile || cpuCores <= 4 || deviceMemory <= 4 || gpuTier === 'low';

  // Set optimization parameters based on capabilities
  if (isLowPower || gpuTier === 'low') {
    return {
      isMobile,
      isLowPower: true,
      gpuTier,
      maxNodes: 30,
      starCount: 100,
      enableParticles: false,
      enableGlow: false,
      pixelRatio: Math.min(window.devicePixelRatio, 1)
    };
  }

  if (gpuTier === 'medium' || isMobile) {
    return {
      isMobile,
      isLowPower: isMobile,
      gpuTier,
      maxNodes: 50,
      starCount: 500,
      enableParticles: true,
      enableGlow: true,
      pixelRatio: Math.min(window.devicePixelRatio, 1.5)
    };
  }

  // High-end desktop
  return {
    isMobile: false,
    isLowPower: false,
    gpuTier: 'high',
    maxNodes: 100,
    starCount: 1000,
    enableParticles: true,
    enableGlow: true,
    pixelRatio: Math.min(window.devicePixelRatio, 2)
  };
}

export function shouldUse3D(capabilities: DeviceCapabilities): boolean {
  // Use 2D fallback for very low-end devices
  if (capabilities.gpuTier === 'low' && capabilities.isMobile) {
    return false;
  }
  return true;
}
