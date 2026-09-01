import * as THREE from "three";
import { CelestialBody } from "$entities";
import type { Graph3DConfig, SimulationNode } from "../model/types";

interface NodeInstance {
  mesh: THREE.InstancedMesh;
  index: number;
  body: CelestialBody;
  size: number;
}

export class NodeManager {
  private scene: THREE.Scene;
  private config: Graph3DConfig;
  private dummy = new THREE.Object3D();
  private meshes = new Map<string, THREE.InstancedMesh>();
  private nodeInstances = new Map<string, NodeInstance>();
  private instanceMap = new Map<THREE.InstancedMesh, string[]>();
  private selectionMesh: THREE.Mesh | null = null;
  private selectionNodeId: string | null = null;

  constructor(scene: THREE.Scene, config: Graph3DConfig) {
    this.scene = scene;
    this.config = config;
  }

  setNodes(nodes: SimulationNode[]) {
    this.clear();

    const nodesByType = new Map<string, SimulationNode[]>();
    for (const node of nodes) {
      const type = node.type || "unknown";
      const list = nodesByType.get(type) ?? [];
      list.push(node);
      nodesByType.set(type, list);
    }

    // Scale node size with the graph bounding radius so spheres remain visible
    // from the camera distance chosen by autoZoomToFit (otherwise a 50-node
    // graph looks like text labels floating in empty space).
    let boundingRadius = 0;
    if (nodes.length > 0) {
      const center = new THREE.Vector3();
      for (const node of nodes) center.add(new THREE.Vector3(node.x, node.y, node.z));
      center.divideScalar(nodes.length);
      for (const node of nodes) {
        const d = new THREE.Vector3(node.x, node.y, node.z).distanceTo(center);
        if (d > boundingRadius) boundingRadius = d;
      }
    }
    const fitScale = Math.max(1, Math.min(8, boundingRadius * 0.075));

    for (const [type, typeNodes] of nodesByType) {
      const body = CelestialBody.fromString(type);
      const geometry = new THREE.IcosahedronGeometry(1, 0);
      // Use a flat, bright emissive material so nodes are visible even from the
      // camera distance used by autoZoomToFit. MeshStandardMaterial relies on
      // scene lighting and was too dim to be seen at the zoomed-out 3D view.
      const material = new THREE.MeshBasicMaterial({
        color: new THREE.Color(body.glowColor),
        transparent: true,
        opacity: 0.95,
      });

      const mesh = new THREE.InstancedMesh(geometry, material, typeNodes.length);
      mesh.userData = { type: "nodes", nodeType: type };
      mesh.instanceMatrix.setUsage(THREE.DynamicDrawUsage);

      const instanceIds: string[] = [];
      const size = Math.max(body.baseRadius * this.config.baseNodeScale * fitScale, 0.8);

      typeNodes.forEach((node, i) => {
        instanceIds.push(node.id);
        this.nodeInstances.set(node.id, { mesh, index: i, body, size });
      });

      this.instanceMap.set(mesh, instanceIds);
      this.meshes.set(type, mesh);
      this.scene.add(mesh);
    }

    this.updatePositions(nodes);
    this.setSelectedNodeId(this.selectionNodeId);
  }

  updatePositions(nodes: SimulationNode[]) {
    const touchedMeshes = new Set<THREE.InstancedMesh>();

    for (const node of nodes) {
      const instance = this.nodeInstances.get(node.id);
      if (!instance) continue;

      this.dummy.position.set(node.x, node.y, node.z);
      this.dummy.scale.set(instance.size, instance.size, instance.size);
      this.dummy.updateMatrix();
      instance.mesh.setMatrixAt(instance.index, this.dummy.matrix);
      touchedMeshes.add(instance.mesh);
    }

    for (const mesh of touchedMeshes) {
      mesh.instanceMatrix.needsUpdate = true;
    }

    if (this.selectionMesh && this.selectionNodeId) {
      const instance = this.nodeInstances.get(this.selectionNodeId);
      if (instance) {
        const offset = instance.size * 1.6;
        this.dummy.position.setFromMatrixPosition(instance.mesh.matrixWorld);
        // Use local matrix of instance to extract position
        const node = nodes.find((n) => n.id === this.selectionNodeId);
        if (node) {
          this.selectionMesh.position.set(node.x, node.y, node.z);
          this.selectionMesh.scale.set(offset, offset, offset);
        }
      }
    }
  }

  setSelectedNodeId(id: string | null | undefined) {
    this.selectionNodeId = id ?? null;

    if (!this.selectionNodeId) {
      if (this.selectionMesh) {
        this.selectionMesh.visible = false;
      }
      return;
    }

    const instance = this.nodeInstances.get(this.selectionNodeId);
    if (!instance) {
      if (this.selectionMesh) this.selectionMesh.visible = false;
      return;
    }

    if (!this.selectionMesh) {
      const geometry = new THREE.IcosahedronGeometry(1, 0);
      const material = new THREE.MeshBasicMaterial({
        color: 0xffffff,
        wireframe: true,
        transparent: true,
        opacity: 0.4,
      });
      this.selectionMesh = new THREE.Mesh(geometry, material);
      this.scene.add(this.selectionMesh);
    }

    this.selectionMesh.visible = true;
    const offset = instance.size * 1.6;
    this.selectionMesh.scale.set(offset, offset, offset);
  }

  getPositionMap(nodes: SimulationNode[]): Map<string, THREE.Vector3> {
    const map = new Map<string, THREE.Vector3>();
    for (const node of nodes) {
      map.set(node.id, new THREE.Vector3(node.x, node.y, node.z));
    }
    return map;
  }

  getNodeSize(node: SimulationNode): number {
    const body = CelestialBody.fromString(node.type);
    return Math.max(body.baseRadius * this.config.baseNodeScale, 0.5);
  }

  getNodeSizeById(id: string): number {
    const instance = this.nodeInstances.get(id);
    return instance?.size ?? 1;
  }

  raycast(raycaster: THREE.Raycaster): { id: string } | null {
    const meshes = Array.from(this.meshes.values());
    const intersects = raycaster.intersectObjects(meshes, false);
    if (intersects.length === 0) return null;

    const hit = intersects[0];
    if (!(hit.object instanceof THREE.InstancedMesh)) return null;

    const instanceId = hit.instanceId ?? -1;
    if (instanceId < 0) return null;

    const ids = this.instanceMap.get(hit.object);
    const nodeId = ids?.[instanceId];
    return nodeId ? { id: nodeId } : null;
  }

  clear() {
    for (const mesh of this.meshes.values()) {
      this.scene.remove(mesh);
      mesh.geometry.dispose();
      const materials = Array.isArray(mesh.material) ? mesh.material : [mesh.material];
      materials.forEach((m) => m.dispose());
    }
    this.meshes.clear();
    this.nodeInstances.clear();
    this.instanceMap.clear();
  }

  dispose() {
    this.clear();
    if (this.selectionMesh) {
      this.scene.remove(this.selectionMesh);
      this.selectionMesh.geometry.dispose();
      const materials = Array.isArray(this.selectionMesh.material)
        ? this.selectionMesh.material
        : [this.selectionMesh.material];
      materials.forEach((m) => m.dispose());
      this.selectionMesh = null;
    }
  }
}
