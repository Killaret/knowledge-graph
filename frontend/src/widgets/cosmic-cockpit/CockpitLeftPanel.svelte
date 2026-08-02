<script lang="ts">
  import { goto } from "$app/navigation";
  import { isAuthenticated, currentUser } from "$shared/stores/auth.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import CockpitTypeFilter from "$components/molecules/CockpitTypeFilter.svelte";

  interface Props {
    typeFilters?: Array<{
      id: string;
      label: string;
      emoji: string;
      description?: string;
      example?: string;
    }>;
    selectedType?: string;
    typeCounts?: Record<string, number>;
    currentView?: "graph" | "list" | "3d";
    layoutProvider?: "d3" | "graph-service";
    onToggleView?: (view: "graph" | "list" | "3d") => void;
    onToggleLayoutProvider?: (provider: "d3" | "graph-service") => void;
    onFilter?: (type: string) => void;
    onFocus?: () => void;
    onReset?: () => void;
    onSearch?: () => void;
    notes?: Array<{ id: string; title: string; type?: string }>;
    onNoteSelect?: (id: string) => void;
  }

  const {
    typeFilters = [],
    selectedType = "all",
    typeCounts = {},
    currentView = "graph",
    layoutProvider = "graph-service",
    onToggleView,
    onToggleLayoutProvider,
    onFilter,
    onFocus,
    onReset,
    onSearch,
    notes = [],
    onNoteSelect,
  }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const authenticated = $derived(isAuthenticated());
  const user = $derived(currentUser());

  const navItems = [
    { href: "/", label: t("cockpit.nav.home"), emoji: "🏠" },
    { href: "/graph", label: t("cockpit.nav.graph"), emoji: "🌌" },
    { href: "/notes", label: t("cockpit.nav.notes"), emoji: "📝" },
    { href: "/search", label: t("cockpit.nav.search"), emoji: "🔍" },
    { href: "/settings", label: t("cockpit.nav.settings"), emoji: "⚙️" },
  ];

  function navigate(href: string) {
    goto(href);
  }

  function handleFilter(id: string) {
    onFilter?.(id);
  }

  function toggleLayoutProvider(provider: "d3" | "graph-service") {
    onToggleLayoutProvider?.(provider);
  }

  function getNoteEmoji(type: string | undefined): string {
    // fallback to star; real mapping via CelestialBody can be added later
    const map: Record<string, string> = {
      star: "⭐",
      planet: "🪐",
      moon: "🌙",
      comet: "☄️",
      galaxy: "🌌",
      asteroid: "🌑",
      dust: "💫",
      unknown: "❓",
    };
    return map[type ?? "unknown"] || "⭐";
  }
</script>

