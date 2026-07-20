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
      localError =
        e instanceof Error ? e.message : t("auth.forgotPasswordFailed");
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
      <ApiErrorDisplay
        error={{ message: localError, code: "FORGOT_PASSWORD_ERROR" }}
      />
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
    color: var(--color-text-dark, #e0e0e0);
    font-size: 1.5rem;
    font-weight: 600;
    letter-spacing: 0.02em;
  }

  .description {
    text-align: center;
    color: var(--color-text-dark, #94a3b8);
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

  .success-message {
    text-align: center;
    padding: 1rem;
    background: rgba(34, 197, 94, 0.1);
    border-radius: var(--radius-md, 8px);
    border: 1px solid rgba(34, 197, 94, 0.3);
    box-shadow: 0 0 20px rgba(34, 197, 94, 0.1);
  }

  .success-message p {
    margin: 0.5rem 0;
    color: var(--color-text-dark, #e0e0e0);
  }

  .back-link {
    text-align: center;
    color: var(--color-glow, #ffcc00);
    text-decoration: none;
    font-weight: 500;
    transition: all 0.2s ease;
  }

  .back-link:hover {
    text-shadow: 0 0 10px rgba(255, 204, 0, 0.5);
  }

  .form-links {
    text-align: center;
    margin-top: 1rem;
    font-size: 0.875rem;
  }

  a {
    color: var(--color-glow-blue, #40a9ff);
    text-decoration: none;
    transition: all 0.2s ease;
  }

  a:hover {
    color: var(--color-glow, #ffcc00);
    text-shadow: 0 0 10px rgba(255, 204, 0, 0.5);
  }
</style>
