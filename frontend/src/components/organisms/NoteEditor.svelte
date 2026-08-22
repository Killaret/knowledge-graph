<script lang="ts">
  import { goto } from "$app/navigation";
  import { createNote, updateNote, getNote, publishNote, unpublishNote } from "$shared/api/notes";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import type { ErrorResponse } from "$shared/types/errors";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  interface Props {
    noteId?: string | null;
    onCancel?: () => void;
  }

  const { noteId = null, onCancel }: Props = $props();

  let title = $state("");
  let content = $state("");
  let noteType = $state("star");
  let isPublic = $state(false);
  let isLoading = $state(false);
  let isSaving = $state(false);
  let isTogglingPublic = $state(false);
  let apiError = $state<ErrorResponse | null>(null);
  let titleError = $state<string | null>(null);

  function extractApiError(err: unknown): ErrorResponse | undefined {
    if (typeof err !== "object" || err === null || !("response" in err)) return undefined;
    const response = (err as { response?: unknown }).response;
    if (typeof response !== "object" || response === null || !("data" in response))
      return undefined;
    const data = (response as { data?: unknown }).data;
    return typeof data === "object" && data !== null ? (data as ErrorResponse) : undefined;
  }

  // Загрузка данных при редактировании
  $effect(() => {
    if (noteId) {
      loadNote(noteId);
    }
  });

  async function loadNote(id: string) {
    isLoading = true;
    apiError = null;
    try {
      const note = await getNote(id);
      title = note.title;
      content = note.content || "";
      noteType = note.type || "star";
      isPublic = note.is_public ?? false;
    } catch (err: unknown) {
      apiError = extractApiError(err) || {
        code: "API_ERROR",
        message: t("noteEditor.loadError"),
      };
    } finally {
      isLoading = false;
    }
  }

  function validate(): boolean {
    titleError = null;
    if (!title.trim()) {
      titleError = t("noteEditor.titleRequired");
      return false;
    }
    return true;
  }

  async function handleSave() {
    if (!validate()) return;

    isSaving = true;
    apiError = null;

    try {
      const noteData = {
        title: title.trim(),
        content: content.trim(),
        type: noteType,
      };

      if (noteId) {
        await updateNote(noteId, noteData);
      } else {
        const newNote = await createNote(noteData);
        await goto(`/notes/${newNote.id}`);
      }
    } catch (err: unknown) {
      apiError = extractApiError(err) || {
        code: "API_ERROR",
        message: t("noteEditor.saveError"),
      };
    } finally {
      isSaving = false;
    }
  }

  async function togglePublic(e: Event) {
    const target = e.target as HTMLInputElement;
    const nextPublic = target.checked;
    if (!noteId) return;

    isTogglingPublic = true;
    apiError = null;
    try {
      const updated = nextPublic ? await publishNote(noteId) : await unpublishNote(noteId);
      isPublic = updated.is_public ?? nextPublic;
    } catch (err: unknown) {
      target.checked = isPublic;
      apiError = extractApiError(err) || {
        code: "API_ERROR",
        message: t("noteEditor.publicToggleError"),
      };
    } finally {
      isTogglingPublic = false;
    }
  }

  function handleCancel() {
    if (onCancel) {
      onCancel();
    } else {
      goto("/");
    }
  }
</script>

<div class="note-editor" data-testid="note-editor">
  {#if isLoading}
    <div class="loading" data-testid="loading">{t("noteEditor.loading")}</div>
  {:else}
    <ApiErrorDisplay error={apiError} onClose={() => (apiError = null)} />

    <form
      onsubmit={(e) => {
        e.preventDefault();
        handleSave();
      }}
    >
      <div class="field">
        <label for="title">{t("noteEditor.titleLabel")}</label>
        <input
          id="title"
          type="text"
          bind:value={title}
          placeholder={t("noteEditor.titlePlaceholder")}
          data-testid="title-input"
          disabled={isSaving}
        />
        {#if titleError}
          <span class="field-error" data-testid="title-error">{titleError}</span>
        {/if}
      </div>

      <div class="field">
        <label for="type">{t("noteEditor.typeLabel")}</label>
        <select id="type" bind:value={noteType} data-testid="type-select" disabled={isSaving}>
          <option value="star">{t("celestialBody.type.star")}</option>
          <option value="planet">{t("celestialBody.type.planet")}</option>
          <option value="comet">{t("celestialBody.type.comet")}</option>
        </select>
      </div>

      <div class="field">
        <label for="content">{t("noteEditor.contentLabel")}</label>
        <textarea
          id="content"
          bind:value={content}
          placeholder={t("noteEditor.contentPlaceholder")}
          rows="10"
          data-testid="content-input"
          disabled={isSaving}
        ></textarea>
      </div>

      {#if noteId}
        <div class="field field--inline">
          <label for="is-public" class="toggle-label">
            <input
              id="is-public"
              type="checkbox"
              checked={isPublic}
              onchange={togglePublic}
              disabled={isTogglingPublic}
              data-testid="public-toggle"
            />
            {t("noteEditor.publicLabel")}
          </label>
          <span class="hint">{t("noteEditor.publicHint")}</span>
        </div>
      {/if}

      <div class="actions">
        <button type="submit" class="btn-primary" disabled={isSaving} data-testid="save-button">
          {isSaving
            ? t("noteEditor.saving")
            : noteId
              ? t("noteEditor.update")
              : t("noteEditor.create")}
        </button>
        <button
          type="button"
          class="btn-secondary"
          onclick={handleCancel}
          disabled={isSaving}
          data-testid="cancel-button"
        >
          {t("noteEditor.cancel")}
        </button>
      </div>
    </form>
  {/if}
</div>

<style>
  .note-editor {
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem;
  }

  .loading {
    text-align: center;
    padding: 2rem;
    color: var(--color-text-secondary);
  }

  .field {
    margin-bottom: 1.5rem;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
  }

  input,
  select,
  textarea {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid var(--color-border);
    border-radius: 4px;
    font-size: 1rem;
  }

  input:focus,
  select:focus,
  textarea:focus {
    outline: none;
    border-color: var(--color-primary);
  }

  .field-error {
    color: var(--color-danger);
    font-size: 0.875rem;
    margin-top: 0.25rem;
  }

  .actions {
    display: flex;
    gap: 1rem;
    margin-top: 2rem;
  }

  button {
    padding: 0.75rem 1.5rem;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
    transition: opacity 0.2s;
  }

  button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-primary {
    background: var(--color-primary);
    color: white;
  }

  .btn-secondary {
    background: var(--color-surface-elevated);
    color: var(--color-text);
    border: 1px solid var(--color-border);
  }

  .field--inline {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .toggle-label {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-weight: 500;
  }

  .toggle-label input[type="checkbox"] {
    width: auto;
  }

  .hint {
    font-size: 0.875rem;
    color: var(--color-text-secondary);
  }
</style>
