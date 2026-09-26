import { describe, it, expect, vi, afterEach } from "vitest";
import * as THREE from "three";
import { LabelManager } from "./labels";
import { Graph3DEngine } from "./engine";
import type { Graph3DConfig, SimulationNode } from "../model/types";
import type { GraphNode, GraphLink } from "../model/types";

const config = { enableLabels: true } as Graph3DConfig;

function simNode(id: string): SimulationNode {
  return { id, title: `Node ${id}`, type: "planet", x: 0, y: 0, z: 0 } as SimulationNode;
}

describe("LabelManager — UI-GRAPH-1 selective labels", () => {
  let scene: THREE.Scene;
  let manager: LabelManager;

  afterEach(() => {
    manager?.dispose();
  });

  it("labels every node when no selection set is given (legacy behaviour)", () => {
    scene = new THREE.Scene();
    manager = new LabelManager(scene, config);
    manager.setLabels([simNode("a"), simNode("b"), simNode("c")]);
    expect(manager.size).toBe(3);
  });

  it("labels only the ids in the selection set", () => {
    scene = new THREE.Scene();
    manager = new LabelManager(scene, config);
    manager.setLabels([simNode("a"), simNode("b"), simNode("c")], new Set(["b"]));
    expect(manager.size).toBe(1);
  });
});

describe("Graph3DEngine — UI-GRAPH-1 selective labels", () => {
  function createContainer() {
    const container = document.createElement("div");
    container.style.width = "300px";
    container.style.height = "300px";
    return container;
  }

  function stubRenderer(engine: Graph3DEngine) {
    const bundle = (engine as any).sceneBundle;
    bundle.renderer.render = vi.fn();
    bundle.labelRenderer.render = vi.fn();
    bundle.controls.update = vi.fn();
  }

  // 25 nodes (> small-graph threshold): one hub with degree 4, rest isolated.
  const bigNodes: GraphNode[] = [
    { id: "hub", title: "Hub", type: "star", x: 0, y: 0, z: 0 },
    ...Array.from({ length: 24 }, (_, i) => ({
      id: `n${i}`,
      title: `N${i}`,
      type: "planet" as const,
      x: i,
      y: 0,
      z: 0,
    })),
  ];
  const bigLinks: GraphLink[] = [
    { source: "hub", target: "n0", weight: 0.5, link_type: "related" },
    { source: "hub", target: "n1", weight: 0.5, link_type: "related" },
    { source: "hub", target: "n2", weight: 0.5, link_type: "related" },
    { source: "hub", target: "n3", weight: 0.5, link_type: "related" },
  ];

  it("labels only hub nodes on a dense graph", () => {
    const engine = new Graph3DEngine(createContainer(), {}, {});
    stubRenderer(engine);
    engine.setData(bigNodes, bigLinks);

    const labelManager = (engine as any).labelManager as LabelManager;
    expect(labelManager.size).toBe(1); // only the hub
    engine.dispose();
  });

  it("adds a label for the selected node even when it is not a hub", () => {
    const engine = new Graph3DEngine(createContainer(), {}, {});
    stubRenderer(engine);
    engine.setData(bigNodes, bigLinks);

    engine.setSelectedNodeId("n10");
    const labelManager = (engine as any).labelManager as LabelManager;
    expect(labelManager.size).toBe(2); // hub + selected
    engine.dispose();
  });

  it("labels every node on a small graph", () => {
    const engine = new Graph3DEngine(createContainer(), {}, {});
    stubRenderer(engine);
    engine.setData(bigNodes.slice(0, 5), []);
    const labelManager = (engine as any).labelManager as LabelManager;
    expect(labelManager.size).toBe(5);
    engine.dispose();
  });
});
