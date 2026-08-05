<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { isAuthenticated, initAuth } from "$shared/stores/auth.svelte.js";
  import {
    previewBookmarks,
    createBookmarksImport,
    getImportStatus,
    type ImportItem,
    type ImportPreviewItem,
    type ImportTaskStatus,
  } from "$shared/api/import";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const noteTypes = [
    "star",
    "planet",
    "comet",
    "galaxy",
    "asteroid",
    "satellite",
    "debris",
    "nebula",
    "blackhole",
    "moon",
    "technical",
    "unknown",
    "reality_rift",
    "chromatic_maw",
    "void_whisper",
    "cosmic_abomination",
  ];

  let status = $state<"idle" | "loading" | "preview" | "importing" | "done" | "error" | "unauthorized">("idle");
  let input = $state("");
  let dragOver = $state(false);
  let previewItems = $state<ImportPreviewItem[]>([]);
  let taskId = $state<string | null>(null);
  let taskStatus = $state<ImportTaskStatus | null>(null);
  let errorMessage = $state("");
  let pollInterval = $state<ReturnType<typeof setInterval> | null>(null);

  onMount(async () => {
    await initAuth();

    if (!isAuthenticated()) {
      const redirect = encodeURIComponent(window.location.pathname + window.location.search);
      void goto(`/auth/login?redirect=${redirect}`);
      status = "unauthorized";
      return;
    }

    status = "idle";
  });

  function parseLine(line: string): ImportItem | null {
    line = line.trim();
    if (!line || line.startsWith("#")) return null;

    const sepIdx = line.indexOf("|");
    if (sepIdx > 0) {
      const title = line.slice(0, sepIdx).trim();
      const url = line.slice(sepIdx + 1).trim();
      return { title, url };
    }

    return { title: line, url: line };
  }

  function parseInput(): ImportItem[] {
    const items: ImportItem[] = [];
    const seen = new Set<string>();
    for (const line of input.split("\n")) {
      const it = parseLine(line);
      if (!it) continue;
      if (seen.has(it.url)) continue;
      seen.add(it.url);
      items.push(it);
    }
    return items;
  }

  function parseBookmarksHTML(html: string): ImportItem[] {
    const doc = new DOMParser().parseFromString(html, "text/html");
    const links = doc.querySelectorAll("a[href]");
    const items: ImportItem[] = [];
    const seen = new Set<string>();
    links.forEach((a) => {
      const url = a.getAttribute("href");
      const title = a.textContent?.trim() || url || "";
      if (!url) return;
      if (seen.has(url)) return;
      seen.add(url);
      items.push({ title, url });
    });
    return items;
  }

  async function buildPreview() {
    const items = parseInput();
    if (items.length === 0) {
      errorMessage = t("import.emptyList");
      status = "error";
      return;
    }

    status = "loading";
    try {
      const res = await previewBookmarks(items);
      previewItems = res.items;
      status = "preview";
    } catch (e) {
      status = "error";
      errorMessage = t("import.error");
      if (import.meta.env.DEV) console.error("Preview error:", e);
    }
  }

  function removeItem(index: number) {
    previewItems = previewItems.filter((_, i) => i !== index);
  }

  function updateItemType(index: number, type: string) {
    previewItems = previewItems.map((it, i) => (i === index ? { ...it, type } : it));
  }

  function updateItemTitle(index: number, title: string) {
    previewItems = previewItems.map((it, i) => (i === index ? { ...it, title } : it));
  }

  async function startImport() {
    const items: ImportItem[] = previewItems
      .filter((it) => it.is_new && !it.error)
      .map(({ title, url, text, type }) => ({
        title,
        url,
        text,
        type: type || "asteroid",
      }));

    if (items.length === 0) {
      errorMessage = t("import.noNewItems");
      status = "error";
      return;
    }

    status = "importing";
    try {
      const res = await createBookmarksImport(items);
      taskId = res.task_id;
      pollStatus();
    } catch (e) {
      status = "error";
      errorMessage = t("import.error");
      if (import.meta.env.DEV) console.error("Import error:", e);
    }
  }

  function pollStatus() {
    if (!taskId) return;
    if (pollInterval) clearInterval(pollInterval);

    const load = async () => {
      try {
        taskStatus = await getImportStatus(taskId!);
        if (taskStatus?.status === "done" || taskStatus?.status === "failed") {
          status = "done";
          if (pollInterval) clearInterval(pollInterval);
        }
      } catch (e) {
        if (import.meta.env.DEV) console.error("Status poll error:", e);
      }
    };

    load();
    pollInterval = setInterval(load, 2000);
  }

  function handleFileDrop(event: DragEvent) {
    event.preventDefault();
    dragOver = false;
    const file = event.dataTransfer?.files[0];
    if (file) readFile(file);
  }

  function handleFileSelect(event: Event) {
    const target = event.target as HTMLInputElement;
    const file = target.files?.[0];
    if (file) readFile(file);
  }

  function readFile(file: File) {
    const reader = new FileReader();
    reader.onload = () => {
      const html = reader.result as string;
      const items = parseBookmarksHTML(html);
      if (items.length === 0) {
        errorMessage = t("import.emptyList");
        status = "error";
        return;
      }
      input = items.map((it) => `${it.title} | ${it.url}`).join("\n");
      buildPreview();
    };
    reader.onerror = () => {
      errorMessage = t("import.error");
      status = "error";
    };
    reader.readAsText(file);
  }

  function openGraph() {
    void goto("/");
  }

  function reset() {
    status = "idle";
    input = "";
    previewItems = [];
    taskId = null;
    taskStatus = null;
    errorMessage = "";
    if (pollInterval) clearInterval(pollInterval);
  }
