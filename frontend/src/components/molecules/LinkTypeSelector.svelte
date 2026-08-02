<script lang="ts">
  /* eslint-disable prefer-const -- Svelte 5 $props() with $bindable requires let */
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  export interface LinkTypeOption {
    type: string;
    icon: string;
    label: string;
    color: string;
    description?: string;
    example?: string;
  }

  interface Props {
    types: LinkTypeOption[];
    defaultSelected?: string;
    selected?: string;
    size?: "sm" | "md";
    showDescription?: boolean;
    showIcon?: boolean;
    disabled?: boolean;
    id?: string;
    onSelect?: (type: string) => void;
  }

  let {
    types,
    defaultSelected = types[0]?.type ?? "",
    selected = defaultSelected,
    size = "md",
    showDescription = true,
    showIcon = true,
    disabled = false,
    id = "link-type-selector",
    onSelect,
  }: Props = $props();

  let container: HTMLDivElement | null = $state(null);

  function selectType(type: string) {
    if (disabled) return;
    onSelect?.(type);
  }

  function handleKeydown(event: KeyboardEvent, index: number) {
    if (disabled) return;
    if (!["ArrowDown", "ArrowUp", "ArrowLeft", "ArrowRight"].includes(event.key)) return;

    event.preventDefault();
    const isHorizontal = size === "sm";
    const direction = ["ArrowRight", "ArrowDown"].includes(event.key) ? 1 : -1;
    const nextIndex = (index + direction + types.length) % types.length;
    const nextButton = container?.querySelector<HTMLButtonElement>(`[data-index="${nextIndex}"]`);
    nextButton?.focus();
    selectType(types[nextIndex].type);
  }
</script>

<div
  bind:this={container}
  class="link-type-selector"
  class:sm={size === "sm"}
  {id}
  role="radiogroup"
  aria-label={t("linkTypeSelector.ariaLabel")}
  aria-disabled={disabled}
>
  {#each types as type, index}
    <button
      type="button"
      class="link-type-btn"
      class:active={selected === type.type}
      class:sm={size === "sm"}
      data-type={type.type}
      data-index={index}
      role="radio"
      aria-checked={selected === type.type}
      tabindex={selected === type.type ? 0 : -1}
      {disabled}
      style="--type-color: {type.color}; --type-bg: {type.color}33"
      onclick={() => selectType(type.type)}
      onkeydown={(e) => handleKeydown(e, index)}
    >
      {#if showIcon}
        <span class="link-type-icon" aria-hidden="true">{type.icon}</span>
      {/if}
      <span class="link-type-content">
        <span class="link-type-label">{type.label}</span>
        {#if showDescription && size === "md" && type.description}
          <span class="link-type-description">{type.description}</span>
        {/if}
        {#if showDescription && size === "md" && type.example}
          <span class="link-type-example">{t("linkTypeSelector.example")}: {type.example}</span>
        {/if}
      </span>
    </button>
  {/each}
</div>

<style>
  .link-type-selector {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .link-type-selector.sm {
    flex-direction: row;
    flex-wrap: wrap;
    gap: 6px;
  }

  .link-type-btn {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    width: 100%;
    padding: 10px 12px;
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-left: 3px solid var(--type-color);
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.3);
    color: rgba(255, 255, 255, 0.85);
    cursor: pointer;
    font-size: 13px;
    text-align: left;
    transition:
      background 0.15s ease,
      border-color 0.15s ease,
      transform 0.1s ease;
    outline: none;
  }

  .link-type-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .link-type-btn:hover:not(:disabled) {
    background: var(--type-bg);
    border-color: var(--type-color);
  }

  .link-type-btn:focus-visible:not(:disabled) {
    box-shadow: 0 0 0 2px var(--type-color);
  }

  .link-type-btn.active {
    background: var(--type-bg);
    border-color: var(--type-color);
    color: white;
  }

  .link-type-btn.sm {
    align-items: center;
    width: auto;
    padding: 6px 10px;
    border-left-width: 1px;
    border-left-color: rgba(255, 255, 255, 0.15);
  }

  .link-type-btn.sm.active {
    border-color: var(--type-color);
  }

  .link-type-icon {
    flex-shrink: 0;
    font-size: 18px;
    line-height: 1.2;
  }

  .link-type-content {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .link-type-label {
    font-weight: 600;
    line-height: 1.3;
  }

  .link-type-description {
    font-size: 12px;
    color: rgba(255, 255, 255, 0.6);
    line-height: 1.3;
  }

  .link-type-example {
    font-size: 11px;
    color: rgba(255, 255, 255, 0.5);
    font-style: italic;
  }

  @media (max-width: 420px) {
    .link-type-selector.sm {
      gap: 4px;
    }

    .link-type-btn.sm {
      padding: 6px 8px;
    }

    .link-type-btn.sm .link-type-label {
      display: none;
    }
  }
</style>
