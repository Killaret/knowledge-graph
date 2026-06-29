// Конфигурация Vite для SvelteKit
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
  plugins: [sveltekit()],
  resolve: {
    alias: {
      // Алиас для проекта (корневой knowledge-graph.config.json)
      '$config': path.resolve(__dirname, '../knowledge-graph.config.json')
    }
  },
  build: {
    sourcemap: false,
    minify: 'esbuild'
  },
  server: {
    // Vite dev server proxy - only used in dev mode, not in production SSR
    proxy: {
      '/api/v1': {
        target: process.env.VITE_API_TARGET || process.env.VITE_API_URL || 'http://127.0.0.1:8085',
        changeOrigin: true
      },
      '/graph-service/api': {
        target: process.env.VITE_GRAPH_SERVICE_URL || 'http://127.0.0.1:9092',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/graph-service/, '')
      }
    }
  }
});
