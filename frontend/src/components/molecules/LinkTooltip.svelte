<script lang="ts">
  import { fade } from "svelte/transition";
  import { LinkType } from "$shared/lib/domain";
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
    onEdit?: () => void;
    onDelete?: () => void;
  } = $props();

  const linkTypeLabel = $derived(LinkType.fromString(linkType).label);

  // Smart positioning to keep tooltip within viewport
  let adjustedX = $derived(x);
  let adjustedY = $derived(y);

  $effect(() => {
    if (!visible) return;

    const tooltipWidth = 220; // approximate width
    const tooltipHeight = 150; // approximate height
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;

    // Adjust X if tooltip would go off right edge
    if (x + tooltipWidth > viewportWidth - 20) {
      adjustedX = x - tooltipWidth - 10;
    }

    // Adjust Y if tooltip would go off bottom edge
    if (y + tooltipHeight > viewportHeight - 20) {
      adjustedY = y - tooltipHeight - 10;
    }

    // Adjust Y if tooltip would go off top edge
    if (y < 20) {
      adjustedY = 20;
    }
  });
</script>

{#if visible}
  <div
    class="link-tooltip"
    style="left: {adjustedX}px; top: {adjustedY}px;"
    transition:fade={{ duration: 200 }}
  >
    <div class="tooltip-header">
      <span class="link-type-badge">{linkTypeLabel}</span>
      {#if sourceType === "gamma"}
        <span class="gamma-badge">{t("linkTooltip.recommended")}</span>
      {/if}
    </div>
    <div class="tooltip-body">
      <div class="tooltip-row">
        <span class="label">{t("linkTooltip.weight")}</span>
        <span class="value">{weight.toFixed(2)}</span>
      </div>
      <div class="tooltip-row">
        <span class="label">{t("linkTooltip.from")}</span>
        <span class="value">{sourceTitle}</span>
      </div>
      <div class="tooltip-row">
        <span class="label">{t("linkTooltip.to")}</span>
        <span class="value">{targetTitle}</span>
      </div>
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
    border: 1px solid rgba(139, 92, 246, 0.5);
    border-radius: 8px;
    padding: 12px;
    min-width: 200px;
    max-width: 280px;
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
    margin-bottom: 8px;
  }

  .link-type-badge {
    background: linear-gradient(135deg, #3b82f6, #8b5cf6);
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
    text-transform: uppercase;
  }

  .gamma-badge {
    background: linear-gradient(135deg, #8b5cf6, #a855f7);
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
  }

  .tooltip-body {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 8px;
  }

  .tooltip-row {
    display: flex;
    justify-content: space-between;
    gap: 12px;
  }

  .label {
    color: rgba(255, 255, 255, 0.6);
  }

  .value {
    color: white;
    font-weight: 500;
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
