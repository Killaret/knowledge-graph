<script lang="ts">
  import type { Snippet } from "svelte";
  import {
    COCKPIT_DEFAULT_SIZES,
    COCKPIT_EDGE_SIZE,
    COCKPIT_PANEL_GAP,
    cockpitStore,
    type CockpitPanelPosition,
  } from "$features/cosmic-cockpit";
  import { CockpitFrame } from "$features/cosmic-ui";
  import CockpitPanel from "./CockpitPanel.svelte";
  import CockpitTopPanel from "./CockpitTopPanel.svelte";
  import CockpitBottomPanel from "./CockpitBottomPanel.svelte";
  import CockpitLeftPanel from "./CockpitLeftPanel.svelte";
  import CockpitRightPanel from "./CockpitRightPanel.svelte";
  import CockpitFirstPersonButton from "./CockpitFirstPersonButton.svelte";

  interface NoteItem {
    id: string;
    title: string;
    type?: string;
  }

  interface TypeFilter {
    id: string;
    label: string;
    emoji: string;
    description?: string;
    example?: string;
  }

  interface Props {
    isAuthenticated: boolean;
    children?: Snippet;
    onNodeSelect?: (id: string | null) => void;
    onNoteCreate?: () => void;
    onNoteDelete?: (id: string) => void;
    onNoteEdit?: (id: string) => void;
    onCreateChildNote?: (note: NoteItem) => void;
    onOpenAuth?: (tab: "login" | "register") => void;
    onSearch?: (query: string) => void;
    onFilter?: (type: string) => void;
    onToggleView?: (view: "graph" | "list" | "3d") => void;
    onToggleLayoutProvider?: (provider: "d3" | "graph-service") => void;
    onToggleFullGraph?: (value: boolean) => void;
    onImport?: () => void;
    onExport?: () => void;
    currentView?: "graph" | "list" | "3d";
    layoutProvider?: "d3" | "graph-service";
    selectedNodeId?: string | null;
    nodeCount?: number;
    linkCount?: number;
    typeFilters?: TypeFilter[];
    selectedType?: string;
    typeCounts?: Record<string, number>;
    notes?: NoteItem[];
    showFullGraph?: boolean;
  }

  const {
    isAuthenticated,
    children,
    onNodeSelect,
    onNoteCreate,
    onNoteDelete,
    onNoteEdit,
    onCreateChildNote,
    onOpenAuth,
    onSearch,
    onFilter,
    onToggleView,
    onToggleLayoutProvider,
    onToggleFullGraph,
    onImport,
    onExport,
    currentView = "graph",
    layoutProvider = "graph-service",
    selectedNodeId = null,
    nodeCount = 0,
    linkCount = 0,
    typeFilters = [],
    selectedType = "all",
    typeCounts = {},
    notes = [],
    showFullGraph = true,
  }: Props = $props();

  function handleNodeSelect(id: string | null) {
    if (id) {
      cockpitStore.openPanel("right");
    } else {
      cockpitStore.closePanel("right");
    }
    onNodeSelect?.(id);
  }

  function visibleSize(position: CockpitPanelPosition): number {
    if (cockpitStore.firstPerson) return 0;
    const panel = cockpitStore.panels[position];
    return panel.open || panel.pinned || panel.hovering
      ? COCKPIT_DEFAULT_SIZES[position]
      : COCKPIT_EDGE_SIZE;
  }

  const topInset = $derived(visibleSize("top") + COCKPIT_PANEL_GAP);
  const bottomInset = $derived(visibleSize("bottom") + COCKPIT_PANEL_GAP);
  const leftInset = $derived(visibleSize("left") + COCKPIT_PANEL_GAP);
  const rightInset = $derived(visibleSize("right") + COCKPIT_PANEL_GAP);

  const cockpitStyle = $derived(
    `--inset-top:${topInset}px;--inset-bottom:${bottomInset}px;--inset-left:${leftInset}px;--inset-right:${rightInset}px;`
  );

  $effect(() => {
    cockpitStore.saveSettings();
  });
</script>

<div
  class="cosmic-cockpit"
  class:first-person={cockpitStore.firstPerson}
  data-testid="cosmic-cockpit"
  style={cockpitStyle}
>
  {#if isAuthenticated}
    <div class="cockpit-frame-wrapper">
      <CockpitFrame>
        {@render children?.()}
      </CockpitFrame>
    </div>

    <CockpitPanel position="top" size={COCKPIT_DEFAULT_SIZES.top} title="Navigation">
      <CockpitTopPanel
        {onSearch}
        {onToggleView}
        {onToggleLayoutProvider}
        {currentView}
        {layoutProvider}
        {onNoteCreate}
      />
    </CockpitPanel>

    <CockpitPanel position="bottom" size={COCKPIT_DEFAULT_SIZES.bottom} title="System View">
      <CockpitBottomPanel {nodeCount} {linkCount} />
    </CockpitPanel>

    <CockpitPanel position="left" size={COCKPIT_DEFAULT_SIZES.left} title="Operations">
      <CockpitLeftPanel
        {typeFilters}
        {selectedType}
        {typeCounts}
        {onFilter}
        onNoteSelect={handleNodeSelect}
        {notes}
        {onImport}
        {onExport}
        {showFullGraph}
        {onToggleFullGraph}
      />
    </CockpitPanel>

    <CockpitPanel position="right" size={COCKPIT_DEFAULT_SIZES.right} title="Details">
      <CockpitRightPanel
        nodeId={selectedNodeId}
        onNodeSelect={handleNodeSelect}
        {onNoteDelete}
        {onNoteEdit}
        {onCreateChildNote}
      />
    </CockpitPanel>

    {#if cockpitStore.firstPerson}
      <CockpitFirstPersonButton />
    {/if}
  {:else}
    <div class="public-cockpit">
      {@render children?.()}
    </div>
  {/if}
</div>

<style>
  .cosmic-cockpit {
    position: fixed;
    inset: 0;
    overflow: hidden;
    z-index: 10;
    pointer-events: none;
  }

  .cosmic-cockpit > :global(*) {
    pointer-events: auto;
  }

  .cockpit-frame-wrapper {
    position: absolute;
    top: var(--inset-top, 0);
    right: var(--inset-right, 0);
    bottom: var(--inset-bottom, 0);
    left: var(--inset-left, 0);
    z-index: 40;
    pointer-events: none;
  }

  .cockpit-frame-wrapper > :global(*) {
    pointer-events: auto;
  }

  .public-cockpit {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    overflow: hidden;
  }

  @media (prefers-reduced-motion: reduce) {
    .cosmic-cockpit {
      transition: none;
    }
  }
</style>
