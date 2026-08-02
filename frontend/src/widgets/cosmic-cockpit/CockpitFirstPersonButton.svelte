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
  {t("cockpit.exitFirstPerson")}
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
    color: #2dd4bf;
    font-size: 13px;
    font-weight: 600;
    letter-spacing: 0.05em;
    text-transform: uppercase;
    cursor: pointer;
    opacity: 0;
    transition: opacity 0.3s ease, transform 0.3s ease;
    pointer-events: none;
  }

  .first-person-exit:hover {
    opacity: 1;
    pointer-events: auto;
  }

  .first-person-exit:focus-visible {
    opacity: 1;
    pointer-events: auto;
    outline: 2px solid #2dd4bf;
  }

  @media (hover: hover) {
    .first-person-exit {
      opacity: 0;
      pointer-events: none;
    }
  }

  .glow-bar {
    width: 40px;
    height: 3px;
    border-radius: 2px;
    background: #2dd4bf;
    box-shadow: 0 0 12px #2dd4bf;
  }
</style>
