<script lang="ts">
  import type { Snippet } from "svelte";
  import { COCKPIT_DEFAULT_SIZES, cockpitStore } from "$features/cosmic-cockpit";
  import CockpitPanel from "./CockpitPanel.svelte";
  import CockpitTopPanel from "./CockpitTopPanel.svelte";
  import CockpitBottomPanel from "./CockpitBottomPanel.svelte";
  import CockpitLeftPanel from "./CockpitLeftPanel.svelte";
  import CockpitRightPanel from "./CockpitRightPanel.svelte";
  import CockpitFirstPersonButton from "./CockpitFirstPersonButton.svelte";

  interface Props {
    /** Main graph/content slot. */
    children?: Snippet;
    /** Callbacks wired to the graph canvas. */
    onNodeSelect?: (id: string | null) => void;
    onNoteCreate?: () => void;
    onNoteDelete?: (id: string) => void;
    onNoteEdit?: (id: string) => void;
    onCreateChildNote?: (note: { id: string; title: string; type?: string }) => void;
    onOpenAuth?: (tab: "login" | "register") => void;
    onSearch?: (query: string) => void;
    onFilter?: (type: string) => void;
    onToggleView?: (view: "graph" | "list" | "3d") => void;
    onToggleLayoutProvider?: (provider: "d3" | "graph-service") => void;
    currentView?: "graph" | "list" | "3d";
    layoutProvider?: "d3" | "graph-service";
    selectedNodeId?: string | null;
    nodeCount?: number;
    linkCount?: number;
    typeFilters?: Array<{
      id: string;
      label: string;
      emoji: string;
      description?: string;
      example?: string;
    }>;
    selectedType?: string;
    typeCounts?: Record<string, number>;
    notes?: Array<{ id: string; title: string; type?: string }>;
  }

  const {
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
    currentView = "graph",
    layoutProvider = "graph-service",
    selectedNodeId = null,
    nodeCount = 0,
    linkCount = 0,
    typeFilters = [],
    selectedType = "all",
    typeCounts = {},
    notes = [],
  }: Props = $props();

  const topHeight = $derived(
    cockpitStore.panels.top.open || cockpitStore.panels.top.pinned ? COCKPIT_DEFAULT_SIZES.top : 0
  );
  const bottomHeight = $derived(
    cockpitStore.panels.bottom.open || cockpitStore.panels.bottom.pinned
      ? COCKPIT_DEFAULT_SIZES.bottom
      : 0
  );
  const leftWidth = $derived(
    cockpitStore.panels.left.open || cockpitStore.panels.left.pinned
      ? COCKPIT_DEFAULT_SIZES.left
      : 0
  );
  const rightWidth = $derived(
    cockpitStore.panels.right.open || cockpitStore.panels.right.pinned
      ? COCKPIT_DEFAULT_SIZES.right
      : 0
  );

  $effect(() => {
    // Persist cockpit settings whenever any of them changes.
    cockpitStore.saveSettings();
  });
</script>

<div
  class="cosmic-cockpit"
  class:first-person={cockpitStore.firstPerson}
  style="--top-height:{topHeight}px;--bottom-height:{bottomHeight}px;--left-width:{leftWidth}px;--right-width:{rightWidth}px;"
  data-testid="cosmic-cockpit"
>
  <CockpitPanel position="top" size={COCKPIT_DEFAULT_SIZES.top} title="Navigation">
    <CockpitTopPanel
      {onSearch}
      {onToggleView}
      {onToggleLayoutProvider}
      {currentView}
      {layoutProvider}
      {onNoteCreate}
      {onOpenAuth}
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
      {currentView}
      {layoutProvider}
      {onToggleView}
      {onToggleLayoutProvider}
      {onFilter}
      onNoteSelect={onNodeSelect}
      {notes}
    />
  </CockpitPanel>

  <CockpitPanel position="right" size={COCKPIT_DEFAULT_SIZES.right} title="Details">
    <CockpitRightPanel
      nodeId={selectedNodeId}
      {onNodeSelect}
      {onNoteDelete}
      {onNoteEdit}
      {onCreateChildNote}
    />
  </CockpitPanel>

  <section class="cockpit-graph" data-testid="cockpit-graph">
    {@render children?.()}
  </section>

  {#if cockpitStore.firstPerson}
    <CockpitFirstPersonButton />
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

  .cockpit-graph {
    position: absolute;
    top: var(--top-height, 0);
    right: var(--right-width, 0);
    bottom: var(--bottom-height, 0);
    left: var(--left-width, 0);
    overflow: hidden;
    transition:
      top 0.3s ease,
      right 0.3s ease,
      bottom 0.3s ease,
      left 0.3s ease;
  }

  .first-person .cockpit-graph {
    top: 0;
    right: 0;
    bottom: 0;
    left: 0;
  }

  @media (prefers-reduced-motion: reduce) {
    .cockpit-graph {
      transition: none;
    }
  }
</style>
