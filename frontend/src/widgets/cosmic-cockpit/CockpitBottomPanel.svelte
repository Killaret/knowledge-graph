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
  // NOTE-HEALTH-1 этап 0: decorative HEALTH is hidden; этап 2 passes a real metric.
</script>

<div class="cockpit-bottom-panel" data-testid="cockpit-bottom-panel">
  <CockpitHUD {nodeCount} {linkCount} cluster={t("cockpit.hud.defaultCluster")} />
</div>

<style>
  .cockpit-bottom-panel {
    display: flex;
    align-items: center;
    height: auto;
    padding: 10px 16px;
    box-sizing: border-box;
  }
</style>
