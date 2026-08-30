<script lang="ts">
  import { onMount, type Snippet } from "svelte";
  import {
    COCKPIT_DEFAULT_SIZES,
    COCKPIT_EDGE_SIZE,
    COCKPIT_PANEL_GAP,
    cockpitStore,
    type CockpitPanelPosition,
  } from "$features/cosmic-cockpit";
  import { CockpitFrame } from "$features/cosmic-ui";
  import CockpitPanel from "./CockpitPanel.svelte";
  import CockpitBottomPanel from "./CockpitBottomPanel.svelte";
  import { GraphTopBar } from "$features/graph-ui";
  import CockpitLeftPanel from "./CockpitLeftPanel.svelte";
  import CockpitRightPanel from "./CockpitRightPanel.svelte";
  import CockpitFirstPersonButton from "./CockpitFirstPersonButton.svelte";
  import QuickCaptureWidget from "$widgets/quick-capture/QuickCaptureWidget.svelte";

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
    onSearch?: (query: string) => void;
    onFilter?: (type: string) => void;
    onSignIn?: () => void;
    onRegister?: () => void;
    canvasController?: {
      focusMode: boolean;
      fogEnabled: boolean;
      resetView: () => void;
      openSearch: () => void;
      toggleFocus: () => void;
      toggleFog: () => void;
    };
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
    searchQuery?: string;
    notes?: NoteItem[];
    links?: Array<{ source: string; target: string; link_type?: string; weight?: number }>;
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
    onSearch,
    onFilter,
    onSignIn,
    onRegister,
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
    searchQuery = "",
    notes = [],
    links = [],
    showFullGraph = true,
    canvasController,
  }: Props = $props();

  // Content-based sizing for the top/bottom panels: their height genuinely
  // reflects content (nav bar height, stats row height), so we let them
  // measure themselves and report their natural size, clamped to a sane
  // range, instead of always reserving a fixed height.
  //
  // Left/right panels are NOT measured this way: their width is what their
  // content is laid out *into* (scrollWidth of a block-level container ends
  // up reflecting the container's own width, not an independent "natural"
  // width), so a self-referential measurement there would be misleading.
  // They keep their fixed default width for now (see roadmap for a real
  // fix, e.g. a temporary max-content measurement pass).
  let panelSizes = $state<Record<CockpitPanelPosition, number>>({ ...COCKPIT_DEFAULT_SIZES });
  const sizeBounds: Partial<Record<CockpitPanelPosition, { min: number; max: number }>> = {
    top: { min: 56, max: 120 },
    bottom: { min: 48, max: 260 },
  };

  function handlePanelSizeChange(position: CockpitPanelPosition, measured: number) {
    if (panelSizes[position] === measured) return;
    panelSizes = { ...panelSizes, [position]: measured };
  }

  function handleNodeSelect(id: string | null) {
    if (id) {
      cockpitStore.openPanel("right");
    } else {
      cockpitStore.closePanel("right");
    }
    onNodeSelect?.(id);
  }

  // The graph canvas selects a node by mutating the shared `selectedNodeId`
  // store/prop directly (not through handleNodeSelect above), so the right
  // panel needs to react to that prop changing on its own — otherwise
  // clicking a node on the canvas updates CockpitRightPanel's content but
  // never actually opens the panel. Also: clicking an object while in
  // first-person mode should exit first-person and reveal the panel.
  let previousSelectedNodeId: string | null = null;
  $effect(() => {
    const id = selectedNodeId;
    if (id === previousSelectedNodeId) return;
    previousSelectedNodeId = id;

    if (id) {
      if (cockpitStore.firstPerson) {
        cockpitStore.exitFirstPerson();
      }
      cockpitStore.openPanel("right");
    } else {
      cockpitStore.closePanel("right");
    }
  });

  function visibleSize(position: CockpitPanelPosition): number {
    if (cockpitStore.firstPerson) return 0;
    const panel = cockpitStore.panels[position];
    return panel.open || panel.pinned || panel.hovering ? panelSizes[position] : COCKPIT_EDGE_SIZE;
  }

  const topInset = $derived(visibleSize("top") + COCKPIT_PANEL_GAP);
  const bottomInset = $derived(visibleSize("bottom") + COCKPIT_PANEL_GAP);
  const leftInset = $derived(visibleSize("left") + COCKPIT_PANEL_GAP);
  const rightInset = $derived(visibleSize("right") + COCKPIT_PANEL_GAP);

  const cockpitStyle = $derived(
    `--inset-top:${topInset}px;--inset-bottom:${bottomInset}px;--inset-left:${leftInset}px;--inset-right:${rightInset}px;`
  );

  // Keep the frame's resize in lockstep with CockpitPanel's own slide
  // animation (see getTransition() in CockpitPanel.svelte) so the canvas
  // area doesn't snap to its new size while the panel is still sliding —
  // that mismatch is what caused both the "abrupt" feel and the stray
  // empty gap between the frame edge and an still-opening/closing panel.
  const frameTransition = $derived(
    cockpitStore.reducedMotion
      ? "none"
      : "top 0.3s ease, right 0.3s ease, bottom 0.3s ease, left 0.3s ease"
  );

  $effect(() => {
    cockpitStore.saveSettings();
  });

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape" && cockpitStore.firstPerson) {
      cockpitStore.exitFirstPerson();
      e.preventDefault();
    }
  }

  onMount(() => {
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  });
</script>

