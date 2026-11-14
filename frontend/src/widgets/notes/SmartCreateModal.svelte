<script lang="ts">
  /* eslint-disable prefer-const -- Svelte 5 bindable props require let */
  import { goto } from "$app/navigation";
  import { HTTPError } from "ky";
  import Modal from "$components/atoms/Modal.svelte";
  import Button from "$components/atoms/Button.svelte";
  import {
    createSmartImport,
    getImportStatus,
    type ImportTaskStatus,
    type SmartImportType,
  } from "$shared/api/import";
  import { getApiErrorMessage } from "$shared/api/errorMessage";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  // ru has three plural forms (1 фрагмент, 2–4 фрагмента, 5+ фрагментов),
  // en has two — the locale picks the key, the dictionaries carry all three.
  function rawChunksKey(count: number): string {
    if (locale === "ru") {
      const mod10 = count % 10;
      const mod100 = count % 100;
      if (mod10 === 1 && mod100 !== 11) return "import.smartRawChunksOne";
      if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) {
        return "import.smartRawChunksFew";
      }
      return "import.smartRawChunks";
    }
    return count === 1 ? "import.smartRawChunksOne" : "import.smartRawChunks";
  }

  async function submitErrorMessage(e: unknown): Promise<string> {
    // A 503 on this endpoint means the document handler went offline between
    // the availability poll and the submit — show the dedicated text, not a
    // raw ky error string.
    if (e instanceof HTTPError && e.response.status === 503) {
      return t("import.smartCreateUnavailable");
    }
    return getApiErrorMessage(e, locale);
  }

  interface Props {
    open: boolean;
    /** A task is queued and the window is collapsed — the bar shows a spinner. */
    pending?: boolean;
    /** A result arrived while collapsed — the bar shows a "1" badge. */
    unread?: boolean;
    onClose?: () => void;
  }

  let {
    open = $bindable(false),
    pending = $bindable(false),
    unread = $bindable(false),
    onClose,
  }: Props = $props();

  type Source = SmartImportType;

  const tiles: { id: Source; labelKey: string; testid: string }[] = [
    { id: "TEXT", labelKey: "import.smartTileText", testid: "smart-create-tile-text" },
    { id: "URL", labelKey: "import.smartTileUrl", testid: "smart-create-tile-url" },
    { id: "FILE", labelKey: "import.smartTileFile", testid: "smart-create-tile-file" },
  ];

  let source = $state<Source>("TEXT");
  let text = $state("");
  let url = $state("");
  let file = $state<File | null>(null);
  let submitting = $state(false);
  let submitError = $state("");
  let taskId = $state<string | null>(null);
  let result = $state<ImportTaskStatus | null>(null);

  const isTerminal = $derived(result?.status === "done" || result?.status === "failed");
  const waiting = $derived(taskId !== null && !isTerminal);

  // The task keeps polling while the window is collapsed; a result that lands
  // unseen flips the bar badge to "1".
  $effect(() => {
    const id = taskId;
    if (!id || isTerminal) return;
    let cancelled = false;
    const poll = async () => {
      try {
        const status = await getImportStatus(id);
        if (!cancelled) result = status;
      } catch {
        // A 404 mid-flight is transient; the task lives an hour in cache.
      }
    };
    void poll();
    const timer = setInterval(poll, 2000);
    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  });

  $effect(() => {
    if (taskId && isTerminal && !open) {
      pending = false;
      unread = true;
    }
    if (open) {
      unread = false;
      if (isTerminal) pending = false;
    }
  });

  function collapse() {
    open = false;
    pending = waiting;
    onClose?.();
  }

  function reset() {
    text = "";
    url = "";
    file = null;
    submitting = false;
    submitError = "";
    taskId = null;
    result = null;
    pending = false;
    unread = false;
  }

  async function submit() {
    if (submitting) return;
    submitError = "";
    submitting = true;
    try {
      if (source === "TEXT") {
        if (!text.trim()) {
          submitError = t("import.smartEmptyText");
          return;
        }
        const res = await createSmartImport({ type: "TEXT", content: text });
        taskId = res.task_id;
      } else if (source === "URL") {
        if (!url.trim()) {
          submitError = t("import.smartEmptyUrl");
          return;
        }
        const res = await createSmartImport({ type: "URL", content: url.trim() });
        taskId = res.task_id;
      } else {
        if (!file) {
          submitError = t("import.smartChooseFile");
          return;
        }
        const res = await createSmartImport({ type: "FILE", file });
        taskId = res.task_id;
      }
    } catch (e) {
      submitError = await submitErrorMessage(e);
    } finally {
      submitting = false;
    }
  }

  function handleFileChange(e: Event) {
    const input = e.target as HTMLInputElement;
    file = input.files?.[0] ?? null;
  }

  function openGraph() {
    open = false;
    void goto("/");
  }
</script>

