<script lang="ts">
  import { goto } from "$app/navigation";
  import { LinkType, GraphMode } from "$entities";
  import { graphStore } from "$shared/stores/graph.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import GraphStats from "$features/graph-ui/GraphStats.svelte";
  import LangSwitcher from "$components/atoms/LangSwitcher.svelte";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  interface Filter {
    id: string;
    label: string;
    emoji: string;
    description?: string;
    example?: string;
  }

  interface Props {
    isAuthenticated: boolean;
    currentView: "graph" | "list" | "3d";
    layoutProvider?: "d3" | "graph-service";
    searchQuery?: string;
    selectedType?: string;
    typeFilters?: Filter[];
    typeCounts?: Record<string, number>;
    nodeCount?: number;
    linkCount?: number;
    onSearch?: (query: string) => void;
    onToggleView?: (view: "graph" | "list" | "3d") => void;
    onToggleLayoutProvider?: (provider: "d3" | "graph-service") => void;
    onFilter?: (type: string) => void;
    onNoteCreate?: () => void;
    onSignIn?: () => void;
    onRegister?: () => void;
    canvasController?: {
      focusMode: boolean;
      fogEnabled: boolean;
      resetView: () => void;
      openSearch: () => void;
      toggleFocus: () => void;
      toggleFog: () => void;
    };
    variant?: "docked" | "floating";
  }

  const {
    isAuthenticated,
    currentView,
    layoutProvider = "d3",
    searchQuery = "",
    selectedType = "all",
    typeFilters = [],
    typeCounts = {},
    nodeCount,
    linkCount,
    onSearch = () => {},
    onToggleView = () => {},
    onToggleLayoutProvider,
    onFilter = () => {},
    onNoteCreate,
    onSignIn,
    onRegister,
    canvasController,
    variant = "docked",
  }: Props = $props();

  let typeDropdownOpen = $state(false);
  let linkDropdownOpen = $state(false);

  const viewOptions: { id: "graph" | "list" | "3d"; label: string; icon: string }[] = [
    { id: "graph", label: t("controls.view2D"), icon: "◯" },
    { id: "3d", label: t("controls.view3D"), icon: "△" },
    { id: "list", label: t("controls.viewList"), icon: "☰" },
  ];

  const layoutOptions: { id: "d3" | "graph-service"; label: string; title: string }[] = [
    { id: "d3", label: t("controls.layoutD3"), title: t("controls.layoutD3Title") },
    {
      id: "graph-service",
      label: t("controls.layoutGraphService"),
      title: t("controls.layoutGraphServiceTitle"),
    },
  ];

  const linkTypes = $derived(LinkType.ALL_TYPES);
  const hiddenLinkSet = $derived(new Set(graphStore.hiddenLinkTypes));
  const areAllLinkTypesVisible = $derived(graphStore.hiddenLinkTypes.length === 0);
  const areAllLinkTypesHidden = $derived(graphStore.hiddenLinkTypes.length === linkTypes.length);
  const graphMode = $derived(
    canvasController ? GraphMode.fromFocus(canvasController.focusMode) : GraphMode.normal()
  );

  function handleSearchInput(event: Event) {
    const value = (event.target as HTMLInputElement).value;
    onSearch(value);
  }

  function handleMinWeightInput(event: Event) {
    const value = Number((event.target as HTMLInputElement).value);
    graphStore.minLinkWeight = value;
  }

  function toggleLinkType(type: string) {
    graphStore.toggleLinkType(type);
  }

  function showAllLinkTypes() {
    graphStore.hiddenLinkTypes = [];
  }

  function hideAllLinkTypes() {
    graphStore.hiddenLinkTypes = linkTypes.map((lt) => lt.type);
  }

  function selectedTypeLabel(): string {
    const filter = typeFilters.find((f) => f.id === selectedType);
    return filter ? `${filter.emoji} ${filter.label}` : t("filter.all");
  }

  function toggleTypeDropdown() {
    const wasOpen = typeDropdownOpen;
    closeDropdowns();
    typeDropdownOpen = !wasOpen;
  }

  function toggleLinkDropdown() {
    const wasOpen = linkDropdownOpen;
    closeDropdowns();
    linkDropdownOpen = !wasOpen;
  }

  function handleTypeSelect(type: string) {
    typeDropdownOpen = false;
    onFilter(type);
  }

  function handleLayoutToggle(provider: "d3" | "graph-service") {
    onToggleLayoutProvider?.(provider);
  }

  function closeDropdowns() {
    typeDropdownOpen = false;
    linkDropdownOpen = false;
  }
