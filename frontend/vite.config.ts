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
  }
});
