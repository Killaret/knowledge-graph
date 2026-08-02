<script lang="ts">
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import CockpitViewport from "./viewport/CockpitViewport.svelte";
  import CockpitHUD from "./CockpitHUD.svelte";

  interface Props {
    nodeCount?: number;
    linkCount?: number;
  }

  const { nodeCount = 0, linkCount = 0 }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  // Phase 1: decorative health; later compute from graph metrics.
  const health = $derived(() => {
    if (nodeCount === 0) return 100;
    const linkRatio = linkCount / nodeCount;
    return Math.min(100, Math.round((linkRatio / 1.5) * 100));
  });
</script>

<div class="cockpit-bottom-panel" data-testid="cockpit-bottom-panel">
  <div class="bottom-viewport">
    <CockpitViewport />
  </div>
  <div class="bottom-hud">
    <CockpitHUD
      {nodeCount}
      {linkCount}
      health={health()}
      cluster={t("cockpit.hud.defaultCluster")}
    />
  </div>
</div>

<style>
  .cockpit-bottom-panel {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
    height: 100%;
    padding: 10px 12px;
    box-sizing: border-box;
  }

  .bottom-viewport {
    min-width: 0;
    min-height: 0;
  }

  .bottom-hud {
    min-width: 0;
    min-height: 0;
    display: flex;
    align-items: center;
  }

  @media (max-width: 768px) {
    .cockpit-bottom-panel {
      grid-template-columns: 1fr;
      grid-template-rows: 1fr 1fr;
    }
  }
</style>
