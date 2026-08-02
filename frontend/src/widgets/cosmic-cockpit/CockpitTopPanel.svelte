<script lang="ts">
  import { goto } from "$app/navigation";
  import { isAuthenticated, currentUser, logout } from "$shared/stores/auth.svelte";
  import { SearchQuery } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import LangSwitcher from "$components/atoms/LangSwitcher.svelte";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string>) => formatMessage(key, locale, params);

  interface Props {
    onSearch?: (query: string) => void;
    onToggleView?: (view: "graph" | "list" | "3d") => void;
    onToggleLayoutProvider?: (provider: "d3" | "graph-service") => void;
    onNoteCreate?: () => void;
    onOpenAuth?: (tab: "login" | "register") => void;
    currentView?: "graph" | "list" | "3d";
    layoutProvider?: "d3" | "graph-service";
    onImport?: () => void;
    onExport?: () => void;
  }

  const {
    onSearch,
    onToggleView,
    onToggleLayoutProvider,
    onNoteCreate,
    onOpenAuth,
    currentView = "graph",
    layoutProvider = "graph-service",
    onImport,
    onExport,
  }: Props = $props();

  let searchQuery = $state("");
  let showMenu = $state(false);

  function handleSearch() {
    const q = new SearchQuery(searchQuery);
    onSearch?.(q.value);
  }

  function toggleView(view: "graph" | "list" | "3d") {
    onToggleView?.(view);
  }

  function toggleLayoutProvider(provider: "d3" | "graph-service") {
    onToggleLayoutProvider?.(provider);
  }

  function handleLogin() {
    if (onOpenAuth) {
      onOpenAuth("login");
    } else {
      goto("/auth/login");
    }
    showMenu = false;
  }

  function handleRegister() {
    if (onOpenAuth) {
      onOpenAuth("register");
    } else {
      goto("/auth/register");
    }
    showMenu = false;
  }

  function handleLogout() {
    logout();
    goto("/auth/login");
    showMenu = false;
  }

  const authenticated = $derived(isAuthenticated());
  const user = $derived(currentUser());
</script>

