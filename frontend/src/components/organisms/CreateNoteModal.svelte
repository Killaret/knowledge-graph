<script lang="ts">
  import { createNote, type Note } from "$shared/api/notes";
  import Button from "$components/atoms/Button.svelte";
  import Modal from "$components/atoms/Modal.svelte";
  import TypeSelector from "$components/molecules/TypeSelector.svelte";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import type { ErrorResponse } from "$shared/types/errors";
  import { getMessage, mode } from "$shared/stores/lexicon-settings";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { CelestialBody, Theme } from "$shared/lib/domain";

  /* eslint-disable prefer-const -- Svelte 5 $bindable() requires let, not const, see: https://svelte.dev/docs/svelte/$bindable */
  let {
    open = $bindable(false),
    onSuccess,
  }: {
    open: boolean;
    onSuccess?: (note: Note) => void;
  } = $props();

  let title = $state("");
  let content = $state("");
  let type = $state<string>(CelestialBody.STAR.type);
  let loading = $state(false);
  let apiError = $state<ErrorResponse | null>(null);
  let currentMode = $state("standard");
  const theme = $derived(Theme.fromString(currentMode));

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);
  const tx = (standardKey: string, galacticKey: string) =>
    theme.isGalactic ? t(galacticKey) : t(standardKey);
  const toErrorResponse = (e: unknown): ErrorResponse => {
    if (e && typeof e === "object") {
      const err = e as { response?: { data?: ErrorResponse } };
      if (err.response?.data) {
        return err.response.data;
      }
    }
    return { code: "API_ERROR", message: t("note.createError") };
  };

  // Subscribe to mode changes
  $effect(() => {
    const unsubscribe = mode.subscribe((m) => (currentMode = m));
    return unsubscribe;
  });

  // Computed labels based on theme
  const modalTitle = $derived(
    tx("note.createTitle", "note.createTitleGalactic"),
  );
  const titleLabel = $derived(tx("note.titleLabel", "note.titleLabelGalactic"));
  const typeLabel = $derived(tx("note.typeLabel", "note.typeLabelGalactic"));
  const contentLabel = $derived(
    tx("note.contentLabel", "note.contentLabelGalactic"),
  );
  const cancelText = $derived(tx("note.cancel", "note.cancelGalactic"));
  const createText = $derived(tx("note.create", "note.createGalactic"));
  const creatingText = $derived(tx("note.creating", "note.creatingGalactic"));
  const titlePlaceholder = $derived(
    tx("note.titlePlaceholder", "note.titlePlaceholderGalactic"),
  );
  const contentPlaceholder = $derived(
    tx("note.contentPlaceholder", "note.contentPlaceholderGalactic"),
  );

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!title.trim()) {
      const msg = await getMessage("error", "validation", "title");
      apiError = { code: "VALIDATION_ERROR", message: msg };
      return;
    }

    loading = true;
    apiError = null;

    try {
      const note = await createNote({
        title: title.trim(),
        content: content.trim(),
        type: type,
        metadata: {},
      });

      onSuccess?.(note);
      close();
    } catch (err: unknown) {
      apiError = toErrorResponse(err);
    } finally {
      loading = false;
    }
  }

  function close() {
    open = false;
    title = "";
    content = "";
    type = CelestialBody.STAR.type; // reset to default
    apiError = null;
  }
</script>

<Modal bind:open title={modalTitle} onClose={close}>
  <form onsubmit={handleSubmit}>
    <div class="form-group">
      <label for="note-title">{titleLabel}</label>
      <input
        id="note-title"
        name="title"
        type="text"
        bind:value={title}
        placeholder={titlePlaceholder}
        disabled={loading}
        data-testid="create-note-title"
      />
    </div>

    <div class="form-group">
      <label for="note-type">{typeLabel}</label>
      <TypeSelector id="note-type" bind:selected={type} />
    </div>

    <div class="form-group">
      <label for="note-content">{contentLabel}</label>
      <textarea
        id="note-content"
        name="content"
        bind:value={content}
        placeholder={contentPlaceholder}
        rows={6}
        disabled={loading}
        data-testid="create-note-content"
      ></textarea>
    </div>

    <ApiErrorDisplay error={apiError} onClose={() => (apiError = null)} />

    <div class="form-actions">
      <Button variant="secondary" onClick={close} disabled={loading}>
        {cancelText}
      </Button>
      <Button
        variant="primary"
        type="submit"
        disabled={loading}
        data-testid="create-note-submit"
      >
        {loading ? creatingText : createText}
      </Button>
    </div>
  </form>
</Modal>

<style>
  .form-group {
    margin-bottom: 20px;
  }

  label {
    display: block;
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text, #374151);
    margin-bottom: 6px;
  }

  input,
  textarea {
    width: 100%;
    padding: 10px 14px;
    border: 1px solid var(--color-border, #d1d5db);
    border-radius: 8px;
    font-size: 15px;
    color: var(--color-text, #1f2937);
    background: var(--color-surface, white);
    transition: border-color 0.2s;
  }

  input:focus,
  textarea:focus {
    outline: none;
    border-color: var(--color-primary, #3b82f6);
    box-shadow: 0 0 0 3px var(--color-primary-light, rgba(59, 130, 246, 0.1));
  }

  textarea {
    resize: vertical;
    font-family: inherit;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
</style>
