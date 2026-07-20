<script lang="ts">
  import { searchNotes } from "$shared/api/notes";
  import { createLink, type CreateLinkData } from "$shared/api/links";
  import { LinkType, SearchQuery } from "$shared/lib/domain";
  import { addJitter } from "$shared/utils/jitter";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const {
    sourceNoteId,
    onSuccess,
    onCancel,
  }: {
    sourceNoteId: string;
    onSuccess?: (link: { id: string; target_note_id: string }) => void;
    onCancel?: () => void;
  } = $props();

  let searchQuery = $state("");
  let searchResults = $state<Array<{ id: string; title: string }>>([]);
  let isSearching = $state(false);
  let selectedTarget = $state<{ id: string; title: string } | null>(null);
  let linkType = $state(LinkType.REFERENCE.type);
  let isSubmitting = $state(false);
  let error = $state<string | null>(null);
  let showTypeDropdown = $state(false);
  let isFocused = $state(false); // eslint-disable-line @typescript-eslint/no-unused-vars -- Reserved for future focus tracking

  const linkTypes = LinkType.UI_TYPES;

  let debounceTimer: ReturnType<typeof setTimeout>;

  async function handleSearch() {
    const q = new SearchQuery(searchQuery);
    if (q.isEmpty()) {
      searchResults = [];
      return;
    }

    isSearching = true;
    error = null;

    try {
      const result = await searchNotes(q.value, 1, 10);
      // Исключаем текущую заметку из результатов
      searchResults = result.data.filter((note) => note.id !== sourceNoteId);
    } catch (err) {
      error =
        err instanceof Error ? err.message : t("linkCreator.searchFailed");
      searchResults = [];
    } finally {
      isSearching = false;
    }
  }

  function onSearchInput() {
    clearTimeout(debounceTimer);
    selectedTarget = null;
    debounceTimer = setTimeout(handleSearch, addJitter(300));
  }

  function selectTarget(note: { id: string; title: string }) {
    selectedTarget = note;
    searchQuery = note.title;
    searchResults = [];
  }

  async function handleSubmit() {
    if (!selectedTarget) {
      error = t("linkCreator.selectTargetError");
      return;
    }

    isSubmitting = true;
    error = null;

    try {
      const data: CreateLinkData = {
        source_note_id: sourceNoteId,
        target_note_id: selectedTarget.id,
        link_type: linkType,
        weight: 1.0,
      };

      const link = await createLink(data);
      onSuccess?.(link);
    } catch (err) {
      if (err instanceof Error && err.message.includes("409")) {
        error = t("linkCreator.linkExists");
      } else {
        error =
          err instanceof Error ? err.message : t("linkCreator.createFailed");
      }
    } finally {
      isSubmitting = false;
    }
  }

  function handleCancel() {
    onCancel?.();
  }

  function handleBlur() {
    isFocused = false;
  }
</script>

