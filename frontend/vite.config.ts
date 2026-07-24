// Конфигурация Vite для SvelteKit
import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";
import path from "path";

export default defineConfig({
  plugins: [sveltekit()],
  resolve: {
    alias: {
      // FSD aliases (must match svelte.config.js)
      $shared: path.resolve(__dirname, "src/shared"),
      $entities: path.resolve(__dirname, "src/entities"),
      $features: path.resolve(__dirname, "src/features"),
      $widgets: path.resolve(__dirname, "src/widgets"),
      $components: path.resolve(__dirname, "src/components"),
      // Алиас для проекта (корневой knowledge-graph.config.json)
      $config: path.resolve(__dirname, "../knowledge-graph.config.json"),
    },
  },
  build: {
    sourcemap: false,
    minify: "esbuild",
  },
  server: {
    // Vite dev server proxy - only used in dev mode, not in production SSR
    // Defaults target the dev stack (docker-compose.yml) services on host ports
    proxy: {
      "/api/v1": {
        target: process.env.VITE_API_TARGET || process.env.VITE_API_URL || "http://127.0.0.1:9000",
        changeOrigin: true,
      },
      "/graph-service/api": {
        target: process.env.VITE_GRAPH_SERVICE_URL || "http://127.0.0.1:9091",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/graph-service/, ""),
      },
    },
  },
});
