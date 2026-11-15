<script lang="ts">
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { onMount } from "svelte";

  interface PickerNode {
    id: string;
    title: string;
    type?: string;
  }

  interface Props {
    x: number;
    y: number;
    visible: boolean;
    sourceId: string;
    nodes: PickerNode[];
    onSelect: (node: PickerNode) => void;
    onClose: () => void;
  }

  const { x, y, visible, sourceId, nodes, onSelect, onClose }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  let query = $state("");
  let pickerEl: HTMLDivElement | null = $state(null);
  let inputEl: HTMLInputElement | null = $state(null);

  // UX-1: the technical Knowledge Core node and the source itself can never be
  // a link target — same exclusion as the drag-to-link path.
  const candidates = $derived.by(() => {
    const q = query.trim().toLowerCase();
    return nodes.filter(
      (n) =>
        n.id !== sourceId &&
        n.type !== "technical" &&
        (q === "" || n.title.toLowerCase().includes(q))
    );
  });

  function handleClickOutside(e: MouseEvent) {
    if (pickerEl && !pickerEl.contains(e.target as Node)) {
      onClose();
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      onClose();
    }
  }

  onMount(() => {
    window.addEventListener("click", handleClickOutside);
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("click", handleClickOutside);
      window.removeEventListener("keydown", handleKeyDown);
    };
  });

  $effect(() => {
    if (visible && pickerEl) {
      const rect = pickerEl.getBoundingClientRect();
      const winW = window.innerWidth;
      const winH = window.innerHeight;
      let adjustedX = x;
      let adjustedY = y;
      if (adjustedX + rect.width > winW) {
        adjustedX = Math.max(8, winW - rect.width - 8);
      }
      if (adjustedY + rect.height > winH) {
        adjustedY = Math.max(8, winH - rect.height - 8);
      }
      pickerEl.style.left = `${adjustedX}px`;
      pickerEl.style.top = `${adjustedY}px`;
      inputEl?.focus();
    }
    if (visible) {
      query = "";
    }
  });
</script>

{#if visible}
  <div
    bind:this={pickerEl}
    class="link-target-picker"
    role="dialog"
    aria-label={t("graph.linkPicker.title")}
    data-testid="link-target-picker"
    style="position: fixed; left: {x}px; top: {y}px;"
  >
    <div class="picker-header">{t("graph.linkPicker.title")}</div>
    <input
      bind:this={inputEl}
      bind:value={query}
      class="picker-input"
      type="text"
      placeholder={t("graph.linkPicker.placeholder")}
      aria-label={t("graph.linkPicker.placeholder")}
      data-testid="link-target-search"
    />
    <div class="picker-list" role="listbox">
      {#if candidates.length === 0}
        <div class="picker-empty" data-testid="link-target-empty">
          {t("graph.linkPicker.empty")}
        </div>
      {:else}
        {#each candidates as candidate (candidate.id)}
          <button
            type="button"
            class="picker-option"
            role="option"
            aria-selected="false"
            onclick={() => {
              onSelect(candidate);
              onClose();
            }}
            data-testid="link-target-option"
          >
            {candidate.title}
          </button>
        {/each}
      {/if}
    </div>
  </div>
{/if}

<style>
  .link-target-picker {
    min-width: 220px;
    max-width: 280px;
    background: rgba(10, 15, 30, 0.96);
    border: 1px solid rgba(45, 212, 191, 0.3);
    border-radius: 10px;
    padding: 8px 0;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
    z-index: 210;
    backdrop-filter: blur(12px);
    color: #e0e0e0;
  }

  .picker-header {
    padding: 8px 14px;
    font-size: 13px;
    font-weight: 600;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  }

  .picker-input {
    width: calc(100% - 28px);
    margin: 8px 14px;
    padding: 8px 10px;
    background: rgba(255, 255, 255, 0.06);
    border: 1px solid rgba(45, 212, 191, 0.3);
    border-radius: 6px;
    color: inherit;
    font-size: 13px;
    outline: none;
  }

  .picker-input:focus {
    border-color: #2dd4bf;
  }

  .picker-list {
    max-height: 240px;
    overflow-y: auto;
  }

  .picker-option {
    display: block;
    width: 100%;
    text-align: left;
    background: transparent;
    border: none;
    color: inherit;
    padding: 10px 14px;
    font-size: 13px;
    cursor: pointer;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    transition: background 0.15s ease;
  }

  .picker-option:hover,
  .picker-option:focus {
    background: rgba(45, 212, 191, 0.12);
    outline: none;
  }

  .picker-empty {
    padding: 12px 14px;
    font-size: 13px;
    color: #94a3b8;
  }
</style>
