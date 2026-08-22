<script lang="ts">
  import Button from "$components/atoms/Button.svelte";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import { forgotPassword } from "$shared/api/auth";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  let email = $state("");
  let isLoading = $state(false);
  let isSent = $state(false);
  let localError = $state<string | null>(null);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    localError = null;

    if (!email.trim()) {
      localError = t("auth.emailRequired");
      return;
    }

    isLoading = true;

    try {
      await forgotPassword(email.trim());
      isSent = true;
    } catch (e) {
      localError = e instanceof Error ? e.message : t("auth.forgotPasswordFailed");
    } finally {
      isLoading = false;
    }
  }
</script>

<form class="forgot-form" onsubmit={handleSubmit}>
  <h2>{t("auth.forgotPasswordTitle")}</h2>

  {#if isSent}
    <div class="success-message">
      <p>✅ {t("auth.forgotPasswordSuccess1")}</p>
      <p>{t("auth.forgotPasswordSuccess2")}</p>
    </div>
    <a href="/auth/login" class="back-link">{t("auth.backToLogin")}</a>
  {:else}
    <p class="description">
      {t("auth.forgotPasswordDescription")}
    </p>

    <div class="form-group">
      <label for="email">{t("auth.emailLabel")}</label>
      <input
        type="email"
        id="email"
        bind:value={email}
        placeholder={t("auth.emailPlaceholder")}
        required
      />
    </div>

    {#if localError}
      <ApiErrorDisplay error={{ message: localError, code: "FORGOT_PASSWORD_ERROR" }} />
    {/if}

    <Button type="submit" variant="primary" disabled={isLoading}>
      {isLoading ? t("auth.sendingButton") : t("auth.sendButton")}
    </Button>

    <div class="form-links">
      <a href="/auth/login">{t("auth.rememberedPasswordLink")}</a>
    </div>
  {/if}
</form>

<style>
  .forgot-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 100%;
  }

  h2 {
    margin: 0 0 0.5rem;
    text-align: center;
    color: var(--carbon-text, #f0f0f5);
    font-size: 1.5rem;
    font-weight: 600;
    letter-spacing: 0.02em;
  }

  .description {
    text-align: center;
    color: var(--carbon-text-muted, #8b8b9e);
    margin: 0 0 1rem;
    font-size: 0.875rem;
  }

  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--carbon-text-muted, #8b8b9e);
  }

  input {
    padding: 0.875rem 1rem;
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: var(--radius-md, 8px);
    background: rgba(0, 0, 0, 0.2);
    color: var(--carbon-text, #f0f0f5);
    font-size: 1rem;
    transition: all 0.3s ease;
  }

  input::placeholder {
    color: var(--carbon-text-dim, #5a5a6e);
  }

  input:focus {
    outline: none;
    border-color: rgba(34, 211, 238, 0.5);
    box-shadow: var(--carbon-focus-ring, 0 0 0 3px rgba(34, 211, 238, 0.15));
    background: rgba(0, 0, 0, 0.3);
  }

  .success-message {
    text-align: center;
    padding: 1rem;
    background: rgba(52, 211, 153, 0.1);
    border-radius: var(--radius-md, 8px);
    border: 1px solid rgba(52, 211, 153, 0.3);
    box-shadow: 0 0 20px rgba(52, 211, 153, 0.1);
  }

  .success-message p {
    margin: 0.5rem 0;
    color: var(--carbon-text, #f0f0f5);
  }

  .back-link {
    text-align: center;
    color: var(--carbon-glow-amber, #f59e0b);
    text-decoration: none;
    font-weight: 500;
    transition: all 0.2s ease;
  }

  .back-link:hover {
    text-shadow: 0 0 10px rgba(245, 158, 11, 0.5);
  }

  .form-links {
    text-align: center;
    margin-top: 1rem;
    font-size: 0.875rem;
  }

  a {
    color: var(--carbon-glow-cyan, #22d3ee);
    text-decoration: none;
    transition: all 0.2s ease;
  }

  a:hover {
    color: var(--carbon-glow-amber, #f59e0b);
    text-shadow: 0 0 10px rgba(245, 158, 11, 0.5);
  }
</style>