<Modal bind:open title={t("import.smartTitle")} onClose={collapse}>
  <div class="smart-create" data-testid="smart-create-modal">
    {#if !taskId}
      <p class="smart-subtitle">{t("import.smartSubtitle")}</p>

      <div class="smart-tiles" role="group" aria-label={t("import.smartTitle")}>
        {#each tiles as tile}
          <button
            type="button"
            class="smart-source-tile"
            class:active={source === tile.id}
            aria-pressed={source === tile.id}
            onclick={() => (source = tile.id)}
            data-testid={tile.testid}
          >
            {t(tile.labelKey)}
          </button>
        {/each}
      </div>

      {#if source === "TEXT"}
        <textarea
          class="smart-input"
          bind:value={text}
          placeholder={t("import.smartTextPlaceholder")}
          rows="8"
          data-testid="smart-create-text"></textarea>
      {:else if source === "URL"}
        <input
          class="smart-input"
          type="url"
          bind:value={url}
          placeholder={t("import.smartUrlPlaceholder")}
          data-testid="smart-create-url"
        />
      {:else}
        <input
          class="smart-input"
          type="file"
          onchange={handleFileChange}
          data-testid="smart-create-file"
        />
      {/if}

      {#if submitError}
        <p class="smart-error" data-testid="smart-create-submit-error">{submitError}</p>
      {/if}

      <div class="smart-actions">
        <Button onClick={submit} disabled={submitting} data-testid="smart-create-submit">
          {t("import.smartSubmit")}
        </Button>
      </div>
    {:else if waiting}
      <div class="smart-state" data-testid="smart-create-pending">
        <div class="smart-spinner" aria-hidden="true"></div>
        <p>{t("import.smartQueued")}</p>
        <Button variant="secondary" onClick={collapse} data-testid="smart-create-collapse">
          {t("import.smartCollapse")}
        </Button>
      </div>
    {:else if result?.status === "done"}
      <div class="smart-state" data-testid="smart-create-done">
        <p class="smart-done-title">
          {t("import.smartDone", { count: result.notes?.length ?? 0 })}
        </p>
        {#if result.notes && result.notes.length > 0}
          <ul class="smart-notes" data-testid="smart-create-notes">
            {#each result.notes as note}
              <li>{note.title || note.id}</li>
            {/each}
          </ul>
        {/if}
        {#if result.raw_chunks_count}
          <p class="smart-raw" data-testid="smart-create-raw">
            {t(rawChunksKey(result.raw_chunks_count), { count: result.raw_chunks_count })}
          </p>
        {/if}
        <div class="smart-actions">
          <Button onClick={openGraph} data-testid="smart-create-open-graph">
            {t("import.smartOpenInGraph")}
          </Button>
          <Button variant="secondary" onClick={reset} data-testid="smart-create-reset">
            {t("import.smartTitle")}
          </Button>
        </div>
      </div>
    {:else}
      <div class="smart-state" data-testid="smart-create-failed">
        <p class="smart-error">
          {#if result?.reason === "timeout"}
            {t("import.smartFailedTimeout")}
          {:else}
            {result?.errors?.join("; ") || result?.status}
          {/if}
        </p>
        <Button variant="secondary" onClick={reset} data-testid="smart-create-reset">
          {t("import.smartTitle")}
        </Button>
      </div>
    {/if}
  </div>
</Modal>

<style>
  .smart-create {
    display: flex;
    flex-direction: column;
    gap: 14px;
    min-width: 340px;
  }

  .smart-subtitle {
    margin: 0;
    color: var(--carbon-text-muted, #9aa3b2);
    font-size: 13px;
    line-height: 1.45;
  }

  .smart-tiles {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
  }

  .smart-source-tile {
    padding: 10px 8px;
    border: 1px solid rgba(139, 92, 246, 0.45);
    border-radius: 10px;
    background: rgba(139, 92, 246, 0.06);
    color: var(--carbon-text, #e6e9f0);
    font-size: 12px;
    line-height: 1.3;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .smart-source-tile:hover {
    border-color: rgba(139, 92, 246, 0.7);
    background: rgba(139, 92, 246, 0.14);
  }

  .smart-source-tile.active {
    border-color: #2dd4bf;
    background: rgba(45, 212, 191, 0.1);
  }

  .smart-input {
    width: 100%;
    box-sizing: border-box;
    padding: 10px 12px;
    background: var(--carbon-black, #0d0f14);
    border: 1px solid var(--carbon-border, #2a2f3a);
    border-radius: 10px;
    color: var(--carbon-text, #e6e9f0);
    font-size: 13px;
    font-family: inherit;
    resize: vertical;
  }

  .smart-input:focus {
    outline: none;
    border-color: rgba(34, 211, 238, 0.6);
  }

  .smart-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    padding: 12px 0;
    text-align: center;
  }

  .smart-spinner {
    width: 22px;
    height: 22px;
    border: 3px solid rgba(139, 92, 246, 0.25);
    border-top-color: #2dd4bf;
    border-radius: 50%;
    animation: smart-spin 0.9s linear infinite;
  }

  @keyframes smart-spin {
    to {
      transform: rotate(360deg);
    }
  }

  .smart-done-title {
    margin: 0;
    font-size: 15px;
    font-weight: 600;
  }

  .smart-notes {
    margin: 0;
    padding: 0 0 0 18px;
    text-align: left;
    max-height: 200px;
    overflow-y: auto;
    align-self: stretch;
  }

  .smart-raw {
    margin: 0;
    color: var(--carbon-text-muted, #9aa3b2);
    font-size: 12px;
  }

  .smart-error {
    margin: 0;
    color: #f87171;
    font-size: 13px;
  }

  .smart-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }
</style>
