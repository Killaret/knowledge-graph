/**
 * Base types for Graph3D component
 */

export interface GraphNode {
  id: string;
  title: string;
  type?: string;
  x?: number;
  y?: number;
  z?: number;
}

export interface GraphLink {
  source: string | GraphNode;
  target: string | GraphNode;
  weight?: number;
  link_type?: string;
}

export interface GraphData {
  nodes: GraphNode[];
  links: GraphLink[];
}

export interface Graph3DState {
  scene: THREE.Scene | null;
  camera: THREE.PerspectiveCamera | null;
  renderer: THREE.WebGLRenderer | null;
  controls: OrbitControls | null;
  labelRenderer: CSS2DRenderer | null;
  simulation: any | null;
  objectManager: ObjectManager | null;
  isInitialized: boolean;
}

export interface HoverState {
  hoveredNodeId: string | null;
}

// Types from external modules (for re-export)
import type { ObjectManager } from '$lib/three/rendering/objectManager';
import type { OrbitControls } from 'three/examples/jsm/controls/OrbitControls.js';
import type { CSS2DRenderer } from 'three/examples/jsm/renderers/CSS2DRenderer.js';
import * as THREE from 'three';

export type { ObjectManager, OrbitControls, CSS2DRenderer, THREE };
