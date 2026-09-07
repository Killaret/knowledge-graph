import * as THREE from 'three';
import type { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';

/**
 * Center camera on a specific node (for local graph view)
 * @param nodeId - ID of the central node to focus on
 * @param nodes - All nodes in the graph
 * @param camera - The camera to position
 * @param controls - The orbit controls
 * @param animate - Whether to animate the transition
 */
export function centerCameraOnNode(
  nodeId: string,
  nodes: { id: string; x: number; y: number; z: number }[],
  camera: THREE.PerspectiveCamera,
  controls: OrbitControls,
  animate: boolean = false
) {
  const centerNode = nodes.find(n => n.id === nodeId);
  if (!centerNode) {
    // Fall back to auto zoom to fit all nodes
    return autoZoomToFit(nodes, camera, controls, animate);
  }

  const center = new THREE.Vector3(centerNode.x, centerNode.y, centerNode.z);
  
  // Calculate distance based on number of connected nodes
  // More nodes = need to see more of the graph = farther away
  const connectedCount = nodes.length;
  const baseDistance = connectedCount > 10 ? 50 : connectedCount > 5 ? 40 : 30;
  
  // Direction: slightly elevated angle for better view
  const direction = new THREE.Vector3(0.5, 0.6, 1).normalize();
  const dist = Math.max(baseDistance, 25); // Minimum 25 units
  
  const newPos = center.clone().add(direction.multiplyScalar(dist));
  
  if (animate) {
    lerpCamera(camera, controls, newPos, center, 1500);
  } else {
    camera.position.copy(newPos);
    controls.target.copy(center);
    controls.update();
  }
}

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

  // Increase margin for larger graphs and ensure minimum distance
  const marginMultiplier = nodes.length > 50 ? 1.8 : nodes.length > 20 ? 1.5 : 1.3;
  const radius = maxDist * marginMultiplier;
  const fov = camera.fov * Math.PI / 180;
  const calculatedDist = radius / Math.tan(fov / 2);

  // Ensure minimum distance to prevent camera being too close
  const minDistance = 30;
  const dist = Math.max(calculatedDist, minDistance);

  const direction = new THREE.Vector3(1, 0.8, 1).normalize();
  const newPos = center.clone().add(direction.multiplyScalar(dist));

  if (animate) {
    lerpCamera(camera, controls, newPos, center, 1500);
  } else {
    camera.position.copy(newPos);
    controls.target.copy(center);
    controls.update();
  }
}
