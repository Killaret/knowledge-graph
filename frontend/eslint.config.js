import js from "@eslint/js";
import tseslint from "typescript-eslint";
import svelteParser from "svelte-eslint-parser";
import globals from "globals";
import jsxA11y from "eslint-plugin-jsx-a11y";

export default [
  js.configs.recommended,
  ...tseslint.configs.recommended,
  jsxA11y.flatConfigs.recommended,
  {
    ignores: ["build/", ".svelte-kit/", "dist/"],
  },
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node,
      },
    },
  },
  {
    files: ["**/*.svelte"],
    languageOptions: {
      parser: svelteParser,
      parserOptions: {
        parser: tseslint.parser,
      },
    },
  },
  {
    files: ["**/*.cjs"],
    rules: {
      "@typescript-eslint/no-require-imports": "off",
    },
  },
  // FSD layer import boundaries (mirrors .windsurfrules). In these glob
  // patterns "**" also matches ".." segments, so relative-path escapes such
  // as "../../widgets/x" are covered together with the $alias form.
  {
    files: ["src/shared/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            "$entities/**",
            "$features/**",
            "$widgets/**",
            "$components/**",
            "../**/entities/**",
            "../**/features/**",
            "../**/widgets/**",
            "../**/components/**",
          ],
        },
      ],
    },
  },
  {
    files: ["src/entities/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            "$features/**",
            "$widgets/**",
            "$components/**",
            "../**/features/**",
            "../**/widgets/**",
            "../**/components/**",
          ],
        },
      ],
    },
  },
  {
    files: ["src/components/atoms/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            "$components/molecules/**",
            "$components/organisms/**",
            "$features/**",
            "$widgets/**",
            "../**/molecules/**",
            "../**/organisms/**",
            "../**/features/**",
            "../**/widgets/**",
          ],
        },
      ],
    },
  },
  {
    files: ["src/components/molecules/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: [
            "$components/organisms/**",
            "$features/**",
            "$widgets/**",
            "../**/organisms/**",
            "../**/features/**",
            "../**/widgets/**",
          ],
        },
      ],
    },
  },
  {
    files: ["src/features/**"],
    rules: {
      "no-restricted-imports": [
        "error",
        {
          patterns: ["$widgets/**", "../**/widgets/**"],
        },
      ],
    },
  },
  // src/routes/** stays unrestricted: the norm allows it to import any layer.
  {
    rules: {
      "@typescript-eslint/no-unused-vars": ["error", { argsIgnorePattern: "^_" }],
      "@typescript-eslint/no-explicit-any": "warn",
      "prefer-const": "error",
      "no-var": "error",
    },
  },
  {
    files: [
      "**/*.spec.ts",
      "**/*.test.ts",
      "**/__mocks__/**",
      "**/shared/test-utils/**",
      "**/test-canvas-mock.ts",
      "tests/**",
      "vitest-*.ts",
    ],
    rules: {
      "@typescript-eslint/no-explicit-any": "off",
      "@typescript-eslint/no-unused-vars": ["warn", { argsIgnorePattern: "^_" }],
    },
  },
];
