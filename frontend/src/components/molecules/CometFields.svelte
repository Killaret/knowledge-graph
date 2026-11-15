<script lang="ts">
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { REMIND_CHOICES, type RemindChoice } from "$shared/utils/comet";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  interface Props {
    /** datetime-local value; "" = без даты */
    dueAtLocal?: string;
    remindChoice?: RemindChoice;
    customMinutes?: number;
    done?: boolean;
    /** «Сделано» — только в форме редактирования существующей кометы */
    showDone?: boolean;
    disabled?: boolean;
    testIdPrefix?: string;
  }

  /* eslint-disable prefer-const -- Svelte 5 $bindable() requires let, not const, see: https://svelte.dev/docs/svelte/$bindable */
  let {
    dueAtLocal = $bindable(""),
    remindChoice = $bindable("none" as RemindChoice),
    customMinutes = $bindable(60),
    done = $bindable(false),
    showDone = false,
    disabled = false,
    testIdPrefix = "comet",
  }: Props = $props();
  /* eslint-enable prefer-const */
</script>

<div class="comet-fields" data-testid="{testIdPrefix}-fields">
  <div class="comet-row">
    <div class="form-group">
      <label for="{testIdPrefix}-due">{t("comet.dueLabel")}</label>
      <input
        id="{testIdPrefix}-due"
        type="datetime-local"
        bind:value={dueAtLocal}
        {disabled}
        data-testid="{testIdPrefix}-due"
      />
    </div>

    {#if dueAtLocal}
      <div class="form-group">
        <label for="{testIdPrefix}-remind">{t("comet.remindLabel")}</label>
        <select
          id="{testIdPrefix}-remind"
          value={remindChoice}
          onchange={(e) => (remindChoice = e.currentTarget.value as RemindChoice)}
          {disabled}
          data-testid="{testIdPrefix}-remind"
        >
          {#each REMIND_CHOICES as choice (choice)}
            <option value={choice}>{t(`comet.remind.${choice}`)}</option>
          {/each}
        </select>
        {#if remindChoice === "custom"}
          <div class="custom-minutes">
            <input
              type="number"
              min="1"
              bind:value={customMinutes}
              {disabled}
              data-testid="{testIdPrefix}-custom-minutes"
            />
            <span>{t("comet.customMinutes")}</span>
          </div>
        {/if}
      </div>
    {/if}
  </div>

  {#if showDone}
    <label class="done-row">
      <input type="checkbox" bind:checked={done} {disabled} data-testid="{testIdPrefix}-done" />
      {t("comet.done")}
    </label>
  {/if}
</div>

<style>
  .comet-fields {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    padding: 0.75rem;
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 10px;
    background: var(--carbon-graphite, #0b0b10);
  }

  .comet-row {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
    flex: 1;
    min-width: 140px;
  }

  label {
    font-size: 13px;
    font-weight: 500;
    color: var(--carbon-text-muted, #8b8b9e);
  }

  input,
  select {
    padding: 8px 10px;
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 8px;
    background: var(--carbon-black, #050508);
    color: var(--carbon-text, #f0f0f5);
    font-size: 13px;
    font-family: inherit;
    box-sizing: border-box;
  }

  input:focus,
  select:focus {
    outline: none;
    border-color: var(--carbon-glow-cyan, #22d3ee);
    box-shadow: 0 0 0 3px rgba(34, 211, 238, 0.15);
  }

  input:disabled,
  select:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .custom-minutes {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.375rem;
    font-size: 13px;
    color: var(--carbon-text-dim, #7a7a8e);
  }

  .custom-minutes input {
    width: 90px;
  }

  .done-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 14px;
    color: var(--carbon-text, #f0f0f5);
    cursor: pointer;
  }

  .done-row input {
    width: auto;
  }
</style>
