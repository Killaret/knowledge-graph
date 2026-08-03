<script lang="ts">
  /* eslint-disable prefer-const -- Svelte 5 $props() with $bindable requires let */
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  export interface TypeOption {
    type: string;
    emoji: string;
    label: string;
    color: string;
    description?: string;
    example?: string;
    toCSSColor(): string;
  }

  interface Props {
    types: TypeOption[];
    defaultSelected?: string;
    selected?: string;
    id?: string;
  }

  let {
    types,
    defaultSelected = types[0]?.type ?? "star",
    selected = $bindable(defaultSelected),
    id = "type-selector",
  }: Props = $props();

  function selectType(type: string) {
    selected = type;
  }
</script>

<div class="type-selector" {id} role="group" aria-label={t("typeSelector.ariaLabel")}>
  {#each types as type}
    <button
      type="button"
      class="type-btn"
      data-type={type.type}
      class:active={selected === type.type}
      onclick={() => selectType(type.type)}
      style="--type-color: {type.toCSSColor()}; --type-bg: {type.color}33"
      aria-pressed={selected === type.type}
      title={type.example
        ? `${type.description}\n${t("typeSelector.example")}: ${type.example}`
        : type.description}
    >
      <span class="emoji">{type.emoji}</span>
      <span class="type-text">
        <span class="label">{type.label}</span>
        {#if type.description}
          <span class="description">{type.description}</span>
        {/if}
      </span>
    </button>
  {/each}
</div>

<style>
  .type-selector {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .type-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    padding: 8px 14px;
    border: 1px solid var(--carbon-border, #2d2d3d);
    background: var(--carbon-graphene, #12121a);
    border-radius: 20px;
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    color: var(--carbon-text-muted, #8b8b9e);
    transition: all var(--carbon-transition, 0.25s ease);
    outline: none;
  }

  .type-btn:hover {
    border-color: var(--type-color);
    background: var(--type-bg);
    color: var(--carbon-text, #f0f0f5);
    transform: translateY(-1px);
    box-shadow: 0 0 14px color-mix(in srgb, var(--type-color) 25%, transparent);
  }

  .type-btn:focus-visible {
    box-shadow: 0 0 0 3px color-mix(in srgb, var(--type-color) 25%, transparent);
  }

  .type-btn.active {
    border-color: var(--type-color);
    background: color-mix(in srgb, var(--type-color) 18%, var(--carbon-c70, #1a1a24));
    color: var(--carbon-text, #f0f0f5);
    box-shadow:
      0 0 0 1px var(--type-color),
      0 0 18px color-mix(in srgb, var(--type-color) 30%, transparent);
  }

  .type-btn.active:hover {
    filter: brightness(1.1);
  }

  .emoji {
    font-size: 16px;
    line-height: 1;
    filter: drop-shadow(0 0 4px color-mix(in srgb, var(--type-color) 50%, transparent));
  }

  .type-text {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 2px;
    line-height: 1.2;
  }

  .label {
    line-height: 1;
  }

  .description {
    max-width: 160px;
    font-size: 10px;
    font-weight: 400;
    color: var(--carbon-text-dim, #5a5a6e);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .type-btn:hover .description,
  .type-btn.active .description {
    color: var(--carbon-text-muted, #8b8b9e);
  }

  @media (max-width: 420px) {
    .type-selector {
      gap: 6px;
    }

    .type-btn {
      padding: 6px 8px;
      min-width: 36px;
      min-height: 36px;
    }

    .label {
      display: none;
    }
  }
</style>
