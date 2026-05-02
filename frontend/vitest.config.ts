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
	define: {
		'import.meta.env.DEV': 'false',
		'import.meta.env.PROD': 'true',
		'import.meta.env.PUBLIC_API_URL': '"http://localhost:8080/api"',
		'import.meta.env.VITEST': 'true',
		'import.meta.env.MODE': '"test"'
	},
	resolve: {
		alias: [
			{ find: /^\$app\/environment$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/environment.ts') },
			{ find: /^\$app\/navigation$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/navigation.ts') },
			{ find: /^\$app\/stores$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/stores.ts') },
			{ find: /^\$lib/, replacement: path.resolve(__dirname, './src/lib') },
			{ find: /^\$config$/, replacement: path.resolve(__dirname, '../knowledge-graph.config.json') }
		],
		conditions: ['browser', 'default']
	},
	test: {
		environment: 'jsdom',
		globals: true,
		include: ['src/**/*.{test,spec}.{js,ts}'],
		exclude: [],
		setupFiles: ['./vitest-setup.ts'],
		testTimeout: 15000,
		coverage: {
			provider: 'v8',
			reporter: ['text', 'json', 'html'],
			include: [
				'src/lib/**/*.{ts,svelte}'
			],
			exclude: [
				'node_modules/',
				'vitest-setup.ts',
				'src/lib/mocks/**/*',
				'src/lib/**/*.spec.ts',
				'src/lib/**/*.test.ts',
				'src/lib/**/__mocks__/**/*',
				'**/*.d.ts',
				'src/lib/three/**/*'
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
