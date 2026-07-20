import { defineConfig } from "vitest/config";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import path from "path";
import { fileURLToPath } from "url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [
    svelte({
      compilerOptions: {
        dev: !process.env.VITEST,
      },
    }),
  ],
  define: {
    "import.meta.env.DEV": "false",
    "import.meta.env.PROD": "true",
    "import.meta.env.PUBLIC_API_URL": '"http://localhost:8080/api"',
    // api/client.ts reads VITE_API_URL; must match MSW handlers in vitest-setup.ts
    "import.meta.env.VITE_API_URL": '"http://localhost:8080"',
    "import.meta.env.VITE_GRAPH_SERVICE_URL": '"http://localhost:9091"',
    "import.meta.env.VITEST": "true",
    "import.meta.env.MODE": '"test"',
  },
  resolve: {
    alias: [
      {
        find: /^\$app\/environment$/,
        replacement: path.resolve(
          __dirname,
          "./src/shared/mocks/app/environment.ts",
        ),
      },
      {
        find: /^\$app\/navigation$/,
        replacement: path.resolve(
          __dirname,
          "./src/shared/mocks/app/navigation.ts",
        ),
      },
      {
        find: /^\$app\/stores$/,
        replacement: path.resolve(
          __dirname,
          "./src/shared/mocks/app/stores.ts",
        ),
      },
      {
        find: /^\$config$/,
        replacement: path.resolve(__dirname, "../knowledge-graph.config.json"),
      },
      // FSD aliases (must match svelte.config.js)
      {
        find: /^\$shared/,
        replacement: path.resolve(__dirname, "./src/shared"),
      },
      {
        find: /^\$features/,
        replacement: path.resolve(__dirname, "./src/features"),
      },
      {
        find: /^\$components/,
        replacement: path.resolve(__dirname, "./src/components"),
      },
    ],
    conditions: ["browser", "default"],
  },
  test: {
    environment: "jsdom",
    pool: "threads",
    globals: true,
    include: ["src/**/*.{test,spec}.{js,ts}"],
    exclude: [],
    setupFiles: ["./vitest-setup.ts"],
    testTimeout: 15000,
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      include: [
        "src/shared/**/*.{ts,svelte}",
        "src/features/**/*.{ts,svelte}",
        "src/components/**/*.{ts,svelte}",
      ],
      exclude: [
        "node_modules/",
        "vitest-setup.ts",
        ".svelte-kit/**",
        "dist/**",
        "src/shared/mocks/**/*",
        "src/shared/test-utils/**/*",
        "src/**/*.spec.ts",
        "src/**/*.test.ts",
        "src/**/__mocks__/**/*",
        "**/*.d.ts",
      ],
      thresholds: {
        lines: 60,
        functions: 55,
        branches: 60,
        statements: 60,
      },
      all: true,
    },
    server: {
      deps: {
        inline: [/svelte/],
      },
    },
  },
});
