<script lang="ts">
  import { getNote, type Note } from "$shared/api/notes";
  import {
    getNoteLinks,
    deleteAllNoteLinks,
    updateLink,
    deleteLink,
    type Link,
  } from "$shared/api/links";
  import { goto } from "$app/navigation";
  import { formatDate } from "$shared/utils/date";
  import { CelestialBody, LinkType } from "$entities";
  import LinkTypeSelector from "$components/molecules/LinkTypeSelector.svelte";
  import IconButton from "$components/atoms/IconButton.svelte";
  import Chip from "$components/atoms/Chip.svelte";
  import ConfirmModal from "$widgets/confirm/ConfirmModal.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  interface Props {
    nodeId: string;
    onClose?: () => void;
    onEdit?: (id: string) => void;
    onDelete?: (id: string) => void;
    onCreateChildNote?: (note: Note) => void;
  }

  const { nodeId, onClose, onEdit, onDelete, onCreateChildNote }: Props = $props();

  let note = $state<Note | null>(null);
  let links = $state<Link[]>([]);
  let loading = $state(true);
  let error = $state("");
  let showDeleteLinksConfirm = $state(false);
  let editingLinkId = $state<string | null>(null);
  let editDraft = $state<{ link_type: string; weight: number } | null>(null);
  let linkActionError = $state("");
  let savingLink = $state(false);
  let deletingLinkId = $state<string | null>(null);

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
    try {
      await deleteAllNoteLinks(nodeId);
      links = [];
      showDeleteLinksConfirm = false;
    } catch {
      error = t("cockpit.noteDetails.deleteLinksError");
    }
  }

  function getNoteTypeIcon(type: string | undefined): string {
    return type ? CelestialBody.fromString(type).emoji : CelestialBody.STAR.emoji;
  }

  function getNoteTypeLabel(type: string | undefined): string {
    return type ? CelestialBody.fromString(type).label : t("cockpit.noteDetails.fallbackType");
  }

  function getLinkType(link: Link): LinkType {
    return LinkType.fromString(link.link_type);
  }

  function startEditLink(link: Link) {
    if (savingLink || deletingLinkId) return;
    editingLinkId = link.id;
    editDraft = { link_type: link.link_type, weight: link.weight };
    linkActionError = "";
  }

  function cancelEditLink() {
    editingLinkId = null;
    editDraft = null;
    linkActionError = "";
  }

  async function saveEditLink(link: Link) {
    if (!editDraft) return;
    savingLink = true;
    linkActionError = "";
    try {
      await updateLink(link.id, {
        link_type: editDraft.link_type,
        weight: editDraft.weight,
      });
      await loadLinks(nodeId);
      editingLinkId = null;
      editDraft = null;
    } catch {
      linkActionError = t("cockpit.noteDetails.linkUpdateError");
    } finally {
      savingLink = false;
    }
  }

  async function handleDeleteLink(link: Link) {
    deletingLinkId = link.id;
    linkActionError = "";
    try {
      await deleteLink(link.id);
      links = links.filter((l) => l.id !== link.id);
    } catch {
      linkActionError = t("cockpit.noteDetails.linkDeleteError");
    } finally {
      deletingLinkId = null;
    }
  }

  function formatRelativeTime(iso: string | undefined): string {
    if (!iso) return "";
    const date = new Date(iso);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMin = Math.floor(diffMs / 60000);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);
    if (diffMin < 1) return t("time.justNow");
    if (diffMin < 60) return t("time.minutesAgo", { count: diffMin });
    if (diffHour < 24) return t("time.hoursAgo", { count: diffHour });
    if (diffDay < 30) return t("time.daysAgo", { count: diffDay });
    return formatDate(iso);
  }
</script>

