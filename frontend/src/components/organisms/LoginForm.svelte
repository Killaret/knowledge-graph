<script lang="ts">
  import { goto } from "$app/navigation";
  import Button from "$components/atoms/Button.svelte";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import YandexLoginButton from "$components/atoms/YandexLoginButton.svelte";
  import { login, loginWithApiKey, isLoading, error } from "$shared/stores/auth.svelte.js";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  interface Props {
    redirectTo?: string;
    onSuccess?: () => void;
    onRegister?: () => void;
  }
  const { redirectTo = "/", onSuccess, onRegister }: Props = $props();

  // Check if API Key mode is enabled
  const apiKeyEnabled = import.meta.env.VITE_API_KEY_ENABLED === "true";
  const yandexEnabled = import.meta.env.VITE_YANDEX_ENABLED === "true";

  let loginValue = $state("");
  let password = $state("");
  let apiKeyValue = $state("");
  let useApiKey = $state(false);
  let localError = $state<string | null>(null);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    localError = null;

    if (useApiKey && apiKeyEnabled) {
      // Login with API Key
      const success = await loginWithApiKey(apiKeyValue.trim());
      if (success) {
        if (onSuccess) {
          onSuccess();
        } else {
          goto(redirectTo);
        }
      } else {
        localError = error() || t("auth.invalidApiKey");
      }
    } else {
      // Normal login
      if (!loginValue.trim() || !password) {
        localError = t("auth.enterLoginAndPassword");
        return;
      }

      const success = await login(loginValue.trim(), password);
      if (success) {
        if (onSuccess) {
          onSuccess();
        } else {
          goto(redirectTo);
        }
      } else {
        localError = error() || t("auth.invalidCredentials");
      }
    }
  }
</script>

<form class="login-form" onsubmit={handleSubmit}>
  <h2>{t("auth.signInTitle")}</h2>

  {#if apiKeyEnabled}
    <div class="auth-mode-toggle">
      <button
        type="button"
        class="mode-btn"
        class:active={!useApiKey}
        onclick={() => (useApiKey = false)}
      >
        {t("auth.loginPasswordMode")}
      </button>
      <button
        type="button"
        class="mode-btn"
        class:active={useApiKey}
        onclick={() => (useApiKey = true)}
      >
        {t("auth.apiKeyMode")}
      </button>
    </div>
  {/if}

  {#if useApiKey && apiKeyEnabled}
    <div class="form-group">
      <label for="api-key">{t("auth.apiKeyLabel")}</label>
      <input
        type="password"
        id="api-key"
        bind:value={apiKeyValue}
        placeholder={t("auth.apiKeyPlaceholder")}
        required
      />
    </div>
  {:else}
    <div class="form-group">
      <label for="login">{t("auth.loginLabel")}</label>
      <input
        type="text"
        id="login"
        name="login"
        bind:value={loginValue}
        placeholder={t("auth.loginPlaceholder")}
        required
      />
    </div>

    <div class="form-group">
      <label for="password">{t("auth.passwordLabel")}</label>
      <input
        type="password"
        id="password"
        name="password"
        bind:value={password}
        placeholder={t("auth.passwordPlaceholder")}
        required
      />
    </div>
  {/if}

  {#if localError}
    <ApiErrorDisplay error={{ message: localError, code: "AUTH_ERROR" }} />
  {/if}

  <Button type="submit" variant="primary" disabled={isLoading()}>
    {isLoading() ? t("auth.signingInButton") : t("auth.signInButton")}
  </Button>

  <div class="form-links">
    <a
      href="/auth/register"
      onclick={(e: MouseEvent) => {
        if (onRegister) {
          e.preventDefault();
          onRegister();
        }
      }}
    >
      {t("auth.registerLink")}
    </a>
    <a href="/auth/forgot-password">{t("auth.forgotPasswordLink")}</a>
    <a href="/">{t("backButton.back")}</a>
  </div>

  {#if yandexEnabled && !useApiKey}
    <div class="divider">
      <span>{t("auth.orDivider")}</span>
    </div>
    <YandexLoginButton />
  {/if}
</form>

<style>
  .login-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 100%;
  }

  h2 {
    margin: 0 0 1rem;
    text-align: center;
    color: var(--color-text-dark, #e0e0e0);
    font-size: 1.5rem;
    font-weight: 600;
    letter-spacing: 0.02em;
  }

  .auth-mode-toggle {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
  }

  .mode-btn {
    flex: 1;
    padding: 0.5rem;
    border: 1px solid rgba(255, 255, 255, 0.1);
    background: rgba(255, 255, 255, 0.05);
    color: var(--color-text-dark, #94a3b8);
    border-radius: var(--radius-md, 8px);
    cursor: pointer;
    transition: all 0.3s ease;
    font-size: 0.875rem;
  }

  .mode-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    border-color: rgba(255, 204, 0, 0.3);
  }

  .mode-btn.active {
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    border-color: transparent;
    box-shadow: 0 4px 15px rgba(59, 130, 246, 0.3);
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-dark, #94a3b8);
  }

  input {
    padding: 0.875rem 1rem;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: var(--radius-md, 8px);
    background: rgba(0, 0, 0, 0.2);
    color: var(--color-text-dark, #e0e0e0);
    font-size: 1rem;
    transition: all 0.3s ease;
  }

  input::placeholder {
    color: rgba(255, 255, 255, 0.3);
  }

  input:focus {
    outline: none;
    border-color: rgba(255, 204, 0, 0.5);
    box-shadow:
      0 0 0 3px rgba(255, 204, 0, 0.1),
      0 0 15px rgba(255, 204, 0, 0.1);
    background: rgba(0, 0, 0, 0.3);
  }

  .form-links {
    display: flex;
    justify-content: space-between;
    margin-top: 1rem;
    font-size: 0.875rem;
  }

  a {
    color: var(--color-glow-blue, #40a9ff);
    text-decoration: none;
    transition: all 0.2s ease;
    position: relative;
  }

  a:hover {
    color: var(--color-glow, #ffcc00);
    text-shadow: 0 0 10px rgba(255, 204, 0, 0.5);
  }

  .divider {
    display: flex;
    align-items: center;
    text-align: center;
    margin: 0.5rem 0;
  }

  .divider::before,
  .divider::after {
    content: "";
    flex: 1;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .divider span {
    padding: 0 0.75rem;
    color: var(--color-text-dark, #64748b);
    font-size: 0.875rem;
  }
</style>
