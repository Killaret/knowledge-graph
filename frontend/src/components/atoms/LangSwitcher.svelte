<script lang="ts">
  import { formatMessage, getCurrentLocale, setLocale, type Locale } from "$shared/utils/i18n";

  interface Props {
    variant?: "light" | "dark";
  }

  const { variant = "light" }: Props = $props();

  const t = (key: string) => formatMessage(key, getCurrentLocale());

  let currentLocale = $state<Locale>(getCurrentLocale());

  function toggleLocale() {
    const next: Locale = currentLocale === "en" ? "ru" : "en";
    currentLocale = next;
    setLocale(next);
    window.location.reload();
  }
</script>

<button
  type="button"
  class="lang-switcher"
  class:lang-switcher--dark={variant === "dark"}
  onclick={toggleLocale}
  title={t("langSwitcher.switchLanguage")}
  aria-label={t("langSwitcher.switchLanguage")}
  data-testid="lang-switcher"
>
  {currentLocale === "en" ? "EN" : "RU"}
</button>

<style>
  .lang-switcher {
    flex-shrink: 0;
    padding: 6px 10px;
    border: 1px solid #e2e8f0;
    background: white;
    border-radius: 14px;
    cursor: pointer;
    font-size: 12px;
    font-weight: 700;
    color: #334155;
    letter-spacing: 0.03em;
    transition: all 0.2s;
    line-height: 1;
  }

  .lang-switcher:hover {
    background: #f1f5f9;
    border-color: #94a3b8;
    color: #1e293b;
  }

  .lang-switcher--dark {
    background: var(--carbon-hex-fill, #14142b);
    border-color: var(--carbon-border, #2d2d3a);
    color: var(--carbon-text, #e0e0e0);
  }

  .lang-switcher--dark:hover {
    background: rgba(139, 92, 246, 0.08);
    border-color: var(--carbon-border-active, #3f3f5a);
    color: #fff;
  }
</style>
