import type { GraphData } from "$shared/api/graph";

/**
 * Cross-view graph UI state shared between 2D canvas, 3D viewer and list view.
 *
 * This store intentionally holds only primitives / DTOs from the shared layer
 * so it can be imported by any FSD layer without pulling in entities or features.
 */
export interface GraphUIState {
  selectedNodeId: string | null;
  searchQuery: string;
  selectedType: string;
  currentView: "graph" | "list" | "3d";
  graphData: GraphData;
  hoveredNodeId: string | null;
  selectedLinkTypes: string[];
  minLinkWeight: number;
}

function createGraphStore(initial: Partial<GraphUIState> = {}) {
  let selectedNodeId = $state<string | null>(initial.selectedNodeId ?? null);
  let searchQuery = $state(initial.searchQuery ?? "");
  let selectedType = $state(initial.selectedType ?? "all");
  let currentView = $state<"graph" | "list" | "3d">(initial.currentView ?? "graph");
  let graphData = $state<GraphData>(initial.graphData ?? { nodes: [], links: [] });
  let hoveredNodeId = $state<string | null>(initial.hoveredNodeId ?? null);
  let selectedLinkTypes = $state<string[]>(initial.selectedLinkTypes ?? []);
  let minLinkWeight = $state(initial.minLinkWeight ?? 0);

  return {
    get selectedNodeId() {
      return selectedNodeId;
    },
    set selectedNodeId(value: string | null) {
      selectedNodeId = value;
    },

    get searchQuery() {
      return searchQuery;
    },
    set searchQuery(value: string) {
      searchQuery = value;
    },

    get selectedType() {
      return selectedType;
    },
    set selectedType(value: string) {
      selectedType = value;
    },

    get currentView() {
      return currentView;
    },
    set currentView(value: "graph" | "list" | "3d") {
      currentView = value;
    },

    get graphData() {
      return graphData;
    },
    set graphData(value: GraphData) {
      graphData = value;
    },

    get hoveredNodeId() {
      return hoveredNodeId;
    },
    set hoveredNodeId(value: string | null) {
      hoveredNodeId = value;
    },

    get selectedLinkTypes() {
      return selectedLinkTypes;
    },
    set selectedLinkTypes(value: string[]) {
      selectedLinkTypes = value;
    },

    get minLinkWeight() {
      return minLinkWeight;
    },
    set minLinkWeight(value: number) {
      minLinkWeight = value;
    },

    toggleLinkType(type: string) {
      if (selectedLinkTypes.includes(type)) {
        selectedLinkTypes = selectedLinkTypes.filter((t) => t !== type);
      } else {
        selectedLinkTypes = [...selectedLinkTypes, type];
      }
    },

    selectNode(id: string | null) {
      selectedNodeId = id;
    },

    clearSearch() {
      searchQuery = "";
    },

    reset() {
      selectedNodeId = null;
      searchQuery = "";
      selectedType = "all";
      currentView = "graph";
      graphData = { nodes: [], links: [] };
      hoveredNodeId = null;
      selectedLinkTypes = [];
      minLinkWeight = 0;
    },
  };
}

export const graphStore = createGraphStore();
