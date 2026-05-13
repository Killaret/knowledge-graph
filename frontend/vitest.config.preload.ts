// Конфигурация Vitest для тестов PreloadService
import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
  plugins: [sveltekit()],
  test: {
    include: [
      'src/lib/services/**/*.test.ts',
      'src/lib/hooks/**/*.test.ts',
      'src/lib/stores/**/*.integration.test.ts'
    ],
    exclude: [
      'node_modules/**',
      '.svelte-kit/**',
      'dist/**',
      'src/lib/stores/auth.svelte.test.ts'
    ],
    environment: 'jsdom',
    setupFiles: ['./tests/setup/preload.setup.ts'],
    globals: true,
    coverage: {
      reporter: ['text', 'json', 'html'],
      include: [
        'src/lib/services/PreloadService.ts',
        'src/lib/hooks/usePreloadedData.ts'
      ],
      exclude: [
        '**/*.test.ts',
        '**/*.spec.ts',
        '**/node_modules/**'
      ]
    }
  }
});
