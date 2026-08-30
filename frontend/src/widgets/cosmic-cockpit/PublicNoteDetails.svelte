<script lang="ts">
  import IconButton from "$components/atoms/IconButton.svelte";
  import Chip from "$components/atoms/Chip.svelte";
  import { CelestialBody, LinkType } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import type { GraphLink } from "$shared/api/graph";

  interface NoteItem {
    id: string;
    title: string;
    type?: string;
  }

  interface RelatedItem {
    id: string;
    title: string;
    type: string;
    linkType: LinkType;
    weight: number;
    direction: "outgoing" | "incoming";
  }

  interface Props {
    nodeId: string;
    notes: NoteItem[];
    links: GraphLink[];
    onClose?: () => void;
    onSignIn?: () => void;
  }

  const { nodeId, notes, links, onClose, onSignIn }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const note = $derived<NoteItem | undefined>(notes.find((n) => n.id === nodeId));
  const noteType = $derived(note ? CelestialBody.fromString(note.type) : CelestialBody.STAR);

  const related = $derived<RelatedItem[]>(
    (() => {
      const nodeMap = new Map<string, NoteItem>();
      for (const n of notes) {
        nodeMap.set(n.id, n);
      }
      const result: RelatedItem[] = [];
      for (const link of links) {
        const isSource = link.source === nodeId;
        const isTarget = link.target === nodeId;
        if (!isSource && !isTarget) continue;
        const otherId = isSource ? link.target : link.source;
        const other = nodeMap.get(otherId);
        if (!other) continue;
        const linkType = LinkType.fromString(link.link_type);
        result.push({
          id: other.id,
          title: other.title,
          type: other.type || "unknown",
          linkType,
          weight: link.weight ?? 0.5,
          direction: isSource ? "outgoing" : "incoming",
        });
      }
      return result;
    })()
  );

  function getNoteTypeIcon(type: string | undefined): string {
    return type ? CelestialBody.fromString(type).emoji : CelestialBody.STAR.emoji;
  }

  function getNoteTypeLabel(type: string | undefined): string {
    return type ? CelestialBody.fromString(type).label : t("noteSidePanel.fallbackType");
  }
</script>

