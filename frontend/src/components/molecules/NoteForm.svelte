<script lang="ts">
  import Button from "$components/atoms/Button.svelte";
  import TypeSelector from "$components/molecules/TypeSelector.svelte";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import type { ErrorResponse } from "$shared/types/errors";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  interface NoteTypeOption {
    type: string;
    emoji: string;
    label: string;
    color: string;
    toCSSColor(): string;
  }

  interface Props {
    title?: string;
    content?: string;
    type?: string;
    defaultType?: string;
    types: NoteTypeOption[];
    loading?: boolean;
    error?: ErrorResponse | null;
    titleLabel?: string;
    typeLabel?: string;
    contentLabel?: string;
    titlePlaceholder?: string;
    contentPlaceholder?: string;
    submitLabel?: string;
    cancelLabel?: string;
    ghost?: boolean;
    onSubmit?: (e?: SubmitEvent) => void;
    onCancel?: () => void;
    onCloseError?: () => void;
    titleTestId?: string;
    contentTestId?: string;
    submitTestId?: string;
    cancelTestId?: string;
  }

  /* eslint-disable prefer-const -- Svelte 5 $bindable() requires let, not const, see: https://svelte.dev/docs/svelte/$bindable */
  let {
    title = $bindable(""),
    content = $bindable(""),
    defaultType = "planet",
    type = $bindable(defaultType),
    types,
    loading = false,
    error = null,
    titleLabel = t("note.titleLabel"),
    typeLabel = t("note.typeLabel"),
    contentLabel = t("note.contentLabel"),
    titlePlaceholder = t("note.titlePlaceholder"),
    contentPlaceholder = t("note.contentPlaceholder"),
    submitLabel = t("note.create"),
    cancelLabel = t("note.cancel"),
    ghost = false,
    onSubmit,
    onCancel,
    onCloseError,
    titleTestId = "note-title-input",
    contentTestId = "note-content-input",
    submitTestId = "note-submit",
    cancelTestId = "note-cancel",
  }: Props = $props();

  const MAX_TITLE_LENGTH = 200;
  const titleLength = $derived(title.length);
  const isTitleWarning = $derived(titleLength > MAX_TITLE_LENGTH * 0.9);
  const isTitleError = $derived(titleLength >= MAX_TITLE_LENGTH);
</script>

<form
  onsubmit={(e) => {
    e.preventDefault();
    onSubmit?.(e);
  }}
  class="note-form"
  class:ghost
>
  <div class="form-group">
    <label for="note-title">{titleLabel}</label>
    <input
      id="note-title"
      name="title"
      type="text"
      bind:value={title}
      placeholder={titlePlaceholder}
      disabled={loading}
      data-testid={titleTestId}
      maxlength={MAX_TITLE_LENGTH}
      class:length-warning={isTitleWarning}
      class:length-error={isTitleError}
      required
    />
    <div
      class="length-indicator"
      class:length-warning={isTitleWarning}
      class:length-error={isTitleError}
    >
      {titleLength}/{MAX_TITLE_LENGTH}
    </div>
  </div>

  <div class="form-group">
    <label for="note-type">{typeLabel}</label>
    <TypeSelector id="note-type" bind:selected={type} {types} />
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
      data-testid={contentTestId}
    ></textarea>
  </div>

  <ApiErrorDisplay {error} onClose={onCloseError} />

  <div class="form-actions" class:ghost>
    <Button
      variant={ghost ? "ghost" : "secondary"}
      onClick={onCancel}
      disabled={loading}
      data-testid={cancelTestId}
    >
      {cancelLabel}
    </Button>
    <Button
      variant="primary"
      type="submit"
      disabled={loading || !title.trim()}
      data-testid={submitTestId}
    >
      {loading ? t("note.creating") : submitLabel}
    </Button>
  </div>
</form>

<style>
  .note-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  label {
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text, #374151);
  }

  input,
  textarea {
    width: 100%;
    padding: 10px 14px;
    border: 1px solid var(--color-border, #d1d5db);
    border-radius: 8px;
    background: var(--color-surface, white);
    color: var(--color-text, #1f2937);
    font-size: 14px;
    outline: none;
    transition:
      border-color 0.2s,
      box-shadow 0.2s;
    box-sizing: border-box;
  }

  input:focus,
  textarea:focus {
    border-color: var(--color-primary, #3b82f6);
    box-shadow: 0 0 0 3px var(--color-primary-light, rgba(59, 130, 246, 0.2));
  }

  input:disabled,
  textarea:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .note-form.ghost input,
  .note-form.ghost textarea {
    background: rgba(0, 0, 0, 0.4);
    border-color: rgba(255, 255, 255, 0.2);
    color: white;
  }

  .note-form.ghost input::placeholder,
  .note-form.ghost textarea::placeholder {
    color: rgba(255, 255, 255, 0.5);
  }

  .note-form.ghost input:focus,
  .note-form.ghost textarea:focus {
    border-color: rgba(139, 92, 246, 0.8);
    box-shadow: 0 0 0 3px rgba(139, 92, 246, 0.25);
  }

  textarea {
    resize: vertical;
    min-height: 100px;
  }

  .length-indicator {
    font-size: 12px;
    color: var(--color-text-tertiary, #9ca3af);
    text-align: right;
  }

  .length-warning {
    color: var(--color-danger, #ef4444);
  }

  input.length-warning,
  input:focus.length-warning {
    border-color: var(--color-warning, #f59e0b);
  }

  input.length-error,
  input:focus.length-error {
    border-color: var(--color-danger, #ef4444);
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    margin-top: 0.5rem;
  }
</style>
