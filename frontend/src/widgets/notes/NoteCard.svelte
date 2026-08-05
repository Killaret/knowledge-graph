<script lang="ts">
  import type { Note } from "$shared/api/notes";
  import { goto } from "$app/navigation";
  import { formatDate } from "$shared/utils/date";
  import { onMount, onDestroy } from "svelte";
  import type { Instance } from "tippy.js";
  import "tippy.js/dist/tippy.css";
  import { CelestialBody } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const HOUR_MS = 60 * 60 * 1000;
  const DAY_MS = 24 * HOUR_MS;

  const {
    note,
    highlightQuery = "",
    selected = false,
    selectMode = false,
    animationIndex = 0,
    keywords = [],
    linkCount = 0,
    onClick,
    onSelect,
    onEdit,
    onDelete,
  }: {
    note: Note;
    highlightQuery?: string;
    selected?: boolean;
    selectMode?: boolean;
    animationIndex?: number;
    keywords?: string[];
    linkCount?: number;
    onClick?: (note: Note) => void;
    onSelect?: (note: Note, selected: boolean) => void;
    onEdit?: (note: Note) => void;
    onDelete?: (note: Note) => void;
  } = $props();

  let cardRef: HTMLElement | null = $state(null);
  let tippyInstance: Instance | null = $state(null);
  const isExiting = $state(false);

  function highlightText(text: string, query: string): string {
    if (!query.trim()) return text;
    const escapedQuery = query.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    const regex = new RegExp(`(${escapedQuery})`, "gi");
    return text.replace(regex, "<mark>$1</mark>");
  }

  function truncateText(text: string, maxLength: number): string {
    if (!text || text.length <= maxLength) return text;
    return text.slice(0, maxLength) + "...";
  }

  function getTypeColor(type?: string): string {
    return CelestialBody.fromString(type).toCSSColor();
  }

  function getTypeEmoji(type?: string): string {
    return CelestialBody.fromString(type).emoji;
  }

  function isNew(): boolean {
    const created = new Date(note.created_at).getTime();
    return Date.now() - created < DAY_MS;
  }

  function isRecentlyUpdated(): boolean {
    const created = new Date(note.created_at).getTime();
    const updated = new Date(note.updated_at).getTime();
    return updated !== created && Date.now() - updated < HOUR_MS;
  }

  function isDust(): boolean {
    return CelestialBody.fromString(note.type).type === "dust";
  }

  function handleClick() {
    if (onClick) {
      onClick(note);
    } else {
      goto(`/notes/${note.id}`);
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      handleClick();
    }
  }

  function handleSelectToggle(e: Event) {
    e.stopPropagation();
    const target = e.target as HTMLInputElement;
    onSelect?.(note, target.checked);
  }

  function handleEdit(e: MouseEvent) {
    e.stopPropagation();
    onEdit?.(note);
    tippyInstance?.hide();
  }

  function handleDelete(e: MouseEvent) {
    e.stopPropagation();
    onDelete?.(note);
    tippyInstance?.hide();
  }

  function handleView(e: MouseEvent) {
    e.stopPropagation();
    handleClick();
    tippyInstance?.hide();
  }

  const editListener = (e: Event) => handleEdit(e as MouseEvent);
  const deleteListener = (e: Event) => handleDelete(e as MouseEvent);
  const viewListener = (e: Event) => handleView(e as MouseEvent);

  function buildTooltipContent(): string {
    const emoji = getTypeEmoji(note.type);
    const title = note.title; // Полный заголовок без обрезки
    const contentPreview = truncateText(note.content, 200); // Больше контента для превью
    const keywordChips = keywords
      .slice(0, 5)
      .map((k) => `<span class="nc-tooltip-keyword">${k}</span>`)
      .join("");

    return `
      <div class="nc-tooltip" role="tooltip">
        <div class="nc-tooltip-header">
          <span class="nc-tooltip-emoji">${emoji}</span>
          <span class="nc-tooltip-title">${title}</span>
        </div>
        <div class="nc-tooltip-content">${contentPreview}</div>
        <div class="nc-tooltip-meta">
          <span class="nc-tooltip-links">${t("noteCard.links", { count: linkCount })}</span>
          <span class="nc-tooltip-date">${formatDate(note.created_at)}</span>
          ${keywordChips ? `<div class="nc-tooltip-keywords">${keywordChips}</div>` : ""}
        </div>
        <div class="nc-tooltip-actions">
          <button class="nc-tooltip-btn nc-tooltip-btn--view" data-action="view" aria-label="${t("noteCard.viewAria")}">${t("noteCard.view")}</button>
          <button class="nc-tooltip-btn nc-tooltip-btn--edit" data-action="edit" aria-label="${t("noteCard.editAria")}">${t("noteCard.edit")}</button>
          <button class="nc-tooltip-btn nc-tooltip-btn--delete" data-action="delete" aria-label="${t("noteCard.deleteAria")}">${t("noteCard.delete")}</button>
        </div>
      </div>
    `;
  }

  onMount(async () => {
    if (!cardRef) return;

    const { default: tippy } = await import("tippy.js");
    if (typeof document === "undefined") return;
    tippyInstance = tippy(cardRef, {
      content: buildTooltipContent(),
      placement: "right",
      animation: "fade",
      duration: 200,
      arrow: true,
      theme: "translucent",
      hideOnClick: false,
      interactive: true,
      allowHTML: true,
      appendTo: document.body,
      onShown: (instance) => {
        const viewBtn = instance.popper.querySelector('[data-action="view"]') as HTMLElement | null;
        const editBtn = instance.popper.querySelector('[data-action="edit"]') as HTMLElement | null;
        const deleteBtn = instance.popper.querySelector(
          '[data-action="delete"]'
        ) as HTMLElement | null;
        viewBtn?.addEventListener("click", viewListener);
        editBtn?.addEventListener("click", editListener);
        deleteBtn?.addEventListener("click", deleteListener);
      },
      onHidden: (instance) => {
        const viewBtn = instance.popper.querySelector('[data-action="view"]') as HTMLElement | null;
        const editBtn = instance.popper.querySelector('[data-action="edit"]') as HTMLElement | null;
        const deleteBtn = instance.popper.querySelector(
          '[data-action="delete"]'
        ) as HTMLElement | null;
        viewBtn?.removeEventListener("click", viewListener);
        editBtn?.removeEventListener("click", editListener);
        deleteBtn?.removeEventListener("click", deleteListener);
      },
    });
  });

  onDestroy(() => {
    tippyInstance?.destroy();
  });
