/**
 * Scene setup module - wraps $lib/three/core/sceneSetup
 */
import { initScene, setFogDensity } from '$lib/three/core/sceneSetup';

export { initScene, setFogDensity };

/**
 * Update fog density with animation
 */
export function updateFog(
  scene: any,
  targetDensity: number,
  duration: number = 1000,
  onUpdate?: (density: number) => void
): { stop: () => void } {
  const startDensity = scene?.fog?.density ?? 0;
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
export function clearScene(scene: any): void {
  if (!scene) return;

  // Remove all objects except camera (which is not in scene usually)
  while (scene.children.length > 0) {
    const object = scene.children[0];
    scene.remove(object);

    // Dispose geometries and materials
    if (object.geometry) {
      object.geometry.dispose();
    }
    if (object.material) {
      if (Array.isArray(object.material)) {
        object.material.forEach((m: { dispose: () => void }) => m.dispose());
      } else {
        object.material.dispose();
      }
    }
  }
}

/**
 * Resize renderer and camera for container
 */
export function resizeScene(
  container: HTMLElement,
  camera: any,
  renderer: any,
  labelRenderer: any
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
  renderer: any,
  controls: any,
  objectManager: any
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
