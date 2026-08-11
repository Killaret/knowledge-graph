<script lang="ts">
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { isAuthenticated, initAuth } from "$shared/stores/auth.svelte.js";
  import { useRequireAuth } from "$shared/composables/auth";
  import {
    previewBookmarks,
    createBookmarksImport,
    getImportStatus,
    type ImportItem,
    type ImportPreviewItem,
    type ImportTaskStatus,
  } from "$shared/api/import";
  import Button from "$components/atoms/Button.svelte";
  import StateIllustration from "$components/atoms/StateIllustration.svelte";

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

  let status = $state<
    "idle" | "loading" | "preview" | "importing" | "done" | "error" | "unauthorized"
  >("idle");
  let input = $state("");
  let extractContent = $state(false);
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
      useRequireAuth(`/auth/login?redirect=${redirect}`);
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
      const res = await previewBookmarks(items, { extract_content: extractContent });
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

  const progressPercent = $derived(
    taskStatus?.progress && taskStatus.progress.total > 0
      ? Math.min(100, Math.round((taskStatus.progress.processed / taskStatus.progress.total) * 100))
      : 0
  );

  const importableCount = $derived(previewItems.filter((i) => i.is_new && !i.error).length);
</script>

<div class="import-page">
  <nav class="tabs" aria-label={t("import.massTitle")}>
    <a class="tab" href="/import">{t("import.singleImport")}</a>
    <a class="tab active" href="/import/bookmarks">{t("import.massImport")}</a>
  </nav>

  <h1 class="import-title">{t("import.massTitle")}</h1>

  {#if status === "unauthorized"}
    <div class="state-card">
      <StateIllustration type="empty" />
      <p class="status error">{t("unauthorized")}</p>
    </div>
  {:else if status === "idle" || status === "error"}
    <div class="import-card">
      {#if status === "error"}
        <div class="state-header">
          <StateIllustration type="error" />
          <p class="status error">{errorMessage || t("import.error")}</p>
        </div>
      {/if}

      <label for="import-list" class="field-label">{t("import.pasteList")}</label>
      <textarea
        id="import-list"
        bind:value={input}
        rows="10"
        placeholder="Example page | https://example.com\nhttps://another.example.com"
      ></textarea>

      <label class="extract-toggle">
        <input type="checkbox" bind:checked={extractContent} />
        <span>{t("import.extractContent")}</span>
      </label>

      <div
        class="drop-zone"
        class:drag-over={dragOver}
        role="button"
        tabindex="0"
        ondragover={(e) => {
          e.preventDefault();
          dragOver = true;
        }}
        ondragleave={() => (dragOver = false)}
        ondrop={handleFileDrop}
        onkeydown={(e) => e.code === "Enter" && document.getElementById("file-input")?.click()}
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
        <Button onClick={buildPreview} data-testid="preview-import">
          {t("import.preview")}
        </Button>
      </div>
    </div>
  {:else if status === "loading"}
    <div class="state-card">
      <div class="spinner" aria-hidden="true"></div>
      <p class="status loading">{t("import.processing")}</p>
    </div>
  {:else if status === "preview"}
    <div class="import-card">
      <h2 class="section-title">{t("import.preview")}</h2>
      {#if previewItems.length === 0}
        <p class="empty-hint">{t("import.emptyList")}</p>
      {:else}
        <div class="table-wrap">
          <table class="preview-table">
            <thead>
              <tr>
                <th>{t("note.titleLabel")}</th>
                <th>URL</th>
                <th>{t("note.typeLabel")}</th>
                <th>{t("import.status")}</th>
                <th aria-label={t("cockpit.panel.close")}></th>
              </tr>
            </thead>
            <tbody>
              {#each previewItems as item, index (item.url + index)}
                <tr class:duplicate={!item.is_new} class:has-error={!!item.error}>
                  <td>
                    <input
                      type="text"
                      class="table-input"
                      value={item.title}
                      oninput={(e) => updateItemTitle(index, e.currentTarget.value)}
                      disabled={!item.is_new || !!item.error}
                    />
                  </td>
                  <td class="url-cell">{item.url}</td>
                  <td>
                    <select
                      class="table-select"
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
                      <span class="badge error">{item.error}</span>
                    {:else if item.is_new}
                      <span class="badge new">{t("import.newBadge")}</span>
                    {:else}
                      <span class="badge duplicate">{t("import.duplicateBadge")}</span>
                    {/if}
                  </td>
                  <td>
                    <button
                      type="button"
                      class="remove-btn"
                      onclick={() => removeItem(index)}
                      aria-label={t("close")}
                    >
                      ×
                    </button>
                  </td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>

        <div class="actions">
          <Button onClick={startImport} disabled={importableCount === 0} data-testid="start-import">
            {t("import.importAll", { count: importableCount })}
          </Button>
          <Button variant="secondary" onClick={reset} data-testid="reset-import">
            {t("cancel")}
          </Button>
        </div>
      {/if}
    </div>
  {:else if status === "importing" || status === "done"}
    <div class="state-card">
      {#if status === "done"}
        <StateIllustration type={taskStatus?.status === "failed" ? "error" : "success"} />
      {:else}
        <div class="spinner" aria-hidden="true"></div>
      {/if}

      <h2 class="section-title" class:done-title={status === "done"}>
        {status === "done"
          ? taskStatus?.status === "failed"
            ? t("import.failed")
            : t("import.doneTitle")
          : t("import.processing")}
      </h2>

      {#if taskId}
        <p class="task-id">{t("import.taskAccepted", { task_id: taskId })}</p>
      {/if}

      {#if taskStatus}
        <div class="progress-section progress">
          <div class="progress-bar-bg">
            <div
              class="progress-bar-fill"
              style="width: {progressPercent}%"
              class:fill-error={taskStatus.status === "failed"}
            ></div>
          </div>
          <div class="progress-value" class:error={taskStatus.status === "failed"}>
            {progressPercent}%
          </div>
          <div class="progress-counters">
            <span>{t("import.total")}: <strong>{taskStatus.progress.total}</strong></span>
            <span>{t("import.created")}: <strong>{taskStatus.progress.created}</strong></span>
            <span>{t("import.skipped")}: <strong>{taskStatus.progress.skipped}</strong></span>
            <span>{t("import.failed")}: <strong>{taskStatus.progress.failed}</strong></span>
          </div>
        </div>
      {/if}

      {#if status === "done"}
        <div class="actions">
          <Button onClick={openGraph} data-testid="import-open-graph">
            {t("import.openInGraph")}
          </Button>
          <Button variant="secondary" onClick={reset} data-testid="reset-import">
            {t("import.massTitle")}
          </Button>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .import-page {
    max-width: 1100px;
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

  .import-card,
  .state-card {
    background: var(
      --carbon-gradient-card,
      linear-gradient(145deg, rgba(30, 30, 42, 0.7) 0%, rgba(18, 18, 26, 0.9) 100%)
    );
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 14px;
    padding: 1.75rem;
    box-shadow:
      0 8px 32px rgba(0, 0, 0, 0.45),
      0 0 12px rgba(139, 92, 246, 0.06);
  }

  .state-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    text-align: center;
  }

  .state-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
    margin-bottom: 1.25rem;
  }

  .field-label {
    display: block;
    margin-bottom: 0.5rem;
    color: var(--carbon-text-muted, #8b8b9e);
    font-weight: 500;
    font-size: 0.95rem;
  }

  textarea {
    width: 100%;
    padding: 0.875rem 1rem;
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 10px;
    background: var(--carbon-black, #050508);
    color: var(--carbon-text, #f0f0f5);
    font-family: inherit;
    box-sizing: border-box;
    font-size: 0.95rem;
    resize: vertical;
    transition:
      border-color 0.2s,
      box-shadow 0.2s;
  }

  textarea::placeholder {
    color: var(--carbon-text-dim, #5a5a6e);
  }

  textarea:focus {
    outline: none;
    border-color: var(--carbon-glow-cyan, #22d3ee);
    box-shadow: var(--carbon-focus-ring, 0 0 0 3px rgba(34, 211, 238, 0.15));
  }

  .extract-toggle {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 1rem 0 0;
    cursor: pointer;
    color: var(--carbon-text-muted, #8b8b9e);
  }

  .extract-toggle input[type="checkbox"] {
    width: 18px;
    height: 18px;
    accent-color: var(--carbon-glow-cyan, #22d3ee);
    cursor: pointer;
  }

  .drop-zone {
    margin-top: 1.25rem;
    padding: 2.5rem 1.5rem;
    border: 2px dashed var(--carbon-border, #2d2d3d);
    border-radius: 12px;
    text-align: center;
    color: var(--carbon-text-muted, #8b8b9e);
    cursor: pointer;
    background: var(--carbon-graphene, #12121a);
    transition: all 0.2s ease;
  }

  .drop-zone:hover,
  .drop-zone.drag-over {
    border-color: var(--carbon-glow-cyan, #22d3ee);
    background: rgba(34, 211, 238, 0.05);
    color: var(--carbon-glow-cyan, #22d3ee);
  }

  .actions {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
    margin-top: 1.5rem;
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

  .status.error {
    color: var(--carbon-glow-red, #ff3a2f);
    background: rgba(255, 58, 47, 0.08);
    border: 1px solid rgba(255, 58, 47, 0.2);
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

  .section-title {
    margin: 0 0 1rem;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--carbon-text, #f0f0f5);
  }

  .done-title {
    margin-top: 0.5rem;
  }

  .empty-hint {
    color: var(--carbon-text-muted, #8b8b9e);
    margin: 0;
  }

  .table-wrap {
    overflow-x: auto;
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 12px;
    background: var(--carbon-black, #050508);
  }

  .preview-table {
    width: 100%;
    border-collapse: collapse;
    min-width: 640px;
  }

  .preview-table th,
  .preview-table td {
    border-bottom: 1px solid var(--carbon-border, #2d2d3d);
    padding: 0.75rem;
    text-align: left;
    vertical-align: middle;
  }

  .preview-table th {
    color: var(--carbon-text-muted, #8b8b9e);
    font-weight: 600;
    font-size: 0.85rem;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    background: var(--carbon-graphene, #12121a);
  }

  .preview-table tbody tr:last-child td {
    border-bottom: none;
  }

  .preview-table tbody tr.duplicate {
    opacity: 0.65;
  }

  .preview-table tbody tr.has-error {
    background: rgba(255, 58, 47, 0.05);
  }

  .table-input,
  .table-select {
    width: 100%;
    box-sizing: border-box;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 8px;
    background: var(--carbon-black, #050508);
    color: var(--carbon-text, #f0f0f5);
    font-size: 0.9rem;
  }

  .table-input:disabled,
  .table-select:disabled {
    opacity: 0.5;
    background: var(--carbon-graphene, #12121a);
  }

  .table-input:focus,
  .table-select:focus {
    outline: none;
    border-color: var(--carbon-glow-cyan, #22d3ee);
  }

  .url-cell {
    max-width: 280px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    color: var(--carbon-text-dim, #5a5a6e);
    font-size: 0.85rem;
  }

  .badge {
    display: inline-flex;
    padding: 0.25rem 0.6rem;
    border-radius: 6px;
    font-size: 0.8rem;
    font-weight: 500;
  }

  .badge.new {
    color: var(--carbon-glow-emerald, #34d399);
    background: rgba(52, 211, 153, 0.1);
    border: 1px solid rgba(52, 211, 153, 0.25);
  }

  .badge.duplicate {
    color: var(--carbon-text-muted, #8b8b9e);
    background: var(--carbon-graphene, #12121a);
    border: 1px solid var(--carbon-border, #2d2d3d);
  }

  .badge.error {
    color: var(--carbon-glow-red, #ff3a2f);
    background: rgba(255, 58, 47, 0.1);
    border: 1px solid rgba(255, 58, 47, 0.25);
  }

  .remove-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    background: transparent;
    border: 1px solid var(--carbon-border, #2d2d3d);
    color: var(--carbon-text-muted, #8b8b9e);
    border-radius: 6px;
    font-size: 1.25rem;
    line-height: 1;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .remove-btn:hover {
    background: rgba(255, 58, 47, 0.1);
    border-color: var(--carbon-glow-red, #ff3a2f);
    color: var(--carbon-glow-red, #ff3a2f);
  }

  .progress-section {
    width: 100%;
    max-width: 520px;
  }

  .progress-bar-bg {
    height: 12px;
    border-radius: 6px;
    background: var(--carbon-c70, #1a1a24);
    overflow: hidden;
    border: 1px solid var(--carbon-border, #2d2d3d);
  }

  .progress-bar-fill {
    height: 100%;
    background: linear-gradient(
      90deg,
      var(--carbon-glow-cyan, #22d3ee),
      var(--carbon-glow-purple, #8b5cf6)
    );
    box-shadow: 0 0 12px rgba(34, 211, 238, 0.3);
    transition: width 0.4s ease;
  }

  .progress-bar-fill.fill-error {
    background: linear-gradient(90deg, var(--carbon-glow-red, #ff3a2f), #c2410c);
    box-shadow: 0 0 12px rgba(255, 58, 47, 0.3);
  }

  .progress-value {
    margin-top: 0.5rem;
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--carbon-glow-cyan, #22d3ee);
  }

  .progress-value.error {
    color: var(--carbon-glow-red, #ff3a2f);
  }

  .progress-counters {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 0.75rem;
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--carbon-border, #2d2d3d);
    font-size: 0.9rem;
    color: var(--carbon-text-muted, #8b8b9e);
  }

  .progress-counters strong {
    color: var(--carbon-text, #f0f0f5);
    display: block;
    font-size: 1.1rem;
  }

  .task-id {
    margin: 0;
    font-size: 0.85rem;
    color: var(--carbon-text-dim, #5a5a6e);
    word-break: break-all;
  }

  @media (max-width: 720px) {
    .progress-counters {
      grid-template-columns: repeat(2, 1fr);
    }

    .actions {
      flex-direction: column;
      align-items: stretch;
    }
  }

  @media (max-width: 480px) {
    .import-card,
    .state-card {
      padding: 1.25rem;
    }

    .drop-zone {
      padding: 1.5rem 1rem;
    }
  }
</style>
