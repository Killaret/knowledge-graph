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
  build: {
    // Manual chunks disabled - causing conflicts with external modules
    // rollupOptions: { output: { manualChunks: { ... } } }
  },
  server: {
    proxy: {
      // Прокси для запросов к API бэкенда
      // В Docker: используем имя сервиса backend:8080
      // Локально: используем localhost:8080 (через env VITE_API_TARGET)
      '/api': {
        target: process.env.VITE_API_TARGET || 'http://localhost:8080',
        changeOrigin: true,
        // Убираем префикс /api, чтобы не было /api/notes
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
});
