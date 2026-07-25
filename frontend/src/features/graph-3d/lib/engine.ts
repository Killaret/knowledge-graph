import * as THREE from "three";
import { createScene, setFogDensity, resizeScene, setStarfieldCount } from "./scene";
import { NodeManager } from "./nodes";
import { LinkManager } from "./links";
import { LabelManager } from "./labels";
import { createGraphSimulation } from "./simulation";
import { autoZoomToFit, centerCameraOnNode } from "./camera";
import { filterValidLinks } from "$shared/utils/graphUtils";
import { graphConfig3D, graphPerformanceConfig } from "$shared/config/config";
import { createPerformanceMonitor } from "$shared/lib/performance-monitor";
import { toSimulationNodes } from "../config";
import { applyFogPreset, getLowerFogPreset, getHigherFogPreset } from "./fog";
import type {
  Graph3DCallbacks,
  Graph3DConfig,
  GraphNode,
  GraphLink,
  SimulationNode,
} from "../model/types";
import { DEFAULT_GRAPH3D_CONFIG, type FogPresetName } from "../model/types";

export class Graph3DEngine {
  private container: HTMLElement;
  private callbacks: Graph3DCallbacks;
  private config: Graph3DConfig;
  private sceneBundle: ReturnType<typeof createScene>;
  private nodeManager: NodeManager;
  private linkManager: LinkManager;
  private labelManager: LabelManager;
  private sim: ReturnType<typeof createGraphSimulation> | null = null;
  private simNodes: SimulationNode[] = [];
  private simLinks: GraphLink[] = [];
  private rafId: number | null = null;
  private disposed = false;
  private isReady = false;
  private currentFogDensity: number;
  private targetFogDensity: number;
  private fogInitial: number;
  private fogFinal: number;
  private lastFrameTime = 0;
  private readonly frameInterval = 33; // ~30 fps
  private monitor = createPerformanceMonitor();
  private performanceConfig = graphPerformanceConfig;
  private currentPerformanceLevel: "high" | "medium" | "low" = "high";
  private currentFogPreset: FogPresetName = "birth";
  private lowFpsStreak = 0;
  private highFpsStreak = 0;

  constructor(
    container: HTMLElement,
    callbacks: Graph3DCallbacks = {},
    partialConfig?: Partial<Graph3DConfig>
  ) {
    this.container = container;
    this.callbacks = callbacks;
    this.config = {
      ...DEFAULT_GRAPH3D_CONFIG,
      maxNodes: graphConfig3D.max_nodes,
      defaultFogPreset: graphConfig3D.fog.default_preset,
      fog_presets: graphConfig3D.fog.presets,
      ...partialConfig,
    };

    this.sceneBundle = createScene(container, this.config);
    this.nodeManager = new NodeManager(this.sceneBundle.scene, this.config);
    this.linkManager = new LinkManager(this.sceneBundle.scene, this.config);
    this.labelManager = new LabelManager(this.sceneBundle.scene, this.config);
    this.currentFogDensity = this.config.fogDensityInitial;
    this.targetFogDensity = this.config.fogDensityInitial;
    this.fogInitial = this.config.fogDensityInitial;
    this.fogFinal = this.config.fogDensityFinal;
  }

  setData(nodes: GraphNode[], links: GraphLink[]) {
    if (this.disposed) return;

    this.callbacks.onLoadingChange?.(true);
    this.isReady = false;
    this.stopLoop();

    const validLinks = filterValidLinks(nodes, links);
    const hasPositions = nodes.every(
      (n) => typeof n.x === "number" && typeof n.y === "number" && typeof n.z === "number"
    );

    if (nodes.length > this.config.maxNodes) {
      console.warn(
        `[Graph3D] Graph too large: ${nodes.length} nodes, limiting to ${this.config.maxNodes}`
      );
    }

    this.simNodes = toSimulationNodes(nodes, validLinks);
    this.simLinks = validLinks;

    try {
      this.sim = createGraphSimulation(this.simNodes, this.simLinks);
      const warmStartTicks = hasPositions ? 10 : this.config.warmStartTicks;
      this.sim.tick(warmStartTicks);

      this.nodeManager.setNodes(this.simNodes);
      this.linkManager.setLinks(this.simLinks, this.nodeManager.getPositionMap(this.simNodes));
      this.labelManager.setLabels(this.simNodes);

      this.updateScene();
      this.centerCamera();

      this.currentFogPreset = this.config.defaultFogPreset;
      this.currentPerformanceLevel = "high";
      const { initial, final } = applyFogPreset(
        this.sceneBundle.scene,
        this.config.defaultFogPreset,
        this.config
      );
      this.fogInitial = initial;
      this.fogFinal = final;
      this.currentFogDensity = initial;
      this.targetFogDensity = initial;

      if (this.config.disableAnimation) {
        this.simulateToStable();
        this.updateScene();
        this.finishLoading();
      } else {
        this.startLoop();
      }
    } catch (e) {
      this.callbacks.onError?.("Failed to initialize 3D graph simulation");
      if (typeof process !== "undefined" && process.env?.VITEST === "true") {
        throw e;
      }
    }
  }

