<script lang="ts">
  import { goto } from "$app/navigation";
  import { isAuthenticated } from "$shared/stores/auth.svelte";
  import { SearchQuery } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import LangSwitcher from "$components/atoms/LangSwitcher.svelte";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string>) => formatMessage(key, locale, params);

  const {
    onCreate,
    onSearch,
    onToggleView,
    onFilter,
    onImport,
    onExport,
    typeFilters = [],
    selectedType = "all",
    typeCounts = {},
    currentView = "graph",
  }: {
    onCreate?: () => void;
    onSearch?: (query: string) => void;
    onToggleView?: (view: "graph" | "list") => void;
    onFilter?: (type: string) => void;
    onImport?: () => void;
    onExport?: () => void;
    typeFilters?: Array<{ id: string; label: string; emoji: string }>;
    selectedType?: string;
    typeCounts?: Record<string, number>;
    currentView?: "graph" | "list";
  } = $props();

  let searchQuery = $state("");
  let showMenu = $state(false);
  let filtersContainer: HTMLDivElement | null = $state(null);

  function scrollFilters(dir: "left" | "right") {
    if (filtersContainer) {
      filtersContainer.scrollBy({
        left: dir === "right" ? 120 : -120,
        behavior: "smooth",
      });
    }
  }

  function handleSearch() {
    const q = new SearchQuery(searchQuery);
    onSearch?.(q.value);
  }

  function toggleView(targetView: "graph" | "list") {
    onToggleView?.(targetView);
  }

  function handleFilter(typeId: string) {
    onFilter?.(typeId);
  }

  function handleLogin() {
    goto("/auth/login");
    showMenu = false;
  }
</script>

