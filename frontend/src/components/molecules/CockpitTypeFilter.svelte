<script lang="ts">
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  interface Filter {
    id: string;
    label: string;
    emoji: string;
  }

  interface Props {
    filters: Filter[];
    selected: string;
    onSelect: (id: string) => void;
    typeCounts?: Record<string, number>;
  }

  const { filters, selected, onSelect, typeCounts = {} }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string>) => formatMessage(key, locale, params);
</script>

<div class="cockpit-type-filter" data-testid="cockpit-type-filter">
  <div class="filter-list" role="group" aria-label={t("cockpit.typeFilterAria")}>
    {#each filters as filter}
      <button
        type="button"
        class="filter-chip {selected === filter.id ? 'active' : ''}"
        class:has-count={typeCounts[filter.id] !== undefined}
        onclick={() => onSelect(filter.id)}
        aria-pressed={selected === filter.id}
        data-testid="cockpit-type-filter-{filter.id}"
        title={t("filter.filterBy", { type: filter.label })}
      >
        <span class="filter-emoji">{filter.emoji}</span>
        <span class="filter-label">{filter.label}</span>
        {#if typeCounts[filter.id] !== undefined}
          <span class="filter-count" data-testid="filter-count-{filter.id}">{typeCounts[filter.id]}</span>
        {/if}
      </button>
    {/each}
  </div>
</div>

<style>
  .cockpit-type-filter {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .filter-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .filter-chip {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 10px;
    border: 1px solid rgba(45, 212, 191, 0.15);
    border-radius: 10px;
    background: rgba(255, 255, 255, 0.03);
    color: var(--color-text, #e0e0e0);
    font-size: 13px;
    cursor: pointer;
    text-align: left;
    transition: all 0.2s ease;
  }

  .filter-chip:hover {
    background: rgba(45, 212, 191, 0.08);
    border-color: rgba(45, 212, 191, 0.3);
  }

  .filter-chip.active {
    background: rgba(45, 212, 191, 0.2);
    border-color: rgba(45, 212, 191, 0.5);
    color: #2dd4bf;
  }

  .filter-emoji {
    font-size: 16px;
  }

  .filter-label {
    flex: 1;
  }

  .filter-count {
    padding: 2px 8px;
    background: rgba(255, 255, 255, 0.08);
    border-radius: 10px;
    font-size: 11px;
    font-weight: 600;
  }

  .filter-chip.active .filter-count {
    background: rgba(0, 0, 0, 0.2);
  }
</style>
