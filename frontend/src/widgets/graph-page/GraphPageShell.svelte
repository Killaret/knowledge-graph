<script lang="ts">
  import type { Snippet } from "svelte";
  import CosmicCockpitLayout from "$widgets/cosmic-cockpit/CosmicCockpitLayout.svelte";
  import { isAuthenticated } from "$shared/stores/auth.svelte";
  import type { GraphLink } from "$shared/api/graph";

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
    view: "graph" | "list" | "3d";
    layoutProvider?: "d3" | "graph-service";
    searchQuery?: string;
    selectedType?: string;
    typeFilters: TypeFilter[];
    notes?: NoteItem[];
    nodes?: NoteItem[];
    links?: GraphLink[];
    nodeCount?: number;
    linkCount?: number;
    selectedNodeId?: string | null;
    canvasController?: {
      focusMode: boolean;
      fogEnabled: boolean;
      resetView: () => void;
      openSearch: () => void;
      toggleFocus: () => void;
      toggleFog: () => void;
    };
    showFullGraph?: boolean;
    onSearch?: (query: string) => void;
    onFilter?: (type: string) => void;
    onToggleView: (view: "graph" | "list" | "3d") => void;
    onToggleLayoutProvider?: (provider: "d3" | "graph-service") => void;
    onNodeSelect?: (id: string | null) => void;
    onNoteCreate?: () => void;
    onNoteDelete?: (id: string) => void;
    onNoteEdit?: (id: string) => void;
    onCreateChildNote?: (note: NoteItem) => void;
    onImport?: () => void;
    onExport?: () => void;
    onToggleFullGraph?: (value: boolean) => void;
    onSignIn?: () => void;
    onRegister?: () => void;
    children: Snippet;
  }

  const {
    view,
    layoutProvider = "d3",
    searchQuery = "",
    selectedType = "all",
    typeFilters,
    notes,
    nodes,
    links,
    nodeCount,
    linkCount,
    selectedNodeId = null,
    canvasController,
    showFullGraph,
    onSearch,
    onFilter,
    onToggleView,
    onToggleLayoutProvider,
    onNodeSelect,
    onNoteCreate,
    onNoteDelete,
    onNoteEdit,
    onCreateChildNote,
    onImport,
    onExport,
    onToggleFullGraph,
    onSignIn,
    onRegister,
    children,
  }: Props = $props();

  const items = $derived<NoteItem[]>(notes ?? nodes ?? []);
  const effectiveNodeCount = $derived(nodeCount ?? nodes?.length ?? notes?.length ?? 0);
  const effectiveLinkCount = $derived(linkCount ?? links?.length ?? 0);
  const effectiveNotes = $derived<NoteItem[]>(items);

  const typeCounts = $derived<Record<string, number>>(
    Object.fromEntries(
      typeFilters.map((filter) => [
        filter.id,
        filter.id === "all" ? items.length : items.filter((item) => item.type === filter.id).length,
      ])
    )
  );

  const auth = $derived(isAuthenticated());
</script>

<CosmicCockpitLayout
  isAuthenticated={auth}
  currentView={view}
  {layoutProvider}
  {searchQuery}
  {selectedType}
  {typeFilters}
  {typeCounts}
  nodeCount={effectiveNodeCount}
  linkCount={effectiveLinkCount}
  {selectedNodeId}
  notes={effectiveNotes}
  {links}
  {showFullGraph}
  {canvasController}
  {onSearch}
  {onFilter}
  {onToggleView}
  {onToggleLayoutProvider}
  {onNodeSelect}
  onImport={auth ? onImport : undefined}
  onExport={auth ? onExport : undefined}
  {onToggleFullGraph}
  onNoteCreate={auth ? onNoteCreate : undefined}
  onNoteDelete={auth ? onNoteDelete : undefined}
  onNoteEdit={auth ? onNoteEdit : undefined}
  onCreateChildNote={auth ? onCreateChildNote : undefined}
  onSignIn={!auth ? onSignIn : undefined}
  onRegister={!auth ? onRegister : undefined}
>
  {@render children()}
</CosmicCockpitLayout>
