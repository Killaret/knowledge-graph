<script lang="ts">
  import { goto } from "$app/navigation";
  import Button from "$components/atoms/Button.svelte";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import { register, isLoading, error } from "$shared/stores/auth.svelte.js";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  interface Props {
    redirectTo?: string;
    onSuccess?: () => void;
    onLogin?: () => void;
  }
  const { redirectTo = "/", onSuccess, onLogin }: Props = $props();

  let login = $state("");
  let email = $state("");
  let password = $state("");
  let confirmPassword = $state("");
  let localError = $state<string | null>(null);

  // Password validation
  const passwordErrors = $derived(() => {
    const errors: string[] = [];
    if (password.length < 10) {
      errors.push(t("auth.passwordMinChars"));
    }
    if (!/[A-Z]/.test(password)) {
      errors.push(t("auth.passwordUppercase"));
    }
    if (!/[a-z]/.test(password)) {
      errors.push(t("auth.passwordLowercase"));
    }
    if (!/[0-9]/.test(password)) {
      errors.push(t("auth.passwordNumber"));
    }
    if (!/[!@#$%^&*]/.test(password)) {
      errors.push(t("auth.passwordSpecial"));
    }
    return errors;
  });

  const isPasswordValid = $derived(passwordErrors().length === 0);
  const passwordsMatch = $derived(password === confirmPassword && confirmPassword.length > 0);

  async function handleSubmit(e: Event) {
    e.preventDefault();
    localError = null;

    // Validation
    if (!login.trim()) {
      localError = t("auth.loginRequired");
      return;
    }

    if (!email.trim()) {
      localError = t("auth.emailRequired");
      return;
    }

    if (!isPasswordValid) {
      localError = t("auth.passwordRequirementsNotMet");
      return;
    }

    if (password !== confirmPassword) {
      localError = t("auth.passwordsDoNotMatch");
      return;
    }

    const success = await register(login.trim(), password, email.trim());
    if (success) {
      if (onSuccess) {
        onSuccess();
      } else {
        goto(redirectTo);
      }
    } else {
      localError = error() || t("auth.registrationFailed");
    }
  }
</script>

<form class="register-form" onsubmit={handleSubmit}>
  <h2>{t("auth.registerTitle")}</h2>

  <div class="form-group">
    <label for="login">{t("auth.loginLabel")} *</label>
    <input
      type="text"
      id="login"
      bind:value={login}
      placeholder={t("auth.chooseLoginPlaceholder")}
      required
      minlength="3"
    />
  </div>

  <div class="form-group">
    <label for="email">{t("auth.emailLabel")} *</label>
    <input
      type="email"
      id="email"
      bind:value={email}
      placeholder={t("auth.enterEmailPlaceholder")}
      required
    />
  </div>

  <div class="form-group">
    <label for="password">{t("auth.passwordLabel")} *</label>
    <input
      type="password"
      id="password"
      bind:value={password}
      placeholder={t("auth.createPasswordPlaceholder")}
      required
    />

    {#if password.length > 0}
      <div class="password-requirements">
        <p>{t("auth.passwordRequirementsTitle")}</p>
        <ul>
          <li class:valid={password.length >= 10}>
            {t("auth.passwordMinChars")}
          </li>
          <li class:valid={/[A-Z]/.test(password)}>
            {t("auth.passwordUppercase")}
          </li>
          <li class:valid={/[a-z]/.test(password)}>
            {t("auth.passwordLowercase")}
          </li>
          <li class:valid={/[0-9]/.test(password)}>
            {t("auth.passwordNumber")}
          </li>
          <li class:valid={/[!@#$%^&*]/.test(password)}>
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
    <ApiErrorDisplay error={{ message: localError, code: "REGISTER_ERROR" }} />
  {/if}

  <Button
    type="submit"
    variant="primary"
    disabled={isLoading() || !isPasswordValid || !passwordsMatch}
  >
    {isLoading() ? t("auth.registeringButton") : t("auth.registerButton")}
  </Button>

  <div class="form-links">
    <a
      href="/auth/login"
      onclick={(e: MouseEvent) => {
        if (onLogin) {
          e.preventDefault();
          onLogin();
        }
      }}
    >
      {t("auth.alreadyHaveAccount")}
    </a>
  </div>
</form>

<style>
  .register-form {
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
