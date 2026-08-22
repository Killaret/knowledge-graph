<script lang="ts">
  import { cockpitStore } from "$features/cosmic-cockpit";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);
</script>

<button
  type="button"
  class="first-person-exit"
  onclick={() => cockpitStore.exitFirstPerson()}
  aria-label={t("cockpit.exitFirstPerson")}
  data-testid="first-person-exit"
>
  <span class="glow-bar"></span>
  <span class="cockpit-gradient-text" style="--cockpit-text-delay: -2.2s;">
    {t("cockpit.exitFirstPersonHint")}
  </span>
</button>

<style>
  .first-person-exit {
    position: fixed;
    bottom: 24px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 500;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    padding: 10px 24px;
    border: 1px solid rgba(45, 212, 191, 0.4);
    border-radius: 20px;
    background: rgba(10, 10, 15, 0.85);
    color: var(--cockpit-accent, #2dd4bf);
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    cursor: pointer;
    opacity: 1;
    transition:
      opacity 0.3s ease,
      transform 0.3s ease;
    pointer-events: auto;
  }

  .first-person-exit:hover {
    transform: translateX(-50%) scale(1.05);
  }

  .first-person-exit:focus-visible {
    outline: 2px solid var(--cockpit-accent, #2dd4bf);
  }

  .glow-bar {
    width: 40px;
    height: 3px;
    border-radius: 2px;
    background: var(--cockpit-accent, #2dd4bf);
    box-shadow: 0 0 12px var(--cockpit-accent, #2dd4bf);
  }
</style>
