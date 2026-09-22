/**
 * Animation loop management for Graph3D
 */
import * as THREE from 'three';
import type { CSS2DRenderer } from 'three/examples/jsm/renderers/CSS2DRenderer.js';
import type { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';

export interface AnimationState {
  animationId: number | null;
  isRunning: boolean;
  lastFrameTime: number;
}

export interface AnimationCallbacks {
  onBeforeRender?: () => void;
  onRender?: () => void;
  onAfterRender?: () => void;
}

/**
 * Start animation loop
 */
export function startAnimationLoop(
  renderer: THREE.WebGLRenderer,
  scene: THREE.Scene,
  camera: THREE.PerspectiveCamera,
  labelRenderer: CSS2DRenderer | null,
  controls: OrbitControls | null,
  callbacks?: AnimationCallbacks,
  state?: AnimationState
): { stop: () => void } {
  if (state) {
    state.isRunning = true;
    state.lastFrameTime = performance.now();
  }

  let animationId: number;

  function animate() {
    if (state && !state.isRunning) return;

    callbacks?.onBeforeRender?.();

    controls?.update();
    renderer?.render(scene, camera);
    labelRenderer?.render(scene, camera);

    callbacks?.onRender?.();
    callbacks?.onAfterRender?.();

    if (state) {
      state.lastFrameTime = performance.now();
    }

    animationId = requestAnimationFrame(animate);
  }

  animationId = requestAnimationFrame(animate);

  return {
    stop() {
      if (state) {
        state.isRunning = false;
      }
      cancelAnimationFrame(animationId);
    }
  };
}

/**
 * Stop animation loop
 */
export function stopAnimationLoop(state: AnimationState): void {
  state.isRunning = false;
  if (state.animationId !== null) {
    cancelAnimationFrame(state.animationId);
    state.animationId = null;
  }
}

/**
 * Pause animation (without canceling frame)
 */
export function pauseAnimation(state: AnimationState): void {
  state.isRunning = false;
}

/**
 * Resume animation
 */
export function resumeAnimation(state: AnimationState): void {
  state.isRunning = true;
}

/**
 * Check if animation is running
 */
export function isAnimationRunning(state: AnimationState): boolean {
  return state.isRunning;
}

/**
 * Create initial animation state
 */
export function createAnimationState(): AnimationState {
  return {
    animationId: null,
    isRunning: false,
    lastFrameTime: 0
  };
}
