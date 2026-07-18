<script lang="ts">
  import { goto } from "$app/navigation";
  import Button from "$components/atoms/Button.svelte";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import { resetPassword } from "$shared/api/auth";

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
      errors.push("Minimum 10 characters");
    }
    if (!/[A-Z]/.test(newPassword)) {
      errors.push("Uppercase letter");
    }
    if (!/[a-z]/.test(newPassword)) {
      errors.push("Lowercase letter");
    }
    if (!/[0-9]/.test(newPassword)) {
      errors.push("Number");
    }
    if (!/[!@#$%^&*]/.test(newPassword)) {
      errors.push("Special character");
    }
    return errors;
  });

  const isPasswordValid = $derived(passwordErrors().length === 0);
  const passwordsMatch = $derived(
    newPassword === confirmPassword && confirmPassword.length > 0,
  );

  async function handleSubmit(e: Event) {
    e.preventDefault();
    localError = null;

    if (!isPasswordValid) {
      localError = "Password does not meet requirements";
      return;
    }

    if (newPassword !== confirmPassword) {
      localError = "Passwords do not match";
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
      localError = e instanceof Error ? e.message : "Failed to reset password";
    } finally {
      isLoading = false;
    }
  }
</script>

<form class="reset-form" onsubmit={handleSubmit}>
  <h2>Reset Password</h2>

  {#if isSuccess}
    <div class="success-message">
      <p>✅ Password changed successfully!</p>
      <p>You will be redirected to the login page...</p>
    </div>
  {:else}
    <div class="form-group">
      <label for="new-password">New Password *</label>
      <input
        type="password"
        id="new-password"
        bind:value={newPassword}
        placeholder="Create a new password"
        required
      />

      {#if newPassword.length > 0}
        <div class="password-requirements">
          <p>Password requirements:</p>
          <ul>
            <li class:valid={newPassword.length >= 10}>
              Minimum 10 characters
            </li>
            <li class:valid={/[A-Z]/.test(newPassword)}>Uppercase letter</li>
            <li class:valid={/[a-z]/.test(newPassword)}>Lowercase letter</li>
            <li class:valid={/[0-9]/.test(newPassword)}>Number</li>
            <li class:valid={/[!@#$%^&*]/.test(newPassword)}>
              Special character
            </li>
          </ul>
        </div>
      {/if}
    </div>

    <div class="form-group">
      <label for="confirm-password">Confirm Password *</label>
      <input
        type="password"
        id="confirm-password"
        bind:value={confirmPassword}
        placeholder="Repeat new password"
        required
      />
      {#if confirmPassword && !passwordsMatch}
        <span class="error-text">Passwords do not match</span>
      {/if}
    </div>

    {#if localError}
      <ApiErrorDisplay
        error={{ message: localError, code: "RESET_PASSWORD_ERROR" }}
      />
    {/if}

    <Button
      type="submit"
      variant="primary"
      disabled={isLoading || !isPasswordValid || !passwordsMatch}
    >
      {isLoading ? "Saving..." : "Save New Password"}
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
    color: var(--color-text-dark, #e0e0e0);
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

  .password-requirements {
    font-size: 0.75rem;
    color: var(--color-text-dark, #94a3b8);
    margin-top: 0.5rem;
    padding: 0.75rem;
    background: rgba(0, 0, 0, 0.2);
    border-radius: var(--radius-md, 8px);
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .password-requirements p {
    margin: 0 0 0.25rem;
    font-weight: 500;
    color: var(--color-text-dark, #e0e0e0);
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
    color: #22c55e;
    text-shadow: 0 0 8px rgba(34, 197, 94, 0.3);
  }

  .error-text {
    font-size: 0.75rem;
    color: #ef4444;
    text-shadow: 0 0 8px rgba(239, 68, 68, 0.3);
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
</style>
