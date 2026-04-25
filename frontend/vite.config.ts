// Конфигурация Vite для SvelteKit
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [sveltekit()],
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
        target: 'http://localhost:8081',
        changeOrigin: true,
        // Убираем префикс /api, чтобы не было /api/notes
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
});