</script>

<div class="import-page">
  <div class="tabs">
    <a href="/import">{t("import.singleImport")}</a>
    <a class="active" href="/import/bookmarks">{t("import.massImport")}</a>
  </div>

  <h1>{t("import.massTitle")}</h1>

  {#if status === "unauthorized"}
    <p class="status error">{t("unauthorized")}</p>
  {:else if status === "idle" || status === "error"}
    {#if status === "error"}
      <p class="status error">{errorMessage || t("import.error")}</p>
    {/if}

    <label for="import-list">{t("import.pasteList")}</label>
    <textarea
      id="import-list"
      bind:value={input}
      rows="10"
      placeholder="Example page | https://example.com\nhttps://another.example.com"
    ></textarea>

    <div
      class="drop-zone"
      class:drag-over={dragOver}
      role="button"
      tabindex="0"
      ondragover={(e) => { e.preventDefault(); dragOver = true; }}
      ondragleave={() => (dragOver = false)}
      ondrop={handleFileDrop}
      onkeydown={(e) => e.key === "Enter" && document.getElementById("file-input")?.click()}
      onclick={() => document.getElementById("file-input")?.click()}
    >
      {t("import.dropFile")}
    </div>
    <input
      id="file-input"
      type="file"
      accept=".html,.htm,text/html"
      style="display: none"
      onchange={handleFileSelect}
    />

    <div class="actions">
      <button type="button" onclick={buildPreview}>{t("import.preview")}</button>
    </div>
  {:else if status === "loading"}
    <p class="status loading">{t("import.processing")}</p>
  {:else if status === "preview"}
    <h2>{t("import.preview")}</h2>
    {#if previewItems.length === 0}
      <p>{t("import.emptyList")}</p>
    {:else}
      <table class="preview-table">
        <thead>
          <tr>
            <th>Title</th>
            <th>URL</th>
            <th>Type</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each previewItems as item, index (item.url + index)}
            <tr class:duplicate={!item.is_new} class:has-error={!!item.error}>
              <td>
                <input
                  type="text"
                  value={item.title}
                  oninput={(e) => updateItemTitle(index, e.currentTarget.value)}
                  disabled={!item.is_new || !!item.error}
                />
              </td>
              <td class="url-cell">{item.url}</td>
              <td>
                <select
                  value={item.type || "asteroid"}
                  onchange={(e) => updateItemType(index, e.currentTarget.value)}
                  disabled={!item.is_new || !!item.error}
                >
                  {#each noteTypes as nt}
                    <option value={nt}>{t(`filter.type.${nt}`)}</option>
                  {/each}
                </select>
              </td>
              <td>
                {#if item.error}
                  <span class="error-badge">{item.error}</span>
                {:else if item.is_new}
                  <span class="new-badge">New</span>
                {:else}
                  <span class="duplicate-badge">Duplicate</span>
                {/if}
              </td>
              <td>
                <button type="button" class="icon" onclick={() => removeItem(index)}>×</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>

      <div class="actions">
        <button type="button" onclick={startImport}>
          {t("import.importAll", { count: previewItems.filter((i) => i.is_new && !i.error).length })}
        </button>
        <button type="button" class="secondary" onclick={reset}>
          {t("import.noParams")}
        </button>
      </div>
    {/if}
  {:else if status === "importing" || status === "done"}
    <h2>{t("import.processing")}</h2>
    {#if taskId}
      <p>{t("import.taskAccepted", { task_id: taskId })}</p>
    {/if}
    {#if taskStatus}
      <div class="progress">
        <div class="progress-row">
          <span>{t("import.total")}:</span>
          <strong>{taskStatus.progress.total}</strong>
        </div>
        <div class="progress-row">
          <span>{t("import.created")}:</span>
          <strong>{taskStatus.progress.created}</strong>
        </div>
        <div class="progress-row">
          <span>{t("import.skipped")}:</span>
          <strong>{taskStatus.progress.skipped}</strong>
        </div>
        <div class="progress-row">
          <span>{t("import.failed")}:</span>
          <strong>{taskStatus.progress.failed}</strong>
        </div>
      </div>
    {/if}
    {#if status === "done"}
      <div class="actions">
        <button type="button" onclick={openGraph}>{t("import.openInGraph")}</button>
        <button type="button" class="secondary" onclick={reset}>
          {t("import.massTitle")}
        </button>
      </div>
    {/if}
  {/if}
</div>

<style>
  .import-page {
    max-width: 900px;
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
  label {
    display: block;
    margin-bottom: 0.5rem;
  }
  textarea {
    width: 100%;
    padding: 0.5rem;
    border: 1px solid #ccc;
    border-radius: 0.25rem;
    font-family: inherit;
    box-sizing: border-box;
  }
  .drop-zone {
    margin-top: 1rem;
    padding: 2rem;
    border: 2px dashed #ccc;
    border-radius: 0.5rem;
    text-align: center;
    color: #666;
    cursor: pointer;
  }
  .drop-zone.drag-over {
    border-color: #007acc;
    background: #f0f4f8;
  }
  .actions {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin-top: 1rem;
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
  button.icon {
    background: transparent;
    color: #c00;
    border: none;
    font-size: 1.25rem;
    line-height: 1;
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
  .error {
    background: #fff0f0;
    color: #c00;
  }
  .preview-table {
    width: 100%;
    border-collapse: collapse;
    margin-top: 1rem;
  }
  .preview-table th,
  .preview-table td {
    border: 1px solid #ddd;
    padding: 0.5rem;
    text-align: left;
    vertical-align: top;
  }
  .preview-table input,
  .preview-table select {
    width: 100%;
    box-sizing: border-box;
  }
  .url-cell {
    max-width: 250px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .new-badge {
    color: #1a6;
  }
  .duplicate-badge {
    color: #888;
  }
  .error-badge {
    color: #c00;
  }
  .duplicate {
    background: #f9f9f9;
  }
  .has-error {
    background: #fff0f0;
  }
  .progress {
    display: grid;
    grid-template-columns: repeat(4, auto);
    gap: 1rem;
    margin: 1rem 0;
  }
  .progress-row {
    display: flex;
    gap: 0.25rem;
  }
</style>
