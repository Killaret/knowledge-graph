/**
 * SYNC-1 stage C: SSE subscription to graph-service `GET /api/v1/graph/events`.
 *
 * The server pushes a bare "graph changed" signal scoped to the user's own
 * events; on it the client fetches /graph/delta itself. Polling stays the
 * fallback — a dead stream only means the 30 s interval works again.
 */
import { browser } from "$app/environment";
import { accessToken } from "$shared/stores/auth.svelte";

export interface GraphEventsHandle {
  /** Permanently stop the subscription (no further reconnects). */
  stop(): void;
  /** True while the EventSource reports OPEN. */
  isConnected(): boolean;
}

/** Minimal EventSource surface used here — injectable for tests. */
export interface EventSourceLike {
  readonly readyState: number;
  addEventListener(type: string, listener: (event: MessageEvent<string>) => void): void;
  close(): void;
  onerror: ((event: Event) => void) | null;
  onopen: ((event: Event) => void) | null;
}

interface GraphEventsOptions {
  /** Called once per burst of server-side change events. */
  onGraphChanged: () => void;
  /** Burst coalescing window for change events (default 400 ms). */
  debounceMs?: number;
  /** Backoff before recreating a dead stream, doubles up to max (2s/30s). */
  reconnectMinMs?: number;
  reconnectMaxMs?: number;
  /** Test hook — replaces `new EventSource(url)`. */
  createSource?: (url: string) => EventSourceLike;
  /** Test hook — replaces accessToken() lookup. */
  getToken?: () => string | null;
}

/** Base URL of the graph-service HTTP API — mirrors getGraphApi() in api/graph.ts. */
export function graphEventsBaseUrl(): string {
  const isTest = typeof process !== "undefined" && process.env?.VITEST === "true";
  if (isTest) return "http://localhost:9091/api";
  if (import.meta.env.DEV) return "/graph-service/api";
  const configured = import.meta.env.VITE_GRAPH_SERVICE_URL || "/graph-service";
  return configured.endsWith("/api") ? configured : `${configured}/api`;
}

export function subscribeGraphEvents(options: GraphEventsOptions): GraphEventsHandle {
  const debounceMs = options.debounceMs ?? 400;
  const reconnectMin = options.reconnectMinMs ?? 2000;
  const reconnectMax = options.reconnectMaxMs ?? 30000;
  const getToken = options.getToken ?? accessToken;
  // No EventSource in the environment (SSR, partial test DOMs) → inert handle.
  const sourceAvailable = options.createSource !== undefined || typeof EventSource === "function";
  const createSource =
    options.createSource ?? ((url: string) => new EventSource(url) as EventSourceLike);

  let stopped = false;
  let source: EventSourceLike | null = null;
  let reconnectTimer: ReturnType<typeof setTimeout> | undefined;
  let debounceTimer: ReturnType<typeof setTimeout> | undefined;
  let backoff = reconnectMin;

  const scheduleChange = () => {
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      debounceTimer = undefined;
      if (!stopped) options.onGraphChanged();
    }, debounceMs);
  };

  const connect = () => {
    if (stopped || !browser || !sourceAvailable) return;
    const token = getToken();
    if (!token) return; // unauthenticated — polling keeps working
    const url = `${graphEventsBaseUrl()}/v1/graph/events?access_token=${encodeURIComponent(token)}`;

    const es = createSource(url);
    source = es;

    es.onopen = () => {
      backoff = reconnectMin;
    };
    es.addEventListener("graph", () => scheduleChange());
    es.onerror = () => {
      if (stopped || source !== es) return;
      // Any failure: drop the stream and reconnect with a fresh token after
      // backoff. An expired JWT inside the URL heals itself this way, and a
      // dropped connection reconnects per SYNC-1 C.
      es.close();
      source = null;
      if (reconnectTimer) return;
      reconnectTimer = setTimeout(() => {
        reconnectTimer = undefined;
        backoff = Math.min(backoff * 2, reconnectMax);
        connect();
      }, backoff);
    };
  };

  connect();

  return {
    stop() {
      stopped = true;
      if (reconnectTimer) clearTimeout(reconnectTimer);
      if (debounceTimer) clearTimeout(debounceTimer);
      source?.close();
      source = null;
    },
    isConnected() {
      return source !== null && source.readyState === 1;
    },
  };
}
