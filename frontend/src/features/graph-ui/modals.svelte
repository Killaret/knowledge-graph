<script lang="ts">
  import type { NoteFormState } from "$features/graph-forms/note-form";
  import type { LinkFormState } from "$features/graph-forms/link-form";
  import NoteForm from "$components/molecules/NoteForm.svelte";
  import LinkTypeSelector from "$components/molecules/LinkTypeSelector.svelte";
  import Button from "$components/atoms/Button.svelte";
  import Bevel from "$components/atoms/Bevel.svelte";
  import { CelestialBody, LinkType } from "$entities";
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

  function handleLinkTypeSelect(type: string) {
    linkFormState.newLinkType = type;
    linkFormState.newLinkWeight = LinkType.fromString(type).defaultWeight;
  }
</script>

{#if activeForm === "note"}
  <Bevel
    variant="note"
    outer={10}
    inner={4}
    borderColor="rgba(139, 92, 246, 0.45)"
    shadeColor="rgba(139, 92, 246, 0.25)"
    data-testid="ghost-note-form"
    style="position: absolute; left: {noteFormState.noteFormPosition.x}px; top: {noteFormState
      .noteFormPosition.y}px; z-index: 100;"
  >
    <div class="ghost-note-form">
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
        types={CelestialBody.UI_TYPES}
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
  </Bevel>
{/if}

{#if activeForm === "link"}
  <Bevel
    variant="note"
    outer={10}
    inner={4}
    borderColor="rgba(245, 158, 11, 0.45)"
    shadeColor="rgba(245, 158, 11, 0.25)"
    data-testid="link-form"
    style="position: absolute; left: {linkFormState.linkFormPosition.x}px; top: {linkFormState
      .linkFormPosition.y}px; z-index: 100;"
  >
    <div class="link-form">
      <div class="link-form-header">
        <h3 class="link-form-title">{t("graphModals.createLinkTitle")}</h3>
        <button
          class="link-form-close"
          onclick={() => onCancel("link")}
          aria-label={t("close")}
          type="button"
        >
          ×
        </button>
      </div>
      <label for="link-type" class="link-form-label">{t("graphModals.linkTypeLabel")}</label>
      <LinkTypeSelector
        id="link-type"
        types={LinkType.CREATABLE_TYPES}
        selected={linkFormState.newLinkType}
        size="sm"
        showDescription={false}
        onSelect={handleLinkTypeSelect}
      />
      <label for="link-strength" class="link-form-label">
        {t("graphModals.linkStrength", {
          value: linkFormState.newLinkWeight.toFixed(1),
        })}
      </label>
      <input
        id="link-strength"
        type="range"
        min="0.1"
        max="1.0"
        step="0.1"
        bind:value={linkFormState.newLinkWeight}
        class="link-form-slider"
      />
      <div class="link-form-actions">
        <Button variant="ghost" onClick={() => onCancel("link")} data-testid="link-form-cancel">
          {t("cancel")}
        </Button>
        <Button variant="primary" onClick={() => onSave("link")} data-testid="link-form-create">
          {t("graphModals.createLink")}
        </Button>
      </div>
    </div>
  </Bevel>
{/if}

<style>
  .ghost-note-form {
    padding: 20px;
    min-width: 360px;
    max-width: min(420px, calc(100vw - 140px));
    z-index: 100;
    color: var(--carbon-text, #f0f0f5);
  }

  .ghost-note-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .ghost-note-title {
    margin: 0;
    color: var(--carbon-glow-cyan, #22d3ee);
    font-size: 16px;
    font-weight: 600;
  }

  .ghost-note-close {
    display: flex;
    align-items: center;
    justify-content: center;
    background: transparent;
    border: 1px solid var(--carbon-border, #2d2d3d);
    color: var(--carbon-text-muted, #8b8b9e);
    cursor: pointer;
    padding: 6px;
    border-radius: 8px;
    transition: all 0.2s ease;
  }

  .ghost-note-close:hover {
    background: rgba(139, 92, 246, 0.12);
    border-color: var(--carbon-border-active, #4b4b5e);
    color: var(--carbon-text, #f0f0f5);
  }

  .link-form {
    padding: 20px;
    min-width: 300px;
    max-width: min(380px, calc(100vw - 140px));
    z-index: 100;
    color: var(--carbon-text, #f0f0f5);
  }

  .link-form-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
  }

  .link-form-title {
    margin: 0;
    color: var(--carbon-glow-amber, #f59e0b);
    font-size: 16px;
    font-weight: 600;
  }

  .link-form-close {
    background: transparent;
    border: 1px solid var(--carbon-border, #2d2d3d);
    color: var(--carbon-text-muted, #8b8b9e);
    font-size: 20px;
    cursor: pointer;
    padding: 4px 10px;
    border-radius: 8px;
    transition: all 0.2s ease;
    line-height: 1;
  }

  .link-form-close:hover {
    background: rgba(245, 158, 11, 0.12);
    border-color: var(--carbon-border-active, #4b4b5e);
    color: var(--carbon-text, #f0f0f5);
  }

  .link-form-label {
    display: block;
    color: var(--carbon-text-muted, #8b8b9e);
    font-size: 13px;
    margin-bottom: 8px;
    font-weight: 500;
  }

  .link-form-slider {
    width: 100%;
    margin-bottom: 16px;
    accent-color: var(--carbon-glow-amber, #f59e0b);
  }

  .link-form-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }
</style>
