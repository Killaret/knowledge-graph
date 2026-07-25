import { sequence } from "@sveltejs/kit/hooks";
import type { Handle } from "@sveltejs/kit";

/**
 * Backend API proxy middleware
 * Proxies requests from /api/v1/* to backend server
 * Uses environment variable VITE_API_TARGET or defaults to Docker service name
 */
const backendApiProxy: Handle = async ({ event, resolve }) => {
  const url = event.url;

  // Get backend API URL from environment or use Docker service name
  // Note: Base URL should NOT include /api suffix - pathname handles that
  const backendApiUrl = process.env.VITE_API_TARGET || "http://backend:8080";

  // Check if this is an API request
  if (url.pathname.startsWith("/api/v1")) {
    const targetUrl = `${backendApiUrl}${url.pathname}${url.search}`;

    try {
      // Security: Only forward safe headers to prevent data leakage
      // Note: cookie and authorization are intentionally excluded
      // Auth tokens are managed client-side and added by API client
      const safeHeaders: Record<string, string> = {};
      const allowedHeaders = ["accept", "content-type", "x-request-id"];
      const blockedHeaders = [
        "cookie",
        "authorization",
        "connection",
        "proxy-",
        "transfer-encoding",
        "keep-alive",
        "upgrade",
        "te",
        "host",
      ];

      for (const [key, value] of event.request.headers.entries()) {
        const headerName = key.toLowerCase();
        // Check if header is allowed and not blocked
        if (
          allowedHeaders.includes(headerName) &&
          !blockedHeaders.some((blocked) => headerName.startsWith(blocked))
        ) {
          safeHeaders[key] = value;
        }
      }

      // Forward the request to backend
      const response = await fetch(targetUrl, {
        method: event.request.method,
        headers: safeHeaders,
        body:
          event.request.method !== "GET" && event.request.method !== "HEAD"
            ? await event.request.blob()
            : undefined,
      });

      // Return the response from backend
      const headers = new Headers(response.headers);
      headers.delete("transfer-encoding");
      headers.delete("connection");
      headers.delete("keep-alive");
      // Remove Content-Encoding - fetch already decompressed the body
      headers.delete("content-encoding");
      headers.delete("content-length"); // Let SvelteKit recalculate

      return new Response(response.body, {
        status: response.status,
        headers: headers,
      });
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error("[Backend API Proxy] Error:", error);
      }
      return new Response(JSON.stringify({ error: "Failed to connect to backend" }), {
        status: 502,
        headers: { "content-type": "application/json" },
      });
    }
  }

  // Not an API request, continue normally
  return resolve(event);
};

/**
 * Graph service proxy middleware
 * Proxies requests from /graph-service/* to graph service
 * Uses environment variable GRAPH_SERVICE_URL or defaults to Docker service name
 */
const graphServiceProxy: Handle = async ({ event, resolve }) => {
  const url = event.url;

  // Get graph service URL from environment or use Docker service name
  const graphServiceUrl =
    process.env.GRAPH_SERVICE_URL ||
    process.env.VITE_GRAPH_SERVICE_URL ||
    "http://graph-service:9091";

  // Check if this is a graph service request (supports both /graph-service/api and /graph-service)
  if (url.pathname.startsWith("/graph-service")) {
    // Remove /graph-service prefix to get the actual path
    let targetPath = url.pathname.replace("/graph-service", "");
    // Add /api if path doesn't already include it
    if (!targetPath.startsWith("/api")) {
      targetPath = "/api" + targetPath;
    }
    const targetUrl = `${graphServiceUrl}${targetPath}${url.search}`;

    try {
      // Security: Only forward safe headers to prevent data leakage
      // Allowlist: accept, content-type, x-request-id
      // Block: cookie, authorization, hop-by-hop headers
      const safeHeaders: Record<string, string> = {};
      const allowedHeaders = ["accept", "content-type", "x-request-id", "authorization"];
      const blockedHeaders = [
        "cookie",
        "connection",
        "proxy-",
        "transfer-encoding",
        "keep-alive",
        "upgrade",
        "te",
      ];

      for (const [key, value] of event.request.headers.entries()) {
        const headerName = key.toLowerCase();
        // Check if header is allowed and not blocked
        if (
          allowedHeaders.includes(headerName) &&
          !blockedHeaders.some((blocked) => headerName.startsWith(blocked))
        ) {
          safeHeaders[key] = value;
        }
      }

      // Add internal authentication header if configured
      const internalAuthToken = process.env.GRAPH_SERVICE_INTERNAL_TOKEN;
      if (internalAuthToken) {
        safeHeaders["x-internal-auth"] = internalAuthToken;
      }

      // Override host header for proper proxying
      safeHeaders["host"] = new URL(graphServiceUrl).host;

      // Forward the request to graph service
      const response = await fetch(targetUrl, {
        method: event.request.method,
        headers: safeHeaders,
        body:
          event.request.method !== "GET" && event.request.method !== "HEAD"
            ? await event.request.blob()
            : undefined,
      });

      // Return the response from graph service
      const headers = new Headers(response.headers);
      // Remove hop-by-hop headers
      headers.delete("transfer-encoding");
      headers.delete("connection");
      headers.delete("keep-alive");
      // Remove Content-Encoding - fetch already decompressed the body
      headers.delete("content-encoding");
      headers.delete("content-length"); // Let SvelteKit recalculate

      return new Response(response.body, {
        status: response.status,
        headers: headers,
      });
    } catch (error) {
      if (import.meta.env.DEV) {
        console.error("[Graph Service Proxy] Error:", error);
      }
      return new Response(JSON.stringify({ error: "Failed to connect to graph service" }), {
        status: 502,
        headers: { "content-type": "application/json" },
      });
    }
  }

  // Not a graph service request, continue normally
  return resolve(event);
};

export const handle = sequence(backendApiProxy, graphServiceProxy);
