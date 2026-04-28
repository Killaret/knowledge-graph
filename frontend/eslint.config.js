import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import svelteParser from 'svelte-eslint-parser';
import globals from 'globals';
import jsxA11y from 'eslint-plugin-jsx-a11y';

export default [
	js.configs.recommended,
	...tseslint.configs.recommended,
	jsxA11y.flatConfigs.recommended,
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
			parser: svelteParser,
			parserOptions: {
				parser: tseslint.parser
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
