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
    setupFiles: ["./vitest-setup.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json", "html"],
      exclude: [
        "node_modules/",
        "vitest-setup.ts",
        "src/lib/mocks/**/*",
        "src/routes/**/*",
        "**/*.d.ts"
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
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZXN0LmNvbmZpZy50cyJdLAogICJzb3VyY2VzQ29udGVudCI6IFsiY29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2Rpcm5hbWUgPSBcIkQ6XFxcXGtub3dsZWRnZS1ncmFwaFxcXFxmcm9udGVuZFwiO2NvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9maWxlbmFtZSA9IFwiRDpcXFxca25vd2xlZGdlLWdyYXBoXFxcXGZyb250ZW5kXFxcXHZpdGVzdC5jb25maWcudHNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL0Q6L2tub3dsZWRnZS1ncmFwaC9mcm9udGVuZC92aXRlc3QuY29uZmlnLnRzXCI7aW1wb3J0IHsgZGVmaW5lQ29uZmlnIH0gZnJvbSAndml0ZXN0L2NvbmZpZyc7XHJcbmltcG9ydCB7IHN2ZWx0ZSB9IGZyb20gJ0BzdmVsdGVqcy92aXRlLXBsdWdpbi1zdmVsdGUnO1xyXG5pbXBvcnQgcGF0aCBmcm9tICdwYXRoJztcclxuaW1wb3J0IHsgZmlsZVVSTFRvUGF0aCB9IGZyb20gJ3VybCc7XHJcblxyXG5jb25zdCBfX2Rpcm5hbWUgPSBwYXRoLmRpcm5hbWUoZmlsZVVSTFRvUGF0aChpbXBvcnQubWV0YS51cmwpKTtcclxuXHJcbmV4cG9ydCBkZWZhdWx0IGRlZmluZUNvbmZpZyh7XHJcblx0cGx1Z2luczogW1xyXG5cdFx0c3ZlbHRlKHtcclxuXHRcdFx0aG90OiAhcHJvY2Vzcy5lbnYuVklURVNUXHJcblx0XHR9KVxyXG5cdF0sXHJcblx0cmVzb2x2ZToge1xyXG5cdFx0YWxpYXM6IFtcclxuXHRcdFx0eyBmaW5kOiAvXlxcJGFwcFxcL2Vudmlyb25tZW50JC8sIHJlcGxhY2VtZW50OiBwYXRoLnJlc29sdmUoX19kaXJuYW1lLCAnLi9zcmMvbGliL21vY2tzL2FwcC9lbnZpcm9ubWVudC50cycpIH0sXHJcblx0XHRcdHsgZmluZDogL15cXCRhcHBcXC9uYXZpZ2F0aW9uJC8sIHJlcGxhY2VtZW50OiBwYXRoLnJlc29sdmUoX19kaXJuYW1lLCAnLi9zcmMvbGliL21vY2tzL2FwcC9uYXZpZ2F0aW9uLnRzJykgfSxcclxuXHRcdFx0eyBmaW5kOiAvXlxcJGFwcFxcL3N0b3JlcyQvLCByZXBsYWNlbWVudDogcGF0aC5yZXNvbHZlKF9fZGlybmFtZSwgJy4vc3JjL2xpYi9tb2Nrcy9hcHAvc3RvcmVzLnRzJykgfSxcclxuXHRcdFx0eyBmaW5kOiAvXlxcJGxpYi8sIHJlcGxhY2VtZW50OiBwYXRoLnJlc29sdmUoX19kaXJuYW1lLCAnLi9zcmMvbGliJykgfSxcclxuXHRcdFx0eyBmaW5kOiAvXlxcJGNvbmZpZyQvLCByZXBsYWNlbWVudDogcGF0aC5yZXNvbHZlKF9fZGlybmFtZSwgJy4uL2tub3dsZWRnZS1ncmFwaC5jb25maWcuanNvbicpIH1cclxuXHRcdF0sXHJcblx0XHRjb25kaXRpb25zOiBbJ2Jyb3dzZXInLCAnZGVmYXVsdCddXHJcblx0fSxcclxuXHR0ZXN0OiB7XHJcblx0XHRlbnZpcm9ubWVudDogJ2pzZG9tJyxcclxuXHRcdGdsb2JhbHM6IHRydWUsXHJcblx0XHRpbmNsdWRlOiBbJ3NyYy8qKi8qLnt0ZXN0LHNwZWN9Lntqcyx0c30nXSxcclxuXHRcdHNldHVwRmlsZXM6IFsnLi92aXRlc3Qtc2V0dXAudHMnXSxcclxuXHRcdGNvdmVyYWdlOiB7XHJcblx0XHRcdHByb3ZpZGVyOiAndjgnLFxyXG5cdFx0XHRyZXBvcnRlcjogWyd0ZXh0JywgJ2pzb24nLCAnaHRtbCddLFxyXG5cdFx0XHRleGNsdWRlOiBbXHJcblx0XHRcdFx0J25vZGVfbW9kdWxlcy8nLFxyXG5cdFx0XHRcdCd2aXRlc3Qtc2V0dXAudHMnLFxyXG5cdFx0XHRcdCdzcmMvbGliL21vY2tzLyoqLyonLFxyXG5cdFx0XHRcdCdzcmMvcm91dGVzLyoqLyonLFxyXG5cdFx0XHRcdCcqKi8qLmQudHMnXHJcblx0XHRcdF0sXHJcblx0XHRcdHRocmVzaG9sZHM6IHtcclxuXHRcdFx0XHRsaW5lczogNTAsXHJcblx0XHRcdFx0ZnVuY3Rpb25zOiA1MCxcclxuXHRcdFx0XHRicmFuY2hlczogNDAsXHJcblx0XHRcdFx0c3RhdGVtZW50czogNTBcclxuXHRcdFx0fVxyXG5cdFx0fSxcclxuXHRcdHNlcnZlcjoge1xyXG5cdFx0XHRkZXBzOiB7XHJcblx0XHRcdFx0aW5saW5lOiBbL3N2ZWx0ZS9dXHJcblx0XHRcdH1cclxuXHRcdH1cclxuXHR9XHJcbn0pO1xyXG4iXSwKICAibWFwcGluZ3MiOiAiO0FBQTZRLFNBQVMsb0JBQW9CO0FBQzFTLFNBQVMsY0FBYztBQUN2QixPQUFPLFVBQVU7QUFDakIsU0FBUyxxQkFBcUI7QUFIdUksSUFBTSwyQ0FBMkM7QUFLdE4sSUFBTSxZQUFZLEtBQUssUUFBUSxjQUFjLHdDQUFlLENBQUM7QUFFN0QsSUFBTyx3QkFBUSxhQUFhO0FBQUEsRUFDM0IsU0FBUztBQUFBLElBQ1IsT0FBTztBQUFBLE1BQ04sS0FBSyxDQUFDLFFBQVEsSUFBSTtBQUFBLElBQ25CLENBQUM7QUFBQSxFQUNGO0FBQUEsRUFDQSxTQUFTO0FBQUEsSUFDUixPQUFPO0FBQUEsTUFDTixFQUFFLE1BQU0sd0JBQXdCLGFBQWEsS0FBSyxRQUFRLFdBQVcsb0NBQW9DLEVBQUU7QUFBQSxNQUMzRyxFQUFFLE1BQU0sdUJBQXVCLGFBQWEsS0FBSyxRQUFRLFdBQVcsbUNBQW1DLEVBQUU7QUFBQSxNQUN6RyxFQUFFLE1BQU0sbUJBQW1CLGFBQWEsS0FBSyxRQUFRLFdBQVcsK0JBQStCLEVBQUU7QUFBQSxNQUNqRyxFQUFFLE1BQU0sVUFBVSxhQUFhLEtBQUssUUFBUSxXQUFXLFdBQVcsRUFBRTtBQUFBLE1BQ3BFLEVBQUUsTUFBTSxjQUFjLGFBQWEsS0FBSyxRQUFRLFdBQVcsZ0NBQWdDLEVBQUU7QUFBQSxJQUM5RjtBQUFBLElBQ0EsWUFBWSxDQUFDLFdBQVcsU0FBUztBQUFBLEVBQ2xDO0FBQUEsRUFDQSxNQUFNO0FBQUEsSUFDTCxhQUFhO0FBQUEsSUFDYixTQUFTO0FBQUEsSUFDVCxTQUFTLENBQUMsOEJBQThCO0FBQUEsSUFDeEMsWUFBWSxDQUFDLG1CQUFtQjtBQUFBLElBQ2hDLFVBQVU7QUFBQSxNQUNULFVBQVU7QUFBQSxNQUNWLFVBQVUsQ0FBQyxRQUFRLFFBQVEsTUFBTTtBQUFBLE1BQ2pDLFNBQVM7QUFBQSxRQUNSO0FBQUEsUUFDQTtBQUFBLFFBQ0E7QUFBQSxRQUNBO0FBQUEsUUFDQTtBQUFBLE1BQ0Q7QUFBQSxNQUNBLFlBQVk7QUFBQSxRQUNYLE9BQU87QUFBQSxRQUNQLFdBQVc7QUFBQSxRQUNYLFVBQVU7QUFBQSxRQUNWLFlBQVk7QUFBQSxNQUNiO0FBQUEsSUFDRDtBQUFBLElBQ0EsUUFBUTtBQUFBLE1BQ1AsTUFBTTtBQUFBLFFBQ0wsUUFBUSxDQUFDLFFBQVE7QUFBQSxNQUNsQjtBQUFBLElBQ0Q7QUFBQSxFQUNEO0FBQ0QsQ0FBQzsiLAogICJuYW1lcyI6IFtdCn0K
