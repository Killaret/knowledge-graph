/**
 * Camera controller module - wraps $lib/three/camera/cameraUtils
 */
import * as THREE from 'three';
import {
  autoZoomToFit,
  centerCameraOnNode
} from '$lib/three/camera/cameraUtils';

export {
  autoZoomToFit,
  centerCameraOnNode
};

export interface CameraState {
  autoRotate: boolean;
  isAnimating: boolean;
}

/**
 * Reset camera to default position
 */
export function resetCamera(
  camera: any,
  controls: any,
  defaultPosition: { x: number; y: number; z: number } = { x: 0, y: 0, z: 1000 }
): void {
  if (!camera || !controls) return;

  camera.position.set(defaultPosition.x, defaultPosition.y, defaultPosition.z);
  camera.lookAt(0, 0, 0);

  controls.target.set(0, 0, 0);
  controls.update();
}

/**
 * Toggle auto-rotation
 */
export function toggleAutoRotate(
  controls: any,
  state: CameraState
): boolean {
  if (!controls) return false;

  state.autoRotate = !state.autoRotate;
  controls.autoRotate = state.autoRotate;
  controls.update();

  return state.autoRotate;
}

/**
 * Set auto-rotate speed
 */
export function setAutoRotateSpeed(
  controls: any,
  speed: number
): void {
  if (!controls) return;
  controls.autoRotateSpeed = speed;
}

/**
 * Animate camera to position
 */
export function animateCameraTo(
  camera: any,
  controls: any,
  targetPosition: { x: number; y: number; z: number },
  duration: number = 1000,
  onComplete?: () => void
): { stop: () => void } {
  if (!camera || !controls) {
    onComplete?.();
    return { stop: () => {} };
  }

  const startPosition = camera.position.clone();
  const startTarget = controls.target.clone();
  const endTarget = new THREE.Vector3(targetPosition.x, targetPosition.y, targetPosition.z);
  const startTime = performance.now();
  let animationId: number | null = null;

  function animate() {
    const elapsed = performance.now() - startTime;
    const progress = Math.min(elapsed / duration, 1);

    // Ease in-out cubic
    const easeProgress = progress < 0.5
      ? 4 * progress * progress * progress
      : 1 - Math.pow(-2 * progress + 2, 3) / 2;

    camera.position.lerpVectors(startPosition, targetPosition, easeProgress);
    controls.target.lerpVectors(startTarget, endTarget, easeProgress);
    controls.update();

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

/**
 * Focus camera on bounding box of all nodes
 */
export function focusOnNodes(
  camera: any,
  controls: any,
  nodes: any[],
  padding: number = 1.2
): void {
  if (!camera || !controls || nodes.length === 0) return;

  // Calculate bounding box
  let minX = Infinity, minY = Infinity, minZ = Infinity;
  let maxX = -Infinity, maxY = -Infinity, maxZ = -Infinity;

  for (const node of nodes) {
    const x = node.x || 0;
    const y = node.y || 0;
    const z = node.z || 0;

    minX = Math.min(minX, x);
    minY = Math.min(minY, y);
    minZ = Math.min(minZ, z);
    maxX = Math.max(maxX, x);
    maxY = Math.max(maxY, y);
    maxZ = Math.max(maxZ, z);
  }

  const centerX = (minX + maxX) / 2;
  const centerY = (minY + maxY) / 2;
  const centerZ = (minZ + maxZ) / 2;

  const size = Math.max(maxX - minX, maxY - minY, maxZ - minZ);
  const distance = size * padding / Math.tan((camera.fov * Math.PI) / 360);

  controls.target.set(centerX, centerY, centerZ);
  camera.position.set(centerX, centerY, centerZ + distance);
  controls.update();
}
