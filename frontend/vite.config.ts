// Конфигурация Vite для SvelteKit
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [sveltekit()],
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          // Three.js - heavy 3D library
          'three': ['three'],
          // D3 force simulation libraries
          'd3': ['d3-force', 'd3-force-3d'],
          // Vendor chunk for framework and utilities
          'vendor': ['svelte', 'ky', 'svelte-sonner']
        }
      }
    }
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
