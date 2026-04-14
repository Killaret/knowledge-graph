import * as THREE from 'three';
import { CSS2DObject } from 'three/examples/jsm/renderers/CSS2DRenderer.js';
import { createNodeMesh } from './nodeFactory';
import { createLinkLine } from './linkFactory';
import { createLabel } from './labelFactory';
import type { GraphNode, GraphLink } from '$lib/api/graph';

export class ObjectManager {
  private scene: THREE.Scene;
  private nodeMap = new Map<string, THREE.Group>();
  private linkMap = new Map<string, THREE.Line>();
  private labelMap = new Map<string, CSS2DObject>();

  constructor(scene: THREE.Scene) {
    this.scene = scene;
  }

  createAll(nodes: GraphNode[], links: GraphLink[]) {
    console.log(`[ObjectManager] createAll called: ${nodes.length} nodes, ${links.length} links`);
    this.clear();

    nodes.forEach((node) => {
      const mesh = createNodeMesh(node);
      mesh.position.set((node as any).x ?? 0, (node as any).y ?? 0, (node as any).z ?? 0);
      mesh.userData = { type: 'node', id: node.id, nodeData: node };
      this.scene.add(mesh);
      this.nodeMap.set(node.id, mesh);
      console.log(`[ObjectManager] Created node ${node.id} (${node.type || 'default'}) at (${mesh.position.x.toFixed(2)}, ${mesh.position.y.toFixed(2)}, ${mesh.position.z.toFixed(2)})`);

      const label = createLabel(node.title || node.id.substring(0, 6), () => {
        window.location.href = `/notes/${node.id}`;
      });
      label.position.copy(mesh.position);
      label.userData = { type: 'label', nodeId: node.id };
      this.scene.add(label);
      this.labelMap.set(node.id, label);
    });

    links.forEach((link) => {
      const sourceNode = nodes.find((n) => n.id === link.source);
      const targetNode = nodes.find((n) => n.id === link.target);
      if (!sourceNode || !targetNode) {
        console.warn(`[ObjectManager] Skipping link ${link.source}-${link.target}: nodes not found`);
        return;
      }

      const sourcePos = new THREE.Vector3((sourceNode as any).x ?? 0, (sourceNode as any).y ?? 0, (sourceNode as any).z ?? 0);
      const targetPos = new THREE.Vector3((targetNode as any).x ?? 0, (targetNode as any).y ?? 0, (targetNode as any).z ?? 0);
      const line = createLinkLine(sourcePos, targetPos, link.weight ?? 1, link.link_type);
      line.userData = { type: 'link', source: link.source, target: link.target, linkType: link.link_type };
      this.scene.add(line);
      this.linkMap.set(`${link.source}-${link.target}`, line);
    });
    console.log(`[ObjectManager] Total created: ${this.nodeMap.size} nodes, ${this.linkMap.size} links, ${this.labelMap.size} labels`);
  }

  updatePositions(nodes: GraphNode[]) {
    nodes.forEach((node) => {
      const obj = this.nodeMap.get(node.id);
      const label = this.labelMap.get(node.id);
      if (obj) {
        obj.position.set((node as any).x ?? 0, (node as any).y ?? 0, (node as any).z ?? 0);
        if (label) label.position.copy(obj.position);
      }
    });
  }

  updateLinks(links: GraphLink[]) {
    let updatedCount = 0;
    links.forEach((link) => {
      const key = `${link.source}-${link.target}`;
      const line = this.linkMap.get(key);
      if (!line) {
        console.warn(`[ObjectManager] Link ${key} not found in linkMap`);
        return;
      }

      const sourceObj = this.nodeMap.get(link.source);
      const targetObj = this.nodeMap.get(link.target);
      if (!sourceObj || !targetObj) {
        console.warn(`[ObjectManager] Cannot update link ${key}: source or target not found`);
        return;
      }

      const points = [sourceObj.position.clone(), targetObj.position.clone()];
      line.geometry.dispose();
      line.geometry = new THREE.BufferGeometry().setFromPoints(points);
      updatedCount++;
    });
    if (updatedCount > 0) {
      console.log(`[ObjectManager] Updated ${updatedCount}/${links.length} links`);
    }
  }

  clear() {
    this.nodeMap.forEach((obj) => this.scene.remove(obj));
    this.linkMap.forEach((obj) => this.scene.remove(obj));
    this.labelMap.forEach((obj) => this.scene.remove(obj));
    this.nodeMap.clear();
    this.linkMap.clear();
    this.labelMap.clear();
  }

  getNode(id: string) {
    return this.nodeMap.get(id);
  }
}
