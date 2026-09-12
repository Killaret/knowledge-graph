<script lang="ts">
  import GraphPageShell from "./GraphPageShell.svelte";

  interface NoteItem {
    id: string;
    title: string;
    type?: string;
  }

  interface Props {
    notes?: NoteItem[];
    nodes?: NoteItem[];
    onSignIn?: () => void;
    onRegister?: () => void;
    onImport?: () => void;
    onExport?: () => void;
    onNoteCreate?: () => void;
    onNoteDelete?: (id: string) => void;
    onNoteEdit?: (id: string) => void;
    onCreateChildNote?: (note: NoteItem) => void;
    onToggleFullGraph?: (value: boolean) => void;
    showFullGraph?: boolean;
  }

  const {
    notes,
    nodes,
    onSignIn,
    onRegister,
    onImport,
    onExport,
    onNoteCreate,
    onNoteDelete,
    onNoteEdit,
    onCreateChildNote,
    onToggleFullGraph,
    showFullGraph,
  }: Props = $props();
</script>

<GraphPageShell
  view="graph"
  typeFilters={[
    { id: "all", label: "All", emoji: "🔭", description: "All types" },
    { id: "star", label: "Star", emoji: "⭐", description: "Star notes" },
  ]}
  {notes}
  {nodes}
  {onSignIn}
  {onRegister}
  {onImport}
  {onExport}
  {onNoteCreate}
  {onNoteDelete}
  {onNoteEdit}
  {onCreateChildNote}
  {onToggleFullGraph}
  {showFullGraph}
  onToggleView={() => {}}
>
  <div data-testid="shell-children">content</div>
</GraphPageShell>
