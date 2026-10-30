/**
 * Fog manager for progressive fog animation
 */
import * as THREE from 'three';
import { setFogDensity } from '$lib/three/core/sceneSetup';

export interface FogAnimationState {
  animationId: number | null;
  isRunning: boolean;
}

/**
 * Animate fog density from current to target
 */
export function animateFogDensity(
  scene: THREE.Scene,
  targetDensity: number,
  duration: number = 1000,
  state?: FogAnimationState
): { stop: () => void } {
  const startDensity = (scene?.fog as THREE.FogExp2 | null)?.density ?? 0;
  const startTime = performance.now();
  let animationId: number | null = null;

  function animate() {
    const elapsed = performance.now() - startTime;
    const progress = Math.min(elapsed / duration, 1);

    // Ease out quad
    const easeProgress = 1 - (1 - progress) * (1 - progress);
    const currentDensity = startDensity + (targetDensity - startDensity) * easeProgress;

    setFogDensity(scene, currentDensity);

    if (progress < 1) {
      animationId = requestAnimationFrame(animate);
    } else if (state) {
      state.isRunning = false;
    }
  }

  if (state) {
    state.isRunning = true;
  }

  animationId = requestAnimationFrame(animate);

  return {
    stop() {
      if (animationId !== null) {
        cancelAnimationFrame(animationId);
        animationId = null;
      }
      if (state) {
        state.isRunning = false;
      }
    }
  };
}

/**
 * Clear fog animation
 */
export function clearFogAnimation(state: FogAnimationState): void {
  if (state.animationId !== null) {
    cancelAnimationFrame(state.animationId);
    state.animationId = null;
  }
  state.isRunning = false;
}

/**
 * Progressive fog clear for exploration effect
 */
export function progressiveFogClear(
  scene: THREE.Scene,
  maxDistance: number,
  duration: number = 2000,
  onComplete?: () => void
): { stop: () => void } {
  const startTime = performance.now();
  let animationId: number | null = null;

  function animate() {
    const elapsed = performance.now() - startTime;
    const progress = Math.min(elapsed / duration, 1);

    // Ease out cubic
    const easeProgress = 1 - Math.pow(1 - progress, 3);
    const currentMaxDistance = maxDistance * easeProgress;

    // Update fog distance
    const fog = scene?.fog as THREE.Fog | null;
    if (fog) {
      fog.near = currentMaxDistance * 0.3;
      fog.far = currentMaxDistance;
    }

    if (progress < 1) {
      animationId = requestAnimationFrame(animate);
    } else {
      onComplete?.();
    }
  }

  animationId = requestAnimationFrame(animate);

  return {
    stop() {
      if (animationId !== null) {
        cancelAnimationFrame(animationId);
        animationId = null;
      }
    }
  };
}
