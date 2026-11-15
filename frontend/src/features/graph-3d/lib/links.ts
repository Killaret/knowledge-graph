import * as THREE from "three";
import { LinkType, AUTO_LINK_COLOR } from "$entities";
import {
  chainDepthOpacity,
  dependencyLinkKey,
  type DependencyChain,
} from "$entities/graph-canvas/lib/dependency-chain";
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
      // LINK-TYPES-1: gamma links use the dedicated auto-link colour — a hue no
      // manual type has — with brightness still following the link's weight.
      const isAuto = link.source_type === "gamma";
      const color = new THREE.Color(isAuto ? AUTO_LINK_COLOR : linkType.color);

      const material = new THREE.LineBasicMaterial({
        color,
        transparent: true,
        opacity,
      });

      const line = new THREE.Line(geometry, material);
      line.userData = {
        type: "link",
        linkId,
        source: sourceId,
        target: targetId,
        // Base styling kept for highlight restore.
        baseColor: color.clone(),
        baseOpacity: opacity,
      };
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

  /**
   * LINK-TYPES-1: apply/remove the dependency-chain highlight. Chain links
   * brighten (cycle members turn red), everything else dims; `null` restores.
   */
  applyChainHighlight(chain: DependencyChain | null): void {
    for (const line of this.linkObjects.values()) {
      const material = line.material as THREE.LineBasicMaterial;
      const { source, target, baseColor, baseOpacity } = line.userData as {
        source: string;
        target: string;
        baseColor: THREE.Color;
        baseOpacity: number;
      };
      if (!chain) {
        material.color.copy(baseColor);
        material.opacity = baseOpacity;
        continue;
      }
      const key = dependencyLinkKey(source, target);
      const depth = chain.linkDepth.get(key);
      if (depth === undefined) {
        material.color.copy(baseColor);
        material.opacity = 0.08;
        continue;
      }
      if (chain.cycleLinks.has(key)) {
        material.color.set(0xef4444);
        material.opacity = Math.min(1, baseOpacity + 0.3);
      } else {
        material.color.copy(baseColor);
        material.opacity = Math.min(1, baseOpacity * 0.5 + chainDepthOpacity(depth) * 0.7);
      }
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
