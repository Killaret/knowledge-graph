<script lang="ts">
  import { fade } from "svelte/transition";
  import { LinkType } from "$entities";
  import { formatDate } from "$shared/utils/date";
  import Chip from "$components/atoms/Chip.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const {
    visible,
    x,
    y,
    linkType,
    weight,
    sourceType,
    sourceTitle,
    targetTitle,
    lastWeightUpdate,
    onEdit,
    onDelete,
  }: {
    visible: boolean;
    x: number;
    y: number;
    linkType: string;
    weight: number;
    sourceType: string;
    sourceTitle: string;
    targetTitle: string;
    lastWeightUpdate?: string;
    onEdit?: () => void;
    onDelete?: () => void;
  } = $props();

  const resolvedLinkType = $derived(LinkType.fromString(linkType));
  const linkTypeColor = $derived(resolvedLinkType.color);

  // Smart positioning to keep tooltip within viewport
  let adjustedX = $derived(x);
  let adjustedY = $derived(y);

  $effect(() => {
    if (!visible) return;

    const tooltipWidth = 240;
    const tooltipHeight = 170;
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;

    if (x + tooltipWidth > viewportWidth - 20) {
      adjustedX = x - tooltipWidth - 10;
    }

    if (y + tooltipHeight > viewportHeight - 20) {
      adjustedY = y - tooltipHeight - 10;
    }

    if (y < 20) {
      adjustedY = 20;
    }
  });
</script>

{#if visible}
  <div
    class="link-tooltip"
    style="left: {adjustedX}px; top: {adjustedY}px; --link-color: {linkTypeColor}; --link-bg: {linkTypeColor}33"
    transition:fade={{ duration: 200 }}
  >
    <div class="tooltip-header">
      <Chip
        size="sm"
        color={linkTypeColor}
        borderColor={linkTypeColor}
        background="{linkTypeColor}33"
      >
        <span class="link-type-icon">{resolvedLinkType.icon}</span>
        <span>{resolvedLinkType.label}</span>
      </Chip>
      {#if sourceType === "gamma"}
        <Chip
          size="sm"
          color="#c4b5fd"
          borderColor="rgba(139, 92, 246, 0.5)"
          background="linear-gradient(135deg, #8b5cf6, #a855f7)"
        >
          {t("linkTooltip.recommended")}
        </Chip>
      {/if}
    </div>
    <div class="tooltip-body">
      <div class="tooltip-row">
        <span class="label">{t("linkTooltip.weight")}</span>
        <span class="weight-bar">
          <span class="weight-fill" style="width: {weight * 100}%; background: {linkTypeColor}"
          ></span>
          <span class="weight-value">{weight.toFixed(2)}</span>
        </span>
      </div>
      <div class="tooltip-row">
        <span class="label">{t("linkTooltip.from")}</span>
        <span class="value">{sourceTitle}</span>
      </div>
      <div class="tooltip-row">
        <span class="label">{t("linkTooltip.to")}</span>
        <span class="value">{targetTitle}</span>
      </div>
      {#if lastWeightUpdate}
        <div class="tooltip-row last-update">
          <span class="label">{t("linkTooltip.lastWeightUpdate")}</span>
          <span class="value">{formatDate(lastWeightUpdate)}</span>
        </div>
      {/if}
    </div>
    <div class="tooltip-actions">
      {#if onEdit}
        <button class="action-btn edit-btn" onmousedown={onEdit}>{t("linkTooltip.edit")}</button>
      {/if}
      {#if onDelete}
        <button class="action-btn delete-btn" onmousedown={onDelete}
          >{t("linkTooltip.delete")}</button
        >
      {/if}
    </div>
  </div>
{/if}

<style>
  .link-tooltip {
    position: absolute;
    background: rgba(15, 23, 42, 0.95);
    border: 1px solid var(--link-color, rgba(139, 92, 246, 0.5));
    border-radius: 8px;
    padding: 12px;
    min-width: 220px;
    max-width: 300px;
    color: white;
    font-size: 13px;
    pointer-events: auto;
    z-index: 1000;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(8px);
  }

  .tooltip-header {
    display: flex;
    gap: 8px;
    align-items: center;
    margin-bottom: 10px;
  }

  .link-type-icon {
    font-size: 12px;
  }

  .tooltip-body {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-bottom: 10px;
  }

  .tooltip-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
  }

  .tooltip-row.last-update {
    justify-content: flex-start;
    gap: 6px;
    font-size: 11px;
  }

  .label {
    color: rgba(255, 255, 255, 0.6);
    flex-shrink: 0;
  }

  .value {
    color: white;
    font-weight: 500;
    max-width: 160px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .weight-bar {
    position: relative;
    display: inline-flex;
    align-items: center;
    width: 80px;
    height: 18px;
    background: rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    overflow: hidden;
    border: 1px solid rgba(255, 255, 255, 0.1);
  }

  .weight-fill {
    display: block;
    height: 100%;
    border-radius: 10px;
    opacity: 0.7;
  }

  .weight-value {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    font-weight: 600;
    color: white;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
  }

  .tooltip-actions {
    display: flex;
    gap: 8px;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    padding-top: 8px;
  }

  .action-btn {
    flex: 1;
    padding: 6px 12px;
    border: none;
    border-radius: 4px;
    font-size: 12px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .edit-btn {
    background: rgba(59, 130, 246, 0.2);
    color: #60a5fa;
  }

  .edit-btn:hover {
    background: rgba(59, 130, 246, 0.3);
  }

  .delete-btn {
    background: rgba(239, 68, 68, 0.2);
    color: #f87171;
  }

  .delete-btn:hover {
    background: rgba(239, 68, 68, 0.3);
  }
</style>
