/**
 * Base types for Graph3D component
 */

import type { ObjectManager } from '$lib/three/rendering/objectManager';
import type { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import type { CSS2DRenderer } from 'three/examples/jsm/renderers/CSS2DRenderer.js';
import * as THREE from 'three';

// Re-export external types
export type { ObjectManager, OrbitControls, CSS2DRenderer, THREE };

/**
 * Basic graph node with optional coordinates (after simulation)
 */
export interface GraphNode {
  id: string;
  title: string;
  type?: string;
  x?: number;
  y?: number;
  z?: number;
  size?: number;
}

/**
 * Graph link connecting two nodes
 */
export interface GraphLink {
  source: string | GraphNode;
  target: string | GraphNode;
  weight?: number;
  link_type?: string;
}

/**
 * Graph data containing nodes and links
 */
export interface GraphData {
  nodes: GraphNode[];
  links: GraphLink[];
}

/**
 * Node with simulation properties (after d3-force layout)
 * Extends GraphNode with velocity and force simulation properties
 */
export interface SimulationNode extends GraphNode {
  x: number;
  y: number;
  z: number;
  vx?: number;
  vy?: number;
  vz?: number;
  fx?: number | null;
  fy?: number | null;
  fz?: number | null;
  index?: number;
}

/**
 * Link with resolved source/target nodes (after simulation)
 */
export interface SimulationLink {
  source: SimulationNode;
  target: SimulationNode;
  weight?: number;
  link_type?: string;
  index?: number;
}

/**
 * D3 Force Simulation interface
 * Mimics the d3-force-3d simulation object structure
 */
export interface Simulation {
  nodes(): SimulationNode[];
  nodes(nodes: SimulationNode[]): Simulation;
  alpha(): number;
  alpha(alpha: number): Simulation;
  alphaTarget(): number;
  alphaTarget(target: number): Simulation;
  alphaDecay(): number;
  alphaDecay(decay: number): Simulation;
  restart(): Simulation;
  stop(): Simulation;
  tick(): void;
  on(type: 'tick' | 'end', callback: () => void): Simulation;
  force(name: string): ForceFn | undefined;
  force(name: string, force: ForceFn | null): Simulation;
}

/**
 * Force function type for d3-force
 */
export type ForceFn = (alpha: number) => void;

/**
 * Graph3D component state
 */
export interface Graph3DState {
  scene: THREE.Scene | null;
  camera: THREE.PerspectiveCamera | null;
  renderer: THREE.WebGLRenderer | null;
  controls: OrbitControls | null;
  labelRenderer: CSS2DRenderer | null;
  simulation: Simulation | null;
  objectManager: ObjectManager | null;
  isInitialized: boolean;
}

/**
 * Hover state for node interactions
 */
export interface HoverState {
  hoveredNodeId: string | null;
}

/**
 * Position in 3D space
 */
export interface Position3D {
  x: number;
  y: number;
  z: number;
}

/**
 * Camera controller interface for camera manipulation
 */
export interface CameraController {
  centerCameraOnNode: (nodeId: string) => void;
  autoZoomToFit: () => void;
}

/**
 * Node mesh map for raycasting
 */
export type NodeMeshMap = Map<string, THREE.Object3D>;

/**
 * Callback types
 */
export type OnNodeClickCallback = (node: GraphNode) => void;
export type OnNodeDoubleClickCallback = (node: GraphNode) => void;
export type OnHoverChangeCallback = (nodeId: string | null) => void;
export type OnCompleteCallback = () => void;
export type OnUpdateCallback = (value: number) => void;