</script>

<div
  bind:this={cardRef}
  class="note-card"
  class:dust={isDust()}
  class:selected
  class:exiting={isExiting}
  data-testid="note-card"
  data-note-id={note.id}
  data-note-type={note.type || "unknown"}
  style="--type-color: {getTypeColor(note.type)}; --stagger-delay: {animationIndex * 50}ms"
  onclick={handleClick}
  onkeydown={handleKeyDown}
  tabindex="0"
  role="button"
  aria-label={t("noteCard.openNote", { title: note.title })}
>
  <div class="note-card__stripe" aria-hidden="true"></div>

  <div class="note-card__content">
    <div class="note-card__header">
      <div class="note-card__type" aria-hidden="true">
        <span class="note-card__emoji">{getTypeEmoji(note.type)}</span>
        {#if note.is_public}
          <span
            class="note-card__public"
            title={t("noteCard.public")}
            aria-label={t("noteCard.publicAria")}>🌐</span
          >
        {/if}
        {#if isNew()}
          <span
            class="note-card__indicator note-card__indicator--new"
            data-visual-test="transparent"
            aria-label={t("noteCard.newNote")}
          ></span>
        {:else if isRecentlyUpdated()}
          <span
            class="note-card__indicator note-card__indicator--updated"
            data-visual-test="transparent"
            aria-label={t("noteCard.recentlyUpdated")}
          ></span>
        {/if}
      </div>

      <div class="note-card__select" class:visible={selectMode || selected}>
        <input
          type="checkbox"
          class="note-card__checkbox"
          checked={selected}
          aria-checked={selected}
          aria-label={t("noteCard.selectNote", { title: note.title })}
          onclick={handleSelectToggle}
        />
      </div>
    </div>

    <h3 class="note-card__title" data-testid="note-title">
      {@html highlightQuery ? highlightText(note.title, highlightQuery) : note.title}
    </h3>

    <div class="note-card__body" data-testid="note-content">
      {@html highlightQuery
        ? highlightText(truncateText(note.content, 180), highlightQuery)
        : truncateText(note.content, 180)}
    </div>

    <div class="note-card__footer">
      <span class="note-card__date" data-testid="note-date" data-visual-test="transparent">
        {t("noteCard.starLit", { date: formatDate(note.created_at) })}
      </span>
      {#if isRecentlyUpdated()}
        <span
          class="note-card__date note-card__date--updated"
          data-testid="note-updated-date"
          data-visual-test="transparent"
        >
          {t("noteCard.orbitCorrected", { date: formatDate(note.updated_at) })}
        </span>
      {/if}
    </div>
  </div>
</div>

<style>
  .note-card {
    position: relative;
    display: flex;
    background: var(
      --carbon-gradient-card,
      linear-gradient(145deg, rgba(30, 30, 42, 0.7) 0%, rgba(18, 18, 26, 0.9) 100%)
    );
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 14px;
    overflow: hidden;
    cursor: pointer;
    box-shadow:
      0 4px 12px rgba(0, 0, 0, 0.35),
      inset 0 0 24px rgba(139, 92, 246, 0.04);
    transition:
      transform var(--carbon-transition, 0.25s ease),
      box-shadow var(--carbon-transition, 0.25s ease),
      border-color var(--carbon-transition, 0.25s ease),
      opacity 0.2s ease;
    animation: note-card-enter 0.4s ease forwards;
    animation-delay: var(--stagger-delay, 0ms);
    opacity: 0;
  }

  .note-card::before {
    content: "";
    position: absolute;
    inset: 0;
    background: radial-gradient(circle at 15% 25%, rgba(34, 211, 238, 0.06) 0%, transparent 55%);
    pointer-events: none;
    opacity: 0.6;
    transition: opacity 0.25s ease;
  }

  .note-card:hover,
  .note-card:focus {
    transform: translateY(-4px);
    border-color: color-mix(in srgb, var(--type-color) 50%, var(--carbon-border, #2d2d3d));
    box-shadow:
      0 12px 32px rgba(0, 0, 0, 0.5),
      0 0 32px color-mix(in srgb, var(--type-color) 22%, transparent),
      inset 0 0 32px rgba(255, 255, 255, 0.02);
    outline: none;
  }

  .note-card:hover::before,
  .note-card:focus::before {
    opacity: 1;
  }

  .note-card.selected {
    border-color: color-mix(in srgb, var(--type-color) 70%, transparent);
    box-shadow:
      0 0 0 2px color-mix(in srgb, var(--type-color) 35%, transparent),
      0 12px 32px rgba(0, 0, 0, 0.5),
      0 0 32px color-mix(in srgb, var(--type-color) 22%, transparent);
  }

  .note-card.exiting {
    animation: note-card-exit 0.2s ease forwards;
  }

  .note-card.dust {
    border-style: dashed;
    border-color: var(--carbon-border, #2d2d3d);
    background: var(--carbon-graphene, #12121a);
  }

  .note-card.dust::after {
    content: "";
    position: absolute;
    inset: 0;
    background: repeating-linear-gradient(
      45deg,
      rgba(255, 255, 255, 0.02) 0px,
      rgba(255, 255, 255, 0.02) 2px,
      transparent 2px,
      transparent 8px
    );
    pointer-events: none;
    opacity: 0.5;
  }

  .note-card__stripe {
    width: 5px;
    flex-shrink: 0;
    background: var(--type-color);
    box-shadow: 0 0 14px color-mix(in srgb, var(--type-color) 70%, transparent);
  }

  .note-card__content {
    flex: 1;
    padding: 1rem 1.25rem 1rem 1rem;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .note-card__header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .note-card__type {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
  }

  .note-card__emoji {
    font-size: 1.25rem;
    line-height: 1;
    filter: drop-shadow(0 0 5px color-mix(in srgb, var(--type-color) 50%, transparent));
  }

  .note-card__public {
    font-size: 0.875rem;
    line-height: 1;
    opacity: 0.8;
  }

  .note-card__indicator {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .note-card__indicator--new {
    background: var(--color-star, #fbbf24);
    box-shadow: 0 0 8px var(--color-star, #fbbf24);
    animation: pulse-new 1.5s ease-in-out infinite;
  }

  .note-card__indicator--updated {
    background: var(--carbon-glow-cyan, #22d3ee);
    box-shadow: 0 0 8px var(--carbon-glow-cyan, #22d3ee);
  }

  .note-card__select {
    opacity: 0;
    transition: opacity 0.2s ease;
  }

  .note-card:hover .note-card__select,
  .note-card:focus .note-card__select,
  .note-card__select.visible {
    opacity: 1;
  }

  .note-card__checkbox {
    width: 18px;
    height: 18px;
    accent-color: var(--carbon-glow-cyan, #22d3ee);
    cursor: pointer;
  }

  .note-card__title {
    margin: 0;
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--carbon-text, #f0f0f5);
    line-height: 1.4;
    word-break: break-word;
  }

  .note-card__title :global(mark) {
    background: linear-gradient(120deg, rgba(245, 158, 11, 0.35) 0%, rgba(249, 115, 22, 0.35) 100%);
    color: var(--carbon-glow-amber, #f59e0b);
    padding: 0.1em 0.2em;
    border-radius: 0.2em;
    font-weight: 600;
  }

  .note-card__body {
    color: var(--carbon-text-muted, #8b8b9e);
    font-size: 0.875rem;
    line-height: 1.6;
    word-break: break-word;
  }

  .note-card__body :global(mark) {
    background: linear-gradient(120deg, rgba(245, 158, 11, 0.35) 0%, rgba(249, 115, 22, 0.35) 100%);
    color: var(--carbon-glow-amber, #f59e0b);
    padding: 0.1em 0.2em;
    border-radius: 0.2em;
    font-weight: 600;
  }

  .note-card__footer {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem 1rem;
    font-size: 0.75rem;
    color: var(--carbon-text-dim, #5a5a6e);
  }

  .note-card__date--updated {
    color: var(--carbon-glow-cyan, #22d3ee);
  }

  :global(.nc-tooltip) {
    background: var(--carbon-c70, #1a1a24);
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 12px;
    padding: 12px;
    color: var(--carbon-text, #f0f0f5);
    min-width: 180px;
    box-shadow:
      0 12px 32px rgba(0, 0, 0, 0.5),
      0 0 24px rgba(139, 92, 246, 0.12);
  }

  :global(.nc-tooltip-header) {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
  }

  :global(.nc-tooltip-emoji) {
    font-size: 1.1rem;
  }

  :global(.nc-tooltip-title) {
    font-weight: 600;
    font-size: 0.9rem;
  }

  :global(.nc-tooltip-meta) {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-bottom: 10px;
    font-size: 0.8rem;
    color: var(--carbon-text-muted, #8b8b9e);
  }

  :global(.nc-tooltip-content) {
    margin: 8px 0;
    padding: 8px;
    background: var(--carbon-black, #050508);
    border-radius: 8px;
    font-size: 0.85rem;
    color: var(--carbon-text, #f0f0f5);
    line-height: 1.4;
    max-height: 150px;
    overflow-y: auto;
  }

  :global(.nc-tooltip-date) {
    font-size: 0.75rem;
    color: var(--carbon-text-dim, #5a5a6e);
  }

  :global(.nc-tooltip-keywords) {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  :global(.nc-tooltip-keyword) {
    background: var(--carbon-graphene, #12121a);
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 4px;
    padding: 2px 6px;
    font-size: 0.75rem;
    color: var(--carbon-glow-cyan, #22d3ee);
  }

  :global(.nc-tooltip-actions) {
    display: flex;
    gap: 8px;
  }

  :global(.nc-tooltip-btn) {
    flex: 1;
    padding: 6px 10px;
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 8px;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: all var(--carbon-transition, 0.25s ease);
    background: var(--carbon-graphene, #12121a);
    color: var(--carbon-text, #f0f0f5);
  }

  :global(.nc-tooltip-btn:hover) {
    opacity: 1;
    border-color: var(--carbon-border-active, #4b4b5e);
  }

  :global(.nc-tooltip-btn--view) {
    background: var(--carbon-graphene, #12121a);
    color: var(--carbon-glow-cyan, #22d3ee);
  }

  :global(.nc-tooltip-btn--view:hover) {
    background: rgba(34, 211, 238, 0.1);
    box-shadow: var(--carbon-glow-cyan, 0 0 10px rgba(34, 211, 238, 0.2));
  }

  :global(.nc-tooltip-btn--edit) {
    background: var(--carbon-graphene, #12121a);
    color: var(--carbon-glow-amber, #f59e0b);
  }

  :global(.nc-tooltip-btn--edit:hover) {
    background: rgba(245, 158, 11, 0.1);
    box-shadow: var(--carbon-glow-amber, 0 0 10px rgba(245, 158, 11, 0.2));
  }

  :global(.nc-tooltip-btn--delete) {
    background: var(--carbon-graphene, #12121a);
    color: var(--carbon-glow-red, #ff3a2f);
  }

  :global(.nc-tooltip-btn--delete:hover) {
    background: rgba(255, 58, 47, 0.1);
    box-shadow: var(--carbon-glow-red, 0 0 10px rgba(255, 58, 47, 0.2));
  }

  @keyframes note-card-enter {
    from {
      opacity: 0;
      transform: translateY(12px) scale(0.98);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  @keyframes note-card-exit {
    from {
      opacity: 1;
      transform: scale(1);
    }
    to {
      opacity: 0;
      transform: scale(0.95);
    }
  }

  @keyframes pulse-new {
    0%,
    100% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.7;
      transform: scale(1.15);
    }
  }

  /* Tippy translucent theme overrides for dark canvas */
  :global(.tippy-box[data-theme~="translucent"]) {
    background: var(--carbon-c70, #1a1a24);
    border: 1px solid var(--carbon-border, #2d2d3d);
    box-shadow:
      0 12px 32px rgba(0, 0, 0, 0.5),
      0 0 24px rgba(139, 92, 246, 0.12);
  }

  :global(.tippy-box[data-theme~="translucent"] .tippy-arrow) {
    color: var(--carbon-c70, #1a1a24);
  }
</style>
