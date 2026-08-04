<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { page } from "$app/stores";
  import { getNote, getSuggestions, deleteNote, type Note, type Suggestion } from "$shared/api/notes";
  import { getGraphData, type GraphData, type GraphNode, type GraphLink } from "$shared/api/graph";
  import { goto } from "$app/navigation";
  import { fade, fly } from "svelte/transition";
  import { quintOut } from "svelte/easing";
  import BackButton from "$components/atoms/BackButton.svelte";
  import Button from "$components/atoms/Button.svelte";
  import Bevel from "$components/atoms/Bevel.svelte";
  import Chip from "$components/atoms/Chip.svelte";
  import EditNoteModal from "$widgets/notes/EditNoteModal.svelte";
  import CreateNoteModal from "$widgets/notes/CreateNoteModal.svelte";
  import ConfirmModal from "$widgets/confirm/ConfirmModal.svelte";
  import StateIllustration from "$components/atoms/StateIllustration.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { formatDateTime } from "$shared/utils/date";
  import { CelestialBody, LinkType } from "$entities";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  let note = $state<Note | null>(null);
  let graphData = $state<GraphData>({ nodes: [], links: [] });
  let suggestions = $state<Suggestion[]>([]);
  let loading = $state(true);
  let error = $state("");
  let editModalOpen = $state(false);
  let createChildModalOpen = $state(false);
  let deleteConfirmOpen = $state(false);
  let graphLoading = $state(true);

  function getTags(n: Note | null): string[] {
    return ((n?.metadata?.tags as string[]) ?? []).filter(Boolean);
  }

  function getKeywords(n: Note | null): string[] {
    return ((n?.metadata?.keywords as string[]) ?? []).filter(Boolean);
  }

  function getNoteType(n: Note | null): CelestialBody {
    return CelestialBody.fromString(n?.type);
  }

  interface RelatedNote {
    id: string;
    title: string;
    type: string;
    linkId?: string;
    linkType: LinkType;
    weight: number;
    direction: "outgoing" | "incoming";
  }

  const relatedNotes = $derived<RelatedNote[]>(
    (() => {
      if (!note) return [];
      const nodeMap = new Map<string, GraphNode>();
      graphData.nodes.forEach((n) => nodeMap.set(n.id, n));
      const related: RelatedNote[] = [];
      const noteId = note.id;
      graphData.links.forEach((link: GraphLink) => {
        const isSource = link.source === noteId;
        const isTarget = link.target === noteId;
        if (!isSource && !isTarget) return;
        const otherId = isSource ? link.target : link.source;
        const other = nodeMap.get(otherId);
        if (!other) return;
        related.push({
          id: other.id,
          title: other.title,
          type: other.type || "unknown",
          linkId: link.id,
          linkType: LinkType.fromString(link.link_type),
          weight: link.weight ?? 0.5,
          direction: isSource ? "outgoing" : "incoming",
        });
      });
      return related;
    })()
  );

  function getLinkStyle(related: RelatedNote): string {
    return `--link-color: ${related.linkType.color}; --type-color: ${CelestialBody.fromString(related.type).toCSSColor()}`;
  }

  function getRouteId(): string {
    const id = $page.params.id;
    if (!id) throw new Error("Missing route parameter: id");
    return id;
  }

  function getErrorStatus(err: unknown): number | undefined {
    if (typeof err !== "object" || err === null || !("response" in err)) return undefined;
    const response = (err as { response?: unknown }).response;
    if (typeof response !== "object" || response === null || !("status" in response))
      return undefined;
    const status = (response as { status?: unknown }).status;
    return typeof status === "number" ? status : undefined;
  }

  onMount(async () => {
    const id = getRouteId();
    try {
      note = await getNote(id);
    } catch (e: unknown) {
      if (getErrorStatus(e) === 404) {
        error = t("note.notFoundShort");
        setTimeout(() => goto("/"), 3000);
      } else {
        error = t("note.loadError");
      }
      loading = false;
      graphLoading = false;
      return;
    } finally {
      // Keep skeleton for related sections; they load in the background.
    }

    loading = false;

    // Load heavy secondary data in the background so the note renders first.
    Promise.all([
      getSuggestions(id, 6)
        .then((s) => (suggestions = s))
        .catch(() => (suggestions = [])),
      getGraphData(id, 1)
        .then((g) => (graphData = g))
        .catch(() => (graphData = { nodes: [], links: [] }))
        .finally(() => (graphLoading = false)),
    ]);
  });

  async function handleDeleteConfirm() {
    if (!browser || !note) return;
    const id = getRouteId();
    deleteConfirmOpen = false;
    try {
      await deleteNote(id);
      await goto("/");
    } catch {
      error = t("note.deleteError");
    }
  }

  function handleChildCreated(childNote: Note) {
    goto(`/notes/${childNote.id}`);
  }
