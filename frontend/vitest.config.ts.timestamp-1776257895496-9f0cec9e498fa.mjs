// vitest.config.ts
import { defineConfig } from "file:///D:/knowledge-graph/frontend/node_modules/vitest/dist/config.js";
import { svelte } from "file:///D:/knowledge-graph/frontend/node_modules/@sveltejs/vite-plugin-svelte/src/index.js";
import path from "path";
import { fileURLToPath } from "url";
const __vite_injected_original_import_meta_url = "file:///D:/knowledge-graph/frontend/vitest.config.ts";
const __dirname = path.dirname(fileURLToPath(__vite_injected_original_import_meta_url));
const vitest_config_default = defineConfig({
  plugins: [svelte({ hot: !process.env.VITEST })],
  resolve: {
    alias: [
      { find: /^\$app\/environment$/, replacement: path.resolve(__dirname, "./src/lib/mocks/app/environment.ts") },
      { find: /^\$app\/navigation$/, replacement: path.resolve(__dirname, "./src/lib/mocks/app/navigation.ts") },
      { find: /^\$app\/stores$/, replacement: path.resolve(__dirname, "./src/lib/mocks/app/stores.ts") }
    ]
  },
  test: {
    environment: "jsdom",
    globals: true,
    include: ["src/**/*.{test,spec}.{js,ts}"],
    setupFiles: ["./vitest-setup.ts"],
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
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZXN0LmNvbmZpZy50cyJdLAogICJzb3VyY2VzQ29udGVudCI6IFsiY29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2Rpcm5hbWUgPSBcIkQ6XFxcXGtub3dsZWRnZS1ncmFwaFxcXFxmcm9udGVuZFwiO2NvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9maWxlbmFtZSA9IFwiRDpcXFxca25vd2xlZGdlLWdyYXBoXFxcXGZyb250ZW5kXFxcXHZpdGVzdC5jb25maWcudHNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL0Q6L2tub3dsZWRnZS1ncmFwaC9mcm9udGVuZC92aXRlc3QuY29uZmlnLnRzXCI7aW1wb3J0IHsgZGVmaW5lQ29uZmlnIH0gZnJvbSAndml0ZXN0L2NvbmZpZyc7XHJcbmltcG9ydCB7IHN2ZWx0ZSB9IGZyb20gJ0BzdmVsdGVqcy92aXRlLXBsdWdpbi1zdmVsdGUnO1xyXG5pbXBvcnQgcGF0aCBmcm9tICdwYXRoJztcclxuaW1wb3J0IHsgZmlsZVVSTFRvUGF0aCB9IGZyb20gJ3VybCc7XHJcblxyXG5jb25zdCBfX2Rpcm5hbWUgPSBwYXRoLmRpcm5hbWUoZmlsZVVSTFRvUGF0aChpbXBvcnQubWV0YS51cmwpKTtcclxuXHJcbmV4cG9ydCBkZWZhdWx0IGRlZmluZUNvbmZpZyh7XHJcblx0cGx1Z2luczogW3N2ZWx0ZSh7IGhvdDogIXByb2Nlc3MuZW52LlZJVEVTVCB9KV0sXHJcblx0cmVzb2x2ZToge1xyXG5cdFx0YWxpYXM6IFtcclxuXHRcdFx0eyBmaW5kOiAvXlxcJGFwcFxcL2Vudmlyb25tZW50JC8sIHJlcGxhY2VtZW50OiBwYXRoLnJlc29sdmUoX19kaXJuYW1lLCAnLi9zcmMvbGliL21vY2tzL2FwcC9lbnZpcm9ubWVudC50cycpIH0sXHJcblx0XHRcdHsgZmluZDogL15cXCRhcHBcXC9uYXZpZ2F0aW9uJC8sIHJlcGxhY2VtZW50OiBwYXRoLnJlc29sdmUoX19kaXJuYW1lLCAnLi9zcmMvbGliL21vY2tzL2FwcC9uYXZpZ2F0aW9uLnRzJykgfSxcclxuXHRcdFx0eyBmaW5kOiAvXlxcJGFwcFxcL3N0b3JlcyQvLCByZXBsYWNlbWVudDogcGF0aC5yZXNvbHZlKF9fZGlybmFtZSwgJy4vc3JjL2xpYi9tb2Nrcy9hcHAvc3RvcmVzLnRzJykgfVxyXG5cdFx0XVxyXG5cdH0sXHJcblx0dGVzdDoge1xyXG5cdFx0ZW52aXJvbm1lbnQ6ICdqc2RvbScsXHJcblx0XHRnbG9iYWxzOiB0cnVlLFxyXG5cdFx0aW5jbHVkZTogWydzcmMvKiovKi57dGVzdCxzcGVjfS57anMsdHN9J10sXHJcblx0XHRzZXR1cEZpbGVzOiBbJy4vdml0ZXN0LXNldHVwLnRzJ10sXHJcblx0XHRzZXJ2ZXI6IHtcclxuXHRcdFx0ZGVwczoge1xyXG5cdFx0XHRcdGlubGluZTogWy9zdmVsdGUvXVxyXG5cdFx0XHR9XHJcblx0XHR9XHJcblx0fVxyXG59KTtcclxuIl0sCiAgIm1hcHBpbmdzIjogIjtBQUE2USxTQUFTLG9CQUFvQjtBQUMxUyxTQUFTLGNBQWM7QUFDdkIsT0FBTyxVQUFVO0FBQ2pCLFNBQVMscUJBQXFCO0FBSHVJLElBQU0sMkNBQTJDO0FBS3ROLElBQU0sWUFBWSxLQUFLLFFBQVEsY0FBYyx3Q0FBZSxDQUFDO0FBRTdELElBQU8sd0JBQVEsYUFBYTtBQUFBLEVBQzNCLFNBQVMsQ0FBQyxPQUFPLEVBQUUsS0FBSyxDQUFDLFFBQVEsSUFBSSxPQUFPLENBQUMsQ0FBQztBQUFBLEVBQzlDLFNBQVM7QUFBQSxJQUNSLE9BQU87QUFBQSxNQUNOLEVBQUUsTUFBTSx3QkFBd0IsYUFBYSxLQUFLLFFBQVEsV0FBVyxvQ0FBb0MsRUFBRTtBQUFBLE1BQzNHLEVBQUUsTUFBTSx1QkFBdUIsYUFBYSxLQUFLLFFBQVEsV0FBVyxtQ0FBbUMsRUFBRTtBQUFBLE1BQ3pHLEVBQUUsTUFBTSxtQkFBbUIsYUFBYSxLQUFLLFFBQVEsV0FBVywrQkFBK0IsRUFBRTtBQUFBLElBQ2xHO0FBQUEsRUFDRDtBQUFBLEVBQ0EsTUFBTTtBQUFBLElBQ0wsYUFBYTtBQUFBLElBQ2IsU0FBUztBQUFBLElBQ1QsU0FBUyxDQUFDLDhCQUE4QjtBQUFBLElBQ3hDLFlBQVksQ0FBQyxtQkFBbUI7QUFBQSxJQUNoQyxRQUFRO0FBQUEsTUFDUCxNQUFNO0FBQUEsUUFDTCxRQUFRLENBQUMsUUFBUTtBQUFBLE1BQ2xCO0FBQUEsSUFDRDtBQUFBLEVBQ0Q7QUFDRCxDQUFDOyIsCiAgIm5hbWVzIjogW10KfQo=
