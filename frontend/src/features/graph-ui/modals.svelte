<script lang="ts">
  import type { NoteFormState } from "$features/graph-forms/note-form";
  import type { LinkFormState } from "$features/graph-forms/link-form";
  import NoteForm from "$components/molecules/NoteForm.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const {
    activeForm,
    noteFormState = $bindable(),
    linkFormState = $bindable(),
    onSave,
    onCancel,
  }: {
    activeForm: "note" | "link" | null;
    noteFormState: NoteFormState;
    linkFormState: LinkFormState;
    onSave: (form: "note" | "link") => void;
    onCancel: (form: "note" | "link") => void;
  } = $props();
</script>

{#if activeForm === "note"}
  <div
    class="ghost-note-form"
    data-testid="ghost-note-form"
    style="position: absolute; left: {noteFormState.noteFormPosition.x}px; top: {noteFormState
      .noteFormPosition
      .y}px;"
  >
    <div class="ghost-note-header">
      <h3 class="ghost-note-title">{t("graphModals.createNoteTitle")}</h3>
      <button
        class="ghost-note-close"
        data-testid="ghost-note-close"
        onclick={() => onCancel("note")}
        aria-label={t("close")}
        type="button"
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
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      </button>
    </div>
    <NoteForm
      ghost
      bind:title={noteFormState.newNoteTitle}
      bind:content={noteFormState.newNoteContent}
      bind:type={noteFormState.newNoteType}
      onSubmit={() => onSave("note")}
      onCancel={() => onCancel("note")}
      titlePlaceholder={t("graphModals.noteTitlePlaceholder")}
      contentPlaceholder={t("graphModals.noteContentPlaceholder")}
      submitLabel={t("graphModals.create")}
      cancelLabel={t("cancel")}
      titleTestId="ghost-note-title"
      contentTestId="ghost-note-content"
      submitTestId="ghost-note-create"
      cancelTestId="ghost-note-cancel"
    />
  </div>
{/if}

<style>
  .ghost-note-form {
    background: rgba(10, 26, 58, 0.98);
    border: 1px solid rgba(138, 43, 226, 0.6);
    border-radius: 12px;
    padding: 20px;
    min-width: 360px;
    max-width: min(420px, calc(100vw - 140px));
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6);
    z-index: 100;
    backdrop-filter: blur(12px);
    color: white;
  }

  .ghost-note-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .ghost-note-title {
    margin: 0;
    color: #a78bfa;
    font-size: 16px;
    font-weight: 600;
  }

  .ghost-note-close {
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.6);
    cursor: pointer;
    padding: 6px;
    border-radius: 6px;
    transition: all 0.2s ease;
  }

  .ghost-note-close:hover {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.9);
  }
</style>

{#if activeForm === "link"}
  <div
    class="link-form"
    data-testid="link-form"
    style="position: absolute; left: {linkFormState.linkFormPosition.x}px; top: {linkFormState
      .linkFormPosition
      .y}px; background: rgba(10, 26, 58, 0.98); border: 1px solid rgba(255, 204, 0, 0.6); border-radius: 12px; padding: 20px; min-width: 300px; box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6); z-index: 100; backdrop-filter: blur(12px);"
  >
    <div
      style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;"
    >
      <h3 style="margin: 0; color: #fbbf24; font-size: 16px; font-weight: 600;">
        {t("graphModals.createLinkTitle")}
      </h3>
      <button
        onclick={() => onCancel("link")}
        style="background: none; border: none; color: rgba(255,255,255,0.6); font-size: 20px; cursor: pointer; padding: 4px 8px; border-radius: 4px; transition: all 0.2s;"
        aria-label={t("close")}
      >
        ×
      </button>
    </div>
    <label
      for="link-type"
      style="display: block; color: rgba(255,255,255,0.8); font-size: 13px; margin-bottom: 8px; font-weight: 500;"
      >{t("graphModals.linkTypeLabel")}</label
    >
    <select
      id="link-type"
      bind:value={linkFormState.newLinkType}
      style="width: 100%; padding: 12px; margin-bottom: 16px; border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; background: rgba(0,0,0,0.4); color: white; font-size: 14px; cursor: pointer; transition: border-color 0.2s;"
    >
      <option value="reference">📖 {t("linkType.reference")}</option>
      <option value="dependency">🔗 {t("linkType.dependency")}</option>
      <option value="related">🔀 {t("linkType.related")}</option>
      <option value="custom">✨ {t("linkType.custom")}</option>
    </select>
    <label
      for="link-strength"
      style="display: block; color: rgba(255,255,255,0.8); font-size: 13px; margin-bottom: 8px; font-weight: 500;"
      >{t("graphModals.linkStrength", {
        value: linkFormState.newLinkWeight.toFixed(1),
      })}</label
    >
    <input
      id="link-strength"
      type="range"
      min="0.1"
      max="1.0"
      step="0.1"
      bind:value={linkFormState.newLinkWeight}
      style="width: 100%; margin-bottom: 16px; accent-color: #fbbf24;"
    />
    <div style="display: flex; gap: 12px; justify-content: flex-end;">
      <button
        data-testid="link-form-cancel"
        onclick={() => onCancel("link")}
        style="padding: 10px 20px; border: 1px solid rgba(255,255,255,0.3); border-radius: 8px; background: transparent; color: white; cursor: pointer; font-size: 14px; transition: all 0.2s;"
      >
        {t("cancel")}
      </button>
      <button
        data-testid="link-form-create"
        onclick={() => onSave("link")}
        style="padding: 10px 20px; border: none; border-radius: 8px; background: linear-gradient(135deg, #fbbf24, #f59e0b); color: #000; cursor: pointer; font-size: 14px; font-weight: 500; transition: all 0.2s; box-shadow: 0 4px 12px rgba(251, 191, 36, 0.3);"
      >
        {t("graphModals.createLink")}
      </button>
    </div>
  </div>
{/if}
