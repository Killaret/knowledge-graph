// vitest.config.ts
import { defineConfig } from "file:///D:/knowledge-graph/frontend/node_modules/vitest/dist/config.js";
import { svelte } from "file:///D:/knowledge-graph/frontend/node_modules/@sveltejs/vite-plugin-svelte/src/index.js";
import path from "path";
import { fileURLToPath } from "url";
const __vite_injected_original_import_meta_url = "file:///D:/knowledge-graph/frontend/vitest.config.ts";
const __dirname = path.dirname(fileURLToPath(__vite_injected_original_import_meta_url));
const vitest_config_default = defineConfig({
  plugins: [
    svelte({
      hot: !process.env.VITEST
    })
  ],
  resolve: {
    alias: [
      { find: /^\$app\/environment$/, replacement: path.resolve(__dirname, "./src/lib/mocks/app/environment.ts") },
      { find: /^\$app\/navigation$/, replacement: path.resolve(__dirname, "./src/lib/mocks/app/navigation.ts") },
      { find: /^\$app\/stores$/, replacement: path.resolve(__dirname, "./src/lib/mocks/app/stores.ts") },
      { find: /^\$lib/, replacement: path.resolve(__dirname, "./src/lib") },
      { find: /^\$config$/, replacement: path.resolve(__dirname, "../knowledge-graph.config.json") }
    ],
    conditions: ["browser", "default"]
  },
  test: {
    environment: "jsdom",
    globals: true,
    include: ["src/**/*.{test,spec}.{js,ts}"],
    exclude: [],
    setupFiles: ["./vitest-setup.ts"],
    testTimeout: 15e3,
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      include: [
        "src/lib/**/*.{ts,svelte}"
      ],
      exclude: [
        "node_modules/",
        "vitest-setup.ts",
        "src/lib/mocks/**/*",
        "src/lib/**/*.spec.ts",
        "src/lib/**/*.test.ts",
        "src/lib/**/__mocks__/**/*",
        "**/*.d.ts",
        "src/lib/three/**/*",
        "src/lib/components/GraphCanvas/**/*"
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
export {
  vitest_config_default as default
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZXN0LmNvbmZpZy50cyJdLAogICJzb3VyY2VzQ29udGVudCI6IFsiY29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2Rpcm5hbWUgPSBcIkQ6XFxcXGtub3dsZWRnZS1ncmFwaFxcXFxmcm9udGVuZFwiO2NvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9maWxlbmFtZSA9IFwiRDpcXFxca25vd2xlZGdlLWdyYXBoXFxcXGZyb250ZW5kXFxcXHZpdGVzdC5jb25maWcudHNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL0Q6L2tub3dsZWRnZS1ncmFwaC9mcm9udGVuZC92aXRlc3QuY29uZmlnLnRzXCI7aW1wb3J0IHsgZGVmaW5lQ29uZmlnIH0gZnJvbSAndml0ZXN0L2NvbmZpZyc7XHJcbmltcG9ydCB7IHN2ZWx0ZSB9IGZyb20gJ0BzdmVsdGVqcy92aXRlLXBsdWdpbi1zdmVsdGUnO1xyXG5pbXBvcnQgcGF0aCBmcm9tICdwYXRoJztcclxuaW1wb3J0IHsgZmlsZVVSTFRvUGF0aCB9IGZyb20gJ3VybCc7XHJcblxyXG5jb25zdCBfX2Rpcm5hbWUgPSBwYXRoLmRpcm5hbWUoZmlsZVVSTFRvUGF0aChpbXBvcnQubWV0YS51cmwpKTtcclxuXHJcbmV4cG9ydCBkZWZhdWx0IGRlZmluZUNvbmZpZyh7XHJcblx0cGx1Z2luczogW1xyXG5cdFx0c3ZlbHRlKHtcclxuXHRcdFx0aG90OiAhcHJvY2Vzcy5lbnYuVklURVNUXHJcblx0XHR9KVxyXG5cdF0sXHJcblx0cmVzb2x2ZToge1xyXG5cdFx0YWxpYXM6IFtcclxuXHRcdFx0eyBmaW5kOiAvXlxcJGFwcFxcL2Vudmlyb25tZW50JC8sIHJlcGxhY2VtZW50OiBwYXRoLnJlc29sdmUoX19kaXJuYW1lLCAnLi9zcmMvbGliL21vY2tzL2FwcC9lbnZpcm9ubWVudC50cycpIH0sXHJcblx0XHRcdHsgZmluZDogL15cXCRhcHBcXC9uYXZpZ2F0aW9uJC8sIHJlcGxhY2VtZW50OiBwYXRoLnJlc29sdmUoX19kaXJuYW1lLCAnLi9zcmMvbGliL21vY2tzL2FwcC9uYXZpZ2F0aW9uLnRzJykgfSxcclxuXHRcdFx0eyBmaW5kOiAvXlxcJGFwcFxcL3N0b3JlcyQvLCByZXBsYWNlbWVudDogcGF0aC5yZXNvbHZlKF9fZGlybmFtZSwgJy4vc3JjL2xpYi9tb2Nrcy9hcHAvc3RvcmVzLnRzJykgfSxcclxuXHRcdFx0eyBmaW5kOiAvXlxcJGxpYi8sIHJlcGxhY2VtZW50OiBwYXRoLnJlc29sdmUoX19kaXJuYW1lLCAnLi9zcmMvbGliJykgfSxcclxuXHRcdFx0eyBmaW5kOiAvXlxcJGNvbmZpZyQvLCByZXBsYWNlbWVudDogcGF0aC5yZXNvbHZlKF9fZGlybmFtZSwgJy4uL2tub3dsZWRnZS1ncmFwaC5jb25maWcuanNvbicpIH1cclxuXHRcdF0sXHJcblx0XHRjb25kaXRpb25zOiBbJ2Jyb3dzZXInLCAnZGVmYXVsdCddXHJcblx0fSxcclxuXHR0ZXN0OiB7XHJcblx0XHRlbnZpcm9ubWVudDogJ2pzZG9tJyxcclxuXHRcdGdsb2JhbHM6IHRydWUsXHJcblx0XHRpbmNsdWRlOiBbJ3NyYy8qKi8qLnt0ZXN0LHNwZWN9Lntqcyx0c30nXSxcclxuXHRcdGV4Y2x1ZGU6IFtdLFxyXG5cdFx0c2V0dXBGaWxlczogWycuL3ZpdGVzdC1zZXR1cC50cyddLFxyXG5cdFx0dGVzdFRpbWVvdXQ6IDE1MDAwLFxyXG5cdFx0Y292ZXJhZ2U6IHtcclxuXHRcdFx0cHJvdmlkZXI6ICd2OCcsXHJcblx0XHRcdHJlcG9ydGVyOiBbJ3RleHQnLCAnanNvbicsICdodG1sJ10sXHJcblx0XHRcdGluY2x1ZGU6IFtcclxuXHRcdFx0XHQnc3JjL2xpYi8qKi8qLnt0cyxzdmVsdGV9J1xyXG5cdFx0XHRdLFxyXG5cdFx0XHRleGNsdWRlOiBbXHJcblx0XHRcdFx0J25vZGVfbW9kdWxlcy8nLFxyXG5cdFx0XHRcdCd2aXRlc3Qtc2V0dXAudHMnLFxyXG5cdFx0XHRcdCdzcmMvbGliL21vY2tzLyoqLyonLFxyXG5cdFx0XHRcdCdzcmMvbGliLyoqLyouc3BlYy50cycsXHJcblx0XHRcdFx0J3NyYy9saWIvKiovKi50ZXN0LnRzJyxcclxuXHRcdFx0XHQnc3JjL2xpYi8qKi9fX21vY2tzX18vKiovKicsXHJcblx0XHRcdFx0JyoqLyouZC50cycsXHJcblx0XHRcdFx0J3NyYy9saWIvdGhyZWUvKiovKicsXHJcblx0XHRcdFx0J3NyYy9saWIvY29tcG9uZW50cy9HcmFwaENhbnZhcy8qKi8qJ1xyXG5cdFx0XHRdLFxyXG5cdFx0XHR0aHJlc2hvbGRzOiB7XHJcblx0XHRcdFx0bGluZXM6IDUwLFxyXG5cdFx0XHRcdGZ1bmN0aW9uczogNTAsXHJcblx0XHRcdFx0YnJhbmNoZXM6IDQwLFxyXG5cdFx0XHRcdHN0YXRlbWVudHM6IDUwXHJcblx0XHRcdH1cclxuXHRcdH0sXHJcblx0XHRzZXJ2ZXI6IHtcclxuXHRcdFx0ZGVwczoge1xyXG5cdFx0XHRcdGlubGluZTogWy9zdmVsdGUvXVxyXG5cdFx0XHR9XHJcblx0XHR9XHJcblx0fVxyXG59KTtcclxuIl0sCiAgIm1hcHBpbmdzIjogIjtBQUE2USxTQUFTLG9CQUFvQjtBQUMxUyxTQUFTLGNBQWM7QUFDdkIsT0FBTyxVQUFVO0FBQ2pCLFNBQVMscUJBQXFCO0FBSHVJLElBQU0sMkNBQTJDO0FBS3ROLElBQU0sWUFBWSxLQUFLLFFBQVEsY0FBYyx3Q0FBZSxDQUFDO0FBRTdELElBQU8sd0JBQVEsYUFBYTtBQUFBLEVBQzNCLFNBQVM7QUFBQSxJQUNSLE9BQU87QUFBQSxNQUNOLEtBQUssQ0FBQyxRQUFRLElBQUk7QUFBQSxJQUNuQixDQUFDO0FBQUEsRUFDRjtBQUFBLEVBQ0EsU0FBUztBQUFBLElBQ1IsT0FBTztBQUFBLE1BQ04sRUFBRSxNQUFNLHdCQUF3QixhQUFhLEtBQUssUUFBUSxXQUFXLG9DQUFvQyxFQUFFO0FBQUEsTUFDM0csRUFBRSxNQUFNLHVCQUF1QixhQUFhLEtBQUssUUFBUSxXQUFXLG1DQUFtQyxFQUFFO0FBQUEsTUFDekcsRUFBRSxNQUFNLG1CQUFtQixhQUFhLEtBQUssUUFBUSxXQUFXLCtCQUErQixFQUFFO0FBQUEsTUFDakcsRUFBRSxNQUFNLFVBQVUsYUFBYSxLQUFLLFFBQVEsV0FBVyxXQUFXLEVBQUU7QUFBQSxNQUNwRSxFQUFFLE1BQU0sY0FBYyxhQUFhLEtBQUssUUFBUSxXQUFXLGdDQUFnQyxFQUFFO0FBQUEsSUFDOUY7QUFBQSxJQUNBLFlBQVksQ0FBQyxXQUFXLFNBQVM7QUFBQSxFQUNsQztBQUFBLEVBQ0EsTUFBTTtBQUFBLElBQ0wsYUFBYTtBQUFBLElBQ2IsU0FBUztBQUFBLElBQ1QsU0FBUyxDQUFDLDhCQUE4QjtBQUFBLElBQ3hDLFNBQVMsQ0FBQztBQUFBLElBQ1YsWUFBWSxDQUFDLG1CQUFtQjtBQUFBLElBQ2hDLGFBQWE7QUFBQSxJQUNiLFVBQVU7QUFBQSxNQUNULFVBQVU7QUFBQSxNQUNWLFVBQVUsQ0FBQyxRQUFRLFFBQVEsTUFBTTtBQUFBLE1BQ2pDLFNBQVM7QUFBQSxRQUNSO0FBQUEsTUFDRDtBQUFBLE1BQ0EsU0FBUztBQUFBLFFBQ1I7QUFBQSxRQUNBO0FBQUEsUUFDQTtBQUFBLFFBQ0E7QUFBQSxRQUNBO0FBQUEsUUFDQTtBQUFBLFFBQ0E7QUFBQSxRQUNBO0FBQUEsUUFDQTtBQUFBLE1BQ0Q7QUFBQSxNQUNBLFlBQVk7QUFBQSxRQUNYLE9BQU87QUFBQSxRQUNQLFdBQVc7QUFBQSxRQUNYLFVBQVU7QUFBQSxRQUNWLFlBQVk7QUFBQSxNQUNiO0FBQUEsSUFDRDtBQUFBLElBQ0EsUUFBUTtBQUFBLE1BQ1AsTUFBTTtBQUFBLFFBQ0wsUUFBUSxDQUFDLFFBQVE7QUFBQSxNQUNsQjtBQUFBLElBQ0Q7QUFBQSxFQUNEO0FBQ0QsQ0FBQzsiLAogICJuYW1lcyI6IFtdCn0K
