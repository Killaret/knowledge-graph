<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { onMount } from "svelte";
  import { handleYandexCallback, error } from "$shared/stores/auth.svelte.js";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  let isProcessing = $state(true);
  let localError = $state<string | null>(null);

  onMount(async () => {
    const code = $page.url.searchParams.get("code");
    const state = $page.url.searchParams.get("state");

    if (!code || !state) {
      localError = t("yandex.missingParams");
      isProcessing = false;
      return;
    }

    const success = await handleYandexCallback(code, state);

    if (success) {
      goto("/");
    } else {
      localError = error() || t("yandex.authFailed");
      isProcessing = false;
    }
  });
</script>

<div class="yandex-callback-page">
  <div class="callback-container">
    {#if isProcessing}
      <div class="loading">
        <div class="spinner"></div>
        <p>{t("yandex.signingIn")}</p>
      </div>
    {:else if localError}
      <div class="error">
        <p>❌ {localError}</p>
        <a href="/auth/login" class="back-link">{t("yandex.backToLogin")}</a>
      </div>
    {/if}
  </div>
</div>

<style>
  .yandex-callback-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: radial-gradient(ellipse at 50% 100%, #0a0a1a 0%, #000 80%);
    padding: 2rem;
  }

  .callback-container {
    text-align: center;
    color: var(--color-text-dark, #e0e0e0);
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }

  .spinner {
    width: 48px;
    height: 48px;
    border: 3px solid rgba(255, 255, 255, 0.1);
    border-top-color: #40a9ff;
    border-right-color: #ffcc00;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    box-shadow: 0 0 20px rgba(64, 169, 255, 0.3);
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .loading p {
    color: var(--color-text-dark, #94a3b8);
    text-shadow: 0 0 10px rgba(255, 204, 0, 0.3);
  }

  .error {
    padding: 2rem;
    background: rgba(10, 10, 26, 0.7);
    backdrop-filter: blur(12px);
    border-radius: 16px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow:
      0 0 0 1px rgba(255, 204, 0, 0.1),
      0 8px 32px rgba(0, 0, 0, 0.4);
  }

  .error p {
    margin: 0 0 1rem;
    color: #ef4444;
    text-shadow: 0 0 8px rgba(239, 68, 68, 0.3);
  }

  .back-link {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    text-decoration: none;
    border-radius: var(--radius-md, 8px);
    font-weight: 500;
    transition: all 0.2s ease;
    box-shadow: 0 4px 15px rgba(59, 130, 246, 0.3);
  }

  .back-link:hover {
    transform: translateY(-2px);
    box-shadow:
      0 8px 25px rgba(59, 130, 246, 0.4),
      0 0 30px rgba(64, 169, 255, 0.2);
  }
</style>
