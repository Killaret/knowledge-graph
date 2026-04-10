import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import pluginSvelte from 'eslint-plugin-svelte';
import globals from 'globals';

export default [
	js.configs.recommended,
	...tseslint.configs.recommended,
	{
		ignores: ['build/', '.svelte-kit/', 'dist/']
	},
	{
		languageOptions: {
			globals: {
				...globals.browser,
				...globals.node
			}
		}
	},
	{
		files: ['**/*.svelte'],
		languageOptions: {
			parser: pluginSvelte.parser,
			parserOptions: {
				parser: tseslint.parser
			},
			globals: {
				...globals.browser,
				...globals.node
			}
		}
	},
	{
		rules: {
			'@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
			'@typescript-eslint/no-explicit-any': 'warn',
			'prefer-const': 'error',
			'no-var': 'error'
		}
	}
];
