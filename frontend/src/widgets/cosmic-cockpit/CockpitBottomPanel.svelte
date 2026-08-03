<script lang="ts">
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
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
  <CockpitHUD
    {nodeCount}
    {linkCount}
    health={health()}
    cluster={t("cockpit.hud.defaultCluster")}
  />
</div>

<style>
  .cockpit-bottom-panel {
    display: flex;
    align-items: center;
    height: 100%;
    padding: 10px 16px;
    box-sizing: border-box;
  }
</style>
