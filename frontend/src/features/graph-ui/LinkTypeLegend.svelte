<script lang="ts">
  import { LinkType } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const {
    hiddenTypes = [],
    onToggle,
    onMinWeightChange,
    minWeight = 0,
    showMinWeight = false,
    collapsible = true,
  }: {
    hiddenTypes?: string[];
    onToggle?: (type: string) => void;
    onMinWeightChange?: (value: number) => void;
    minWeight?: number;
    showMinWeight?: boolean;
    collapsible?: boolean;
  } = $props();

  let collapsed = $state(false);
  const types = $derived(LinkType.ALL_TYPES);
  const isInteractive = $derived(!!onToggle);
  const areAllVisible = $derived(hiddenTypes.length === 0);
  const areAllHidden = $derived(hiddenTypes.length === types.length);

  // hiddenTypes is now interpreted as the set of HIDDEN link types.
  function showAll() {
    if (!onToggle) return;
    for (const type of types) {
      if (hiddenTypes.includes(type.type)) {
        onToggle(type.type);
      }
    }
  }

  function hideAll() {
    if (!onToggle) return;
    for (const type of types) {
      if (!hiddenTypes.includes(type.type)) {
        onToggle(type.type);
      }
    }
  }

  function handleMinWeightInput(event: Event) {
    const target = event.target as HTMLInputElement;
    onMinWeightChange?.(Number(target.value));
  }
</script>

<div class="link-type-legend" class:collapsed>
  <button
    type="button"
    class="legend-header"
    disabled={!collapsible}
    onclick={() => (collapsed = !collapsed)}
  >
    <span class="legend-title">{t("linkLegend.title")}</span>
    {#if collapsible}
      <span class="legend-toggle">{collapsed ? "▼" : "▲"}</span>
    {/if}
  </button>

  {#if !collapsed}
    <div class="legend-content">
      {#if isInteractive}
        <div class="legend-actions">
          <button type="button" class="legend-action" disabled={areAllVisible} onclick={showAll}>
            {t("linkLegend.showAll")}
          </button>
          <button type="button" class="legend-action" disabled={areAllHidden} onclick={hideAll}>
            {t("linkLegend.hideAll")}
          </button>
        </div>
      {/if}

      {#if showMinWeight}
        <div class="min-weight-control">
          <label for="link-legend-min-weight" class="min-weight-label">
            {t("linkLegend.minWeight", { weight: minWeight.toFixed(1) })}
          </label>
          <input
            id="link-legend-min-weight"
            type="range"
            min="0"
            max="1"
            step="0.1"
            value={minWeight}
            oninput={handleMinWeightInput}
          />
        </div>
      {/if}

      <div class="legend-list" role="list">
        {#each types as type}
          <div class="legend-list-item" role="listitem">
            <button
              type="button"
              class="legend-item"
              class:disabled={isInteractive && hiddenTypes.includes(type.type)}
              style="--type-color: {type.color}"
              onclick={() => onToggle?.(type.type)}
              aria-pressed={isInteractive ? !hiddenTypes.includes(type.type) : undefined}
              disabled={!isInteractive}
            >
              <span class="legend-line" style="background: {type.color}"></span>
              <span class="legend-icon">{type.icon}</span>
              <span class="legend-label">{type.label}</span>
            </button>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<style>
  .link-type-legend {
    position: absolute;
    bottom: 16px;
    right: 16px;
    background: rgba(15, 23, 42, 0.92);
    border: 1px solid rgba(139, 92, 246, 0.3);
    border-radius: 10px;
    padding: 10px 12px;
    min-width: 180px;
    max-width: 260px;
    color: white;
    font-size: 12px;
    z-index: 50;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
    backdrop-filter: blur(8px);
  }

  .legend-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    padding: 0;
    background: transparent;
    border: none;
    color: inherit;
    font: inherit;
    cursor: pointer;
    user-select: none;
  }

  .legend-header:disabled {
    cursor: default;
  }

  .legend-title {
    font-weight: 600;
    font-size: 13px;
  }

  .legend-toggle {
    color: rgba(255, 255, 255, 0.5);
    font-size: 11px;
  }

  .legend-content {
    display: flex;
    flex-direction: column;
    gap: 10px;
    margin-top: 10px;
  }

  .legend-actions {
    display: flex;
    gap: 8px;
  }

  .legend-action {
    flex: 1;
    padding: 4px 6px;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 5px;
    background: rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.8);
    font-size: 11px;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .legend-action:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.1);
  }

  .legend-action:disabled {
    opacity: 0.4;
    cursor: not-allowed;
  }

  .min-weight-control {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .min-weight-label {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.7);
  }

  .min-weight-control input[type="range"] {
    width: 100%;
    accent-color: #fbbf24;
  }

  .legend-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    padding: 5px 8px;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: white;
    font-size: 12px;
    cursor: default;
    text-align: left;
    transition: background 0.15s ease;
  }

  .legend-item:disabled {
    opacity: 0.4;
  }

  .legend-item:not(:disabled) {
    cursor: pointer;
  }

  .legend-item:not(:disabled):hover {
    background: rgba(255, 255, 255, 0.05);
  }

  .legend-item.disabled:not(:disabled) {
    opacity: 0.35;
  }

  .legend-line {
    width: 20px;
    height: 3px;
    border-radius: 2px;
    flex-shrink: 0;
  }

  .legend-icon {
    font-size: 13px;
    width: 16px;
    text-align: center;
  }

  .legend-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>
