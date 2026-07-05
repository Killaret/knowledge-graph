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
			'$entities': path.resolve('src/entities'),
			'$features': path.resolve('src/features'),
			'$widgets': path.resolve('src/widgets')
		}
	}
};

export default config;
