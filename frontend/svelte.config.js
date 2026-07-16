import adapter from '@sveltejs/adapter-node';
import path from 'path';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		runes: true,
		dev: process.env.NODE_ENV === 'development'
	},
	kit: {
		adapter: adapter(),
		alias: {
			'$shared': path.resolve('src/shared'),
			'$features': path.resolve('src/features'),
			'$components': path.resolve('src/components'),
			'$config': path.resolve('../knowledge-graph.config.json')
		}
	}
};

export default config;