</script>

{#if loading}
  <div class="page-loading" role="status" aria-live="polite">
    <div class="spinner" aria-hidden="true"></div>
    <p>{t("note.loading")}</p>
  </div>
{:else if error}
  <div class="note-error" in:fade={{ duration: 300 }}>
    <StateIllustration type={error === t("note.notFoundShort") ? "404" : "error"} />
    <p class="error-message">{error}</p>
  </div>
{:else if note}
  <div class="note-page">
    <div class="note-top-bar">
      <BackButton href="/" />
    </div>

    <article
      class="note-article"
      style="--type-color: {getNoteType(note).toCSSColor()}"
      in:fly={{ y: 16, duration: 350, easing: quintOut }}
    >
      <Bevel
        class="note-article-surface"
        variant="note"
        shadeColor={getNoteType(note).toCSSColor()}
        fullHeight
      >
        <header class="note-hero">
          <div class="hero-top">
            <div class="chips">
              <Chip
                color={getNoteType(note).toCSSColor()}
                borderColor={getNoteType(note).toCSSColor()}
                glow
                title={getNoteType(note).description}
              >
                <span class="type-emoji">{getNoteType(note).emoji}</span>
                <span>{getNoteType(note).label}</span>
              </Chip>
              <Chip
                color={note.is_public ? "var(--carbon-glow-cyan, #22d3ee)" : undefined}
                borderColor={note.is_public ? "rgba(34, 211, 238, 0.3)" : undefined}
                glow={note.is_public}
              >
                {note.is_public ? "🌐" : "🔒"}
                {note.is_public ? t("note.public") : t("note.private")}
              </Chip>
            </div>
          <div class="actions">
            <Button
              variant="primary"
              onClick={() => (editModalOpen = true)}
              data-testid="edit-note-btn"
            >
              {t("note.editButton")}
            </Button>
            <Button
              variant="danger"
              onClick={() => (deleteConfirmOpen = true)}
              data-testid="delete-note-btn"
            >
              {t("note.deleteButton")}
            </Button>
            <Button variant="ghost" onClick={() => (createChildModalOpen = true)}>
              {t("note.createChildButton")}
            </Button>
            <Button
              variant="secondary"
              onClick={() => note && goto("/graph/" + note.id)}
              data-testid="view-graph-btn"
            >
              {t("note.showGraph")}
            </Button>
            <Button
              variant="secondary"
              onClick={() => note && goto("/graph/3d/" + note.id)}
              data-testid="view-3d-btn"
            >
              {t("note.showConstellation")}
            </Button>
          </div>
        </div>

        <h1 class="note-title" data-testid="note-detail-title">{note.title}</h1>

        <div class="meta-dates">
          <span title={formatDateTime(note.created_at)}>
            {t("note.createdLabel")}{formatDateTime(note.created_at)}
          </span>
          <span class="dot" aria-hidden="true">·</span>
          <span title={formatDateTime(note.updated_at)}>
            {t("note.updatedLabel")}{formatDateTime(note.updated_at)}
          </span>
        </div>
      </header>

      <section class="note-content" data-testid="note-detail-content">
        {note.content || t("note.noContent")}
      </section>

      {#if getTags(note).length > 0 || getKeywords(note).length > 0}
        <section class="meta-section" in:fly={{ y: 12, duration: 400, delay: 100 }}>
          {#if getTags(note).length > 0}
            <div class="tag-list">
              <span class="meta-label">{t("note.tags")}</span>
              {#each getTags(note) as tag}
                <a class="tag" href={"/search?q=" + encodeURIComponent(tag)}>#{tag}</a>
              {/each}
            </div>
          {/if}
          {#if getKeywords(note).length > 0}
            <div class="keyword-list">
              <span class="meta-label">{t("note.keywords")}</span>
              {#each getKeywords(note) as kw}
                <a class="keyword" href={"/search?q=" + encodeURIComponent(kw)}>{kw}</a>
              {/each}
            </div>
          {/if}
        </section>
      {/if}

      {#if graphLoading}
        <div class="links-loading" role="status">
          <div class="spinner--sm" aria-hidden="true"></div>
          <span>{t("note.linksLoading")}</span>
        </div>
      {:else if relatedNotes.length > 0}
        <section class="links-section" in:fly={{ y: 12, duration: 400, delay: 150 }}>
          <h2 class="section-title">{t("note.relatedNotes")}</h2>
          <div class="links-list">
            {#each relatedNotes as related}
              <a
                class="link-card"
                href={"/notes/" + related.id}
                style={getLinkStyle(related)}
              >
                <span class="link-icon">{related.linkType.icon}</span>
                <span class="link-direction" title={t("note.direction." + related.direction)}>
                  {related.direction === "outgoing" ? "→" : "←"}
                </span>
                <span class="link-emoji">{CelestialBody.fromString(related.type).emoji}</span>
                <span class="link-title">{related.title}</span>
                <span
                  class="link-weight"
                  title={t("note.linkWeight", { weight: related.weight.toFixed(2) })}
                >
                  <span class="weight-fill" style="width: {related.weight * 100}%"></span>
                </span>
              </a>
            {/each}
          </div>
        </section>
      {:else}
        <section class="links-section links-empty">
          <p>{t("note.noLinks")}</p>
          <Button variant="ghost" onClick={() => (createChildModalOpen = true)}>
            {t("note.createFirstLink")}
          </Button>
        </section>
      {/if}

      {#if suggestions.length > 0}
        <section class="suggestions-section" in:fly={{ y: 12, duration: 400, delay: 200 }}>
          <h2 class="section-title">{t("note.similarNotes")}</h2>
          <ul class="suggestions-list">
            {#each suggestions as s}
              <li class="suggestion">
                <a href={"/notes/" + s.note_id} class="suggestion-link">
                  <span class="suggestion-title">{s.title}</span>
                  <span class="score">{t("note.score", { score: s.score.toFixed(3) })}</span>
                </a>
              </li>
            {/each}
          </ul>
        </section>
      {/if}
      </Bevel>
    </article>

    <EditNoteModal
      bind:open={editModalOpen}
      noteId={note?.id ?? ""}
      onSuccess={(updatedNote: Note) => (note = updatedNote)}
    />

    <CreateNoteModal
      bind:open={createChildModalOpen}
      parentNote={
        note
          ? { id: note.id, title: note.title, type: note.type }
          : undefined
      }
      defaultType={note ? CelestialBody.getChildSuggestion(note.type) : "planet"}
      onSuccess={handleChildCreated}
      onClose={() => (createChildModalOpen = false)}
    />

    <ConfirmModal
      bind:open={deleteConfirmOpen}
      title={t("note.deleteConfirmTitle")}
      message={t("note.deleteConfirmMessage", { title: note?.title ?? "" })}
      confirmText={t("note.deleteButton")}
      cancelText={t("note.cancel")}
      danger={true}
      onConfirm={handleDeleteConfirm}
      onCancel={() => (deleteConfirmOpen = false)}
    />
  </div>
{/if}

<style>
  .note-page {
    max-width: 960px;
    margin: 0 auto;
    padding: 0;
    min-height: 100vh;
    color: var(--carbon-text, #f0f0f5);
    background:
      radial-gradient(ellipse at 20% 0%, rgba(45, 212, 191, 0.04) 0%, transparent 50%),
      radial-gradient(ellipse at 80% 100%, rgba(192, 38, 211, 0.04) 0%, transparent 50%),
      var(--carbon-bg, #0b0b11);
    display: flex;
    flex-direction: column;
  }

  .note-top-bar {
    padding: 1rem 1rem 0.5rem;
    flex-shrink: 0;
  }

  .page-loading,
  .note-error {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    padding: 2rem;
    color: var(--carbon-text-muted, #8b8b9e);
  }

  .spinner,
  .spinner--sm {
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--carbon-border, #2d2d3d);
    border-top-color: var(--carbon-glow-cyan, #22d3ee);
    box-shadow: 0 0 14px var(--carbon-glow-cyan, rgba(34, 211, 238, 0.25));
  }

  .spinner--sm {
    width: 20px;
    height: 20px;
    border: 2px solid var(--carbon-border, #2d2d3d);
    border-top-color: var(--carbon-glow-cyan, #22d3ee);
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .error-message {
    color: var(--carbon-glow-red, #ff3a2f);
    font-weight: 500;
  }

  .note-article {
    flex: 1;
    display: flex;
    flex-direction: column;
    margin: 0;
  }

  .note-hero {
    padding: 1.75rem 2rem;
    background: linear-gradient(
      180deg,
      color-mix(in srgb, var(--type-color) 10%, transparent) 0%,
      transparent 100%
    );
    border-bottom: 1px solid var(--carbon-border, #2d2d3d);
  }

  .hero-top {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 1rem;
    justify-content: space-between;
    margin-bottom: 1.25rem;
  }

  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .type-emoji {
    filter: drop-shadow(0 0 4px color-mix(in srgb, var(--type-color) 60%, transparent));
  }

  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .note-title {
    margin: 0 0 0.5rem;
    font-size: clamp(1.5rem, 5vw, 2.25rem);
    font-weight: 700;
    color: var(--carbon-text, #f0f0f5);
    text-shadow: 0 0 20px color-mix(in srgb, var(--type-color) 25%, transparent);
    line-height: 1.2;
  }

  .meta-dates {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
    color: var(--carbon-text-dim, #5a5a6e);
  }

  .dot {
    color: var(--carbon-text-dim, #5a5a6e);
  }

  .note-content {
    padding: 2rem;
    white-space: pre-wrap;
    line-height: 1.75;
    color: var(--carbon-text, #f0f0f5);
    font-size: 1rem;
    border-bottom: 1px solid var(--carbon-border, #2d2d3d);
  }

  .meta-section {
    padding: 1.25rem 2rem;
    border-bottom: 1px solid var(--carbon-border, #2d2d3d);
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .tag-list,
  .keyword-list {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
  }

  .meta-label {
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--carbon-text-dim, #5a5a6e);
    margin-right: 0.25rem;
  }

  .tag,
  .keyword {
    display: inline-flex;
    align-items: center;
    padding: 0.25rem 0.6rem;
    border-radius: 6px;
    font-size: 0.8rem;
    text-decoration: none;
    transition: all var(--carbon-transition, 0.25s ease);
    border: 1px solid var(--carbon-border, #2d2d3d);
    background: var(--carbon-graphene, #12121a);
  }

  .tag {
    color: var(--carbon-glow-cyan, #22d3ee);
  }

  .tag:hover {
    background: rgba(34, 211, 238, 0.1);
    box-shadow: var(--carbon-glow-cyan, 0 0 10px rgba(34, 211, 238, 0.2));
  }

  .keyword {
    color: var(--carbon-glow-amber, #f59e0b);
  }

  .keyword:hover {
    background: rgba(245, 158, 11, 0.1);
    box-shadow: var(--carbon-glow-amber, 0 0 10px rgba(245, 158, 11, 0.2));
  }

  .links-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    padding: 2rem;
    color: var(--carbon-text-muted, #8b8b9e);
  }

  .links-section,
  .suggestions-section {
    padding: 1.5rem 2rem;
    border-bottom: 1px solid var(--carbon-border, #2d2d3d);
  }

  .links-empty {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }

  .section-title {
    margin: 0 0 1rem;
    font-size: 1.125rem;
    font-weight: 600;
    color: var(--carbon-text, #f0f0f5);
  }

  .links-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 0.75rem;
  }

  .link-card {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border-radius: 10px;
    border: 1px solid var(--carbon-border, #2d2d3d);
    background: var(--carbon-graphene, #12121a);
    text-decoration: none;
    color: var(--carbon-text, #f0f0f5);
    transition: all var(--carbon-transition, 0.25s ease);
    box-shadow: inset 0 0 12px rgba(139, 92, 246, 0.03);
  }

  .link-card:hover {
    border-color: color-mix(in srgb, var(--link-color) 50%, var(--carbon-border, #2d2d3d));
    background: color-mix(in srgb, var(--link-color) 8%, var(--carbon-graphene, #12121a));
    box-shadow:
      0 4px 12px rgba(0, 0, 0, 0.3),
      0 0 18px color-mix(in srgb, var(--link-color) 20%, transparent);
    transform: translateY(-2px);
  }

  .link-icon {
    font-size: 1rem;
    filter: drop-shadow(0 0 4px color-mix(in srgb, var(--link-color) 60%, transparent));
  }

  .link-direction {
    font-weight: 700;
    color: var(--carbon-text-dim, #5a5a6e);
  }

  .link-emoji {
    filter: drop-shadow(0 0 4px color-mix(in srgb, var(--type-color) 60%, transparent));
  }

  .link-title {
    flex: 1;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-weight: 500;
  }

  .link-weight {
    position: relative;
    width: 48px;
    height: 6px;
    border-radius: 3px;
    background: var(--carbon-c70, #1a1a24);
    overflow: hidden;
    flex-shrink: 0;
  }

  .weight-fill {
    display: block;
    height: 100%;
    background: var(--link-color);
    box-shadow: 0 0 8px color-mix(in srgb, var(--link-color) 70%, transparent);
    transition: width 0.4s ease;
  }

  .suggestions-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .suggestion {
    border-radius: 10px;
    background: var(--carbon-graphene, #12121a);
    border: 1px solid var(--carbon-border, #2d2d3d);
    transition: all var(--carbon-transition, 0.25s ease);
  }

  .suggestion:hover {
    border-color: var(--carbon-glow-purple, #8b5cf6);
    box-shadow: 0 0 14px rgba(139, 92, 246, 0.2);
  }

  .suggestion-link {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.75rem 1rem;
    text-decoration: none;
    color: var(--carbon-text, #f0f0f5);
  }

  .suggestion-title {
    font-weight: 500;
  }

  .score {
    flex-shrink: 0;
    font-size: 0.8rem;
    color: var(--carbon-glow-amber, #f59e0b);
    font-weight: 600;
  }

  @media (max-width: 640px) {
    .note-page {
      padding: 1rem 0.75rem 3rem;
    }

    .note-hero {
      padding: 1.25rem;
    }

    .hero-top {
      flex-direction: column;
      align-items: flex-start;
    }

    .actions {
      width: 100%;
    }

    .note-content,
    .meta-section,
    .links-section,
    .suggestions-section {
      padding-left: 1.25rem;
      padding-right: 1.25rem;
    }

    .links-list {
      grid-template-columns: 1fr;
    }
  }
</style>
