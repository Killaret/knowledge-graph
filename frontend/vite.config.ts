// Конфигурация Vite для SvelteKit
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import path from 'path';

export default defineConfig({
  plugins: [sveltekit()],
  resolve: {
    alias: {
      // Алиас для конфигурации проекта (корневой knowledge-graph.config.json)
      '$config': path.resolve(__dirname, '../knowledge-graph.config.json')
    }
  },
  build: {},
  server: {
    proxy: {
      '/v1': {
        target: process.env.VITE_API_TARGET || 'http://localhost:8080/api',
        changeOrigin: true
      }
    }
  }
});
