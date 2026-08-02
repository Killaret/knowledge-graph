export { cockpitStore } from "./model/cockpit.svelte";
export type {
  CockpitPanelPosition,
  CockpitPanelState,
  CockpitPanelsState,
  CockpitSettings,
  CockpitState,
} from "./model/cockpit.types";
export { COCKPIT_DEFAULT_SIZES, COCKPIT_EDGE_SIZE } from "./model/cockpit.types";
export { isInsideEdge, dragDistance, getPanelOffset } from "./lib/panel-geometry";
