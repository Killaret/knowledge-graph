<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { isAuthenticated, initAuth } from "$shared/stores/auth.svelte.js";
  import { useRequireAuth } from "$shared/composables/auth";
  import { createBookmarkletNote, type BookmarkletResult } from "$shared/api/import";
  import Button from "$components/atoms/Button.svelte";
  import StateIllustration from "$components/atoms/StateIllustration.svelte";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  let status = $state<"idle" | "loading" | "success" | "error" | "unauthorized">("idle");
  let result = $state<BookmarkletResult | null>(null);
  let errorMessage = $state("");

  onMount(async () => {
    await initAuth();

    if (!isAuthenticated()) {
      const redirect = encodeURIComponent(window.location.pathname + window.location.search);
      useRequireAuth(`/auth/login?redirect=${redirect}`);
      status = "unauthorized";
      return;
    }

    const params = $page.url.searchParams;
    const title = params.get("title") || "";
    const url = params.get("url") || "";
    const text = params.get("text") || "";

    if (!title || !url) {
      status = "error";
      errorMessage = t("import.noParams");
      return;
    }

    status = "loading";
    try {
      const note = await createBookmarkletNote({ title, url, text });
      result = note;
      status = "success";
    } catch (e) {
      status = "error";
      errorMessage = t("import.error");
      if (import.meta.env.DEV) {
        console.error("Bookmarklet import error:", e);
      }
    }
  });

  function openGraph() {
    void goto("/");
  }

  function openNote() {
    if (result?.note_id) {
      void goto(`/notes/${result.note_id}`);
    }
  }
</script>

<div class="import-page">
  <nav class="tabs" aria-label={t("import.massTitle")}>
    <a class="tab active" href="/import">{t("import.singleImport")}</a>
    <a class="tab" href="/import/bookmarks">{t("import.massImport")}</a>
  </nav>

  <h1 class="import-title">{t("controls.import")}</h1>

  <div class="import-card">
    {#if status === "loading"}
      <div class="state">
        <div class="spinner" aria-hidden="true"></div>
        <p class="status loading">{t("import.creating")}</p>
      </div>
    {:else if status === "success" && result}
      <div class="state">
        <StateIllustration type="success" />
        <p class="status success">{t("import.success", { title: result.title })}</p>
      </div>
      <div class="actions">
        <Button onClick={openGraph} data-testid="import-open-graph">
          {t("import.openInGraph")}
        </Button>
        <Button variant="secondary" onClick={openNote} data-testid="import-open-note">
          {t("note.openNote")}
        </Button>
      </div>
    {:else if status === "error"}
      <div class="state">
        <StateIllustration type="error" />
        <p class="status error">{errorMessage || t("import.error")}</p>
      </div>
    {:else if status === "unauthorized"}
      <div class="state">
        <StateIllustration type="empty" />
        <p class="status error">{t("unauthorized")}</p>
      </div>
    {:else}
      <div class="state">
        <StateIllustration type="empty" />
        <p class="status idle">{t("import.noParams")}</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .import-page {
    max-width: 720px;
    margin: 0 auto;
    padding: 2rem 1rem 4rem;
    min-height: 100vh;
    color: var(--carbon-text, #f0f0f5);
  }

  .tabs {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
    border-bottom: 1px solid var(--carbon-border, #2d2d3d);
  }

  .tab {
    padding: 0.6rem 1rem;
    text-decoration: none;
    color: var(--carbon-text-muted, #8b8b9e);
    border-bottom: 2px solid transparent;
    transition: all 0.2s ease;
    font-weight: 500;
  }

  .tab:hover {
    color: var(--carbon-text, #f0f0f5);
  }

  .tab.active {
    color: var(--carbon-glow-cyan, #22d3ee);
    border-bottom-color: var(--carbon-glow-cyan, #22d3ee);
    font-weight: 600;
  }

  .import-title {
    margin: 0 0 1.25rem;
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--carbon-text, #f0f0f5);
  }

  .import-card {
    background: var(
      --carbon-gradient-card,
      linear-gradient(145deg, rgba(30, 30, 42, 0.7) 0%, rgba(18, 18, 26, 0.9) 100%)
    );
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 14px;
    padding: 2rem;
    box-shadow:
      0 8px 32px rgba(0, 0, 0, 0.45),
      0 0 12px rgba(139, 92, 246, 0.06);
  }

  .state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    text-align: center;
    margin-bottom: 1.5rem;
  }

  .status {
    margin: 0;
    padding: 0.875rem 1.25rem;
    border-radius: 10px;
    font-weight: 500;
  }

  .status.loading {
    color: var(--carbon-glow-cyan, #22d3ee);
    background: rgba(34, 211, 238, 0.08);
    border: 1px solid rgba(34, 211, 238, 0.2);
  }

  .status.success {
    color: var(--carbon-glow-emerald, #34d399);
    background: rgba(52, 211, 153, 0.08);
    border: 1px solid rgba(52, 211, 153, 0.2);
  }

  .status.error {
    color: var(--carbon-glow-red, #ff3a2f);
    background: rgba(255, 58, 47, 0.08);
    border: 1px solid rgba(255, 58, 47, 0.2);
  }

  .status.idle {
    color: var(--carbon-text-muted, #8b8b9e);
    background: var(--carbon-graphene, #12121a);
    border: 1px solid var(--carbon-border, #2d2d3d);
  }

  .spinner {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    border: 3px solid var(--carbon-border, #2d2d3d);
    border-top-color: var(--carbon-glow-cyan, #22d3ee);
    box-shadow: 0 0 14px var(--carbon-glow-cyan, rgba(34, 211, 238, 0.25));
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .actions {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
    justify-content: center;
  }

  @media (max-width: 480px) {
    .import-card {
      padding: 1.25rem;
    }

    .actions {
      flex-direction: column;
      align-items: stretch;
    }
  }
</style>