  private startLoop() {
    this.stopLoop();
    this.lastFrameTime = performance.now();
    this.rafId = requestAnimationFrame(() => this.frame());
  }

  private stopLoop() {
    if (this.rafId !== null) {
      cancelAnimationFrame(this.rafId);
      this.rafId = null;
    }
  }

  private frame() {
    if (this.disposed || this.config.disableAnimation) return;

    const now = performance.now();

    // Sample every animation frame so the monitor can see the actual FPS,
    // even when rendering is throttled to ~30 fps.
    this.monitor.tick(now);
    this.adaptPerformance();

    if (now - this.lastFrameTime < this.frameInterval) {
      this.rafId = requestAnimationFrame(() => this.frame());
      return;
    }
    this.lastFrameTime = now;

    if (this.sim && this.sim.alpha() > this.sim.alphaMin()) {
      this.sim.tick(1);
      this.updateScene();

      const progress = Math.min(1 - this.sim.alpha(), 1);
      this.targetFogDensity = this.fogInitial - (this.fogInitial - this.fogFinal) * progress;
    } else if (!this.isReady) {
      this.simulateToStable();
      this.updateScene();
      this.finishLoading();
    }

    if (Math.abs(this.targetFogDensity - this.currentFogDensity) > 0.0001) {
      this.currentFogDensity += (this.targetFogDensity - this.currentFogDensity) * 0.1;
      setFogDensity(this.sceneBundle.scene, this.currentFogDensity);
    }

    this.sceneBundle.controls.update();
    this.sceneBundle.renderer.render(this.sceneBundle.scene, this.sceneBundle.camera);
    this.sceneBundle.labelRenderer.render(this.sceneBundle.scene, this.sceneBundle.camera);

    this.rafId = requestAnimationFrame(() => this.frame());
  }

  private adaptPerformance() {
    const fps = this.monitor.fps;
    if (fps < this.performanceConfig.fps_threshold_low) {
      this.lowFpsStreak++;
      this.highFpsStreak = 0;
    } else if (fps > this.performanceConfig.fps_threshold_high) {
      this.highFpsStreak++;
      this.lowFpsStreak = 0;
    } else {
      this.lowFpsStreak = 0;
      this.highFpsStreak = 0;
    }

    if (this.lowFpsStreak >= this.performanceConfig.low_fps_sample_count) {
      this.downgradePerformance();
      this.lowFpsStreak = 0;
    }

    if (this.highFpsStreak >= this.performanceConfig.low_fps_sample_count) {
      this.upgradePerformance();
      this.highFpsStreak = 0;
    }
  }

  private downgradePerformance() {
    const levels: ("high" | "medium" | "low")[] = ["high", "medium", "low"];
    const presets: FogPresetName[] = ["birth", "nebula", "deep-space"];
    const idx = levels.indexOf(this.currentPerformanceLevel);
    if (idx >= levels.length - 1) return;

    this.currentPerformanceLevel = levels[idx + 1];
    this.currentFogPreset = presets[idx + 1];
    this.applyPerformancePreset();

    this.sceneBundle.controls.autoRotate = false;
    this.sceneBundle.controls.autoRotateSpeed = 0;
  }

