import * as THREE from 'three';
import type { ObjectManager } from '$lib/three/rendering/objectManager';

export function handleNodeClick(
  _event: MouseEvent,
  _camera: THREE.PerspectiveCamera,
  _scene: THREE.Scene,
  _objectManager: ObjectManager
) {
  // Реализация рейкастинга для определения узла под мышью
  // Возвращает ID узла или null
  // Вызывается из внешнего обработчика
}

export function handleCanvasClick(
  event: MouseEvent,
  camera: THREE.PerspectiveCamera,
  scene: THREE.Scene,
  objectManager: ObjectManager,
  onEmptyClick?: () => void
) {
  const raycaster = new THREE.Raycaster();
  const mouse = new THREE.Vector2();
  mouse.x = (event.clientX / window.innerWidth) * 2 - 1;
  mouse.y = -(event.clientY / window.innerHeight) * 2 + 1;
  raycaster.setFromCamera(mouse, camera);
  const intersects = raycaster.intersectObjects(scene.children, true);
  const nodeIntersect = intersects.find((i) => i.object.userData.type === 'node');
  if (!nodeIntersect && onEmptyClick) {
    onEmptyClick();
  }
}
