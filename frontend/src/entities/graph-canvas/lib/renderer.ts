/**
 * Canvas renderer public API for GraphCanvas.
 */
export type { SimulationNode, SimulationLink } from "./types";
export type { BlackHoleState } from "./black-hole";
export type { GhostNodeState } from "./ghost-node";

export { drawRealityRift } from "$shared/lib/graph/renderer/anomalies/reality-rift";
export { drawChromaticMaw } from "$shared/lib/graph/renderer/anomalies/chromatic-maw";
export { drawVoidWhisper } from "$shared/lib/graph/renderer/anomalies/void-whisper";
export { drawCosmicAbomination } from "$shared/lib/graph/renderer/anomalies/cosmic-abomination";

export * from "./renderer-utils";
export * from "./node-renderers";
export * from "./link-renderers";
export * from "./background";
export * from "./node-registration";
export * from "./renderer-orchestrator";