<div class="link-creator">
  <h3 class="title">{t("linkCreator.title")}</h3>

  <div class="search-section">
    <label for="target-search">{t("linkCreator.targetLabel")}</label>
    <div class="search-wrapper">
      <input
        id="target-search"
        type="text"
        bind:value={searchQuery}
        onfocus={() => (isFocused = true)}
        onblur={handleBlur}
        oninput={onSearchInput}
        placeholder={t("linkCreator.searchPlaceholder")}
        class="search-input"
        disabled={isSubmitting}
        aria-label={t("linkCreator.searchAria")}
      />
      {#if isSearching}
        <span class="spinner" aria-label={t("linkCreator.searching")}></span>
      {/if}
    </div>

    {#if searchResults.length > 0}
      <ul
        class="search-results"
        role="listbox"
        aria-label={t("linkCreator.searchResults")}
      >
        {#each searchResults as note}
          <li role="option" aria-selected="false">
            <button
              type="button"
              class="result-item"
              onclick={() => selectTarget(note)}
              aria-label={t("linkCreator.selectTarget", { title: note.title })}
            >
              {note.title}
            </button>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  <div class="type-section">
    <label for="link-type">{t("linkCreator.linkTypeLabel")}</label>
    <div class="type-selector">
      <button
        type="button"
        id="link-type"
        class="type-dropdown-btn"
        onclick={() => (showTypeDropdown = !showTypeDropdown)}
        aria-expanded={showTypeDropdown}
        aria-haspopup="listbox"
      >
        {linkTypes.find((t) => t.type === linkType)?.label ||
          t("linkCreator.selectType")}
        <span class="dropdown-arrow">▼</span>
      </button>

      {#if showTypeDropdown}
        <ul
          class="type-dropdown"
          role="listbox"
          aria-label={t("linkCreator.linkTypes")}
        >
          {#each linkTypes as type}
            <li role="option" aria-selected={type.type === linkType}>
              <button
                type="button"
                class="type-option"
                class:selected={type.type === linkType}
                onclick={() => {
                  linkType = type.type;
                  showTypeDropdown = false;
                }}
              >
                {type.label}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </div>

  {#if error}
    <div class="error" role="alert" aria-live="polite">
      {error}
    </div>
  {/if}

  <div class="actions">
    <button
      type="button"
      class="btn-primary"
      onclick={handleSubmit}
      disabled={isSubmitting || !selectedTarget}
      aria-busy={isSubmitting}
    >
      {#if isSubmitting}
        {t("linkCreator.creating")}
      {:else}
        {t("linkCreator.createLink")}
      {/if}
    </button>
    <button
      type="button"
      class="btn-secondary"
      onclick={handleCancel}
      disabled={isSubmitting}
    >
      {t("linkCreator.cancel")}
    </button>
  </div>
</div>

<style>
  .link-creator {
    background: white;
    border-radius: 12px;
    padding: 24px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
    min-width: 320px;
  }

  .title {
    margin: 0 0 20px 0;
    font-size: 18px;
    font-weight: 600;
    color: #1e293b;
  }

  .search-section,
  .type-section {
    margin-bottom: 16px;
  }

  label {
    display: block;
    margin-bottom: 6px;
    font-size: 13px;
    font-weight: 500;
    color: #64748b;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .search-wrapper {
    position: relative;
  }

  .search-input {
    width: 100%;
    padding: 10px 40px 10px 12px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 14px;
    background: #f8fafc;
    transition: all 0.2s;
  }

  .search-input:focus {
    outline: none;
    border-color: #3b82f6;
    background: white;
  }

  .spinner {
    position: absolute;
    right: 12px;
    top: 50%;
    transform: translateY(-50%);
    width: 16px;
    height: 16px;
    border: 2px solid #e2e8f0;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: translateY(-50%) rotate(360deg);
    }
  }

  .search-results {
    list-style: none;
    margin: 8px 0 0 0;
    padding: 0;
    max-height: 200px;
    overflow-y: auto;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    background: white;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .result-item {
    width: 100%;
    padding: 10px 12px;
    text-align: left;
    border: none;
    background: none;
    font-size: 14px;
    color: #334155;
    cursor: pointer;
    transition: background 0.15s;
  }

  .result-item:hover {
    background: #f1f5f9;
  }

  .type-selector {
    position: relative;
  }

  .type-dropdown-btn {
    width: 100%;
    padding: 10px 12px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    background: #f8fafc;
    font-size: 14px;
    color: #334155;
    cursor: pointer;
    display: flex;
    justify-content: space-between;
    align-items: center;
    transition: all 0.2s;
  }

  .type-dropdown-btn:hover {
    background: #f1f5f9;
  }

  .dropdown-arrow {
    font-size: 10px;
    color: #64748b;
  }

  .type-dropdown {
    list-style: none;
    margin: 4px 0 0 0;
    padding: 0;
    position: absolute;
    top: 100%;
    left: 0;
    right: 0;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    background: white;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    z-index: 10;
  }

  .type-option {
    width: 100%;
    padding: 10px 12px;
    text-align: left;
    border: none;
    background: none;
    font-size: 14px;
    color: #334155;
    cursor: pointer;
    transition: background 0.15s;
  }

  .type-option:hover,
  .type-option.selected {
    background: #f1f5f9;
  }

  .error {
    padding: 10px 12px;
    background: #fee2e2;
    color: #ef4444;
    border-radius: 8px;
    font-size: 13px;
    margin-bottom: 16px;
  }

  .actions {
    display: flex;
    gap: 12px;
    margin-top: 20px;
  }

  .btn-primary,
  .btn-secondary {
    flex: 1;
    padding: 12px 16px;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
    border: none;
  }

  .btn-primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .btn-primary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .btn-secondary {
    background: white;
    color: #64748b;
    border: 1px solid #e2e8f0;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #f8fafc;
  }
</style>
