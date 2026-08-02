<script lang="ts">
  import { getNote, type Note } from "$shared/api/notes";
  import { getNoteLinks, deleteAllNoteLinks, type Link } from "$shared/api/links";
  import { goto } from "$app/navigation";
  import { formatDate } from "$shared/utils/date";
  import { CelestialBody } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) => formatMessage(key, locale, params);

  interface Props {
    nodeId: string;
    onClose?: () => void;
    onEdit?: (id: string) => void;
    onDelete?: (id: string) => void;
  }

  const { nodeId, onClose, onEdit, onDelete }: Props = $props();

  let note = $state<Note | null>(null);
  let links = $state<Link[]>([]);
  let loading = $state(true);
  let error = $state("");
  let showDeleteLinksConfirm = $state(false);
  let deletingLinks = $state(false);

  const tags = $derived((note?.metadata?.tags ?? []) as string[]);

  $effect(() => {
    const id = nodeId;
    loadNote(id);
    loadLinks(id);
  });

  async function loadNote(id: string) {
    loading = true;
    error = "";
    try {
      note = await getNote(id);
    } catch {
      error = t("cockpit.noteDetails.loadError");
      note = null;
    } finally {
      loading = false;
    }
  }

  async function loadLinks(id: string) {
    try {
      links = await getNoteLinks(id);
    } catch {
      links = [];
    }
  }

  async function handleDeleteAllLinks() {
    deletingLinks = true;
    try {
      await deleteAllNoteLinks(nodeId);
      links = [];
      showDeleteLinksConfirm = false;
    } catch {
      error = t("cockpit.noteDetails.deleteLinksError");
    } finally {
      deletingLinks = false;
    }
  }

  function getTypeIcon(type: string | undefined): string {
    return type ? CelestialBody.fromString(type).emoji : CelestialBody.STAR.emoji;
  }

  function getTypeLabel(type: string | undefined): string {
    return type ? CelestialBody.fromString(type).label : t("cockpit.noteDetails.fallbackType");
  }
</script>

