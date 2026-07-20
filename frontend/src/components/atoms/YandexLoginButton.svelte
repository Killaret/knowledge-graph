<script lang="ts">
  import { getYandexLoginUrl } from "$shared/api/auth.js";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  let isLoading = $state(false);
  let error = $state<string | null>(null);

  async function handleYandexLogin() {
    isLoading = true;
    error = null;

    try {
      const result = await getYandexLoginUrl();
      if (result.url) {
        window.location.href = result.url;
      } else {
        error = t("yandex.authUrlError");
      }
    } catch (err) {
      error =
        err instanceof Error ? err.message : t("yandex.initError");
    } finally {
      isLoading = false;
    }
  }
</script>

<button
  type="button"
  class="yandex-login-button"
  onclick={handleYandexLogin}
  disabled={isLoading}
>
  {#if isLoading}
    <span class="spinner"></span>
    <span>{t("yandex.connecting")}</span>
  {:else}
    <svg
      class="yandex-icon"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <path
        d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"
      />
      <circle cx="12" cy="12" r="3" />
    </svg>
    <span>{t("yandex.signIn")}</span>
  {/if}
</button>

{#if error}
  <div class="error-message">
    {error}
  </div>
{/if}

<style>
  .yandex-login-button {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    width: 100%;
    padding: 0.875rem 1.5rem;
    background: #fc3f1d;
    color: white;
    border: none;
    border-radius: var(--radius-md);
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .yandex-login-button:hover:not(:disabled) {
    background: #e63917;
    transform: translateY(-2px) scale(1.02);
    box-shadow:
      0 8px 25px rgba(252, 63, 29, 0.4),
      0 0 20px rgba(252, 63, 29, 0.3),
      0 0 40px rgba(255, 204, 0, 0.1);
  }

  .yandex-login-button:disabled {
    opacity: 0.7;
    cursor: not-allowed;
  }

  .yandex-icon {
    width: 20px;
    height: 20px;
  }

  .spinner {
    width: 18px;
    height: 18px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .error-message {
    margin-top: 0.75rem;
    padding: 0.75rem;
    background: var(--color-error-bg, #fee2e2);
    color: var(--color-error, #dc2626);
    border-radius: var(--radius-sm);
    font-size: 0.875rem;
    text-align: center;
  }
</style>
