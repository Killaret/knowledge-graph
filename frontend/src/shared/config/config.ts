// Centralized configuration module
// Imports settings from knowledge-graph.config.json at project root

import configData from '$config';

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
      max_nodes: number;
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
    test: {
      debounce_timeout_ms: number;
      max_retry_count: number;
      mock_goto_delay_ms: number;
    };
    graph: {
      '2d': {
        max_nodes: number;
        /** Node count below which CSS drop-shadows are rendered (performance) */
        shadows_threshold: number;
        /** Link count above which animated link drawing falls back to static (performance) */
        animated_links_threshold: number;
        /** Node count above which the gravity attraction system is disabled (performance) */
        gravity_nodes_threshold: number;
        /** Max world-unit radius for gravity pull between nodes */
        gravity_max_distance: number;
        /** Screen-space pixel radius of the ghost-node create button */
        ghost_node_radius: number;
        /** Delay in milliseconds before node/link hover dimming and tooltips activate */
        hover_delay_ms: number;
      };
      '3d': {
        max_nodes: number;
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
    api: {
      default_limit: number;
      link_limit: number;
    };
    achievements: {
      poll_interval_ms: number;
    };
  };
  ci_cd: {
    integration_test: {
      migrate_all: boolean;
      truncate_list: string[];
    };
  };
  nlp: {
    model_name: string;
    max_text_length: number;
  };
}

// Export the typed config
export const config: Config = configData as Config;

// Convenience exports for common values
export const graphConfig2D = config.frontend.graph['2d'];
export const graphConfig3D = config.frontend.graph['3d'];
export const anomalyConfig = config.frontend.graph.anomaly;
export const apiConfig = config.frontend.api;
export const testConfig = config.frontend.test;
export const ciCdConfig = config.ci_cd;
export const ACHIEVEMENT_POLL_INTERVAL_MS = config.frontend.achievements.poll_interval_ms;

export default config;