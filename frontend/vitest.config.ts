import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
	plugins: [
		svelte({
			hot: !process.env.VITEST
		})
	],
	resolve: {
		alias: [
			{ find: /^\$app\/environment$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/environment.ts') },
			{ find: /^\$app\/navigation$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/navigation.ts') },
			{ find: /^\$app\/stores$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/stores.ts') },
			{ find: /^\$lib/, replacement: path.resolve(__dirname, './src/lib') }
		],
		conditions: ['browser', 'default']
	},
	test: {
		environment: 'jsdom',
		globals: true,
		include: ['src/**/*.{test,spec}.{js,ts}'],
		setupFiles: ['./vitest-setup.ts'],
		coverage: {
			provider: 'v8',
			reporter: ['text', 'json', 'html'],
			exclude: [
				'node_modules/',
				'vitest-setup.ts',
				'src/lib/mocks/**/*',
				'src/routes/**/*',
				'**/*.d.ts'
			],
			thresholds: {
				lines: 50,
				functions: 50,
				branches: 40,
				statements: 50
			}
		},
		server: {
			deps: {
				inline: [/svelte/]
			}
		}
	}
});
