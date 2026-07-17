<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import tippy from 'tippy.js';
  import type { Instance } from 'tippy.js';
  import 'tippy.js/dist/tippy.css';

  const {
    target
  }: {
    target: HTMLElement;
  } = $props();

  let tippyInstance: Instance | null = $state(null);
  let currentContent: string = $state('');

  onMount(() => {
    tippyInstance = tippy(target, {
      content: '',
      placement: 'top',
      followCursor: true,
      animation: 'fade',
      duration: 200,
      arrow: true,
      theme: 'light',
      hideOnClick: false,
      interactive: true,
      allowHTML: true
    });
  });

  onDestroy(() => {
    tippyInstance?.destroy();
  });

  export function showNodeTooltip(title: string, type: string, emoji: string): void {
    if (!tippyInstance) return;
    currentContent = `
      <div style="padding: 8px 12px;">
        <div style="font-weight: 600; margin-bottom: 4px;">${emoji} ${type}</div>
        <div style="color: #666;">${title}</div>
      </div>
    `;
    tippyInstance.setContent(currentContent);
    tippyInstance.show();
  }

  export function showLinkTooltip(linkType: string, weight: number): void {
    if (!tippyInstance) return;
    currentContent = `
      <div style="padding: 8px 12px;">
        <div style="font-weight: 600; margin-bottom: 4px;">${linkType}</div>
        <div style="color: #666;">Weight: ${weight.toFixed(2)}</div>
      </div>
    `;
    tippyInstance.setContent(currentContent);
    tippyInstance.show();
  }

  export function hide(): void {
    tippyInstance?.hide();
  }
</script>
