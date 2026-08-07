<script lang="ts">
  import { goto } from "$app/navigation";
  import { browser } from "$app/environment";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const t = (key: string) => formatMessage(key, getCurrentLocale());

  const { href = "/", text = t("backButton.back") }: { href?: string; text?: string } = $props();

  function handleBack() {
    if (browser && window.history.length > 1) {
      window.history.back();
    } else {
      goto(href);
    }
  }
</script>

<button class="back-button" onclick={handleBack} aria-label={t("backButton.goBack")}>
  <svg
    class="back-icon"
    width="16"
    height="16"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="2.2"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <path d="M19 12H5" />
    <path d="M12 19l-7-7 7-7" />
  </svg>
  <span>{text}</span>
</button>

<style>
  .back-button {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    background: var(--carbon-graphene, #12121a);
    color: var(--carbon-text, #f0f0f5);
    padding: 0.5rem 1rem;
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 8px;
    text-decoration: none;
    font-size: 0.9rem;
    font-weight: 500;
    cursor: pointer;
    transition: all var(--carbon-transition, 0.25s ease);
    margin-bottom: 1rem;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
  }

  .back-button:hover {
    background: var(--carbon-c70, #1a1a24);
    border-color: var(--carbon-glow-cyan, #22d3ee);
    color: var(--carbon-glow-cyan, #22d3ee);
    transform: translateY(-1px);
    box-shadow: var(--carbon-glow-cyan, 0 0 12px rgba(34, 211, 238, 0.2));
  }

  .back-button:active {
    transform: translateY(0);
  }

  .back-icon {
    flex-shrink: 0;
  }
</style>
