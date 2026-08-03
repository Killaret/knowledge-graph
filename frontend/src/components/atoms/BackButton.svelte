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
  {text}
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

  .back-button::before {
    content: "«";
    font-weight: bold;
  }
</style>
