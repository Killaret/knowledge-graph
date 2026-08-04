<script lang="ts">
  import { goto } from "$app/navigation";
  import Button from "$components/atoms/Button.svelte";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import { resetPassword } from "$shared/api/auth";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  interface Props {
    token: string;
  }

  const { token }: Props = $props();

  let newPassword = $state("");
  let confirmPassword = $state("");
  let isLoading = $state(false);
  let isSuccess = $state(false);
  let localError = $state<string | null>(null);

  // Password validation
  const passwordErrors = $derived(() => {
    const errors: string[] = [];
    if (newPassword.length < 10) {
      errors.push(t("auth.passwordMinChars"));
    }
    if (!/[A-Z]/.test(newPassword)) {
      errors.push(t("auth.passwordUppercase"));
    }
    if (!/[a-z]/.test(newPassword)) {
      errors.push(t("auth.passwordLowercase"));
    }
    if (!/[0-9]/.test(newPassword)) {
      errors.push(t("auth.passwordNumber"));
    }
    if (!/[!@#$%^&*]/.test(newPassword)) {
      errors.push(t("auth.passwordSpecial"));
    }
    return errors;
  });

  const isPasswordValid = $derived(passwordErrors().length === 0);
  const passwordsMatch = $derived(newPassword === confirmPassword && confirmPassword.length > 0);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    localError = null;

    if (!isPasswordValid) {
      localError = t("auth.passwordRequirementsNotMet");
      return;
    }

    if (newPassword !== confirmPassword) {
      localError = t("auth.passwordsDoNotMatch");
      return;
    }

    isLoading = true;

    try {
      await resetPassword(token, newPassword);
      isSuccess = true;

      // Redirect to login after 3 seconds
      setTimeout(() => {
        goto("/auth/login");
      }, 3000);
    } catch (e) {
      localError = e instanceof Error ? e.message : t("auth.resetPasswordFailed");
    } finally {
      isLoading = false;
    }
  }
</script>

<form class="reset-form" onsubmit={handleSubmit}>
  <h2>{t("auth.resetPasswordTitle")}</h2>

  {#if isSuccess}
    <div class="success-message">
      <p>✅ {t("auth.resetPasswordSuccess")}</p>
      <p>{t("auth.resetPasswordRedirect")}</p>
    </div>
  {:else}
    <div class="form-group">
      <label for="new-password">{t("auth.newPasswordLabel")} *</label>
      <input
        type="password"
        id="new-password"
        bind:value={newPassword}
        placeholder={t("auth.createPasswordPlaceholder")}
        required
      />

      {#if newPassword.length > 0}
        <div class="password-requirements">
          <p>{t("auth.passwordRequirementsTitle")}</p>
          <ul>
            <li class:valid={newPassword.length >= 10}>
              {t("auth.passwordMinChars")}
            </li>
            <li class:valid={/[A-Z]/.test(newPassword)}>
              {t("auth.passwordUppercase")}
            </li>
            <li class:valid={/[a-z]/.test(newPassword)}>
              {t("auth.passwordLowercase")}
            </li>
            <li class:valid={/[0-9]/.test(newPassword)}>
              {t("auth.passwordNumber")}
            </li>
            <li class:valid={/[!@#$%^&*]/.test(newPassword)}>
              {t("auth.passwordSpecial")}
            </li>
          </ul>
        </div>
      {/if}
    </div>

    <div class="form-group">
      <label for="confirm-password">{t("auth.confirmPasswordLabel")} *</label>
      <input
        type="password"
        id="confirm-password"
        bind:value={confirmPassword}
        placeholder={t("auth.repeatPasswordPlaceholder")}
        required
      />
      {#if confirmPassword && !passwordsMatch}
        <span class="error-text">{t("auth.passwordsDoNotMatch")}</span>
      {/if}
    </div>

    {#if localError}
      <ApiErrorDisplay error={{ message: localError, code: "RESET_PASSWORD_ERROR" }} />
    {/if}

    <Button
      type="submit"
      variant="primary"
      disabled={isLoading || !isPasswordValid || !passwordsMatch}
    >
      {isLoading ? t("auth.savingButton") : t("auth.saveNewPasswordButton")}
    </Button>
  {/if}
</form>

<style>
  .reset-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 100%;
  }

  h2 {
    margin: 0 0 1rem;
    text-align: center;
    color: var(--carbon-text, #f0f0f5);
    font-size: 1.5rem;
    font-weight: 600;
    letter-spacing: 0.02em;
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

  .password-requirements {
    font-size: 0.75rem;
    color: var(--carbon-text-muted, #8b8b9e);
    margin-top: 0.5rem;
    padding: 0.75rem;
    background: rgba(0, 0, 0, 0.2);
    border-radius: var(--radius-md, 8px);
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .password-requirements p {
    margin: 0 0 0.25rem;
    font-weight: 500;
    color: var(--carbon-text, #f0f0f5);
  }

  .password-requirements ul {
    margin: 0;
    padding-left: 1rem;
  }

  .password-requirements li {
    color: rgba(255, 255, 255, 0.4);
    transition: color 0.3s ease;
  }

  .password-requirements li.valid {
    color: var(--carbon-glow-green, #34d399);
    text-shadow: 0 0 8px rgba(52, 211, 153, 0.3);
  }

  .error-text {
    font-size: 0.75rem;
    color: var(--carbon-glow-red, #ff3a2f);
    text-shadow: 0 0 8px rgba(255, 58, 47, 0.3);
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
</style>