<div
  class="cosmic-cockpit"
  class:first-person={cockpitStore.firstPerson}
  data-testid="cosmic-cockpit"
  style={cockpitStyle}
>
  {#if isAuthenticated}
    <div class="cockpit-frame-wrapper" style="transition: {frameTransition};">
      <CockpitFrame>
        {@render children?.()}
      </CockpitFrame>
    </div>

    <CockpitPanel
      position="top"
      size={panelSizes.top}
      minSize={sizeBounds.top?.min}
      maxSize={sizeBounds.top?.max}
      onSizeChange={(s) => handlePanelSizeChange("top", s)}
      title="Navigation"
    >
      <GraphTopBar
        {isAuthenticated}
        {currentView}
        {layoutProvider}
        {searchQuery}
        {selectedType}
        {typeFilters}
        {typeCounts}
        {nodeCount}
        {linkCount}
        {onSearch}
        {onToggleView}
        {onToggleLayoutProvider}
        {onFilter}
        {onNoteCreate}
        {canvasController}
        variant="docked"
      />
    </CockpitPanel>

    <CockpitPanel
      position="bottom"
      size={panelSizes.bottom}
      minSize={sizeBounds.bottom?.min}
      maxSize={sizeBounds.bottom?.max}
      onSizeChange={(s) => handlePanelSizeChange("bottom", s)}
      title="System View"
    >
      <CockpitBottomPanel {nodeCount} {linkCount} />
    </CockpitPanel>

    <CockpitPanel position="left" size={panelSizes.left} title="Operations">
      <CockpitLeftPanel
        onNoteSelect={handleNodeSelect}
        {notes}
        {onImport}
        {onExport}
        {showFullGraph}
        {onToggleFullGraph}
      />
    </CockpitPanel>

    <CockpitPanel position="right" size={panelSizes.right} title="Details" onClose={() => handleNodeSelect(null)}>
      <CockpitRightPanel
        nodeId={selectedNodeId}
        isAuthenticated={true}
        {notes}
        {links}
        onNodeSelect={handleNodeSelect}
        {onNoteDelete}
        {onNoteEdit}
        {onCreateChildNote}
      />
    </CockpitPanel>

    {#if cockpitStore.firstPerson}
      <CockpitFirstPersonButton />
    {/if}

    {#if isAuthenticated}
      <QuickCaptureWidget docked />
    {/if}
  {:else}
    <div class="public-cockpit">
      <div class="graph-top-bar-wrapper">
        <GraphTopBar
          isAuthenticated={false}
          {currentView}
          {layoutProvider}
          {searchQuery}
          {selectedType}
          {typeFilters}
          {typeCounts}
          {nodeCount}
          {linkCount}
          {onSearch}
          {onToggleView}
          {onToggleLayoutProvider}
          {onFilter}
          {onSignIn}
          {onRegister}
          {canvasController}
          variant="floating"
        />
      </div>
      {@render children?.()}
      <CockpitPanel
        position="right"
        size={panelSizes.right}
        title="Details"
        onClose={() => handleNodeSelect(null)}
      >
        <CockpitRightPanel
          nodeId={selectedNodeId}
          isAuthenticated={false}
          {notes}
          {links}
          onNodeSelect={handleNodeSelect}
          {onSignIn}
        />
      </CockpitPanel>
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

  .graph-top-bar-wrapper {
    position: absolute;
    top: 16px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 60;
    pointer-events: auto;
    width: fit-content;
    max-width: calc(100% - 32px);
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
