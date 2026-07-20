<script lang="ts">
  import { getNote, type Note } from "$shared/api/notes";
  import {
    getNoteLinks,
    deleteAllNoteLinks,
    type Link,
  } from "$shared/api/links";
  import { goto } from "$app/navigation";
  import { formatDate } from "$shared/utils/date";
  import { CelestialBody } from "$shared/lib/domain";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import ShareModal from "$components/organisms/ShareModal.svelte";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const {
    nodeId,
    onClose,
    onEdit,
    onDelete,
  }: {
    nodeId: string;
    onClose: () => void;
    onEdit?: (id: string) => void;
    onDelete?: (id: string) => void;
  } = $props();

  let note = $state<Note | null>(null);
  let links = $state<Link[]>([]);
  let loading = $state(true);
  let error = $state("");
  let showShareModal = $state(false);
  let showDeleteLinksConfirm = $state(false);
  let deletingLinks = $state(false);

  // Load note when nodeId changes
  $effect(() => {
    const id = nodeId; // track dependency
    loadNote(id);
    loadLinks(id);
  });

  async function loadNote(id: string) {
    loading = true;
    error = "";
    try {
      note = await getNote(id);
    } catch {
      error = t("noteSidePanel.loadError");
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
      error = t("noteSidePanel.deleteLinksError");
    } finally {
      deletingLinks = false;
    }
  }

  function getTypeIcon(type: string | undefined): string {
    return type
      ? CelestialBody.fromString(type).emoji
      : CelestialBody.STAR.emoji;
  }

  function getTypeLabel(type: string | undefined): string {
    return type
      ? CelestialBody.fromString(type).label
      : t("noteSidePanel.fallbackType");
  }
</script>