<div class="cockpit-left-panel" data-testid="cockpit-left-panel">
  <section class="left-section" aria-labelledby="nav-heading">
    <h3 class="section-heading" id="nav-heading">{t("cockpit.left.navigation")}</h3>
    <nav class="nav-list">
      {#each navItems as item}
        <a
          href={item.href}
          class="nav-link"
          onclick={(e: MouseEvent) => {
            e.preventDefault();
            navigate(item.href);
          }}
        >
          <span class="nav-emoji">{item.emoji}</span>
          <span class="nav-label">{item.label}</span>
        </a>
      {/each}
    </nav>
    {#if authenticated}
      <div class="user-badge">
        <span class="user-avatar">{user?.email?.[0]?.toUpperCase() ?? "?"}</span>
        <span class="user-email" title={user?.email ?? ""}
          >{user?.email ?? t("cockpit.left.user")}</span
        >
      </div>
    {/if}
  </section>

  <section class="left-section" aria-labelledby="graph-heading">
    <h3 class="section-heading" id="graph-heading">{t("cockpit.left.graphControls")}</h3>
    <div class="graph-controls">
      <button
        type="button"
        class="control-btn"
        onclick={() => onReset?.()}
        title={t("cockpit.left.reset")}
      >
        🔄 {t("cockpit.left.reset")}
      </button>
      <button
        type="button"
        class="control-btn"
        onclick={() => onSearch?.()}
        title={t("cockpit.left.search")}
      >
        🎯 {t("cockpit.left.search")}
      </button>
      <button
        type="button"
        class="control-btn"
        onclick={() => onFocus?.()}
        title={t("cockpit.left.focus")}
      >
        👁 {t("cockpit.left.focus")}
      </button>
    </div>

    <h3 class="section-heading sub-heading">{t("cockpit.left.view")}</h3>
    <div class="view-toggle">
      {#each ["graph", "list", "3d"] as view}
        <button
          type="button"
          class="toggle-btn {currentView === view ? 'active' : ''}"
          onclick={() => onToggleView?.(view as "graph" | "list" | "3d")}
          aria-pressed={currentView === view}
        >
          {view.toUpperCase()}
        </button>
      {/each}
    </div>

    {#if currentView === "3d"}
      <h3 class="section-heading sub-heading">{t("cockpit.left.layoutProvider")}</h3>
      <div class="view-toggle">
        <button
          type="button"
          class="toggle-btn {layoutProvider === 'd3' ? 'active' : ''}"
          onclick={() => toggleLayoutProvider("d3")}
          aria-pressed={layoutProvider === "d3"}
        >
          D3
        </button>
        <button
          type="button"
          class="toggle-btn {layoutProvider === 'graph-service' ? 'active' : ''}"
          onclick={() => toggleLayoutProvider("graph-service")}
          aria-pressed={layoutProvider === "graph-service"}
        >
          Service
        </button>
      </div>
    {/if}
  </section>

  {#if typeFilters.length > 0}
    <section class="left-section" aria-labelledby="filter-heading">
      <h3 class="section-heading" id="filter-heading">{t("cockpit.left.filters")}</h3>
      <CockpitTypeFilter
        filters={typeFilters}
        selected={selectedType}
        onSelect={handleFilter}
        {typeCounts}
      />
    </section>
  {/if}

  <section class="left-section" aria-labelledby="tree-heading">
    <h3 class="section-heading" id="tree-heading">{t("cockpit.left.noteTree")}</h3>
    <div class="note-tree">
      {#each notes as note}
        <button
          type="button"
          class="tree-item"
          onclick={() => onNoteSelect?.(note.id)}
          data-testid="cockpit-note-tree-item"
        >
          <span class="tree-emoji">{getNoteEmoji(note.type)}</span>
          <span class="tree-title">{note.title}</span>
        </button>
      {:else}
        <p class="empty-note-tree">{t("cockpit.left.noNotes")}</p>
      {/each}
    </div>
  </section>
</div>

<style>
  .cockpit-left-panel {
    display: flex;
    flex-direction: column;
    gap: 18px;
    height: 100%;
    overflow-y: auto;
    padding: 10px 12px 20px;
    box-sizing: border-box;
  }

  .left-section {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .section-heading {
    margin: 0;
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: rgba(45, 212, 191, 0.8);
  }

  .sub-heading {
    margin-top: 8px;
  }

  .nav-list,
  .graph-controls,
  .note-tree {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .nav-link,
  .control-btn,
  .tree-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    border: 1px solid rgba(45, 212, 191, 0.1);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.03);
    color: var(--color-text, #e0e0e0);
    font-size: 13px;
    text-decoration: none;
    text-align: left;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .nav-link:hover,
  .control-btn:hover,
  .tree-item:hover {
    background: rgba(45, 212, 191, 0.08);
    border-color: rgba(45, 212, 191, 0.25);
  }

  .nav-emoji,
  .tree-emoji {
    width: 20px;
    text-align: center;
  }

  .view-toggle {
    display: flex;
    gap: 4px;
    padding: 4px;
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(45, 212, 191, 0.1);
    border-radius: 8px;
  }

  .toggle-btn {
    flex: 1;
    padding: 6px 8px;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: rgba(255, 255, 255, 0.6);
    font-size: 11px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .toggle-btn:hover {
    color: white;
  }

  .toggle-btn.active {
    background: rgba(45, 212, 191, 0.2);
    color: #2dd4bf;
  }

  .user-badge {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px;
    margin-top: 6px;
    border: 1px solid rgba(45, 212, 191, 0.1);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.03);
  }

  .user-avatar {
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    background: rgba(45, 212, 191, 0.2);
    color: #2dd4bf;
    font-weight: 700;
    font-size: 12px;
  }

  .user-email {
    font-size: 12px;
    color: rgba(255, 255, 255, 0.7);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 180px;
  }

  .note-tree {
    max-height: 180px;
    overflow-y: auto;
  }

  .tree-title {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .empty-note-tree {
    margin: 0;
    padding: 10px;
    color: rgba(255, 255, 255, 0.4);
    font-size: 12px;
    text-align: center;
  }
</style>
