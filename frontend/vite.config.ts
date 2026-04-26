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
      // Все запросы, начинающиеся с /api, перенаправляются на http://localhost:8080
      // Например, /api/notes → http://localhost:8080/notes
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // Убираем префикс /api, чтобы не было /api/notes
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
});
