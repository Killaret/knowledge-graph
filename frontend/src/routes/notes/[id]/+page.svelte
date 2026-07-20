<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { page } from "$app/stores";
  import { getNote, getSuggestions, deleteNote } from "$shared/api/notes";
  import type { Note, Suggestion } from "$shared/api/notes";
  import { goto } from "$app/navigation";
  import { formatDateTime } from "$shared/utils/date";
  import BackButton from "$components/atoms/BackButton.svelte";
  import EditNoteModal from "$components/organisms/EditNoteModal.svelte";
  import StateIllustration from "$components/atoms/StateIllustration.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  let note: Note | null = $state(null);
  let suggestions: Suggestion[] = $state([]);
  let loading = $state(true);
  let error = $state("");
  let editModalOpen = $state(false);

  function getRouteId(): string {
    const id = $page.params.id;
    if (!id) throw new Error("Missing route parameter: id");
    return id;
  }

  onMount(async () => {
    try {
      const id = getRouteId();
      note = await getNote(id);
      suggestions = await getSuggestions(id, 5);
    } catch (e: any) {
      if (e.response?.status === 404) {
        error = t("note.notFoundShort");
        setTimeout(() => goto("/"), 3000);
      } else {
        error = t("note.loadError");
      }
    } finally {
      loading = false;
    }
  });

  async function handleDelete() {
    if (!browser) return;
    if (!confirm(t("note.deleteConfirm"))) return;
    const id = getRouteId();
    await deleteNote(id);
    await goto("/");
  }
</script>

{#if loading}
  <p>{t("note.loading")}</p>
{:else if error}
  <div class="note-error">
    <StateIllustration
      type={error === t("note.notFoundShort") ? "404" : "error"}
    />
    <p class="error">{error}</p>
  </div>
{:else if note}
  <div class="note-container">
    <BackButton href="/" />
    <h1 data-testid="note-detail-title">{note.title}</h1>
    <div class="meta">
      {t("note.createdLabel")}{formatDateTime(note.created_at)}
    </div>
    <div class="content" data-testid="note-detail-content">{note.content}</div>
    <div class="actions">
      <button
        onclick={() => (editModalOpen = true)}
        class="edit-btn"
        data-testid="edit-note-btn">{t("note.editButton")}</button
      >
      <button onclick={handleDelete} data-testid="delete-note-btn"
        >{t("note.deleteButton")}</button
      >
      <a href={`/graph/3d/${note.id}`} class="graph-link"
        >{t("note.showConstellation")}</a
      >
    </div>

    <EditNoteModal
      bind:open={editModalOpen}
      noteId={note.id}
      onSuccess={(updatedNote: Note) => (note = updatedNote)}
    />

    {#if suggestions.length}
      <h2>{t("note.similarNotes")}</h2>
      <ul class="suggestions">
        {#each suggestions as s}
          <li>
            <a href={`/notes/${s.note_id}`}>{s.title}</a>
            <span class="score"
              >{t("note.score", { score: s.score.toFixed(3) })}</span
            >
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}

<style>
  .note-container {
    max-width: 800px;
    margin: 0 auto;
  }
  .content {
    white-space: pre-wrap;
    margin: 1rem 0;
  }
  .actions {
    display: flex;
    gap: 1rem;
    margin: 1rem 0;
  }
  .actions button {
    padding: 0.5rem 1rem;
    border: 1px solid var(--color-border);
    border-radius: 4px;
    cursor: pointer;
    background: var(--color-surface-elevated);
  }
  .actions button:hover {
    background: var(--color-background);
  }

  .note-error {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    text-align: center;
  }
  .edit-btn {
    background: var(--color-primary) !important;
    color: white;
    border-color: var(--color-primary) !important;
  }
  .edit-btn:hover {
    background: var(--color-primary-hover) !important;
  }
  .suggestions li {
    margin-bottom: 0.5rem;
  }
  .score {
    margin-left: 1rem;
    color: var(--color-text-secondary);
    font-size: 0.9rem;
  }
  .error {
    color: var(--color-danger);
  }
  .graph-link {
    background: var(--color-galaxy);
    color: white;
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    text-decoration: none;
  }
</style>
