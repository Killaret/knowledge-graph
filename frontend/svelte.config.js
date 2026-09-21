import adapter from "@sveltejs/adapter-node";
import path from "path";

const dev = process.env.NODE_ENV === "development";

// Content-Security-Policy (CSP-1). The policy is emitted by SvelteKit so the
// hydration script nonce is computed by the framework; do not duplicate the
// header in nginx — two CSP headers mean the intersection of both policies.
const cspDirectives = {
  "default-src": ["self"],
  // No 'unsafe-inline'/'unsafe-eval' here: scripts are the primary XSS vector.
  // Vite dev mode needs 'unsafe-eval'; it must never reach the production policy.
  "script-src": dev ? ["self", "unsafe-eval"] : ["self"],
  // TD-CSP-STYLES (docs/BACKLOG.md): ~76 inline style attributes; nonce cannot
  // apply to attributes, so 'unsafe-inline' is a documented concession by the
  // owner (2026-09-08). Report-only measurement showed Chromium attributes
  // inline-style use to style-src itself (not only style-src-attr), so the
  // concession lives here. Rewriting the call sites is a separate backlog task.
  "style-src": ["self", "unsafe-inline"],
  "img-src": ["self", "data:"],
  "font-src": ["self"],
  // Same-origin only: the browser talks to /api and /graph-service/api on the
  // same host. Dev adds ws: for the Vite HMR websocket.
  "connect-src": dev ? ["self", "ws:"] : ["self"],
  // Mirrors X-Frame-Options: SAMEORIGIN (owner decision 2026-09-08); the future
  // embeddable public-graph widget keeps this from going stricter.
  "frame-ancestors": ["self"],
  "base-uri": ["self"],
  "form-action": ["self"],
  "object-src": ["none"],
};

/** @type {import('@sveltejs/kit').Config} */
const config = {
  compilerOptions: {
    runes: true,
    dev,
  },
  kit: {
    csp: {
      mode: "auto",
      // CSP-1 measurement: CSP_REPORT_ONLY=1 emits the policy as Report-Only
      // so violations can be collected without breaking the app.
      ...(process.env.CSP_REPORT_ONLY === "true"
        ? {
            reportOnly: {
              ...cspDirectives,
              // SvelteKit requires a report destination for Report-Only; the
              // endpoint need not exist for console-based collection.
              "report-uri": ["/api/csp-report"],
            },
            // Permissive enforcing policy so the app keeps working while the
            // target policy is only reported. Never used in production.
            directives: {
              "default-src": ["*", "data:", "blob:", "unsafe-inline", "unsafe-eval"],
            },
          }
        : { directives: cspDirectives }),
    },
    adapter: adapter(),
    alias: {
      $shared: path.resolve("src/shared"),
      $features: path.resolve("src/features"),
      $components: path.resolve("src/components"),
      $entities: path.resolve("src/entities"),
      $widgets: path.resolve("src/widgets"),
      $config: path.resolve("../knowledge-graph.config.json"),
    },
  },
};

export default config;