<div class="floating-controls">
  <!-- View Toggle: 2D / 3D / List -->
  <div class="view-toggle">
    <button
      type="button"
      class="toggle-btn {currentView === 'graph' ? 'active' : ''}"
      onclick={() => toggleView("graph")}
      title={t("controls.graph2DTitle")}
      data-testid="view-toggle-graph"
      aria-pressed={currentView === "graph"}
      aria-label={t("controls.graph2DAria")}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <circle cx="12" cy="12" r="3" />
        <circle cx="5" cy="5" r="2" />
        <circle cx="19" cy="5" r="2" />
        <circle cx="5" cy="19" r="2" />
        <circle cx="19" cy="19" r="2" />
        <line x1="7" y1="7" x2="10" y2="10" />
        <line x1="14" y1="10" x2="17" y2="7" />
        <line x1="7" y1="17" x2="10" y2="14" />
        <line x1="14" y1="14" x2="17" y2="17" />
      </svg>
      <span class="btn-label">{t("controls.view2D")}</span>
    </button>
    <!-- 3D functionality frozen for v1 - see CHANGELOG.md -->
    <!--
    <button
      type="button"
      class="toggle-btn"
      onclick={handleToggle3D}
      title={noteId ? "3D Graph for selected note" : "Full 3D Graph"}
      data-testid="view-toggle-3d"
      aria-label={noteId ? "Open 3D graph for selected note" : "Open full 3D graph"}
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M12 2L2 7l10 5 10-5-10-5z"/>
        <path d="M2 17l10 5 10-5"/>
        <path d="M2 12l10 5 10-5"/>
      </svg>
      <span class="btn-label">3D</span>
    </button>
    -->
    <button
      type="button"
      class="toggle-btn {currentView === 'list' ? 'active' : ''}"
      onclick={() => toggleView("list")}
      title={t("controls.listViewTitle")}
      data-testid="view-toggle-list"
      aria-pressed={currentView === "list"}
      aria-label={t("controls.listViewAria")}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <line x1="3" y1="6" x2="21" y2="6" />
        <line x1="3" y1="12" x2="21" y2="12" />
        <line x1="3" y1="18" x2="21" y2="18" />
      </svg>
      <span class="btn-label">{t("controls.viewList")}</span>
    </button>
  </div>

  <!-- Type Filters -->
  {#if typeFilters.length > 0}
    <div class="type-filters-wrapper">
      <button
        type="button"
        class="scroll-arrow scroll-left"
        onclick={() => scrollFilters("left")}
        aria-label={t("controls.scrollLeft")}>‹</button
      >
      <div class="type-filters" bind:this={filtersContainer}>
        {#each typeFilters as filter}
          <button
            type="button"
            class="filter-chip {selectedType === filter.id ? 'active' : ''}"
            onclick={() => handleFilter(filter.id)}
            title={filter.label}
            data-testid="filter-chip-{filter.id}"
            aria-pressed={selectedType === filter.id}
            aria-label={t("filter.filterBy", { type: filter.label })}
          >
            <span class="filter-emoji">{filter.emoji}</span>
            <span class="filter-label">{filter.label}</span>
            {#if typeCounts[filter.id] !== undefined}
              <span class="filter-count" data-testid="filter-count-{filter.id}"
                >{typeCounts[filter.id]}</span
              >
            {/if}
          </button>
        {/each}
      </div>
      <button
        type="button"
        class="scroll-arrow scroll-right"
        onclick={() => scrollFilters("right")}
        aria-label={t("controls.scrollRight")}>›</button
      >
    </div>
  {/if}

  <!-- Search -->
  <div class="search-container">
    <input
      type="text"
      bind:value={searchQuery}
      placeholder={t("search.placeholder")}
      onkeyup={(e) => e.key === "Enter" && handleSearch()}
      class="search-input"
      data-testid="search-input"
      aria-label={t("search.inputAriaLabel")}
    />
    <button type="button" class="search-btn" onclick={handleSearch} aria-label={t("search.label")}>
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <circle cx="11" cy="11" r="8" />
        <line x1="21" y1="21" x2="16.65" y2="16.65" />
      </svg>
    </button>
  </div>

  <!-- Login Button (visible when not authenticated) -->
  {#if !isAuthenticated()}
    <button
      type="button"
      class="login-btn"
      onclick={handleLogin}
      title={t("auth.signInButton")}
      data-testid="floating-login-button"
      aria-label={t("auth.loginAriaLabel")}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" />
        <polyline points="10 17 15 12 10 7" />
        <line x1="15" y1="12" x2="3" y2="12" />
      </svg>
    </button>
  {/if}

  <!-- Menu -->
  <div class="menu-container">
    <button
      type="button"
      class="menu-btn"
      data-testid="menu-button"
      onclick={() => (showMenu = !showMenu)}
      title={t("controls.menuTitle")}
      aria-expanded={showMenu}
      aria-haspopup="true"
      aria-label={t("controls.menuAria")}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <line x1="3" y1="12" x2="21" y2="12" />
        <line x1="3" y1="6" x2="21" y2="6" />
        <line x1="3" y1="18" x2="21" y2="18" />
      </svg>
    </button>

    {#if showMenu}
      <div class="dropdown-menu" role="menu">
        {#if !isAuthenticated()}
          <button
            type="button"
            class="menu-item"
            role="menuitem"
            onclick={handleLogin}
            data-testid="menu-login"
          >
            🔑 {t("auth.loginMenuItem")}
          </button>
        {/if}
        <button
          type="button"
          class="menu-item"
          role="menuitem"
          onclick={() => {
            onImport?.();
            showMenu = false;
          }}
          data-testid="menu-import"
        >
          {t("controls.import")}
        </button>
        <button
          type="button"
          class="menu-item"
          role="menuitem"
          onclick={() => {
            onExport?.();
            showMenu = false;
          }}
          data-testid="menu-export"
        >
          {t("controls.export")}
        </button>
      </div>
    {/if}
  </div>

  <!-- Language Toggle -->
  <LangSwitcher />

  <!-- Create Button -->
  <button
    type="button"
    class="create-btn"
    onclick={() => onCreate?.()}
    title={t("controls.createTitle")}
    data-testid="create-note-button"
    aria-label={t("controls.createAria")}
  >
    <svg
      xmlns="http://www.w3.org/2000/svg"
      width="24"
      height="24"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
    >
      <line x1="12" y1="5" x2="12" y2="19" />
      <line x1="5" y1="12" x2="19" y2="12" />
    </svg>
  </button>
</div>

<style>
  .floating-controls {
    position: fixed;
    top: 16px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border-radius: 50px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
    z-index: 100;
    width: clamp(320px, 90vw, 900px);
    /* Allow natural height — chips must not be clipped */
    min-height: 56px;
    box-sizing: border-box;
    flex-wrap: nowrap;
    overflow: visible;
  }

  .floating-controls::-webkit-scrollbar {
    display: none;
  }

  .view-toggle {
    display: flex;
    gap: 4px;
    padding: 4px;
    background: #f1f5f9;
    border-radius: 25px;
    flex-shrink: 0;
  }

  .toggle-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 8px 12px;
    border: none;
    background: transparent;
    border-radius: 20px;
    cursor: pointer;
    transition: all 0.2s;
    color: #64748b;
    font-size: 12px;
  }

  .toggle-btn:hover {
    background: rgba(255, 255, 255, 0.5);
  }

  .toggle-btn.active {
    background: white;
    color: #3b82f6;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .btn-label {
    font-weight: 500;
  }

  .search-container {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 8px 4px 14px;
    background: #f8fafc;
    border-radius: 25px;
    border: 1px solid #e2e8f0;
    flex-shrink: 0;
    height: 38px;
    box-sizing: border-box;
  }

  .search-input {
    border: none;
    background: transparent;
    outline: none;
    font-size: 13px;
    width: clamp(100px, 15vw, 180px);
    color: #334155;
  }

  .search-input::placeholder {
    color: #94a3b8;
  }

  .search-btn {
    padding: 6px;
    border: none;
    background: transparent;
    cursor: pointer;
    color: #64748b;
    border-radius: 50%;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .search-btn:hover {
    background: #e2e8f0;
    color: #334155;
  }

  .menu-container {
    position: relative;
    flex-shrink: 0;
  }

  .menu-btn {
    padding: 8px;
    border: none;
    background: transparent;
    cursor: pointer;
    color: #64748b;
    border-radius: 50%;
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .menu-btn:hover {
    background: #f1f5f9;
    color: #334155;
  }

  .dropdown-menu {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    background: white;
    border-radius: 12px;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
    padding: 8px;
    min-width: 140px;
    z-index: 101;
  }

  .menu-item {
    display: block;
    width: 100%;
    padding: 10px 16px;
    border: none;
    background: transparent;
    text-align: left;
    cursor: pointer;
    border-radius: 8px;
    font-size: 14px;
    color: #334155;
    transition: background 0.2s;
  }

  .menu-item:hover {
    background: #f1f5f9;
  }

  .create-btn {
    padding: 10px;
    border: none;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .create-btn:hover {
    transform: scale(1.05);
    box-shadow: 0 6px 16px rgba(59, 130, 246, 0.5);
  }

  .login-btn {
    padding: 10px;
    border: none;
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
    color: white;
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .login-btn:hover {
    transform: scale(1.05);
    box-shadow: 0 6px 16px rgba(16, 185, 129, 0.5);
  }

  /* Type Filters wrapper — scroll arrows + scrollable row.
     flex: 1 1 0 but capped so right-side buttons (lang, create) always stay visible */
  .type-filters-wrapper {
    display: flex;
    align-items: center;
    flex: 1 1 0;
    min-width: 0;
    max-width: 380px;
    gap: 2px;
  }

  .type-filters {
    display: flex;
    gap: 6px;
    align-items: center;
    flex: 1 1 0;
    min-width: 0;
    flex-wrap: nowrap;
    overflow-x: auto;
    overflow-y: visible;
    scrollbar-width: none;
    padding: 4px 2px;
    -webkit-overflow-scrolling: touch;
  }

  .type-filters::-webkit-scrollbar {
    display: none;
  }

  .scroll-arrow {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 24px;
    height: 28px;
    border: 1px solid #e2e8f0;
    background: white;
    border-radius: 50%;
    cursor: pointer;
    font-size: 16px;
    color: #64748b;
    line-height: 1;
    padding: 0;
    transition: all 0.15s;
    user-select: none;
  }

  .scroll-arrow:hover {
    background: #f1f5f9;
    border-color: #cbd5e0;
    color: #334155;
  }

  .filter-chip {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 5px 12px;
    border: 1px solid #e2e8f0;
    background: white;
    border-radius: 16px;
    cursor: pointer;
    transition: all 0.2s;
    font-size: 13px;
    color: #334155;
    white-space: nowrap;
    flex-shrink: 0;
    min-height: 32px;
  }

  .filter-chip:hover {
    background: #f8fafc;
    border-color: #cbd5e0;
  }

  .filter-chip.active {
    background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
    color: white;
    border-color: transparent;
  }

  .filter-emoji {
    font-size: 14px;
  }

  .filter-label {
    font-weight: 500;
  }

  .filter-count {
    font-size: 11px;
    background: rgba(0, 0, 0, 0.1);
    padding: 2px 6px;
    border-radius: 10px;
    margin-left: 2px;
    font-weight: 600;
  }

  .filter-chip.active .filter-count {
    background: rgba(255, 255, 255, 0.2);
  }

  /* Tablet */
  @media (max-width: 1024px) {
    .floating-controls {
      gap: 6px;
      padding: 8px 12px;
      border-radius: 40px;
      width: clamp(280px, 92vw, 700px);
    }
  }

  /* Mobile — hide labels, show only emojis */
  @media (max-width: 768px) {
    .floating-controls {
      top: 12px;
      left: 50%;
      transform: translateX(-50%);
      width: clamp(260px, 95vw, 600px);
      padding: 8px 10px;
      gap: 6px;
      border-radius: 40px;
      height: 50px;
    }

    .search-container {
      display: none;
    }

    .btn-label {
      display: none;
    }

    .toggle-btn {
      padding: 8px;
    }

    .filter-label {
      display: none;
    }

    .filter-chip {
      padding: 6px 8px;
    }

    .filter-count {
      display: none;
    }

    .create-btn {
      padding: 9px;
    }
  }

  /* Very small mobile — hide view toggle, compact filters */
  @media (max-width: 480px) {
    .floating-controls {
      gap: 4px;
      padding: 6px 8px;
      border-radius: 30px;
      top: 10px;
      height: 46px;
    }

    .view-toggle {
      display: none;
    }

    .filter-chip {
      padding: 5px 6px;
      font-size: 11px;
    }

    .filter-emoji {
      font-size: 12px;
    }

    .menu-btn,
    .create-btn {
      padding: 7px;
    }

    .menu-btn svg,
    .create-btn svg {
      width: 18px;
      height: 18px;
    }
  }
</style>
