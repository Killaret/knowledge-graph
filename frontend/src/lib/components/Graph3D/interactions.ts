/**
 * Interaction handlers for Graph3D
 */
import { goto } from '$app/navigation';
import type { GraphNode } from './types';

export interface RaycasterState {
  raycaster: any;
  mouse: any;
  hoveredNodeId: string | null;
}

export interface ClickCallbacks {
  onNodeClick?: (node: GraphNode) => void;
  onNodeDoubleClick?: (node: GraphNode) => void;
}

/**
 * Initialize raycaster for mouse interactions
 */
export function initRaycaster(THREE: any): RaycasterState {
  return {
    raycaster: new THREE.Raycaster(),
    mouse: new THREE.Vector2(),
    hoveredNodeId: null
  };
}

/**
 * Update mouse coordinates from event
 */
export function updateMouseCoordinates(
  event: MouseEvent,
  container: HTMLElement,
  mouse: any
): void {
  const rect = container.getBoundingClientRect();
  mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
  mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
}

/**
 * Find intersected node from mouse position
 */
export function findIntersectedNode(
  raycaster: any,
  mouse: any,
  camera: any,
  nodeMeshes: Map<string, any>
): GraphNode | null {
  raycaster.setFromCamera(mouse, camera);

  const meshes = Array.from(nodeMeshes.values());
  const intersects = raycaster.intersectObjects(meshes);

  if (intersects.length > 0) {
    const intersectedMesh = intersects[0].object;
    // Find node by mesh
    for (const [id, mesh] of nodeMeshes.entries()) {
      if (mesh === intersectedMesh) {
        return { id, title: mesh.userData?.title || id, type: mesh.userData?.type };
      }
    }
  }

  return null;
}

/**
 * Handle single click on node
 */
export function handleNodeClick(
  node: GraphNode,
  callbacks: ClickCallbacks
): void {
  if (callbacks.onNodeClick) {
    callbacks.onNodeClick(node);
  } else {
    // Default: navigate to note
    goto(`/notes/${node.id}`);
  }
}

/**
 * Handle double click on node
 */
export function handleNodeDoubleClick(
  node: GraphNode,
  callbacks: ClickCallbacks,
  cameraController: any
): void {
  if (callbacks.onNodeDoubleClick) {
    callbacks.onNodeDoubleClick(node);
  }

  // Center camera on node
  if (cameraController?.centerCameraOnNode) {
    cameraController.centerCameraOnNode(node.id);
  }
}

/**
 * Handle hover state change
 */
export function handleNodeHover(
  node: GraphNode | null,
  state: RaycasterState,
  nodeMeshes: Map<string, any>,
  onHoverChange?: (nodeId: string | null) => void
): void {
  const previousId = state.hoveredNodeId;
  const newId = node?.id || null;

  if (previousId !== newId) {
    state.hoveredNodeId = newId;

    // Reset previous mesh scale
    if (previousId && nodeMeshes.has(previousId)) {
      const prevMesh = nodeMeshes.get(previousId);
      if (prevMesh) {
        prevMesh.scale.set(1, 1, 1);
      }
    }

    // Scale up new mesh
    if (newId && nodeMeshes.has(newId)) {
      const newMesh = nodeMeshes.get(newId);
      if (newMesh) {
        newMesh.scale.set(1.2, 1.2, 1.2);
      }
    }

    onHoverChange?.(newId);
  }
}

/**
 * Setup all interaction handlers
 */
export function setupInteractions(
  container: HTMLElement,
  state: RaycasterState,
  camera: any,
  nodeMeshes: Map<string, any>,
  callbacks: ClickCallbacks,
  cameraController: any
): () => void {
  let clickTimeout: ReturnType<typeof setTimeout> | null = null;
  let lastClickedNode: GraphNode | null = null;

  function onMouseMove(event: MouseEvent) {
    updateMouseCoordinates(event, container, state.mouse);
    const node = findIntersectedNode(state.raycaster, state.mouse, camera, nodeMeshes);
    handleNodeHover(node, state, nodeMeshes);
  }

  function onClick(event: MouseEvent) {
    updateMouseCoordinates(event, container, state.mouse);
    const node = findIntersectedNode(state.raycaster, state.mouse, camera, nodeMeshes);

    if (node) {
      // Check for double click
      if (lastClickedNode?.id === node.id && clickTimeout) {
        clearTimeout(clickTimeout);
        clickTimeout = null;
        handleNodeDoubleClick(node, callbacks, cameraController);
        lastClickedNode = null;
      } else {
        lastClickedNode = node;
        clickTimeout = setTimeout(() => {
          handleNodeClick(node, callbacks);
          lastClickedNode = null;
          clickTimeout = null;
        }, 250);
      }
    }
  }

  container.addEventListener('mousemove', onMouseMove);
  container.addEventListener('click', onClick);

  return () => {
    container.removeEventListener('mousemove', onMouseMove);
    container.removeEventListener('click', onClick);
    if (clickTimeout) {
      clearTimeout(clickTimeout);
    }
  };
}
