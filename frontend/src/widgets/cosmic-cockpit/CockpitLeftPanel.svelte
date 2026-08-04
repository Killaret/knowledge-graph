<script lang="ts">
  import { goto } from "$app/navigation";
  import { isAuthenticated, currentUser, logout } from "$shared/stores/auth.svelte";
  import { graphStore } from "$shared/stores/graph.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { LinkType } from "$entities";
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
    onFilter?: (type: string) => void;
    onFocus?: () => void;
    onReset?: () => void;
    onSearch?: () => void;
    notes?: Array<{ id: string; title: string; type?: string }>;
    onNoteSelect?: (id: string) => void;
    onImport?: () => void;
    onExport?: () => void;
    showFullGraph?: boolean;
    onToggleFullGraph?: (value: boolean) => void;
  }

  const {
    typeFilters = [],
    selectedType = "all",
    typeCounts = {},
    onFilter,
    onFocus,
    onReset,
    onSearch,
    notes = [],
    onNoteSelect,
    onImport,
    onExport,
    showFullGraph = true,
    onToggleFullGraph,
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

  function handleLogout() {
    logout();
    goto("/auth/login");
  }

  const linkTypes = LinkType.ALL_TYPES;
  const areAllLinkTypesVisible = $derived(graphStore.hiddenLinkTypes.length === 0);
  const areAllLinkTypesHidden = $derived(graphStore.hiddenLinkTypes.length === linkTypes.length);

  function toggleLinkType(type: string) {
    graphStore.toggleLinkType(type);
  }

  function showAllLinkTypes() {
    graphStore.hiddenLinkTypes = [];
  }

  function hideAllLinkTypes() {
    graphStore.hiddenLinkTypes = linkTypes.map((lt) => lt.type);
  }

  function handleMinWeightInput(event: Event) {
    graphStore.minLinkWeight = Number((event.currentTarget as HTMLInputElement).value);
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
    <h3 class="section-heading cockpit-gradient-text" id="nav-heading">
      {t("cockpit.left.navigation")}
    </h3>
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
    <h3 class="section-heading cockpit-gradient-text" id="graph-heading">
      {t("cockpit.left.graphControls")}
    </h3>
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

    {#if onToggleFullGraph}
      <h3 class="section-heading sub-heading cockpit-gradient-text">
        {t("cockpit.left.graphScope")}
      </h3>
      <label class="full-graph-toggle">
        <input
          type="checkbox"
          checked={showFullGraph}
          onchange={(e) => onToggleFullGraph?.(e.currentTarget.checked)}
          data-testid="full-graph-toggle"
        />
        <span>{t("graph.showAllNotes")}</span>
      </label>
    {/if}
  </section>

  {#if typeFilters.length > 0}
    <section class="left-section" aria-labelledby="filter-heading">
      <h3 class="section-heading cockpit-gradient-text" id="filter-heading">
        {t("cockpit.left.filters")}
      </h3>
      <CockpitTypeFilter
        filters={typeFilters}
        selected={selectedType}
        onSelect={handleFilter}
        {typeCounts}
      />
    </section>
  {/if}

  <section class="left-section" aria-labelledby="link-types-heading">
    <h3 class="section-heading cockpit-gradient-text" id="link-types-heading">
      {t("linkLegend.title")}
    </h3>
    <div class="link-types-actions">
      <button
        type="button"
        class="control-btn control-btn--small"
        disabled={areAllLinkTypesVisible}
        onclick={showAllLinkTypes}
        data-testid="link-types-show-all"
      >
        {t("linkLegend.showAll")}
      </button>
      <button
        type="button"
        class="control-btn control-btn--small"
        disabled={areAllLinkTypesHidden}
        onclick={hideAllLinkTypes}
        data-testid="link-types-hide-all"
      >
        {t("linkLegend.hideAll")}
      </button>
    </div>
    <div class="link-types-list">
      {#each linkTypes as linkType}
        {@const active = !graphStore.hiddenLinkTypes.includes(linkType.type)}
        <button
          type="button"
          class="link-type-chip"
          class:active
          onclick={() => toggleLinkType(linkType.type)}
          data-testid="link-type-chip-{linkType.type}"
        >
          <span class="link-type-icon">{linkType.icon}</span>
          <span>{linkType.label}</span>
        </button>
      {/each}
    </div>
    <label class="min-weight-label" for="cockpit-min-weight">
      {t("linkLegend.minWeight", { weight: graphStore.minLinkWeight.toFixed(1) })}
    </label>
    <input
      id="cockpit-min-weight"
      type="range"
      min="0"
      max="1"
      step="0.1"
      value={graphStore.minLinkWeight}
      oninput={handleMinWeightInput}
      data-testid="cockpit-min-weight"
    />
  </section>

  <section class="left-section" aria-labelledby="note-list-heading">
    <h3 class="section-heading cockpit-gradient-text" id="note-list-heading">
      {t("cockpit.left.noteList")}
    </h3>
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

  <section class="left-section" aria-labelledby="system-heading">
    <h3 class="section-heading cockpit-gradient-text" id="system-heading">
      {t("cockpit.left.system")}
    </h3>
    <div class="system-controls">
      {#if onImport}
        <button
          type="button"
          class="control-btn"
          onclick={() => onImport?.()}
          data-testid="menu-import"
        >
          📥 {t("controls.import")}
        </button>
      {/if}
      {#if onExport}
        <button
          type="button"
          class="control-btn"
          onclick={() => onExport?.()}
          data-testid="menu-export"
        >
          📤 {t("controls.export")}
        </button>
      {/if}
      <button
        type="button"
        class="control-btn control-btn--danger"
        onclick={handleLogout}
        data-testid="menu-logout"
      >
        ↩ {t("auth.logout")}
      </button>
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
  }

  .sub-heading {
    margin-top: 8px;
  }

  /* Desynchronize the gradient shimmer across sections so headings don't
     all shimmer in lockstep. */
  .left-section:nth-of-type(1) .cockpit-gradient-text {
    --cockpit-text-delay: 0s;
  }
  .left-section:nth-of-type(2) .cockpit-gradient-text {
    --cockpit-text-delay: -1.2s;
  }
  .left-section:nth-of-type(3) .cockpit-gradient-text {
    --cockpit-text-delay: -2.4s;
  }
  .left-section:nth-of-type(4) .cockpit-gradient-text {
    --cockpit-text-delay: -3.6s;
  }
  .left-section:nth-of-type(5) .cockpit-gradient-text {
    --cockpit-text-delay: -4.8s;
  }
  .left-section:nth-of-type(6) .cockpit-gradient-text {
    --cockpit-text-delay: -6s;
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

  .full-graph-toggle {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    border: 1px solid rgba(45, 212, 191, 0.1);
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.03);
    color: var(--color-text, #e0e0e0);
    font-size: 13px;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .full-graph-toggle:hover {
    background: rgba(45, 212, 191, 0.08);
    border-color: rgba(45, 212, 191, 0.25);
  }

  .full-graph-toggle input {
    width: 18px;
    height: 18px;
    accent-color: #2dd4bf;
    cursor: pointer;
  }

  .system-controls {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .link-types-actions {
    display: flex;
    gap: 6px;
  }

  .control-btn--small {
    flex: 1;
    padding: 6px 8px;
    font-size: 11px;
    justify-content: center;
  }

  .control-btn--small:disabled {
    opacity: 0.4;
    cursor: default;
  }

  .link-types-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
  }

  .link-type-chip {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    border-radius: 6px;
    border: 1px solid rgba(45, 212, 191, 0.15);
    background: transparent;
    color: rgba(224, 224, 224, 0.5);
    font-size: 11px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .link-type-chip.active {
    border-color: rgba(45, 212, 191, 0.5);
    background: rgba(45, 212, 191, 0.1);
    color: var(--color-text, #e0e0e0);
  }

  .link-type-chip:hover {
    border-color: rgba(45, 212, 191, 0.5);
  }

  .link-type-icon {
    font-size: 12px;
  }

  .min-weight-label {
    font-size: 11px;
    color: rgba(224, 224, 224, 0.6);
  }

  .control-btn--danger:hover {
    background: rgba(248, 113, 113, 0.12);
    border-color: rgba(248, 113, 113, 0.25);
    color: #f87171;
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