<div class="side-panel" data-testid="note-side-panel" class:open={true}>
  <div class="panel-header">
    <button
      class="close-btn"
      onclick={onClose}
      aria-label={t("noteSidePanel.closeAria")}
      data-testid="sidepanel-close-btn"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <line x1="18" y1="6" x2="6" y2="18" />
        <line x1="6" y1="6" x2="18" y2="18" />
      </svg>
    </button>

    {#if !loading && note}
      <div class="actions">
        <button
          class="action-btn share"
          onclick={() => (showShareModal = true)}
          aria-label={t("noteSidePanel.shareAria")}
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <circle cx="18" cy="5" r="3" />
            <circle cx="6" cy="12" r="3" />
            <circle cx="18" cy="19" r="3" />
            <line x1="8.59" y1="13.51" x2="15.42" y2="17.49" />
            <line x1="15.41" y1="6.51" x2="8.59" y2="10.49" />
          </svg>
        </button>
        <button
          class="action-btn"
          onclick={() => onEdit?.(nodeId)}
          aria-label={t("noteSidePanel.editAria")}
          data-testid="sidepanel-edit-btn"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <path
              d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"
            />
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
          </svg>
        </button>
        <button
          class="action-btn delete"
          onclick={() => onDelete?.(nodeId)}
          aria-label={t("noteSidePanel.deleteAria")}
          data-testid="sidepanel-delete-btn"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <polyline points="3 6 5 6 21 6" />
            <path
              d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"
            />
          </svg>
        </button>
      </div>
    {/if}
  </div>

  <div class="panel-content">
    {#if loading}
      <div class="loading" role="status" aria-live="polite">
        <div class="spinner" aria-hidden="true"></div>
        <p>{t("noteSidePanel.loading")}</p>
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
        <span class="date"
          >{t("noteSidePanel.created", {
            date: formatDate(note.created_at),
          })}</span
        >
        <span class="date"
          >{t("noteSidePanel.updated", {
            date: formatDate(note.updated_at),
          })}</span
        >
      </div>

      <div class="content">
        {note.content}
      </div>

      {#if note.metadata?.tags?.length > 0}
        <div class="tags">
          {#each note.metadata.tags as tag}
            <span class="tag">#{tag}</span>
          {/each}
        </div>
      {/if}

      <div class="links-section">
        <div class="links-header">
          <h3>{t("noteSidePanel.linksTitle", { count: links.length })}</h3>
          {#if links.length > 0}
            <button
              class="delete-all-links-btn"
              onclick={() => (showDeleteLinksConfirm = true)}
              aria-label={t("noteSidePanel.deleteAllAria")}
            >
              {t("noteSidePanel.deleteAll")}
            </button>
          {/if}
        </div>
        {#if links.length === 0}
          <p class="no-links">{t("noteSidePanel.noLinks")}</p>
        {:else}
          <div class="links-list">
            {#each links as link}
              <div class="link-item">
                <span class="link-type">{link.link_type}</span>
                <span class="link-weight"
                  >{t("noteSidePanel.weight", {
                    weight: link.weight.toFixed(1),
                  })}</span
                >
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <div class="panel-footer">
        <button
          type="button"
          class="view-full-btn"
          onclick={() => note && goto(`/notes/${note.id}`)}
          aria-label={t("noteSidePanel.viewFullAria", { title: note.title })}
        >
          {t("noteSidePanel.viewFullPage")}
        </button>
      </div>
    {/if}
  </div>
</div>

{#if showShareModal && note}
  <ShareModal
    noteId={note.id}
    noteTitle={note.title}
    on:close={() => (showShareModal = false)}
  />
{/if}

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
    <div
      class="modal"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <h3>{t("noteSidePanel.deleteLinksTitle")}</h3>
      <p>
        {t("noteSidePanel.deleteLinksMessage", { count: links.length })}
      </p>
      <div class="modal-actions">
        <button
          class="modal-btn cancel"
          onclick={() => (showDeleteLinksConfirm = false)}
          disabled={deletingLinks}
        >
          {t("noteSidePanel.cancel")}
        </button>
        <button
          class="modal-btn delete"
          onclick={handleDeleteAllLinks}
          disabled={deletingLinks}
        >
          {deletingLinks
            ? t("noteSidePanel.delete") + "..."
            : t("noteSidePanel.deleteAll")}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .side-panel {
    position: fixed;
    top: 0;
    right: -400px;
    width: 400px;
    max-height: 100vh;
    background: white;
    box-shadow: -4px 0 20px rgba(0, 0, 0, 0.1);
    z-index: 200;
    transition: right 0.3s ease;
    display: flex;
    flex-direction: column;
  }

  .side-panel.open {
    right: 0;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid #e2e8f0;
  }

  .close-btn {
    padding: 8px;
    border: none;
    background: #f1f5f9;
    border-radius: 8px;
    cursor: pointer;
    color: #64748b;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: #e2e8f0;
    color: #334155;
  }

  .actions {
    display: flex;
    gap: 8px;
  }

  .action-btn {
    padding: 8px 12px;
    border: none;
    background: #f1f5f9;
    border-radius: 8px;
    cursor: pointer;
    color: #64748b;
    transition: all 0.2s;
  }

  .action-btn:hover {
    background: #e2e8f0;
    color: #334155;
  }

  .action-btn.delete:hover {
    background: #fee2e2;
    color: #ef4444;
  }

  .action-btn.share:hover {
    background: #dbeafe;
    color: #3b82f6;
  }

  .panel-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px;
    color: #94a3b8;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid #e2e8f0;
    border-top-color: #3b82f6;
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
    padding: 20px;
    background: #fee2e2;
    color: #ef4444;
    border-radius: 8px;
    text-align: center;
  }

  .note-header {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 20px;
  }

  .type-icon {
    font-size: 32px;
  }

  .title {
    font-size: 24px;
    font-weight: 600;
    color: #1e293b;
    line-height: 1.3;
    margin: 0;
  }

  .type-badge {
    display: inline-block;
    padding: 4px 12px;
    background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
    color: white;
    font-size: 12px;
    font-weight: 500;
    text-transform: uppercase;
    border-radius: 20px;
    width: fit-content;
  }

  .meta {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid #e2e8f0;
  }

  .date {
    font-size: 13px;
    color: #94a3b8;
  }

  .content {
    font-size: 15px;
    line-height: 1.7;
    color: #475569;
    white-space: pre-wrap;
    margin-bottom: 24px;
  }

  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 24px;
  }

  .tag {
    padding: 4px 10px;
    background: #f1f5f9;
    color: #64748b;
    font-size: 12px;
    border-radius: 16px;
  }

  .panel-footer {
    padding-top: 20px;
    border-top: 1px solid #e2e8f0;
  }

  .view-full-btn {
    width: 100%;
    padding: 12px;
    border: 1px solid #e2e8f0;
    background: white;
    border-radius: 8px;
    color: #3b82f6;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .view-full-btn:hover {
    background: #f8fafc;
    border-color: #3b82f6;
  }

  .links-section {
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid #e2e8f0;
  }

  .links-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .links-header h3 {
    margin: 0;
    font-size: 14px;
    font-weight: 600;
    color: #334155;
  }

  .delete-all-links-btn {
    padding: 4px 8px;
    border: 1px solid #fee2e2;
    background: #fef2f2;
    color: #ef4444;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .delete-all-links-btn:hover {
    background: #fee2e2;
  }

  .no-links {
    color: #94a3b8;
    font-size: 13px;
    margin: 0;
  }

  .links-list {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .link-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    background: #f8fafc;
    border-radius: 6px;
    font-size: 12px;
  }

  .link-type {
    font-weight: 500;
    color: #334155;
  }

  .link-weight {
    color: #64748b;
  }

  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
  }

  .modal {
    background: white;
    border-radius: 12px;
    padding: 24px;
    max-width: 400px;
    width: 90%;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  }

  .modal h3 {
    margin: 0 0 12px 0;
    font-size: 18px;
    color: #1e293b;
  }

  .modal p {
    margin: 0 0 20px 0;
    color: #64748b;
    line-height: 1.5;
  }

  .modal-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  .modal-btn {
    padding: 8px 16px;
    border: none;
    border-radius: 6px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .modal-btn.cancel {
    background: #f1f5f9;
    color: #64748b;
  }

  .modal-btn.cancel:hover {
    background: #e2e8f0;
  }

  .modal-btn.delete {
    background: #ef4444;
    color: white;
  }

  .modal-btn.delete:hover {
    background: #dc2626;
  }

  .modal-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  @media (max-width: 768px) {
    .side-panel {
      width: 100%;
      right: -100%;
    }
  }
</style>
