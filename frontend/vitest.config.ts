import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
	plugins: [
		svelte({
			compilerOptions: {
				dev: !process.env.VITEST
			}
		})
	],
	define: {
		'import.meta.env.DEV': 'false',
		'import.meta.env.PROD': 'true',
		'import.meta.env.PUBLIC_API_URL': '"http://localhost:8080/api"',
		// api/client.ts reads VITE_API_URL; must match MSW handlers in vitest-setup.ts
		'import.meta.env.VITE_API_URL': '"http://localhost:8080"',
		'import.meta.env.VITE_GRAPH_SERVICE_URL': '"http://localhost:9091"',
		'import.meta.env.VITEST': 'true',
		'import.meta.env.MODE': '"test"'
	},
	resolve: {
		alias: [
			{ find: /^\$app\/environment$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/environment.ts') },
			{ find: /^\$app\/navigation$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/navigation.ts') },
			{ find: /^\$app\/stores$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/stores.ts') },
			{ find: /^\$lib/, replacement: path.resolve(__dirname, './src/lib') },
			{ find: /^\$config$/, replacement: path.resolve(__dirname, '../knowledge-graph.config.json') },
			// FSD aliases (must match svelte.config.js)
			{ find: /^\$shared/, replacement: path.resolve(__dirname, './src/shared') },
			{ find: /^\$entities/, replacement: path.resolve(__dirname, './src/entities') },
			{ find: /^\$features/, replacement: path.resolve(__dirname, './src/features') },
			{ find: /^\$widgets/, replacement: path.resolve(__dirname, './src/widgets') }
		],
		conditions: ['browser', 'default']
	},
	test: {
		environment: 'jsdom',
		pool: 'threads',
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
				lines: 40,
				functions: 40,
				branches: 30,
				statements: 40
			},
			all: true
		},
		server: {
			deps: {
				inline: [/svelte/]
			}
		}
	}
});
