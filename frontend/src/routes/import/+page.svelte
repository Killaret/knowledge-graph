<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { isAuthenticated, initAuth } from "$shared/stores/auth.svelte.js";
  import { createBookmarkletNote, type BookmarkletResult } from "$shared/api/import";

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
      void goto(`/auth/login?redirect=${redirect}`);
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
  <div class="tabs">
    <a class="active" href="/import">{t("import.singleImport")}</a>
    <a href="/import/bookmarks">{t("import.massImport")}</a>
  </div>

  <h1>{t("controls.import")}</h1>

  {#if status === "loading"}
    <p class="status loading">{t("import.creating")}</p>
  {:else if status === "success" && result}
    <p class="status success">
      {t("import.success", { title: result.title })}
    </p>
    <div class="actions">
      <button type="button" onclick={openGraph}>{t("import.openInGraph")}</button>
      <button type="button" class="secondary" onclick={openNote}>
        {t("note.openNote")}
      </button>
    </div>
  {:else if status === "error"}
    <p class="status error">{errorMessage || t("import.error")}</p>
  {:else if status === "unauthorized"}
    <p class="status error">{t("unauthorized")}</p>
  {/if}
</div>

<style>
  .import-page {
    max-width: 600px;
    margin: 2rem auto;
    padding: 1rem;
  }
  .tabs {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
    border-bottom: 1px solid #ddd;
  }
  .tabs a {
    padding: 0.5rem 1rem;
    text-decoration: none;
    color: #333;
    border-bottom: 2px solid transparent;
  }
  .tabs a.active {
    border-bottom-color: #007acc;
    font-weight: bold;
  }
  h1 {
    margin-bottom: 1rem;
  }
  .status {
    padding: 1rem;
    border-radius: 0.5rem;
    margin-bottom: 1rem;
  }
  .loading {
    background: #f0f4f8;
    color: #333;
  }
  .success {
    background: #e6f7e6;
    color: #1a6;
  }
  .error {
    background: #fff0f0;
    color: #c00;
  }
  .actions {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }
  button {
    padding: 0.5rem 1rem;
    border: 1px solid #ccc;
    border-radius: 0.25rem;
    cursor: pointer;
    background: #007acc;
    color: #fff;
  }
  button.secondary {
    background: #f0f0f0;
    color: #333;
  }
</style>
