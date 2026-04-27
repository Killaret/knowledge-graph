/**
 * Shared types for GraphCanvas modules
 */

export interface SimulationNode {
  id: string;
  title: string;
  type?: string;
  x?: number;
  y?: number;
}

export interface SimulationLink {
  source: string | SimulationNode;
  target: string | SimulationNode;
  weight?: number;
  link_type?: string;
}

export interface SimulationState {
  simulation: any | null;
  simLinks: SimulationLink[];
  isRunning: boolean;
}

export interface TransformState {
  x: number;
  y: number;
  k: number;
}

export interface DragState {
  dragging: boolean;
  dragStart: { x: number; y: number };
}

export interface ResizeState {
  width: number;
  height: number;
}
