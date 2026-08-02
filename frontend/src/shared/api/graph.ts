import ky, { HTTPError, TimeoutError } from "ky";
import { api, refreshAccessToken } from "./client";
import { apiConfig } from "$shared/config";
import { accessToken, clearAuthState, isAuthenticated } from "$shared/stores/auth-session.svelte";
import { formatMessage } from "$shared/utils/i18n";

const userLocale = "ru";

// Узел графа – заметка (звезда)
export interface GraphNode {
  id: string;
  title: string;
  type?: string;
  x?: number;
  y?: number;
  z?: number;
  size?: number;
}

// Ребро графа – связь между заметками
export interface GraphLink {
  source: string; // ID исходной заметки
  target: string; // ID целевой заметки
  weight?: number; // вес связи (толщина линии)
  link_type?: string; // тип связи: reference, dependency, related, custom
}

// Данные графа: список узлов и рёбер
export interface GraphData {
  nodes: GraphNode[];
  links: GraphLink[];
  // Layout hash returned by graph-service for delta updates
  hash?: string;
}

/** Normalize a graph node to the canonical shape expected by the UI. */
export function normalizeNode(node: Partial<GraphNode>): GraphNode {
  return {
    id: node.id ?? "",
    title: node.title ?? "",
    type: node.type || "unknown",
    x: node.x,
    y: node.y,
    z: node.z,
    size: node.size,
  };
}

/** Normalize a graph link to the canonical shape expected by the UI. */
export function normalizeLink(link: Partial<GraphLink>): GraphLink {
  return {
    source: link.source ?? "",
    target: link.target ?? "",
    weight: link.weight ?? 0.5,
    link_type: link.link_type ?? "related",
  };
}

/**
 * Обрабатывает ошибки API графа и возвращает понятное сообщение
 */
function handleGraphError(error: unknown, context: string): never {
  if (import.meta.env.DEV) {
    console.error(`[Graph API] ${context}:`, error);
  }

  if (error instanceof Error) {
    if (error.message.includes("404") || error.message.includes("Not Found")) {
      throw new Error(formatMessage("graph.notFound", userLocale));
    }
    if (error.message.includes("500")) {
      throw new Error(formatMessage("graph.serverError", userLocale));
    }
    if (error.message.includes("Failed to fetch") || error.message.includes("NetworkError")) {
      throw new Error(formatMessage("graph.networkError", userLocale));
    }
    throw new Error(formatMessage("graph.loadError", userLocale, { message: error.message }));
  }

  throw new Error(formatMessage("graph.unknownError", userLocale));
}

// API response wrapper structure used by graph-service
interface GraphApiResponse {
  data: GraphData;
  meta?: {
    total_nodes?: number;
    total_links?: number;
    limit?: number;
    hash?: string;
  };
}

function buildQuery(params: Record<string, string | number | undefined>): string {
  return Object.entries(params)
    .filter(([, value]) => value !== undefined && value !== "")
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)
    .join("&");
}

function shouldFallback(error: unknown): boolean {
  if (error instanceof TimeoutError) return true;
  if (error instanceof HTTPError) {
    return (
      error.response.status >= 500 || error.response.status === 408 || error.response.status === 429
    );
  }
  if (error instanceof Error) {
    return (
      error.message.includes("Failed to fetch") ||
      error.message.includes("NetworkError") ||
      error.message.includes("network") ||
      error.message.includes("timeout")
    );
  }
  return false;
}

function normalizeGraphData(raw: unknown): GraphData {
  if (
    raw &&
    typeof raw === "object" &&
    "data" in raw &&
    raw.data &&
    typeof raw.data === "object" &&
    "nodes" in (raw as { data: object }).data
  ) {
    return (raw as GraphApiResponse).data;
  }
  if (
    raw &&
    typeof raw === "object" &&
    "nodes" in (raw as GraphData) &&
    Array.isArray((raw as GraphData).nodes)
  ) {
    return raw as GraphData;
  }
  return { nodes: [], links: [] };
}

function getGraphApi() {
  const isTest = typeof process !== "undefined" && process.env?.VITEST === "true";

  let baseUrl: string;
  if (isTest) {
    baseUrl = "http://localhost:9091/api";
  } else if (import.meta.env.DEV) {
    baseUrl = "/graph-service/api";
  } else {
    const configured = import.meta.env.VITE_GRAPH_SERVICE_URL || "/graph-service";
    baseUrl = configured.endsWith("/api") ? configured : `${configured}/api`;
  }

  return ky.create({
    prefixUrl: baseUrl,
    timeout: 30000,
    credentials: "include",
    retry: {
      limit: 0,
    },
    hooks: {
      beforeRequest: [
        (request) => {
          const token = accessToken();
          if (token) {
            request.headers.set("Authorization", `Bearer ${token}`);
          }
        },
      ],
      afterResponse: [
        async (request, options, response) => {
          if (response.status !== 401) return response;
          if (request.headers.get("X-Graph-Retry")) return response;

          // If the request was made without credentials, the caller is either
          // an anonymous user on a public page or an unauthenticated visitor
          // hitting a protected endpoint. Do not try to refresh the token and
          // do not redirect to login; let the caller handle the 401.
          const hadAuth = request.headers.has("Authorization") || request.headers.has("X-API-Key");
          if (!hadAuth) {
            return response;
          }

          const refreshed = await refreshAccessToken();
          if (refreshed) {
            request.headers.set("X-Graph-Retry", "1");
            return ky(request);
          }
          clearAuthState();
          return response;
        },
      ],
    },
  });
}

