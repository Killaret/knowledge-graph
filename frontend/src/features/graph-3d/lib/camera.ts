import * as THREE from "three";
import type { OrbitControls } from "three/examples/jsm/controls/OrbitControls.js";
import type { SimulationNode } from "../model/types";

function getP95Distance(nodes: SimulationNode[], center: THREE.Vector3): number {
  const distances = nodes
    .map((n) => new THREE.Vector3(n.x, n.y, n.z).distanceTo(center))
    .sort((a, b) => a - b);
  const index = Math.min(nodes.length - 1, Math.floor(nodes.length * 0.95));
  return distances[index] || 0;
}

export function centerCameraOnNode(
  nodeId: string,
  nodes: SimulationNode[],
  camera: THREE.PerspectiveCamera,
  controls: OrbitControls
): void {
  const centerNode = nodes.find((n) => n.id === nodeId);
  if (!centerNode || nodes.length === 0) {
    autoZoomToFit(nodes, camera, controls);
    return;
  }

  const center = new THREE.Vector3(centerNode.x, centerNode.y, centerNode.z);
  const maxDist = getP95Distance(nodes, center);

  const marginMultiplier = nodes.length > 50 ? 1.8 : nodes.length > 20 ? 1.5 : 1.3;
  const radius = Math.max(maxDist * marginMultiplier, 20);
  const dist = Math.max(radius / Math.tan((camera.fov * Math.PI) / 360), 30);
  const direction = new THREE.Vector3(0.5, 0.6, 1).normalize();
  const newPos = center.clone().add(direction.multiplyScalar(dist));

  controls.target.copy(center);
  camera.position.copy(newPos);
  camera.lookAt(center);
  controls.update();
}

export function autoZoomToFit(
  nodes: SimulationNode[],
  camera: THREE.PerspectiveCamera,
  controls: OrbitControls
): void {
  if (nodes.length === 0) return;

  const center = new THREE.Vector3();
  nodes.forEach((n) => center.add(new THREE.Vector3(n.x, n.y, n.z)));
  center.divideScalar(nodes.length);

  const maxDist = getP95Distance(nodes, center);

  const marginMultiplier = nodes.length > 50 ? 1.8 : nodes.length > 20 ? 1.5 : 1.3;
  const radius = Math.max(maxDist * marginMultiplier, 10);
  const dist = Math.max(radius / Math.tan((camera.fov * Math.PI) / 360), 30);
  const direction = new THREE.Vector3(1, 0.8, 1).normalize();
  const newPos = center.clone().add(direction.multiplyScalar(dist));

  controls.target.copy(center);
  camera.position.copy(newPos);
  camera.lookAt(center);
  controls.update();
}