<div class="note-details" data-testid="cockpit-note-details">
  <div class="details-header">
    <button type="button" class="close-btn" onclick={() => onClose?.()} aria-label={t("cockpit.noteDetails.close")}>
      ✕
    </button>
    {#if note}
      <div class="actions">
        <button type="button" class="action-btn" onclick={() => onEdit?.(nodeId)} aria-label={t("cockpit.noteDetails.edit")}>
          ✎
        </button>
        <button type="button" class="action-btn share" onclick={() => { /* share */ }} aria-label={t("cockpit.noteDetails.share")}>
          ⇄
        </button>
        <button type="button" class="action-btn delete" onclick={() => onDelete?.(nodeId)} aria-label={t("cockpit.noteDetails.deleteNote")}>
          🗑
        </button>
      </div>
    {/if}
  </div>

  <div class="details-content">
    {#if loading}
      <div class="loading" role="status" aria-live="polite">
        <div class="spinner" aria-hidden="true"></div>
        <p>{t("cockpit.noteDetails.loading")}</p>
      </div>
    {:else if error}
      <div class="error">{error}</div>
    {:else if note}
      <div class="note-header">
        <span class="type-icon">{getTypeIcon(note.type)}</span>
        <h2 class="title">{note.title}</h2>
        <span class="type-badge">{getTypeLabel(note.type)}</span>
      </div>

      <div class="meta">
        <span class="date">{t("cockpit.noteDetails.created", { date: formatDate(note.created_at) })}</span>
        <span class="date">{t("cockpit.noteDetails.updated", { date: formatDate(note.updated_at) })}</span>
      </div>

      <div class="content">{note.content}</div>

      {#if tags.length > 0}
        <div class="tags">
          {#each tags as tag}
            <span class="tag">#{tag}</span>
          {/each}
        </div>
      {/if}

      <div class="links-section">
        <div class="links-header">
          <h3>{t("cockpit.noteDetails.linksTitle", { count: links.length })}</h3>
          {#if links.length > 0}
            <button type="button" class="delete-all-links-btn" onclick={() => (showDeleteLinksConfirm = true)} aria-label={t("cockpit.noteDetails.deleteAllAria")}>
              {t("cockpit.noteDetails.deleteAll")}
            </button>
          {/if}
        </div>
        {#if links.length === 0}
          <p class="no-links">{t("cockpit.noteDetails.noLinks")}</p>
        {:else}
          <div class="links-list">
            {#each links as link}
              <div class="link-item">
                <span class="link-type">{link.link_type}</span>
                <span class="link-weight">{t("cockpit.noteDetails.weight", { weight: link.weight.toFixed(1) })}</span>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <div class="panel-footer">
        <button type="button" class="view-full-btn" onclick={() => note && goto(`/notes/${note.id}`)}>
          {t("cockpit.noteDetails.viewFull")}
        </button>
      </div>
    {/if}
  </div>
</div>

{#if showDeleteLinksConfirm}
  <div
    class="modal-overlay"
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    onclick={() => (showDeleteLinksConfirm = false)}
    onkeydown={(e) => e.key === "Escape" && (showDeleteLinksConfirm = false)}
  >
    <!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
    <div class="modal" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.stopPropagation()}>
      <h3>{t("cockpit.noteDetails.deleteLinksTitle")}</h3>
      <p>{t("cockpit.noteDetails.deleteLinksMessage", { count: links.length })}</p>
      <div class="modal-actions">
        <button type="button" class="modal-btn cancel" onclick={() => (showDeleteLinksConfirm = false)} disabled={deletingLinks}>
          {t("cockpit.noteDetails.cancel")}
        </button>
        <button type="button" class="modal-btn delete" onclick={handleDeleteAllLinks} disabled={deletingLinks}>
          {deletingLinks ? t("cockpit.noteDetails.delete") + "..." : t("cockpit.noteDetails.deleteAll")}
          <!-- `cockpit.noteDetails.delete` = "Delete" used for in-progress ellipsis -->
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .note-details {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }

  .details-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 14px;
    border-bottom: 1px solid rgba(45, 212, 191, 0.15);
  }

  .close-btn,
  .action-btn {
    width: 28px;
    height: 28px;
    border: none;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.05);
    color: var(--color-text, #e0e0e0);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    transition: background 0.2s ease;
  }

  .close-btn:hover,
  .action-btn:hover {
    background: rgba(45, 212, 191, 0.2);
  }

  .actions {
    display: flex;
    gap: 6px;
  }

  .action-btn.delete:hover {
    background: rgba(248, 113, 113, 0.25);
    color: #f87171;
  }

  .action-btn.share:hover {
    background: rgba(96, 165, 250, 0.25);
    color: #60a5fa;
  }

  .details-content {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px;
    color: rgba(255, 255, 255, 0.5);
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid rgba(255, 255, 255, 0.1);
    border-top-color: #2dd4bf;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 12px;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .error {
    padding: 16px;
    background: rgba(248, 113, 113, 0.1);
    color: #f87171;
    border-radius: 8px;
    text-align: center;
  }

  .note-header {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 20px;
  }

  .type-icon {
    font-size: 32px;
  }

  .title {
    font-size: 22px;
    font-weight: 700;
    color: white;
    margin: 0;
    line-height: 1.3;
  }

  .type-badge {
    display: inline-block;
    padding: 4px 12px;
    background: rgba(45, 212, 191, 0.15);
    color: #2dd4bf;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
    border-radius: 20px;
    width: fit-content;
    border: 1px solid rgba(45, 212, 191, 0.25);
  }

  .meta {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 20px;
    padding-bottom: 16px;
    border-bottom: 1px solid rgba(45, 212, 191, 0.1);
  }

  .date {
    font-size: 12px;
    color: rgba(255, 255, 255, 0.5);
  }

  .content {
    font-size: 14px;
    line-height: 1.7;
    color: rgba(255, 255, 255, 0.85);
    white-space: pre-wrap;
    margin-bottom: 20px;
  }

  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-bottom: 20px;
  }

  .tag {
    padding: 4px 10px;
    background: rgba(45, 212, 191, 0.08);
    color: #2dd4bf;
    font-size: 12px;
    border-radius: 16px;
  }

  .links-section {
    margin-top: 20px;
    padding-top: 16px;
    border-top: 1px solid rgba(45, 212, 191, 0.1);
  }

  .links-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .links-header h3 {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    color: white;
  }

  .delete-all-links-btn {
    padding: 4px 8px;
    border: 1px solid rgba(248, 113, 113, 0.3);
    background: rgba(248, 113, 113, 0.08);
    color: #f87171;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s ease;
  }

  .delete-all-links-btn:hover {
    background: rgba(248, 113, 113, 0.15);
  }

  .no-links {
    color: rgba(255, 255, 255, 0.4);
    font-size: 13px;
    margin: 0;
  }

  .links-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .link-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 10px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(45, 212, 191, 0.1);
    border-radius: 6px;
    font-size: 12px;
  }

  .link-type {
    font-weight: 600;
    color: white;
  }

  .link-weight {
    color: rgba(255, 255, 255, 0.5);
  }

  .panel-footer {
    padding-top: 20px;
    border-top: 1px solid rgba(45, 212, 191, 0.1);
  }

  .view-full-btn {
    width: 100%;
    padding: 12px;
    border: 1px solid rgba(45, 212, 191, 0.3);
    background: rgba(45, 212, 191, 0.1);
    border-radius: 8px;
    color: #2dd4bf;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s ease;
  }

  .view-full-btn:hover {
    background: rgba(45, 212, 191, 0.2);
  }

  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 500;
  }

  .modal {
    background: rgba(10, 10, 15, 0.95);
    border: 1px solid rgba(45, 212, 191, 0.25);
    border-radius: 12px;
    padding: 20px;
    max-width: 360px;
    width: 90%;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  }

  .modal h3 {
    margin: 0 0 12px 0;
    font-size: 16px;
    color: white;
  }

  .modal p {
    margin: 0 0 20px 0;
    color: rgba(255, 255, 255, 0.7);
    line-height: 1.5;
  }

  .modal-actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
  }

  .modal-btn {
    padding: 8px 14px;
    border: none;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.2s ease;
  }

  .modal-btn.cancel {
    background: rgba(255, 255, 255, 0.08);
    color: var(--color-text, #e0e0e0);
  }

  .modal-btn.cancel:hover {
    background: rgba(255, 255, 255, 0.15);
  }

  .modal-btn.delete {
    background: rgba(248, 113, 113, 0.85);
    color: white;
  }

  .modal-btn.delete:hover {
    background: #f87171;
  }

  .modal-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
