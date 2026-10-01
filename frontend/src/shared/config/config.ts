// Centralized configuration module
// Imports settings from knowledge-graph.config.json at project root

import configData from "$config";

// Type definitions matching the JSON structure
export interface Config {
  backend: {
    recommendation: {
      depth: number;
      decay: number;
      top_n: number;
      alpha: number;
      beta: number;
      gamma: number;
      cache_ttl_seconds: number;
      task_delay_seconds: number;
      fallback_enabled: boolean;
      fallback_ttl_seconds: number;
      fallback_semantic_enabled: boolean;
      bfs_aggregation: string;
      bfs_normalize: boolean;
    };
    graph: {
      load_depth: number;
    };
    embedding: {
      similarity_limit: number;
    };
    asynq: {
      concurrency: number;
      queue_default: number;
      queue_max_len: number;
    };
  };
  frontend: {
    graph: {
      /** Max number of hub (top-degree) labels drawn at low zoom — shared 2D/3D rule */
      label_hub_count: number;
      /** Max BFS depth of the dependency-chain highlight (LINK-TYPES-1), shared 2D/3D */
      dependency_highlight_depth: number;
      /** 2D graph look: "classic" icons or "light" (GRAPH-LIGHT-1, decision 83) */
      style?: "classic" | "light";
      /** How many recommendations the light style draws on hover (decision 81) */
      recommendations_on_hover?: number;
      /** Light style: background motion only up to this many notes (GRAPH-LIGHT-1) */
      ambient_max_nodes?: number;
      "2d": {
        /** Node count below which CSS drop-shadows are rendered (performance) */
        shadows_threshold: number;
        /** Link count above which animated link drawing falls back to static (performance) */
        animated_links_threshold: number;
        /** Node count above which the gravity attraction system is disabled (performance) */
        gravity_nodes_threshold: number;
        /** Max world-unit radius for gravity pull between nodes */
        gravity_max_distance: number;
        /** Delay in milliseconds before node/link hover dimming and tooltips activate */
        hover_delay_ms: number;
        /** Node count above which animated visual effects (glow, particles, star corona, rings, nebula) are simplified */
        visual_fx_threshold: number;
        /** Target frame rate when the graph is stable and no user interaction occurs */
        idle_fps: number;
        /** Zoom level below which nodes are drawn with a simplified, cheaper renderer */
        lod_simplify_zoom: number;
        /** 2D adaptive fog-of-war settings */
        fog: {
          enabled: boolean;
          atmospheric: boolean;
          adaptive: boolean;
          radius_min: number;
          radius_max: number;
          fps_low: number;
          fps_high: number;
          warning_threshold: number;
          transition_ms: number;
          color: string;
          edge_feather: number;
        };
      };
      "3d": {
        /** FREEZE-3D-1 (решение 82): 3D is frozen until 2D is ready. */
        enabled: boolean;
        max_nodes: number;
        layout_provider: "d3" | "graph-service";
        fog: {
          presets: Record<
            "birth" | "nebula" | "deep-space",
            { density_initial: number; density_final: number }
          >;
          default_preset: "birth" | "nebula" | "deep-space";
        };
        performance: {
          fps_threshold_low: number;
          fps_threshold_high: number;
          low_fps_sample_count: number;
          starfield_counts: {
            high: number;
            medium: number;
            low: number;
          };
        };
      };
      anomaly: {
        reality_rift: {
          core_color: string;
          glow_color: string;
          crack_count_min: number;
          crack_count_max: number;
          deform_amount_min: number;
          deform_amount_max: number;
        };
        chromatic_maw: {
          tentacle_count_min: number;
          tentacle_count_max: number;
          hue_shift_base: number;
          hue_shift_range: number;
        };
        void_whisper: {
          particle_count_min: number;
          particle_count_max: number;
          hue_shift_base: number;
          hue_shift_range: number;
          connection_distance_threshold: number;
        };
        cosmic_abomination: {
          particle_count_min: number;
          particle_count_max: number;
          tentacle_count_min: number;
          tentacle_count_max: number;
          crack_count_min: number;
          crack_count_max: number;
        };
      };
    };
    achievements: {
      poll_interval_ms: number;
    };
  };
  ci_cd: Record<string, never>;
  nlp: {
    model_name: string;
    max_text_length: number;
  };
}

// Export the typed config
export const config: Config = configData as Config;

// Convenience exports for common values
export const graphConfig2D = config.frontend.graph["2d"];
export const graphLabelHubCount = config.frontend.graph.label_hub_count;
export const graphDependencyHighlightDepth = config.frontend.graph.dependency_highlight_depth;
export const graphStyle: "classic" | "light" =
  config.frontend.graph.style === "classic" ? "classic" : "light";
export const graphRecommendationsOnHover = config.frontend.graph.recommendations_on_hover ?? 4;
export const graphAmbientMaxNodes = config.frontend.graph.ambient_max_nodes ?? 500;
export const graphConfig3D = config.frontend.graph["3d"];
export const graphPerformanceConfig = config.frontend.graph["3d"].performance;
/** FREEZE-3D-1 (решение 82): 3D view frozen until 2D is ready. Function form
 *  so tests can stub the gate without re-mocking the whole config object. */
export function isGraph3DEnabled(): boolean {
  return config.frontend.graph["3d"].enabled;
}
export const anomalyConfig = config.frontend.graph.anomaly;
export const ACHIEVEMENT_POLL_INTERVAL_MS = config.frontend.achievements.poll_interval_ms;

export default config;