</script>

<div class="graph-top-bar graph-top-bar--{variant}" data-testid="graph-top-bar">
  <div class="left-cluster">
    {#if isAuthenticated}
      <a
        href="/"
        class="logo"
        onclick={(e: MouseEvent) => {
          e.preventDefault();
          goto("/");
        }}
      >
        <span class="logo-icon">🌌</span>
        <span class="logo-text">KG</span>
      </a>
    {/if}

    <GraphStats {nodeCount} {linkCount} />

    <div class="view-toggle" role="group" aria-label={t("cockpit.left.view")}>
      {#each viewOptions as option}
        <button
          type="button"
          class="top-bar-btn top-bar-btn--segment"
          class:active={currentView === option.id}
          aria-pressed={currentView === option.id}
          onclick={() => onToggleView(option.id)}
          data-testid="view-toggle-{option.id}"
          title={option.label}
        >
          <span>{option.icon}</span>
          <span>{option.label}</span>
        </button>
      {/each}
    </div>

    {#if currentView === "3d"}
      <div class="layout-toggle" role="group" aria-label={t("controls.layoutProviderTitle")}>
        {#each layoutOptions as option}
          <button
            type="button"
            class="top-bar-btn top-bar-btn--segment"
            class:active={layoutProvider === option.id}
            aria-pressed={layoutProvider === option.id}
            onclick={() => handleLayoutToggle(option.id)}
            data-testid="layout-provider-{option.id}"
            title={option.title}
          >
            {option.label}
          </button>
        {/each}
      </div>
    {/if}

    <div class="search-box">
      <input
        type="text"
        class="top-bar-input"
        value={searchQuery}
        oninput={handleSearchInput}
        placeholder={t("search.placeholder")}
        aria-label={t("search.inputAriaLabel")}
        data-testid="top-bar-search-input"
      />
    </div>

    <div class="dropdown">
      <button
        type="button"
        class="top-bar-btn"
        onclick={toggleTypeDropdown}
        aria-haspopup="listbox"
        aria-expanded={typeDropdownOpen}
        data-testid="type-dropdown-toggle"
      >
        {selectedTypeLabel()}
        <span class="chevron">▼</span>
      </button>
      {#if typeDropdownOpen}
        <div class="dropdown-panel">
          {#each typeFilters as filter}
            <button
              type="button"
              class="dropdown-item"
              class:active={selectedType === filter.id}
              onclick={() => handleTypeSelect(filter.id)}
              data-testid="filter-chip-{filter.id}"
            >
              <span class="dropdown-item-icon">{filter.emoji}</span>
              <span class="dropdown-item-label">{filter.label}</span>
              {#if typeCounts[filter.id] !== undefined}
                <span class="dropdown-count">
                  {typeCounts[filter.id]}
                </span>
              {/if}
            </button>
          {/each}
        </div>
      {/if}
    </div>

    <div class="dropdown">
      <button
        type="button"
        class="top-bar-btn"
        onclick={toggleLinkDropdown}
        aria-haspopup="true"
        aria-expanded={linkDropdownOpen}
        data-testid="link-dropdown-toggle"
      >
        {t("linkLegend.title")}
        <span class="chevron">▼</span>
      </button>
      {#if linkDropdownOpen}
        <div class="dropdown-panel dropdown-panel--right">
          <div class="dropdown-actions">
            <button
              type="button"
              class="dropdown-action"
              disabled={areAllLinkTypesVisible}
              onclick={showAllLinkTypes}
              data-testid="link-types-show-all"
            >
              {t("linkLegend.showAll")}
            </button>
            <button
              type="button"
              class="dropdown-action"
              disabled={areAllLinkTypesHidden}
              onclick={hideAllLinkTypes}
              data-testid="link-types-hide-all"
            >
              {t("linkLegend.hideAll")}
            </button>
          </div>

          <div class="dropdown-section">
            <label for="top-bar-min-weight" class="dropdown-label">
              {t("linkLegend.minWeight", { weight: graphStore.minLinkWeight.toFixed(1) })}
            </label>
            <input
              id="top-bar-min-weight"
              type="range"
              min="0"
              max="1"
              step="0.1"
              value={graphStore.minLinkWeight}
              oninput={handleMinWeightInput}
              class="top-bar-range"
              data-testid="top-bar-min-weight"
            />
          </div>

          <div class="dropdown-list">
            {#each linkTypes as lt}
              <button
                type="button"
                class="dropdown-item"
                class:active={!hiddenLinkSet.has(lt.type)}
                onclick={() => toggleLinkType(lt.type)}
                style="--type-color: {lt.color}"
                data-testid="link-type-chip-{lt.type}"
              >
                <span class="dropdown-line" style="background: {lt.color};"></span>
                <span class="dropdown-item-icon">{lt.icon}</span>
                <span class="dropdown-item-label">{lt.label}</span>
              </button>
            {/each}
          </div>
        </div>
      {/if}
    </div>

    {#if canvasController}
      <div class="canvas-controls" role="group" aria-label={t("cockpit.left.graphControls")}>
        <button
          type="button"
          class="top-bar-btn top-bar-btn--icon"
          onclick={canvasController.resetView}
          title={t("graphControls.resetView")}
          data-testid="top-bar-reset"
        >
          🔄
        </button>
        <button
          type="button"
          class="top-bar-btn top-bar-btn--icon"
          onclick={canvasController.openSearch}
          title={t("graphControls.search")}
          data-testid="top-bar-open-search"
        >
          🎯
        </button>
        <button
          type="button"
          class="top-bar-btn top-bar-btn--icon"
          class:active={graphMode.isFocus}
          style="border-color: {graphMode.borderColor}; color: {graphMode.textColor};"
          onclick={canvasController.toggleFocus}
          title={graphMode.label}
          data-testid="top-bar-focus"
          aria-pressed={graphMode.isFocus}
        >
          {graphMode.icon}
        </button>
        <button
          type="button"
          class="top-bar-btn top-bar-btn--icon"
          class:active={canvasController.fogEnabled}
          onclick={canvasController.toggleFog}
          title={t("graphOverlay.fogToggle")}
          data-testid="top-bar-fog"
          aria-pressed={canvasController.fogEnabled}
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <path d="M3 15h18M3 10h18M4 5h16M5 20h14" />
          </svg>
        </button>
      </div>
    {/if}
  </div>

  <div class="right-cluster">
    <LangSwitcher />

    {#if isAuthenticated && onNoteCreate}
      <button
        type="button"
        class="create-btn"
        onclick={() => onNoteCreate()}
        title={t("controls.createTitle")}
        data-testid="create-note-button"
        aria-label={t("controls.createAria")}
      >
        <svg
          width="20"
          height="20"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
      </button>
    {:else if !isAuthenticated}
      {#if onSignIn}
        <button
          type="button"
          class="top-bar-btn top-bar-btn--ghost"
          onclick={onSignIn}
          data-testid="top-bar-sign-in"
        >
          {t("auth.signInButton")}
        </button>
      {/if}
      {#if onRegister}
        <button
          type="button"
          class="top-bar-btn top-bar-btn--primary"
          onclick={onRegister}
          data-testid="top-bar-register"
        >
          {t("auth.registerButton")}
        </button>
      {/if}
    {/if}
  </div>
</div>

<style>
  .graph-top-bar {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    padding: 8px 12px;
    background:
      linear-gradient(135deg, var(--carbon-graphene) 0%, rgba(18, 18, 26, 0.92) 100%),
      radial-gradient(circle at 80% 20%, var(--carbon-hex-fill) 0%, transparent 60%);
    border: 1px solid var(--carbon-border);
    border-radius: 14px;
    color: var(--carbon-text);
    font-size: 13px;
    z-index: 60;
    width: fit-content;
    max-width: calc(100% - 32px);
    box-shadow:
      var(--carbon-shadow),
      0 0 24px rgba(139, 92, 246, 0.08);
    backdrop-filter: blur(10px);
  }

  .graph-top-bar--floating {
    position: relative;
    background:
      linear-gradient(135deg, var(--carbon-graphene) 0%, rgba(18, 18, 26, 0.92) 100%),
      radial-gradient(circle at 80% 20%, var(--carbon-hex-fill) 0%, transparent 60%);
    border: 1px solid var(--carbon-border);
    border-radius: 14px;
    box-shadow:
      var(--carbon-shadow),
      0 0 24px rgba(139, 92, 246, 0.08);
    backdrop-filter: blur(10px);
  }

  .graph-top-bar--docked {
    background: transparent;
    border: none;
    border-radius: 0;
    box-shadow: none;
    backdrop-filter: none;
    padding: 8px 16px;
    width: 100%;
    max-width: none;
    z-index: auto;
  }

  .left-cluster,
  .right-cluster {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
  }

  .right-cluster {
    margin-left: auto;
    flex-shrink: 0;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 6px;
    text-decoration: none;
    color: #2dd4bf;
    font-weight: 700;
    letter-spacing: 0.05em;
    font-size: 15px;
    flex-shrink: 0;
  }

  .logo-icon {
    font-size: 20px;
  }

  .view-toggle,
  .layout-toggle,
  .canvas-controls {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 4px;
  }

  .top-bar-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 7px 12px;
    background: var(--carbon-hex-fill);
    border: 1px solid var(--carbon-border);
    border-radius: 10px;
    color: var(--carbon-text);
    font-size: 13px;
    cursor: pointer;
    white-space: nowrap;
    transition: all 0.2s ease;
  }

  .top-bar-btn:hover {
    background: rgba(139, 92, 246, 0.08);
    border-color: var(--carbon-border-active);
  }

  .top-bar-btn.active,
  .top-bar-btn[aria-pressed="true"] {
    background: rgba(245, 158, 11, 0.12);
    border-color: var(--carbon-glow-amber);
    color: var(--carbon-glow-amber);
    box-shadow: var(--carbon-glow-amber);
  }

  .top-bar-btn--icon {
    padding: 7px 8px;
    min-width: 34px;
  }

  .top-bar-btn--primary {
    background: linear-gradient(135deg, rgba(34, 211, 238, 0.18) 0%, rgba(139, 92, 246, 0.18) 100%);
    border-color: rgba(34, 211, 238, 0.4);
    color: var(--carbon-glow-cyan);
  }

  .top-bar-btn--primary:hover {
    background: linear-gradient(135deg, rgba(34, 211, 238, 0.28) 0%, rgba(139, 92, 246, 0.28) 100%);
    border-color: rgba(34, 211, 238, 0.6);
  }

  .top-bar-btn--ghost {
    background: transparent;
    border-color: transparent;
    color: var(--carbon-text-muted);
  }

  .top-bar-btn--ghost:hover {
    background: var(--carbon-hex-fill);
    border-color: var(--carbon-border);
    color: var(--carbon-text);
  }

  .create-btn {
    width: 34px;
    height: 34px;
    border: none;
    border-radius: 50%;
    background: rgba(45, 212, 191, 0.15);
    color: #2dd4bf;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: background 0.2s ease;
  }

  .create-btn:hover {
    background: rgba(45, 212, 191, 0.25);
  }

  .top-bar-input {
    width: 160px;
    padding: 7px 10px;
    background: var(--carbon-black);
    border: 1px solid var(--carbon-border);
    border-radius: 10px;
    color: var(--carbon-text);
    font-size: 13px;
    outline: none;
    transition: all 0.2s ease;
  }

  .top-bar-input::placeholder {
    color: var(--carbon-text-dim);
  }

  .top-bar-input:focus {
    border-color: var(--carbon-glow-cyan);
    box-shadow: 0 0 0 3px rgba(34, 211, 238, 0.1);
    background: var(--carbon-graphite);
  }

  .top-bar-range {
    width: 100%;
    accent-color: var(--carbon-glow-amber);
  }

  .search-box {
    position: relative;
  }

  .dropdown {
    position: relative;
  }

  .chevron {
    margin-left: 4px;
    opacity: 0.7;
    font-size: 10px;
  }

  .dropdown-panel {
    position: absolute;
    top: calc(100% + 6px);
    left: 0;
    min-width: 200px;
    max-height: 320px;
    overflow-y: auto;
    background: var(--carbon-c70);
    border: 1px solid var(--carbon-border);
    border-radius: 12px;
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    box-shadow:
      var(--carbon-shadow),
      0 0 30px rgba(139, 92, 246, 0.1);
    z-index: 70;
  }

  .dropdown-panel--right {
    left: auto;
    right: 0;
    min-width: 240px;
  }

  .dropdown-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 8px 10px;
    background: transparent;
    border: none;
    border-radius: 8px;
    color: var(--carbon-text);
    font-size: 13px;
    text-align: left;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .dropdown-item:hover {
    background: rgba(139, 92, 246, 0.08);
  }

  .dropdown-item.active {
    background: rgba(245, 158, 11, 0.12);
    border-color: var(--carbon-glow-amber);
    color: var(--carbon-glow-amber);
  }

  .dropdown-item-icon {
    width: 18px;
    text-align: center;
  }

  .dropdown-item-label {
    flex: 1;
    text-align: left;
  }

  .dropdown-count {
    padding: 2px 8px;
    background: var(--carbon-hex-fill);
    border: 1px solid var(--carbon-border);
    border-radius: 10px;
    font-size: 11px;
    color: var(--carbon-text-muted);
  }

  .dropdown-line {
    width: 20px;
    height: 3px;
    border-radius: 2px;
    flex-shrink: 0;
  }

  .dropdown-actions {
    display: flex;
    gap: 8px;
  }

  .dropdown-action {
    flex: 1;
    padding: 6px 8px;
    background: var(--carbon-hex-fill);
    border: 1px solid var(--carbon-border);
    border-radius: 8px;
    color: var(--carbon-text-muted);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .dropdown-action:hover:not(:disabled) {
    background: rgba(139, 92, 246, 0.1);
    border-color: var(--carbon-border-active);
    color: var(--carbon-text);
  }

  .dropdown-action:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .dropdown-section {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 8px 4px;
    border-bottom: 1px solid var(--carbon-border);
    border-top: 1px solid var(--carbon-border);
  }

  .dropdown-label {
    font-size: 11px;
    color: var(--carbon-text-muted);
  }

  .dropdown-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  /* Mobile */
  @media (max-width: 640px) {
    .graph-top-bar {
      padding: 6px 8px;
      gap: 6px;
      max-width: calc(100% - 16px);
      border-radius: 12px;
    }

    .top-bar-input {
      width: 120px;
    }

    .top-bar-btn {
      padding: 6px 10px;
      font-size: 12px;
    }

    .top-bar-btn--icon {
      padding: 6px 7px;
      min-width: 30px;
    }

    .dropdown-panel {
      left: 0;
      right: auto;
      min-width: 180px;
    }

    .dropdown-panel--right {
      left: auto;
      right: 0;
    }
  }
</style>
