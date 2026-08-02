<script lang="ts">
  import { cockpitStore } from "$features/cosmic-cockpit";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  interface Props {
    nodeCount?: number;
    linkCount?: number;
    cluster?: string | null;
    health?: number;
  }

  const { nodeCount = 0, linkCount = 0, cluster = null, health = 100 }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) => formatMessage(key, locale, params);

  const syncText = $derived(() => {
    if (cockpitStore.syncing) return t("cockpit.hud.syncing");
    if (cockpitStore.lastSyncAt) {
      const seconds = Math.floor((Date.now() - cockpitStore.lastSyncAt) / 1000);
      return t("cockpit.hud.syncAgo", { seconds });
    }
    return t("cockpit.hud.syncIdle");
  });

  const healthColor = $derived(() => {
    if (health >= 80) return "#2dd4bf";
    if (health >= 50) return "#facc15";
    return "#f87171";
  });

  function formatFps(value: number): string {
    return Number.isFinite(value) ? value.toFixed(0) : "--";
  }
</script>

<div class="cockpit-hud" data-testid="cockpit-hud">
  <div class="hud-row hud-row--primary">
    <div class="hud-item cluster" data-testid="hud-cluster">
      <span class="hud-label">{t("cockpit.hud.cluster")}</span>
      <span class="hud-value" title={cluster ?? t("cockpit.hud.noCluster")}>
        {cluster ?? t("cockpit.hud.noCluster")}
      </span>
    </div>

    <div class="hud-item" data-testid="hud-node-count">
      <span class="hud-label">{t("cockpit.hud.notes")}</span>
      <span class="hud-value">{nodeCount}</span>
    </div>

    <div class="hud-item" data-testid="hud-link-count">
      <span class="hud-label">{t("cockpit.hud.links")}</span>
      <span class="hud-value">{linkCount}</span>
    </div>

    <div class="hud-item health" data-testid="hud-health">
      <span class="hud-label">{t("cockpit.hud.health")}</span>
      <div class="health-bar">
        <div class="health-fill" style="width:{health}%;background:{healthColor()}"></div>
      </div>
      <span class="hud-value">{health}%</span>
    </div>

    <div class="hud-item" data-testid="hud-fps">
      <span class="hud-label">{t("cockpit.hud.fps")}</span>
      <span class="hud-value">{formatFps(cockpitStore.fps)}</span>
    </div>

    <div class="hud-item sync" data-testid="hud-sync">
      <span class="hud-label">{t("cockpit.hud.sync")}</span>
      <span class="hud-value" class:syncing={cockpitStore.syncing}>
        {syncText()}
      </span>
    </div>
  </div>

  <div class="hud-row hud-row--actions">
    <button
      type="button"
      class="first-person-btn"
      onclick={() => cockpitStore.toggleFirstPerson()}
      aria-pressed={cockpitStore.firstPerson}
      data-testid="first-person-toggle"
    >
      {cockpitStore.firstPerson ? t("cockpit.hud.exitFirstPerson") : t("cockpit.hud.enterFirstPerson")}
    </button>
  </div>
</div>

<style>
  .cockpit-hud {
    display: flex;
    flex-direction: column;
    justify-content: center;
    gap: 10px;
    height: 100%;
    padding: 10px 14px;
    color: var(--color-text, #e0e0e0);
    font-family: "SFMono-Regular", Consolas, "Courier New", monospace;
    font-size: 12px;
  }

  .hud-row {
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
  }

  .hud-row--primary {
    flex: 1;
  }

  .hud-row--actions {
    justify-content: flex-end;
  }

  .hud-item {
    display: flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
  }

  .hud-label {
    color: rgba(45, 212, 191, 0.7);
    text-transform: uppercase;
    letter-spacing: 0.06em;
    font-size: 10px;
    white-space: nowrap;
  }

  .hud-value {
    color: white;
    font-weight: 600;
    white-space: nowrap;
  }

  .cluster .hud-value {
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .health {
    flex: 1 1 140px;
    min-width: 120px;
  }

  .health-bar {
    width: 100%;
    height: 6px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 3px;
    overflow: hidden;
  }

  .health-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 0.3s ease, background 0.3s ease;
  }

  .syncing {
    color: #2dd4bf;
    animation: pulse 1.2s ease-in-out infinite;
  }

  .first-person-btn {
    padding: 8px 16px;
    border: 1px solid rgba(45, 212, 191, 0.4);
    border-radius: 16px;
    background: rgba(45, 212, 191, 0.1);
    color: #2dd4bf;
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .first-person-btn:hover {
    background: rgba(45, 212, 191, 0.2);
    box-shadow: 0 0 16px rgba(45, 212, 191, 0.3);
  }

  @keyframes pulse {
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.5;
    }
  }

  @media (max-width: 768px) {
    .hud-row--primary {
      gap: 10px;
    }

    .hud-item {
      flex: 1 1 45%;
    }
  }
</style>