<div class="note-details" data-testid="cockpit-note-details">
  <div class="details-header">
    <IconButton variant="ghost" size="md" onClick={() => onClose?.()} title={t("cockpit.noteDetails.close")}>
      ✕
    </IconButton>
    {#if note}
      <div class="actions">
        <IconButton
          size="md"
          onClick={() => onEdit?.(nodeId)}
          title={t("cockpit.noteDetails.edit")}
          aria-label={t("cockpit.noteDetails.edit")}
        >
          ✎
        </IconButton>
        <IconButton
          size="md"
          onClick={() => note && onCreateChildNote?.(note)}
          title={t("cockpit.noteDetails.createChildNote")}
          data-testid="note-details-create-child"
        >
          ＋
        </IconButton>
        <IconButton
          size="md"
          variant="ghost"
          onClick={() => {
            /* share */
          }}
          title={t("cockpit.noteDetails.share")}
          aria-label={t("cockpit.noteDetails.share")}
        >
          ⇄
        </IconButton>
        <IconButton
          size="md"
          variant="danger"
          onClick={() => onDelete?.(nodeId)}
          title={t("cockpit.noteDetails.deleteNote")}
          aria-label={t("cockpit.noteDetails.deleteNote")}
        >
          🗑
        </IconButton>
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
        <span class="type-icon">{getNoteTypeIcon(note.type)}</span>
        <h2 class="title">{note.title}</h2>
        <Chip size="sm" color={CelestialBody.fromString(note.type).toCSSColor()} borderColor={CelestialBody.fromString(note.type).toCSSColor()}>
          {getNoteTypeLabel(note.type)}
        </Chip>
      </div>

      <div class="meta">
        <span class="date"
          >{t("cockpit.noteDetails.created", { date: formatDate(note.created_at) })}</span
        >
        <span class="date"
          >{t("cockpit.noteDetails.updated", { date: formatDate(note.updated_at) })}</span
        >
      </div>

      <div class="content">{note.content}</div>

      {#if tags.length > 0}
        <div class="tags">
          {#each tags as tag}
            <Chip size="sm" color="var(--carbon-glow-cyan, #22d3ee)">#{tag}</Chip>
          {/each}
        </div>
      {/if}

      <div class="links-section">
        <div class="links-header">
          <h3>{t("cockpit.noteDetails.linksTitle", { count: links.length })}</h3>
          {#if links.length > 0}
            <button
              type="button"
              class="delete-all-links-btn"
              onclick={() => (showDeleteLinksConfirm = true)}
              aria-label={t("cockpit.noteDetails.deleteAllAria")}
            >
              {t("cockpit.noteDetails.deleteAll")}
            </button>
          {/if}
        </div>
        {#if linkActionError}
          <div class="link-action-error" role="alert">{linkActionError}</div>
        {/if}
        {#if links.length === 0}
          <p class="no-links">{t("cockpit.noteDetails.noLinks")}</p>
        {:else}
          <div class="links-list">
            {#each links as link}
              {@const linkType = getLinkType(link)}
              {@const isEditing = editingLinkId === link.id}
              {@const isBusy = savingLink || deletingLinkId === link.id}
              <div class="link-item" class:editing={isEditing} class:busy={isBusy}>
                {#if isEditing && editDraft}
                  <div class="link-edit-form">
                    <LinkTypeSelector
                      types={LinkType.CREATABLE_TYPES}
                      selected={editDraft.link_type}
                      size="sm"
                      showDescription={false}
                      onSelect={(type) => {
                        if (editDraft) {
                          editDraft.link_type = type;
                          editDraft.weight = LinkType.fromString(type).defaultWeight;
                        }
                      }}
                    />
                    <div class="link-weight-edit">
                      <label for="edit-link-weight-{link.id}">
                        {t("cockpit.noteDetails.linkWeight", {
                          weight: editDraft.weight.toFixed(1),
                        })}
                      </label>
                      <input
                        id="edit-link-weight-{link.id}"
                        type="range"
                        min="0.1"
                        max="1.0"
                        step="0.1"
                        bind:value={editDraft.weight}
                      />
                    </div>
                    <div class="link-edit-actions">
                      <button
                        type="button"
                        class="link-edit-save"
                        disabled={savingLink}
                        onclick={() => saveEditLink(link)}
                      >
                        {savingLink
                          ? t("cockpit.noteDetails.saving") + "..."
                          : t("cockpit.noteDetails.save")}
                      </button>
                      <button
                        type="button"
                        class="link-edit-cancel"
                        disabled={savingLink}
                        onclick={cancelEditLink}
                      >
                        {t("cockpit.noteDetails.cancel")}
                      </button>
                    </div>
                  </div>
                {:else}
                  <div class="link-info">
                    <Chip
                      size="sm"
                      color={linkType.color}
                      borderColor={linkType.color}
                      background="{linkType.color}33"
                    >
                      <span class="link-type-icon">{linkType.icon}</span>
                      <span>{linkType.label}</span>
                    </Chip>
                    <span
                      class="link-weight-bar"
                      title={t("cockpit.noteDetails.weight", { weight: link.weight.toFixed(2) })}
                    >
                      <span
                        class="link-weight-fill"
                        style="width: {link.weight * 100}%; background: {linkType.color}"
                      ></span>
                      <span class="link-weight-value">{link.weight.toFixed(1)}</span>
                    </span>
                    {#if link.source_type === "gamma"}
                      <Chip
                        size="sm"
                        color="#c4b5fd"
                        borderColor="rgba(139, 92, 246, 0.3)"
                        background="linear-gradient(135deg, rgba(139, 92, 246, 0.2), rgba(168, 85, 247, 0.2))"
                      >
                        {t("linkTooltip.recommended")}
                      </Chip>
                    {/if}
                    {#if link.last_weight_update}
                      <span class="link-last-update" title={formatDate(link.last_weight_update)}>
                        {formatRelativeTime(link.last_weight_update)}
                      </span>
                    {/if}
                  </div>
                  <div class="link-actions">
                    <IconButton
                      size="sm"
                      onClick={() => startEditLink(link)}
                      title={t("cockpit.noteDetails.editLink")}
                      aria-label={t("cockpit.noteDetails.editLink")}
                      disabled={!!editingLinkId || !!deletingLinkId}
                    >
                      ✎
                    </IconButton>
                    <IconButton
                      size="sm"
                      variant="danger"
                      onClick={() => handleDeleteLink(link)}
                      title={t("cockpit.noteDetails.deleteLink")}
                      aria-label={t("cockpit.noteDetails.deleteLink")}
                      disabled={!!editingLinkId || !!deletingLinkId}
                    >
                      {deletingLinkId === link.id ? t("cockpit.noteDetails.deleting") + "..." : "🗑"}
                    </IconButton>
                  </div>
                {/if}
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
        >
          {t("cockpit.noteDetails.viewFull")}
        </button>
      </div>
    {/if}
  </div>
</div>

<ConfirmModal
  bind:open={showDeleteLinksConfirm}
  title={t("cockpit.noteDetails.deleteLinksTitle")}
  message={t("cockpit.noteDetails.deleteLinksMessage", { count: links.length })}
  confirmText={t("cockpit.noteDetails.deleteAll")}
  cancelText={t("cockpit.noteDetails.cancel")}
  danger={true}
  onConfirm={handleDeleteAllLinks}
  onCancel={() => (showDeleteLinksConfirm = false)}
/>

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

  .actions {
    display: flex;
    gap: 6px;
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

  .link-action-error {
    padding: 8px 10px;
    margin-bottom: 10px;
    background: rgba(248, 113, 113, 0.1);
    color: #f87171;
    border: 1px solid rgba(248, 113, 113, 0.2);
    border-radius: 6px;
    font-size: 12px;
  }

  .link-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(45, 212, 191, 0.1);
    border-radius: 6px;
    font-size: 12px;
    transition: background 0.15s ease;
  }

  .link-item:hover:not(.editing):not(.busy) {
    background: rgba(255, 255, 255, 0.06);
  }

  .link-item.busy {
    opacity: 0.6;
  }

  .link-info {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px;
    min-width: 0;
    flex: 1;
  }

  .link-type-icon {
    font-size: 12px;
  }

  .link-weight-bar {
    position: relative;
    display: inline-flex;
    align-items: center;
    width: 80px;
    height: 18px;
    background: rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .link-weight-fill {
    display: block;
    height: 100%;
    border-radius: 10px;
    opacity: 0.7;
  }

  .link-weight-value {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    font-weight: 600;
    color: white;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
  }

  .link-last-update {
    color: rgba(255, 255, 255, 0.4);
    font-size: 10px;
    white-space: nowrap;
  }

  .link-actions {
    display: flex;
    gap: 4px;
    flex-shrink: 0;
  }

  .link-edit-form {
    display: flex;
    flex-direction: column;
    gap: 10px;
    width: 100%;
  }

  .link-weight-edit {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .link-weight-edit label {
    color: rgba(255, 255, 255, 0.7);
    font-size: 12px;
  }

  .link-weight-edit input[type="range"] {
    width: 100%;
    accent-color: #fbbf24;
  }

  .link-edit-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }

  .link-edit-actions button {
    padding: 6px 12px;
    border: none;
    border-radius: 6px;
    font-size: 12px;
    font-weight: 600;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .link-edit-save {
    background: rgba(45, 212, 191, 0.2);
    color: #2dd4bf;
  }

  .link-edit-save:hover:not(:disabled) {
    background: rgba(45, 212, 191, 0.3);
  }

  .link-edit-cancel {
    background: rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 0.8);
  }

  .link-edit-cancel:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.15);
  }
</style>
