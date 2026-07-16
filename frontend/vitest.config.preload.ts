// Конфигурация Vitest для тестов PreloadService
import { defineConfig } from 'vitest/config';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
  plugins: [sveltekit()],
  test: {
    include: [
      'src/shared/services/**/*.test.ts',
      'src/shared/hooks/**/*.test.ts',
      'src/shared/stores/**/*.integration.test.ts'
    ],
    exclude: [
      'node_modules/**',
      '.svelte-kit/**',
      'dist/**',
      'src/shared/stores/auth.svelte.test.ts'
    ],
    environment: 'jsdom',
    setupFiles: ['./tests/setup/preload.setup.ts'],
    globals: true,
    coverage: {
      reporter: ['text', 'json', 'html'],
      include: [
        'src/shared/services/PreloadService.ts',
        'src/shared/hooks/usePreloadedData.ts'
      ],
      exclude: [
        '**/*.test.ts',
        '**/*.spec.ts',
        '**/node_modules/**'
      ]
    }
  }
});