<div class="public-note-details" data-testid="public-note-details">
  <div class="details-header">
    <h2 class="details-title">{t("publicNoteDetail.title")}</h2>
    <IconButton variant="ghost" size="sm" onClick={() => onClose?.()} title={t("noteSidePanel.closeAria")}>
      ✕
    </IconButton>
  </div>

  <div class="details-content">
    {#if note}
      <div class="note-header" style="--type-color: {noteType.toCSSColor()}">
        <div class="hero-top">
          <span class="type-icon" title={noteType.description}>{noteType.emoji}</span>
          <h3 class="title">{note.title}</h3>
        </div>
        <Chip
          size="sm"
          color={noteType.toCSSColor()}
          borderColor={noteType.toCSSColor()}
          glow
        >
          {getNoteTypeLabel(note.type)}
        </Chip>
      </div>

      <div class="visibility">
        <Chip size="sm" color="#22d3ee" glow>{t("note.public")}</Chip>
      </div>

      <div class="links-section">
        <div class="links-header">
          <h4>{t("publicNoteDetail.relatedNotes", { count: related.length })}</h4>
        </div>
        {#if related.length === 0}
          <p class="no-links">{t("note.noLinks")}</p>
        {:else}
          <ul class="links-list">
            {#each related as item}
              <li class="link-item" style="--link-color: {item.linkType.color}">
                <span
                  class="direction"
                  title={t(`note.direction.${item.direction}`)}
                  aria-label={t(`note.direction.${item.direction}`)}
                >
                  {item.direction === "outgoing" ? "→" : "←"}
                </span>
                <span class="link-type-badge" style="--lt-color: {item.linkType.color}">
                  {item.linkType.icon}
                </span>
                <span class="neighbor-title">
                  <span class="type-icon--sm">{getNoteTypeIcon(item.type)}</span>
                  {item.title}
                </span>
                <span class="weight">{t("note.linkWeight", { weight: item.weight.toFixed(1) })}</span>
              </li>
            {/each}
          </ul>
        {/if}
      </div>

      {#if onSignIn}
        <div class="sign-in-prompt">
          <p>{t("publicNoteDetail.signInPrompt")}</p>
          <button type="button" class="sign-in-btn" onclick={() => onSignIn()}>
            {t("publicNoteDetail.signIn")}
          </button>
        </div>
      {/if}
    {:else}
      <p class="not-found">{t("publicNoteDetail.notFound")}</p>
    {/if}
  </div>
</div>

<style>
  .public-note-details {
    display: flex;
    flex-direction: column;
    height: 100%;
    overflow: hidden;
  }

  .details-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 10px 14px;
    border-bottom: 1px solid rgba(45, 212, 191, 0.15);
  }

  .details-title {
    margin: 0;
    font-size: 1rem;
    font-weight: 600;
    color: var(--carbon-text, #f0f0f5);
  }

  .details-content {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
  }

  .note-header {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-bottom: 16px;
  }

  .hero-top {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .type-icon {
    font-size: 2rem;
    line-height: 1;
    filter: drop-shadow(0 0 8px color-mix(in srgb, var(--type-color) 40%, transparent));
  }

  .title {
    margin: 0;
    font-size: 1.35rem;
    font-weight: 700;
    line-height: 1.3;
    color: var(--carbon-text, #f0f0f5);
    word-break: break-word;
  }

  .visibility {
    margin-bottom: 20px;
  }

  .links-section {
    margin-bottom: 24px;
  }

  .links-header {
    display: flex;
    align-items: center;
    margin-bottom: 10px;
  }

  .links-header h4 {
    margin: 0;
    font-size: 0.95rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.85);
  }

  .no-links {
    color: rgba(255, 255, 255, 0.45);
    font-size: 0.9rem;
    margin: 0;
  }

  .links-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .link-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    border-radius: 8px;
    background: rgba(255, 255, 255, 0.04);
    border: 1px solid rgba(255, 255, 255, 0.06);
    font-size: 0.9rem;
    color: var(--carbon-text, #f0f0f5);
  }

  .direction {
    font-family: monospace;
    color: rgba(255, 255, 255, 0.5);
    font-size: 0.85rem;
    min-width: 1.2em;
    text-align: center;
  }

  .link-type-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 22px;
    height: 22px;
    border-radius: 999px;
    background: color-mix(in srgb, var(--lt-color) 15%, transparent);
    color: var(--lt-color);
    font-size: 0.75rem;
  }

  .neighbor-title {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }

  .type-icon--sm {
    font-size: 0.9rem;
  }

  .weight {
    color: rgba(255, 255, 255, 0.5);
    font-size: 0.8rem;
    white-space: nowrap;
  }

  .sign-in-prompt {
    padding: 16px;
    border-radius: 10px;
    background: rgba(34, 211, 238, 0.06);
    border: 1px solid rgba(34, 211, 238, 0.15);
    text-align: center;
  }

  .sign-in-prompt p {
    margin: 0 0 12px;
    font-size: 0.9rem;
    color: rgba(255, 255, 255, 0.75);
  }

  .sign-in-btn {
    padding: 8px 16px;
    border-radius: 8px;
    border: 1px solid var(--carbon-glow-cyan, #22d3ee);
    background: transparent;
    color: var(--carbon-glow-cyan, #22d3ee);
    font-weight: 600;
    cursor: pointer;
    transition:
      background 0.2s ease,
      color 0.2s ease;
  }

  .sign-in-btn:hover {
    background: var(--carbon-glow-cyan, #22d3ee);
    color: var(--carbon-graphene, #12121a);
  }

  .not-found {
    padding: 16px;
    color: rgba(255, 255, 255, 0.5);
    text-align: center;
  }
</style>
