import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
	plugins: [
		svelte({
			hot: !process.env.VITEST,
			compilerOptions: () => ({
				generate: 'client'
			})
		})
	],
	resolve: {
		alias: [
			{ find: /^\$app\/environment$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/environment.ts') },
			{ find: /^\$app\/navigation$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/navigation.ts') },
			{ find: /^\$app\/stores$/, replacement: path.resolve(__dirname, './src/lib/mocks/app/stores.ts') }
		],
		conditions: ['browser', 'default']
	},
	test: {
		environment: 'jsdom',
		globals: true,
		include: ['src/**/*.{test,spec}.{js,ts}'],
		setupFiles: ['./vitest-setup.ts'],
		server: {
			deps: {
				inline: [/svelte/]
			}
		}
	}
});