async function callBackendFallback<T>(fn: () => Promise<T>, context: string): Promise<T> {
  try {
    return await fn();
  } catch (error) {
    handleGraphError(error, context);
  }
}

// Запросить граф для заметки (возвращает все прямые связи и связанные заметки)
export async function getGraphData(
  noteId: string,
  depth: number = 2,
  _userId?: string,
  layout: "2d" | "3d" = "2d"
): Promise<GraphData> {
  const query = buildQuery({ depth, layout });
  try {
    const raw = await getGraphApi()
      .get(`v1/graph/note/${encodeURIComponent(noteId)}${query ? `?${query}` : ""}`)
      .json<unknown>();
    return normalizeGraphData(raw);
  } catch (error) {
    if (shouldFallback(error)) {
      return callBackendFallback(async () => {
        const raw = await api
          .get(`v1/notes/${encodeURIComponent(noteId)}/graph?${buildQuery({ depth })}`)
          .json<unknown>();
        return normalizeGraphData(raw);
      }, `Failed to load graph for note ${noteId}`);
    }
    handleGraphError(error, `Failed to load graph for note ${noteId}`);
  }
}

// Запросить полный граф всех заметок и связей
export async function getFullGraphData(
  limit: number = apiConfig.default_limit,
  _userId?: string,
  nocache?: boolean
): Promise<GraphData> {
  const query = buildQuery({
    limit,
    nocache: nocache ? 1 : undefined,
  });
  try {
    // graph-service exposes a dedicated public endpoint for unauthenticated users
    const endpoint = isAuthenticated() ? "v1/graph/full" : "v1/graph/public";
    const raw = await getGraphApi()
      .get(`${endpoint}${query ? `?${query}` : ""}`)
      .json<unknown>();
    const result = normalizeGraphData(raw);
    result.hash = (raw as GraphApiResponse).meta?.hash ?? result.hash;
    return result;
  } catch (error) {
    if (shouldFallback(error)) {
      return callBackendFallback(async () => {
        const raw = await api.get(`v1/graph/all?${buildQuery({ limit })}`).json<unknown>();
        return normalizeGraphData(raw);
      }, "Failed to load full graph");
    }
    handleGraphError(error, "Failed to load full graph");
  }
}

// Запросить инкрементальное дельта-обновление графа от graph-service
export async function getGraphDelta(lastHash: string): Promise<GraphDeltaData> {
  const query = buildQuery({ last_hash: lastHash });
  try {
    const raw = await getGraphApi()
      .get(`v1/graph/delta${query ? `?${query}` : ""}`)
      .json<unknown>();

    if (raw && typeof raw === "object") {
      const delta = raw as GraphDeltaData & { current_hash?: string };
      if (delta.current_hash) {
        (delta as GraphDeltaData).current_hash = delta.current_hash;
      }
      return delta;
    }
    return {};
  } catch (error) {
    handleGraphError(error, "Failed to load graph delta");
  }
}

// Graph delta structure for incremental updates
export interface GraphDeltaData {
  added_nodes?: GraphNode[];
  removed_nodes?: string[];
  updated_nodes?: GraphNode[];
  added_links?: GraphLink[];
  removed_links?: GraphLink[];
  // Hash of the graph version this delta transitions to
  current_hash?: string;
}

// Fresh graph response with optional delta
export interface FreshGraphResponse {
  fresh: GraphData;
  delta?: GraphDeltaData;
}

// API wrapper for fresh graph response (backend returns { data: ... })
interface FreshGraphApiResponse {
  data: FreshGraphResponse;
}

// Запросить кэшированный граф пользователя
export async function getCachedGraph(): Promise<GraphData | null> {
  try {
    const response = await api.get("v1/me/graph/cached");
    if (response.status === 204) {
      return null; // No cached data
    }
    const body = await response.json<GraphApiResponse>();
    return body.data || null;
  } catch (error) {
    if (import.meta.env.DEV) {
      console.warn("[Graph API] Failed to get cached graph:", error);
    }
    return null; // Return null on error, fallback to fresh
  }
}

// Запросить свежий граф с опциональным дельта-обновлением
export async function getFreshGraph(): Promise<FreshGraphResponse> {
  try {
    const response = await api
      .get("v1/me/graph/fresh", { cache: "no-store" })
      .json<FreshGraphApiResponse>();
    return response.data || { fresh: { nodes: [], links: [] } };
  } catch (error) {
    return handleGraphError(error, "Failed to load fresh graph");
  }
}
