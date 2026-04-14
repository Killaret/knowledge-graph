import * as THREE from 'three';
import type { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';

export function autoZoomToFit(
  nodes: { x: number; y: number; z: number }[],
  camera: THREE.PerspectiveCamera,
  controls: OrbitControls
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

  console.log('[autoZoomToFit] center:', `(${center.x.toFixed(2)}, ${center.y.toFixed(2)}, ${center.z.toFixed(2)})`, 'radius:', radius.toFixed(2), 'distance:', dist.toFixed(2), 'nodes:', nodes.length);

  const direction = new THREE.Vector3(1, 1, 1).normalize();
  const newPos = center.clone().add(direction.multiplyScalar(dist));
  camera.position.copy(newPos);
  controls.target.copy(center);
  controls.update();
  console.log('[autoZoomToFit] camera position set to:', `(${newPos.x.toFixed(2)}, ${newPos.y.toFixed(2)}, ${newPos.z.toFixed(2)})`);
}
