<script lang="ts">
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import CockpitNoteDetails from "./CockpitNoteDetails.svelte";

  interface Props {
    nodeId: string | null;
    onNodeSelect?: (id: string | null) => void;
    onNoteDelete?: (id: string) => void;
    onNoteEdit?: (id: string) => void;
  }

  const { nodeId, onNodeSelect, onNoteDelete, onNoteEdit }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  function handleClose() {
    onNodeSelect?.(null);
    // keep panel open but clear selection
  }

  function handleEdit(id: string) {
    onNoteEdit?.(id);
  }

  function handleDelete(id: string) {
    onNoteDelete?.(id);
    handleClose();
  }
</script>

<div class="cockpit-right-panel" data-testid="cockpit-right-panel">
  {#if nodeId}
    <CockpitNoteDetails {nodeId} onClose={handleClose} onEdit={handleEdit} onDelete={handleDelete} />
  {:else}
    <div class="empty-state">
      <span class="empty-icon">🌌</span>
      <p>{t("cockpit.right.empty")}</p>
    </div>
  {/if}
</div>

<style>
  .cockpit-right-panel {
    height: 100%;
    overflow: hidden;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 12px;
    color: rgba(255, 255, 255, 0.5);
    text-align: center;
    padding: 20px;
  }

  .empty-icon {
    font-size: 48px;
    opacity: 0.6;
  }

  .empty-state p {
    margin: 0;
    font-size: 14px;
  }
</style>
