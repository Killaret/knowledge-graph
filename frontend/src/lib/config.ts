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
        shadows_threshold: number;
      };
      '3d': {
        max_nodes: number;
      };
    };
    api: {
      default_limit: number;
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
export const apiConfig = config.frontend.api;
export const testConfig = config.frontend.test;
export const ciCdConfig = config.ci_cd;

export default config;
