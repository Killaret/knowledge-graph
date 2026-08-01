<script lang="ts">
  /* eslint-disable prefer-const -- Svelte 5 $props() with $bindable requires let */
  import { CelestialBody } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  interface Props {
    selected: string;
    id?: string;
  }

  let { selected = $bindable(CelestialBody.STAR.type), id = "type-selector" }: Props = $props();

  const types = CelestialBody.UI_TYPES;

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
    >
      <span class="emoji">{type.emoji}</span>
      <span class="label">{type.label}</span>
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
    border: 2px solid var(--color-border, #e5e7eb);
    background: var(--color-surface, white);
    border-radius: 20px;
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    color: var(--color-text-secondary, #6b7280);
    transition: all 0.2s ease;
    outline: none;
  }

  .type-btn:hover {
    border-color: var(--type-color);
    background: var(--type-bg);
    color: var(--color-text, #1f2937);
    transform: translateY(-1px);
  }

  .type-btn:focus {
    box-shadow: 0 0 0 3px var(--type-bg);
  }

  .type-btn.active {
    border-color: var(--color-primary, #3b82f6);
    background: var(--color-primary-light, rgba(59, 130, 246, 0.1));
    color: var(--color-primary, #3b82f6);
    box-shadow: 0 0 0 2px var(--color-primary, #3b82f6);
  }

  .type-btn.active:hover {
    border-color: var(--color-primary-hover, #2563eb);
    background: rgba(59, 130, 246, 0.15);
    color: var(--color-primary-hover, #2563eb);
  }

  .emoji {
    font-size: 16px;
    line-height: 1;
  }

  .label {
    line-height: 1;
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
