<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import type { Instance } from "tippy.js";
  import "tippy.js/dist/tippy.css";
  import { CelestialBody, LinkType } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const {
    target,
  }: {
    target: HTMLElement;
  } = $props();

  let tippyInstance: Instance | null = $state(null);
  let currentContent: string = $state("");

  onMount(async () => {
    const { default: tippy } = await import("tippy.js");
    tippyInstance = tippy(target, {
      content: "",
      placement: "top",
      followCursor: true,
      animation: "fade",
      duration: 200,
      arrow: true,
      theme: "light",
      hideOnClick: false,
      interactive: true,
      allowHTML: true,
    });
  });

  onDestroy(() => {
    tippyInstance?.destroy();
  });

  export function showNodeTooltip(
    title: string,
    type: string,
    emoji: string,
  ): void {
    if (!tippyInstance) return;
    const typeLabel = CelestialBody.fromString(type).label;
    currentContent = `
      <div style="padding: 8px 12px;">
        <div style="font-weight: 600; margin-bottom: 4px;">${emoji} ${typeLabel}</div>
        <div style="color: #666;">${title}</div>
      </div>
    `;
    tippyInstance.setContent(currentContent);
    tippyInstance.show();
  }

  export function showLinkTooltip(linkType: string, weight: number): void {
    if (!tippyInstance) return;
    const typeLabel = LinkType.fromString(linkType).label;
    currentContent = `
      <div style="padding: 8px 12px;">
        <div style="font-weight: 600; margin-bottom: 4px;">${typeLabel}</div>
        <div style="color: #666;">${t("linkTooltip.weight")} ${weight.toFixed(2)}</div>
      </div>
    `;
    tippyInstance.setContent(currentContent);
    tippyInstance.show();
  }

  export function hide(): void {
    tippyInstance?.hide();
  }
</script>
