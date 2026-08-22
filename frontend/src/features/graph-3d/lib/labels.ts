import { CSS2DObject } from "three/examples/jsm/renderers/CSS2DRenderer.js";
import * as THREE from "three";
import { CelestialBody } from "$entities";
import type { Graph3DConfig, SimulationNode } from "../model/types";

export class LabelManager {
  private scene: THREE.Scene;
  private config: Graph3DConfig;
  private labels = new Map<string, CSS2DObject>();

  constructor(scene: THREE.Scene, config: Graph3DConfig) {
    this.scene = scene;
    this.config = config;
  }

  setLabels(nodes: SimulationNode[]) {
    this.clear();

    if (!this.config.enableLabels) return;

    for (const node of nodes) {
      const body = CelestialBody.fromString(node.type);
      const text = `${body.emoji} ${node.title || node.id.slice(0, 6)}`;
      const div = document.createElement("div");
      div.textContent = text;
      div.style.color = "#ffffff";
      div.style.fontSize = "12px";
      div.style.fontWeight = "bold";
      div.style.fontFamily = "Arial, sans-serif";
      div.style.textShadow = "1px 1px 3px rgba(0,0,0,0.9)";
      div.style.padding = "2px 6px";
      div.style.background = "rgba(5, 5, 16, 0.6)";
      div.style.borderRadius = "12px";
      div.style.backdropFilter = "blur(2px)";
      div.style.pointerEvents = "none";
      div.style.whiteSpace = "nowrap";
      div.style.userSelect = "none";

      const label = new CSS2DObject(div);
      label.userData = { nodeId: node.id };
      this.setLabelPosition(label, node, body);
      this.scene.add(label);
      this.labels.set(node.id, label);
    }
  }

  updatePositions(nodes: SimulationNode[]) {
    if (!this.config.enableLabels) return;

    for (const node of nodes) {
      const label = this.labels.get(node.id);
      if (!label) continue;
      const body = CelestialBody.fromString(node.type);
      this.setLabelPosition(label, node, body);
    }
  }

  private setLabelPosition(label: CSS2DObject, node: SimulationNode, body: CelestialBody) {
    const size = body.baseRadius * this.config.baseNodeScale;
    label.position.set(node.x, node.y + size + 1.2, node.z);
  }

  clear() {
    for (const label of this.labels.values()) {
      this.scene.remove(label);
      if (label.element && label.element.parentNode) {
        label.element.parentNode.removeChild(label.element);
      }
    }
    this.labels.clear();
  }

  dispose() {
    this.clear();
  }
}
