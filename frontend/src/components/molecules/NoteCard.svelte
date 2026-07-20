<script lang="ts">
  import type { Note } from "$shared/api/notes";
  import { goto } from "$app/navigation";
  import { formatDate } from "$shared/utils/date";
  import { onMount, onDestroy } from "svelte";
  import type { Instance } from "tippy.js";
  import "tippy.js/dist/tippy.css";
  import { CelestialBody } from "$shared/lib/domain";
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

  const editListener = (e: Event) => handleEdit(e as MouseEvent);
  const deleteListener = (e: Event) => handleDelete(e as MouseEvent);

  function buildTooltipContent(): string {
    const emoji = getTypeEmoji(note.type);
    const title = truncateText(note.title, 60);
    const keywordChips = keywords
      .slice(0, 3)
      .map((k) => `<span class="nc-tooltip-keyword">${k}</span>`)
      .join("");

    return `
      <div class="nc-tooltip" role="tooltip">
        <div class="nc-tooltip-header">
          <span class="nc-tooltip-emoji">${emoji}</span>
          <span class="nc-tooltip-title">${title}</span>
        </div>
        <div class="nc-tooltip-meta">
          <span class="nc-tooltip-links">${t("noteCard.links", { count: linkCount })}</span>
          ${keywordChips ? `<div class="nc-tooltip-keywords">${keywordChips}</div>` : ""}
        </div>
        <div class="nc-tooltip-actions">
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
        const editBtn = instance.popper.querySelector(
          '[data-action="edit"]',
        ) as HTMLElement | null;
        const deleteBtn = instance.popper.querySelector(
          '[data-action="delete"]',
        ) as HTMLElement | null;
        editBtn?.addEventListener("click", editListener);
        deleteBtn?.addEventListener("click", deleteListener);
      },
      onHidden: (instance) => {
        const editBtn = instance.popper.querySelector(
          '[data-action="edit"]',
        ) as HTMLElement | null;
        const deleteBtn = instance.popper.querySelector(
          '[data-action="delete"]',
        ) as HTMLElement | null;
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
  style="--type-color: {getTypeColor(
    note.type,
  )}; --stagger-delay: {animationIndex * 50}ms"
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
      {@html highlightQuery
        ? highlightText(note.title, highlightQuery)
        : note.title}
    </h3>

    <div class="note-card__body" data-testid="note-content">
      {@html highlightQuery
        ? highlightText(truncateText(note.content, 180), highlightQuery)
        : truncateText(note.content, 180)}
    </div>

    <div class="note-card__footer">
      <span
        class="note-card__date"
        data-testid="note-date"
        data-visual-test="transparent"
      >
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
    background: var(--color-surface-elevated, rgba(20, 24, 45, 0.85));
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-radius: 12px;
    overflow: hidden;
    cursor: pointer;
    box-shadow:
      0 1px 3px rgba(0, 0, 0, 0.3),
      inset 0 0 24px rgba(255, 255, 255, 0.02);
    transition:
      transform 0.25s ease,
      box-shadow 0.25s ease,
      border-color 0.25s ease,
      opacity 0.2s ease;
    animation: note-card-enter 0.4s ease forwards;
    animation-delay: var(--stagger-delay, 0ms);
    opacity: 0;
  }

  .note-card::before {
    content: "";
    position: absolute;
    inset: 0;
    background: radial-gradient(
      circle at 20% 30%,
      rgba(255, 255, 255, 0.04) 0%,
      transparent 60%
    );
    pointer-events: none;
    opacity: 0.6;
    transition: opacity 0.25s ease;
  }

  .note-card:hover,
  .note-card:focus {
    transform: translateY(-4px);
    border-color: color-mix(in srgb, var(--type-color) 40%, transparent);
    box-shadow:
      0 8px 28px rgba(0, 0, 0, 0.45),
      0 0 28px color-mix(in srgb, var(--type-color) 18%, transparent),
      inset 0 0 32px rgba(255, 255, 255, 0.03);
    outline: none;
  }

  .note-card:hover::before,
  .note-card:focus::before {
    opacity: 1;
  }

  .note-card.selected {
    border-color: color-mix(in srgb, var(--type-color) 70%, transparent);
    box-shadow:
      0 0 0 2px color-mix(in srgb, var(--type-color) 30%, transparent),
      0 8px 28px rgba(0, 0, 0, 0.45),
      0 0 28px color-mix(in srgb, var(--type-color) 18%, transparent);
  }

  .note-card.exiting {
    animation: note-card-exit 0.2s ease forwards;
  }

  .note-card.dust {
    border-style: dashed;
    border-color: rgba(255, 255, 255, 0.12);
    background: color-mix(
      in srgb,
      var(--color-surface-elevated, rgba(20, 24, 45, 0.85)) 95%,
      transparent
    );
  }

  .note-card.dust::after {
    content: "";
    position: absolute;
    inset: 0;
    background: repeating-linear-gradient(
      45deg,
      rgba(255, 255, 255, 0.01) 0px,
      rgba(255, 255, 255, 0.01) 2px,
      transparent 2px,
      transparent 8px
    );
    pointer-events: none;
    opacity: 0.6;
  }

  .note-card__stripe {
    width: 4px;
    flex-shrink: 0;
    background: var(--type-color);
    box-shadow: 0 0 12px color-mix(in srgb, var(--type-color) 60%, transparent);
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
    filter: drop-shadow(0 0 4px rgba(255, 255, 255, 0.2));
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
    background: var(--color-info, #22d3ee);
    box-shadow: 0 0 8px var(--color-info, #22d3ee);
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
    accent-color: var(--color-primary, #8b5cf6);
    cursor: pointer;
  }

  .note-card__title {
    margin: 0;
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--color-text-dark, #e0e0e0);
    line-height: 1.4;
    word-break: break-word;
  }

  .note-card__body {
    color: rgba(255, 255, 255, 0.62);
    font-size: 0.875rem;
    line-height: 1.6;
    word-break: break-word;
  }

  .note-card__footer {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem 1rem;
    font-size: 0.75rem;
    color: rgba(255, 255, 255, 0.45);
  }

  .note-card__date--updated {
    color: var(--color-info, #22d3ee);
  }

  :global(.nc-tooltip) {
    background: rgba(10, 14, 35, 0.96);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 10px;
    padding: 12px;
    color: white;
    min-width: 180px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
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
    color: rgba(255, 255, 255, 0.7);
  }

  :global(.nc-tooltip-keywords) {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  :global(.nc-tooltip-keyword) {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
    padding: 2px 6px;
    font-size: 0.75rem;
  }

  :global(.nc-tooltip-actions) {
    display: flex;
    gap: 8px;
  }

  :global(.nc-tooltip-btn) {
    flex: 1;
    padding: 6px 10px;
    border: none;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.15s ease;
  }

  :global(.nc-tooltip-btn:hover) {
    opacity: 0.85;
  }

  :global(.nc-tooltip-btn--edit) {
    background: var(--color-primary, #8b5cf6);
    color: white;
  }

  :global(.nc-tooltip-btn--delete) {
    background: rgba(239, 68, 68, 0.85);
    color: white;
  }

  :global(mark) {
    background: linear-gradient(
      120deg,
      rgba(254, 240, 138, 0.3) 0%,
      rgba(253, 224, 71, 0.3) 100%
    );
    color: #ffcc00;
    padding: 0.1em 0.2em;
    border-radius: 0.2em;
    font-weight: 600;
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
      opacity: 0.6;
      transform: scale(1.2);
    }
  }

  /* Tippy translucent theme overrides for dark canvas */
  :global(.tippy-box[data-theme~="translucent"]) {
    background: rgba(10, 14, 35, 0.96);
    border: 1px solid rgba(255, 255, 255, 0.12);
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
  }

  :global(.tippy-box[data-theme~="translucent"] .tippy-arrow) {
    color: rgba(10, 14, 35, 0.96);
  }
</style>
