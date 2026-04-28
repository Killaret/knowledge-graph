/**
 * Scene setup module - wraps $lib/three/core/sceneSetup
 */
import * as THREE from 'three';
import type { CSS2DRenderer } from 'three/examples/jsm/renderers/CSS2DRenderer.js';
import type { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import { initScene, setFogDensity } from '$lib/three/core/sceneSetup';
import type { ObjectManager } from '$lib/three/rendering/objectManager';

export { initScene, setFogDensity };

/**
 * Update fog density with animation
 */
export function updateFog(
  scene: THREE.Scene,
  targetDensity: number,
  duration: number = 1000,
  onUpdate?: (density: number) => void
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
    onUpdate?.(currentDensity);

    if (progress < 1) {
      animationId = requestAnimationFrame(animate);
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

/**
 * Clear scene from all objects
 */
export function clearScene(scene: THREE.Scene): void {
  if (!scene) return;

  // Remove all objects except camera (which is not in scene usually)
  while (scene.children.length > 0) {
    const object = scene.children[0] as THREE.Mesh;
    scene.remove(object);

    // Dispose geometries and materials
    if (object.geometry) {
      object.geometry.dispose();
    }
    if (object.material) {
      if (Array.isArray(object.material)) {
        object.material.forEach((m: THREE.Material) => m.dispose());
      } else {
        (object.material as THREE.Material).dispose();
      }
    }
  }
}

/**
 * Resize renderer and camera for container
 */
export function resizeScene(
  container: HTMLElement,
  camera: THREE.PerspectiveCamera,
  renderer: THREE.WebGLRenderer,
  labelRenderer?: CSS2DRenderer
): void {
  if (!container || !camera || !renderer) return;

  const width = container.clientWidth;
  const height = container.clientHeight;

  camera.aspect = width / height;
  camera.updateProjectionMatrix();

  renderer.setSize(width, height);
  renderer.setPixelRatio(window.devicePixelRatio);

  if (labelRenderer) {
    labelRenderer.setSize(width, height);
  }
}

/**
 * Dispose all Three.js resources
 */
export function disposeScene(
  renderer: THREE.WebGLRenderer,
  controls: OrbitControls,
  objectManager: ObjectManager
): void {
  if (controls) {
    controls.dispose();
  }
  if (objectManager) {
    objectManager.clear();
  }
  if (renderer) {
    renderer.dispose();
  }
}
