<script lang="ts">
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  interface Props {
    nodeCount?: number;
    linkCount?: number;
    loading?: boolean;
  }

  const { nodeCount = 0, linkCount = 0, loading = false }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);
</script>

<div class="graph-stats" data-testid="graph-stats">
  {#if loading}
    <span class="graph-stats__spinner" aria-hidden="true"></span>
  {/if}
  <span class="graph-stats__count">
    <strong>{nodeCount}</strong>
    {t("graphOverlay.nodes")}
  </span>
  <span class="graph-stats__dot" aria-hidden="true">·</span>
  <span class="graph-stats__count">
    <strong>{linkCount}</strong>
    {t("graphOverlay.links")}
  </span>
</div>

<style>
  .graph-stats {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    background: rgba(10, 10, 15, 0.75);
    border: 1px solid rgba(45, 212, 191, 0.25);
    border-radius: 8px;
    color: rgba(255, 255, 255, 0.85);
    font-size: 12px;
    backdrop-filter: blur(8px);
    pointer-events: none;
    user-select: none;
    white-space: nowrap;
  }

  .graph-stats__count strong {
    color: #fff;
    font-weight: 600;
  }

  .graph-stats__dot {
    color: rgba(255, 255, 255, 0.4);
  }

  .graph-stats__spinner {
    display: inline-block;
    width: 12px;
    height: 12px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: #fff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
