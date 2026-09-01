import * as THREE from "three";
import { LinkType } from "$entities";
import type { GraphLink } from "$shared/api/graph";
import type { Graph3DConfig } from "../model/types";

function getLinkEndpointId(value: unknown): string | undefined {
  if (typeof value === "string") return value;
  if (typeof value === "number") return String(value);
  if (value && typeof value === "object" && "id" in value && typeof value.id === "string") {
    return value.id;
  }
  return undefined;
}

export class LinkManager {
  private scene: THREE.Scene;
  private config: Graph3DConfig;
  private linkObjects = new Map<string, THREE.Line>();

  constructor(scene: THREE.Scene, config: Graph3DConfig) {
    this.scene = scene;
    this.config = config;
  }

  setLinks(links: GraphLink[], nodePositions: Map<string, THREE.Vector3>) {
    this.clear();

    for (const link of links) {
      const sourceId = getLinkEndpointId(link.source);
      const targetId = getLinkEndpointId(link.target);
      if (!sourceId || !targetId) continue;
      const linkId = `${sourceId}-${targetId}`;

      const sourcePos = nodePositions.get(sourceId);
      const targetPos = nodePositions.get(targetId);
      if (!sourcePos || !targetPos) continue;

      const geometry = new THREE.BufferGeometry();
      const positions = new Float32Array([
        sourcePos.x,
        sourcePos.y,
        sourcePos.z,
        targetPos.x,
        targetPos.y,
        targetPos.z,
      ]);
      geometry.setAttribute("position", new THREE.BufferAttribute(positions, 3));

      const linkType = LinkType.fromString(link.link_type);
      const weight = Math.max(0, Math.min(1, link.weight ?? linkType.defaultWeight));
      const opacity = 0.6 + weight * 0.4;

      const material = new THREE.LineBasicMaterial({
        color: new THREE.Color(linkType.color),
        transparent: true,
        opacity,
      });

      const line = new THREE.Line(geometry, material);
      line.userData = { type: "link", linkId, source: sourceId, target: targetId };
      this.scene.add(line);
      this.linkObjects.set(linkId, line);
    }
  }

  updatePositions(links: GraphLink[], nodePositions: Map<string, THREE.Vector3>) {
    for (const link of links) {
      const sourceId = getLinkEndpointId(link.source);
      const targetId = getLinkEndpointId(link.target);
      if (!sourceId || !targetId) continue;
      const linkId = `${sourceId}-${targetId}`;

      const line = this.linkObjects.get(linkId);
      if (!line) continue;

      const sourcePos = nodePositions.get(sourceId);
      const targetPos = nodePositions.get(targetId);
      if (!sourcePos || !targetPos) continue;

      const positions = line.geometry.attributes.position.array as Float32Array;
      positions[0] = sourcePos.x;
      positions[1] = sourcePos.y;
      positions[2] = sourcePos.z;
      positions[3] = targetPos.x;
      positions[4] = targetPos.y;
      positions[5] = targetPos.z;
      line.geometry.attributes.position.needsUpdate = true;
    }
  }

  clear() {
    for (const line of this.linkObjects.values()) {
      this.scene.remove(line);
      line.geometry.dispose();
      const materials = Array.isArray(line.material) ? line.material : [line.material];
      materials.forEach((m) => m.dispose());
    }
    this.linkObjects.clear();
  }

  dispose() {
    this.clear();
  }
}
