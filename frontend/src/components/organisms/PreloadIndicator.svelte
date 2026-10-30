<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { browser } from "$app/environment";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import {
    isPreloadingData,
    hasPreloadedData,
    getPreloadedGraph,
  } from "$shared/services/PreloadService";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  let message = $state("");
  let intervalId: ReturnType<typeof setInterval> | null = null;

  function update() {
    if (isPreloadingData()) {
      message = t("preload.loading");
      return;
    }

    if (hasPreloadedData()) {
      const graph = getPreloadedGraph();
      const count = graph?.nodes.length ?? 0;
      message = t("preload.ready", { count });
      return;
    }

    message = "";
  }

  onMount(() => {
    if (!browser) return;

    update();
    intervalId = setInterval(update, 500);
  });

  onDestroy(() => {
    if (intervalId) clearInterval(intervalId);
  });
</script>

{#if message}
  <div class="preload-indicator" role="status" aria-live="polite">
    {message}
  </div>
{/if}

<style>
  .preload-indicator {
    margin-top: 1rem;
    padding: 0.75rem 1rem;
    text-align: center;
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.7);
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 8px;
    backdrop-filter: blur(4px);
  }
</style>