  private upgradePerformance() {
    const levels: ("high" | "medium" | "low")[] = ["high", "medium", "low"];
    const presets: FogPresetName[] = ["birth", "nebula", "deep-space"];
    const idx = levels.indexOf(this.currentPerformanceLevel);
    if (idx <= 0) return;

    this.currentPerformanceLevel = levels[idx - 1];
    this.currentFogPreset = presets[idx - 1];
    this.applyPerformancePreset();

    if (this.config.autoRotate) {
      this.sceneBundle.controls.autoRotate = true;
      this.sceneBundle.controls.autoRotateSpeed = 0.8;
    }
  }

  private applyPerformancePreset() {
    const { initial, final } = applyFogPreset(
      this.sceneBundle.scene,
      this.currentFogPreset,
      this.config
    );
    this.fogInitial = initial;
    this.fogFinal = final;
    this.currentFogDensity = initial;
    this.targetFogDensity = initial;

    const starfieldCount = this.performanceConfig.starfield_counts[this.currentPerformanceLevel];
    this.sceneBundle.starfield = setStarfieldCount(
      this.sceneBundle.scene,
      this.sceneBundle.starfield,
      starfieldCount
    );
  }

  private simulateToStable() {
    if (!this.sim) return;
    let iterations = 0;
    while (this.sim.alpha() > this.sim.alphaMin() && iterations < 500) {
      this.sim.tick(1);
      iterations++;
    }
  }

  private updateScene() {
    const positionMap = this.nodeManager.getPositionMap(this.simNodes);
    this.nodeManager.updatePositions(this.simNodes);
    this.linkManager.updatePositions(this.simLinks, positionMap);
    this.labelManager.updatePositions(this.simNodes);
  }

  private finishLoading() {
    if (this.isReady) return;
    this.isReady = true;
    this.targetFogDensity = this.fogFinal;
    this.callbacks.onReady?.();
    this.callbacks.onLoadingChange?.(false);
  }

  private centerCamera() {
    if (this.simNodes.length === 0) return;
    autoZoomToFit(this.simNodes, this.sceneBundle.camera, this.sceneBundle.controls);
  }

  centerOnNode(nodeId: string | null | undefined) {
    if (this.disposed || !nodeId || this.simNodes.length === 0) return;
    centerCameraOnNode(nodeId, this.simNodes, this.sceneBundle.camera, this.sceneBundle.controls);
  }

  setSelectedNodeId(nodeId: string | null | undefined) {
    if (this.disposed) return;
    this.nodeManager.setSelectedNodeId(nodeId);
  }

  handleResize() {
    if (this.disposed) return;
    resizeScene(
      this.container,
      this.sceneBundle.camera,
      this.sceneBundle.renderer,
      this.sceneBundle.labelRenderer
    );
  }

  handleClick(event: MouseEvent) {
    if (this.disposed || this.simNodes.length === 0) return;
    const nodeId = this.raycastNodeId(event);
    if (!nodeId) return;

    const node = this.simNodes.find((n) => n.id === nodeId);
    if (node && this.callbacks.onNodeClick) {
      this.callbacks.onNodeClick({
        id: node.id,
        title: node.title,
        type: node.type,
      });
    }
  }

  handleDoubleClick(event: MouseEvent) {
    if (this.disposed || this.simNodes.length === 0) return;
    const nodeId = this.raycastNodeId(event);
    if (!nodeId) return;

    const node = this.simNodes.find((n) => n.id === nodeId);
    if (node && this.callbacks.onNodeDoubleClick) {
      this.callbacks.onNodeDoubleClick({
        id: node.id,
        title: node.title,
        type: node.type,
      });
    }
  }

  private raycastNodeId(event: MouseEvent): string | null {
    const rect = this.container.getBoundingClientRect();
    const mouse = new THREE.Vector2(
      ((event.clientX - rect.left) / rect.width) * 2 - 1,
      -((event.clientY - rect.top) / rect.height) * 2 + 1
    );
    const raycaster = new THREE.Raycaster();
    raycaster.setFromCamera(mouse, this.sceneBundle.camera);
    return this.nodeManager.raycast(raycaster)?.id ?? null;
  }

  dispose() {
    if (this.disposed) return;
    this.disposed = true;
    this.stopLoop();
    this.sim?.stop();
    this.nodeManager.dispose();
    this.linkManager.dispose();
    this.labelManager.dispose();
    this.sceneBundle.dispose();
  }
}
