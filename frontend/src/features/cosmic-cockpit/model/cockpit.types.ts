/**
 * Domain types for the Cosmic Cockpit UI.
 *
 * These types are intentionally UI/UX only and live in the feature layer
 * because they describe the cockpit shell, not the graph domain.
 */

export type CockpitPanelPosition = "top" | "bottom" | "left" | "right";

export interface CockpitPanelState {
  /** Whether the panel is currently visible. */
  open: boolean;
  /** If true, the panel stays open on mouse leave. */
  pinned: boolean;
  /** Whether the user is currently hovering the edge/panel. */
  hovering: boolean;
}

export interface CockpitPanelsState {
  top: CockpitPanelState;
  bottom: CockpitPanelState;
  left: CockpitPanelState;
  right: CockpitPanelState;
}

export interface CockpitSettings {
  /** Animation duration in ms; 0 means reduced motion. */
  hoverDelay: number;
  /** Pixel threshold for drag-to-open. */
  edgeSensitivity: number;
  /** If true, non-pinned panels collapse on mouse leave. */
  autoCollapse: boolean;
  /** If true, minimize motion (instant transitions). */
  reducedMotion: boolean;
}

export interface CockpitState extends CockpitSettings {
  /** All panels are hidden and the graph fills the viewport. */
  firstPerson: boolean;
  /** Currently active cluster / system view. */
  activeCluster: string | null;
  /** FPS value emitted by the graph renderer. */
  fps: number;
  /** Last successful delta sync timestamp. */
  lastSyncAt: number | null;
  /** Whether the graph is currently being reloaded/updated. */
  syncing: boolean;
  /** Four cockpit panels. */
  panels: CockpitPanelsState;
}

export const COCKPIT_DEFAULT_SIZES: Record<CockpitPanelPosition, number> = {
  top: 64,
  bottom: 280,
  left: 320,
  right: 320,
};

export const COCKPIT_EDGE_SIZE = 24;
