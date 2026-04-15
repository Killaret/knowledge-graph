import * as THREE from 'three';
import type { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';

/**
 * Animate camera position and target smoothly using lerp
 * @param camera - The camera to animate
 * @param controls - The orbit controls
 * @param targetPos - Target camera position
 * @param targetCenter - Target look-at point
 * @param duration - Animation duration in ms
 */
export function lerpCamera(
  camera: THREE.PerspectiveCamera,
  controls: OrbitControls,
  targetPos: THREE.Vector3,
  targetCenter: THREE.Vector3,
  duration: number = 1000
): void {
  const startPos = camera.position.clone();
  const startCenter = controls.target.clone();
  const startTime = performance.now();

  function animate() {
    const elapsed = performance.now() - startTime;
    const progress = Math.min(elapsed / duration, 1);
    
    // Ease out cubic
    const ease = 1 - Math.pow(1 - progress, 3);
    
    camera.position.lerpVectors(startPos, targetPos, ease);
    controls.target.lerpVectors(startCenter, targetCenter, ease);
    controls.update();
    
    if (progress < 1) {
      requestAnimationFrame(animate);
    }
  }
  
  animate();
}

export function autoZoomToFit(
  nodes: { x: number; y: number; z: number }[],
  camera: THREE.PerspectiveCamera,
  controls: OrbitControls,
  animate: boolean = false
) {
  if (nodes.length === 0) return;

  const center = new THREE.Vector3();
  nodes.forEach((n) => center.add(new THREE.Vector3(n.x, n.y, n.z)));
  center.divideScalar(nodes.length);

  let maxDist = 0;
  nodes.forEach((n) => {
    const dist = new THREE.Vector3(n.x, n.y, n.z).distanceTo(center);
    if (dist > maxDist) maxDist = dist;
  });

  const radius = maxDist * 1.2;
  const fov = camera.fov * Math.PI / 180;
  const dist = radius / Math.tan(fov / 2);

  console.log('[autoZoomToFit] center:', `(${center.x.toFixed(2)}, ${center.y.toFixed(2)}, ${center.z.toFixed(2)})`, 'radius:', radius.toFixed(2), 'distance:', dist.toFixed(2), 'nodes:', nodes.length, 'animate:', animate);

  const direction = new THREE.Vector3(1, 1, 1).normalize();
  const newPos = center.clone().add(direction.multiplyScalar(dist));
  
  if (animate) {
    lerpCamera(camera, controls, newPos, center, 1500);
  } else {
    camera.position.copy(newPos);
    controls.target.copy(center);
    controls.update();
  }
  
  console.log('[autoZoomToFit] camera position set to:', `(${newPos.x.toFixed(2)}, ${newPos.y.toFixed(2)}, ${newPos.z.toFixed(2)})`);
}
