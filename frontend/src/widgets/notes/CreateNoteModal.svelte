<script lang="ts">
  import { createNote, type Note } from "$shared/api/notes";
  import Modal from "$components/atoms/Modal.svelte";
  import NoteForm from "$components/molecules/NoteForm.svelte";
  import type { ErrorResponse } from "$shared/types/errors";
  import { getMessage, mode } from "$shared/stores/lexicon-settings";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { CelestialBody, Theme } from "$entities";

  /* eslint-disable prefer-const -- Svelte 5 $bindable() requires let, not const, see: https://svelte.dev/docs/svelte/$bindable */
  let {
    open = $bindable(false),
    onSuccess,
    onClose,
    parentNote,
    defaultType,
  }: {
    open: boolean;
    onSuccess?: (note: Note) => void;
    onClose?: () => void;
    parentNote?: { id: string; title: string; type?: string };
    defaultType?: string;
  } = $props();

  let title = $state("");
  let content = $state("");
  let type = $state<string>(CelestialBody.PLANET.type);
  let loading = $state(false);
  let apiError = $state<ErrorResponse | null>(null);
  let currentMode = $state("standard");
  const theme = $derived(Theme.fromString(currentMode));
  const MAX_TITLE_LENGTH = 200;

  const initialType = $derived(
    parentNote
      ? (defaultType ?? CelestialBody.getChildSuggestion(parentNote.type))
      : CelestialBody.PLANET.type
  );

  $effect(() => {
    if (open) {
      type = initialType;
    }
  });

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);
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
  const modalTitle = $derived(tx("note.createTitle", "note.createTitleGalactic"));
  const titleLabel = $derived(tx("note.titleLabel", "note.titleLabelGalactic"));
  const typeLabel = $derived(tx("note.typeLabel", "note.typeLabelGalactic"));
  const contentLabel = $derived(tx("note.contentLabel", "note.contentLabelGalactic"));
  const cancelText = $derived(tx("note.cancel", "note.cancelGalactic"));
  const createText = $derived(tx("note.create", "note.createGalactic"));
  const titlePlaceholder = $derived(tx("note.titlePlaceholder", "note.titlePlaceholderGalactic"));
  const contentPlaceholder = $derived(
    tx("note.contentPlaceholder", "note.contentPlaceholderGalactic")
  );

  async function handleSubmit(_e?: SubmitEvent) {
    const trimmedTitle = title.trim();
    if (!trimmedTitle) {
      const msg = await getMessage("error", "validation", "title");
      apiError = { code: "VALIDATION_ERROR", message: msg };
      return;
    }

    if (trimmedTitle.length > MAX_TITLE_LENGTH) {
      const msg = t("note.titleTooLong").replace("{max}", MAX_TITLE_LENGTH.toString());
      apiError = { code: "VALIDATION_ERROR", message: msg };
      return;
    }

    loading = true;
    apiError = null;

    try {
      const note = await createNote({
        title: trimmedTitle,
        content: content.trim(),
        type,
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
    type = initialType;
    apiError = null;
    onClose?.();
  }
</script>

<Modal bind:open title={modalTitle} onClose={close}>
  {#if parentNote}
    <p class="parent-breadcrumb" data-testid="create-note-parent">
      {t("note.childOf", { title: parentNote.title })}
    </p>
  {/if}
  <NoteForm
    bind:title
    bind:content
    bind:type
    types={CelestialBody.UI_TYPES}
    {loading}
    error={apiError}
    onSubmit={handleSubmit}
    onCancel={close}
    onCloseError={() => (apiError = null)}
    {titleLabel}
    {typeLabel}
    {contentLabel}
    {titlePlaceholder}
    {contentPlaceholder}
    submitLabel={createText}
    cancelLabel={cancelText}
    titleTestId="create-note-title"
    contentTestId="create-note-content"
    submitTestId="create-note-submit"
    cancelTestId="create-note-cancel"
  />
</Modal>

<style>
  .parent-breadcrumb {
    margin: 0 0 12px;
    font-size: 13px;
    color: var(--carbon-text-muted, #8b8b9e);
    padding: 8px 12px;
    background: var(--carbon-graphene, #12121a);
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 8px;
    box-shadow: inset 0 0 12px rgba(139, 92, 246, 0.04);
  }
</style>