<div class="cockpit-top-panel" data-testid="cockpit-top-panel">
  <div class="left-cluster">
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

    <div class="view-toggle">
      <button
        type="button"
        class="toggle-btn {currentView === 'graph' ? 'active' : ''}"
        onclick={() => toggleView("graph")}
        title={t("controls.graph2DTitle")}
        data-testid="view-toggle-graph"
        aria-pressed={currentView === "graph"}
      >
        <svg
          width="18"
          height="18"
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
        <span class="btn-label">2D</span>
      </button>
      <button
        type="button"
        class="toggle-btn {currentView === '3d' ? 'active' : ''}"
        onclick={() => toggleView("3d")}
        title={t("controls.graph3DTitle")}
        data-testid="view-toggle-3d"
        aria-pressed={currentView === "3d"}
      >
        <svg
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <path d="M12 2L2 7l10 5 10-5-10-5z" />
          <path d="M2 17l10 5 10-5" />
          <path d="M2 12l10 5 10-5" />
        </svg>
        <span class="btn-label">3D</span>
      </button>
      <button
        type="button"
        class="toggle-btn {currentView === 'list' ? 'active' : ''}"
        onclick={() => toggleView("list")}
        title={t("controls.listViewTitle")}
        data-testid="view-toggle-list"
        aria-pressed={currentView === "list"}
      >
        <svg
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
        >
          <line x1="3" y1="6" x2="21" y2="6" />
          <line x1="3" y1="12" x2="21" y2="12" />
          <line x1="3" y1="18" x2="21" y2="18" />
        </svg>
        <span class="btn-label">List</span>
      </button>
    </div>

    {#if currentView === "3d"}
      <div class="layout-provider-toggle">
        <button
          type="button"
          class="toggle-btn {layoutProvider === 'd3' ? 'active' : ''}"
          onclick={() => toggleLayoutProvider("d3")}
          title={t("controls.layoutD3Title")}
          data-testid="layout-provider-d3"
          aria-pressed={layoutProvider === "d3"}
        >
          {t("controls.layoutD3")}
        </button>
        <button
          type="button"
          class="toggle-btn {layoutProvider === 'graph-service' ? 'active' : ''}"
          onclick={() => toggleLayoutProvider("graph-service")}
          title={t("controls.layoutGraphServiceTitle")}
          data-testid="layout-provider-graph-service"
          aria-pressed={layoutProvider === "graph-service"}
        >
          {t("controls.layoutGraphService")}
        </button>
      </div>
    {/if}
  </div>

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
        width="16"
        height="16"
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

  <div class="right-cluster">
    <button
      type="button"
      class="create-btn"
      onclick={() => onNoteCreate?.()}
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

    <div class="menu-container">
      <button
        type="button"
        class="menu-btn"
        data-testid="menu-button"
        onclick={() => (showMenu = !showMenu)}
        title={t("controls.menuTitle")}
        aria-expanded={showMenu}
        aria-haspopup="true"
      >
        <svg
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
          {#if !authenticated}
            <button
              type="button"
              class="menu-item"
              role="menuitem"
              onclick={handleLogin}
              data-testid="menu-login"
            >
              🔑 {t("auth.signInButton")}
            </button>
            <button
              type="button"
              class="menu-item"
              role="menuitem"
              onclick={handleRegister}
              data-testid="menu-register"
            >
              ✨ {t("auth.registerLink")}
            </button>
          {:else}
            <div class="user-info">{user?.email}</div>
            <button
              type="button"
              class="menu-item"
              role="menuitem"
              onclick={handleLogout}
              data-testid="menu-logout"
            >
              ↩ {t("auth.logout")}
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

    <LangSwitcher />
  </div>
</div>

<style>
  .cockpit-top-panel {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding: 8px 16px;
    height: 100%;
    box-sizing: border-box;
    background: rgba(10, 10, 15, 0.88);
    color: var(--color-text, #e0e0e0);
  }

  .left-cluster,
  .right-cluster {
    display: flex;
    align-items: center;
    gap: 10px;
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
  .layout-provider-toggle {
    display: flex;
    gap: 2px;
    padding: 3px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(45, 212, 191, 0.2);
    border-radius: 20px;
    flex-shrink: 0;
  }

  .toggle-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 10px;
    border: none;
    background: transparent;
    border-radius: 16px;
    color: rgba(255, 255, 255, 0.6);
    font-size: 12px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .toggle-btn:hover {
    background: rgba(45, 212, 191, 0.1);
    color: white;
  }

  .toggle-btn.active {
    background: rgba(45, 212, 191, 0.25);
    color: #2dd4bf;
    box-shadow: 0 0 10px rgba(45, 212, 191, 0.2);
  }

  .btn-label {
    font-weight: 500;
  }

  .search-container {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(45, 212, 191, 0.2);
    border-radius: 20px;
    flex: 1 1 auto;
    min-width: 0;
    max-width: 360px;
  }

  .search-input {
    flex: 1;
    border: none;
    background: transparent;
    outline: none;
    color: var(--color-text, #e0e0e0);
    font-size: 13px;
    min-width: 0;
  }

  .search-input::placeholder {
    color: rgba(255, 255, 255, 0.4);
  }

  .search-btn {
    padding: 4px;
    border: none;
    background: transparent;
    color: #2dd4bf;
    cursor: pointer;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .create-btn,
  .menu-btn {
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

  .create-btn:hover,
  .menu-btn:hover {
    background: rgba(45, 212, 191, 0.25);
  }

  .menu-container {
    position: relative;
  }

  .dropdown-menu {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    background: rgba(10, 10, 15, 0.95);
    border: 1px solid rgba(45, 212, 191, 0.25);
    border-radius: 10px;
    padding: 8px;
    min-width: 160px;
    z-index: 160;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
  }

  .menu-item {
    width: 100%;
    padding: 10px 14px;
    border: none;
    background: transparent;
    color: var(--color-text, #e0e0e0);
    text-align: left;
    cursor: pointer;
    border-radius: 6px;
    font-size: 13px;
    transition: background 0.2s ease;
  }

  .menu-item:hover {
    background: rgba(45, 212, 191, 0.15);
  }

  .user-info {
    padding: 8px 14px;
    font-size: 12px;
    color: rgba(255, 255, 255, 0.6);
    border-bottom: 1px solid rgba(45, 212, 191, 0.1);
    margin-bottom: 4px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 180px;
  }

  @media (max-width: 768px) {
    .btn-label,
    .logo-text {
      display: none;
    }

    .search-container {
      max-width: none;
    }
  }
</style>
