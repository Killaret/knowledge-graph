// Конфигурация Vite для SvelteKit
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import path from 'path';

const graphServiceProxyTarget = (process.env.VITE_GRAPH_SERVICE_URL || 'http://127.0.0.1:9091').replace(/\/graph-service$/, '');

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
    port: 3000,
    proxy: {
      '/api/v1': {
        target: process.env.VITE_API_TARGET || 'http://127.0.0.1:9000',
        changeOrigin: true
      },
      '/graph-service/api': {
        target: graphServiceProxyTarget,
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/graph-service/, '')
      }
    }
  }
});
