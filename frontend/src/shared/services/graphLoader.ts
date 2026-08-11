/**
 * Unified graph loader driven by authentication state.
 *
 * This module is the single place where the home page and the dedicated
 * /graph page coordinate graph + note loading. It must remain in `shared/`
 * and must not import any higher FSD layer.
 */

import { getNotes, getNote, type Note } from "$shared/api/notes";
import {
  getGraphData,
  getFullGraphData,
  type GraphData,
  type GraphNode,
  type GraphLink,
} from "$shared/api/graph";
import { isAuthenticated } from "$shared/stores/auth.svelte";

const KNOWLEDGE_CORE_ID = "00000000-0000-0000-0000-000000000001";

interface RawNode extends Partial<GraphNode> {
  Id?: string;
  ID?: string;
  Title?: string;
  Type?: string;
  created_at?: string;
  createdAt?: string;
  CreatedAt?: string;
}

interface RawLink extends Partial<GraphLink> {
  source_note_id?: string;
  target_note_id?: string;
}

export interface LoadGraphOptions {
  full?: boolean;
  depth?: number;
  nocache?: boolean;
  includeKnowledgeCore?: boolean;
  fallbackToNotes?: boolean;
  ensureNotesInGraph?: boolean;
  fullGraphLoader?: (nocache?: boolean) => Promise<GraphData>;
}

function getNodeId(node: RawNode): string {
  return node.id ?? node.Id ?? node.ID ?? "";
}

function getNodeTitle(node: RawNode): string {
  return node.title ?? node.Title ?? "";
}

function getNodeType(node: RawNode): string {
  return node.type ?? node.Type ?? "star";
}

function getNodeCreatedAt(node: RawNode): string | undefined {
  return node.created_at ?? node.createdAt ?? node.CreatedAt;
}

function getLinkSource(link: RawLink): string {
  return link.source_note_id || link.source || "";
}

function getLinkTarget(link: RawLink): string {
  return link.target_note_id || link.target || "";
}

function createGraphNodeWithCreatedAt(node: RawNode): GraphNode & { createdAt?: string } {
  return {
    id: getNodeId(node),
    title: getNodeTitle(node),
    type: getNodeType(node),
    x: node.x,
    y: node.y,
    z: node.z,
    size: node.size,
    createdAt: getNodeCreatedAt(node),
  };
}

/**
 * Load the user's notes when authenticated, otherwise return an empty list.
 */
export async function loadNotesIfAuthenticated(): Promise<Note[]> {
  if (!isAuthenticated()) {
    return [];
  }
  return getNotes();
}

/**
 * Load raw graph data from the graph-service or backend.
 */
export async function loadRawGraphData(
  options: LoadGraphOptions,
  notes?: Note[]
): Promise<GraphData> {
  if (options.full || !notes || notes.length === 0) {
    const loader =
      options.fullGraphLoader ?? ((nocache?: boolean) => getFullGraphData(0, undefined, nocache));
    return loader(options.nocache);
  }
  return getGraphData(notes[0].id, options.depth ?? 3);
}

/**
 * Load the Knowledge Core system note when authenticated.
 */
export async function loadKnowledgeCoreIfAuthenticated(): Promise<Note | null> {
  if (!isAuthenticated()) {
    return null;
  }
  try {
    return await getNote(KNOWLEDGE_CORE_ID);
  } catch {
    return null;
  }
}

/**
 * Normalize raw graph nodes and links, and optionally inject the
 * Knowledge Core as a `technical` node.
 */
export function transformRawGraph(rawData: GraphData, knowledgeCore?: Note | null): GraphData {
  const normalizedNodes: Array<GraphNode & { createdAt?: string }> = (
    rawData.nodes as RawNode[]
  ).map(createGraphNodeWithCreatedAt);

  if (knowledgeCore && !normalizedNodes.some((n) => n.id === knowledgeCore.id)) {
    normalizedNodes.push({
      id: knowledgeCore.id,
      title: knowledgeCore.title,
      type: "technical",
      createdAt: knowledgeCore.created_at,
    });
  }

  const normalizedLinks: GraphLink[] = (rawData.links as RawLink[]).map((link) => ({
    id: link.id,
    source: getLinkSource(link),
    target: getLinkTarget(link),
    weight: link.weight ?? 0.5,
    link_type: link.link_type ?? "related",
    source_type: link.source_type ?? "user",
    last_weight_update: link.last_weight_update,
  }));

  return {
    ...rawData,
    nodes: normalizedNodes,
    links: normalizedLinks,
  };
}

/**
 * Build a minimal fallback graph directly from a list of notes.
 */
export function buildNotesGraph(notes: Note[]): GraphData {
  return {
    nodes: notes.map((note) => ({
      id: note.id,
      title: note.title,
      type: note.type || "unknown",
      size: 1,
      createdAt: note.created_at,
    })),
    links: [],
  };
}

/**
 * Guarantee that every note in `notes` appears as a node in the graph.
 */
export function ensureNotesInGraph(graph: GraphData, notes: Note[]): GraphData {
  const nodeIds = new Set(graph.nodes.map((n) => n.id));
  const added: Array<GraphNode & { createdAt?: string }> = [];

  for (const note of notes) {
    if (!nodeIds.has(note.id)) {
      added.push({
        id: note.id,
        title: note.title,
        type: note.type || "unknown",
        x: 0,
        y: 0,
        z: 0,
        size: 1,
        createdAt: note.created_at,
      });
      nodeIds.add(note.id);
    }
  }

  if (added.length === 0) {
    return graph;
  }

  return {
    ...graph,
    nodes: [...graph.nodes, ...added],
  };
}

/**
 * Main orchestration: load notes, graph, and optional Knowledge Core,
 * then normalize, merge, and fallback as configured.
 */
export async function loadGraph(
  options: LoadGraphOptions,
  providedNotes?: Note[]
): Promise<{ graph: GraphData; notes: Note[]; knowledgeCore: Note | null }> {
  const notes = providedNotes ?? (await loadNotesIfAuthenticated());

  const [rawData, knowledgeCore] = await Promise.all([
    loadRawGraphData(options, notes),
    options.includeKnowledgeCore
      ? loadKnowledgeCoreIfAuthenticated()
      : Promise.resolve<Note | null>(null),
  ]);

  let graph = transformRawGraph(rawData, knowledgeCore);

  if (options.ensureNotesInGraph && isAuthenticated()) {
    graph = ensureNotesInGraph(graph, notes);
  }

  if (options.fallbackToNotes && graph.nodes.length === 0 && notes.length > 0) {
    graph = buildNotesGraph(notes);
  }

  return { graph, notes, knowledgeCore };
}
