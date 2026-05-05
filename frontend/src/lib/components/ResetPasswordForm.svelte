<script lang="ts">
  import { goto } from '$app/navigation';
  import Button from './Button.svelte';
  import ApiErrorDisplay from './ApiErrorDisplay.svelte';
  import { resetPassword } from '$lib/api/auth';
  
  interface Props {
    token: string;
  }
  
  let { token }: Props = $props();
  
  let newPassword = $state('');
  let confirmPassword = $state('');
  let isLoading = $state(false);
  let isSuccess = $state(false);
  let localError = $state<string | null>(null);
  
  // Password validation
  let passwordErrors = $derived(() => {
    const errors: string[] = [];
    if (newPassword.length < 10) {
      errors.push('Минимум 10 символов');
    }
    if (!/[A-Z]/.test(newPassword)) {
      errors.push('Заглавная буква');
    }
    if (!/[a-z]/.test(newPassword)) {
      errors.push('Строчная буква');
    }
    if (!/[0-9]/.test(newPassword)) {
      errors.push('Цифра');
    }
    if (!/[!@#$%^&*]/.test(newPassword)) {
      errors.push('Специальный символ');
    }
    return errors;
  });
  
  let isPasswordValid = $derived(passwordErrors().length === 0);
  let passwordsMatch = $derived(newPassword === confirmPassword && confirmPassword.length > 0);
  
  async function handleSubmit(e: Event) {
    e.preventDefault();
    localError = null;
    
    if (!isPasswordValid) {
      localError = 'Пароль не соответствует требованиям';
      return;
    }
    
    if (newPassword !== confirmPassword) {
      localError = 'Пароли не совпадают';
      return;
    }
    
    isLoading = true;
    
    try {
      await resetPassword(token, newPassword);
      isSuccess = true;
      
      // Redirect to login after 3 seconds
      setTimeout(() => {
        goto('/auth/login');
      }, 3000);
    } catch (e) {
      localError = e instanceof Error ? e.message : 'Failed to reset password';
    } finally {
      isLoading = false;
    }
  }
</script>

<form class="reset-form" onsubmit={handleSubmit}>
  <h2>Сброс пароля</h2>
  
  {#if isSuccess}
    <div class="success-message">
      <p>✅ Пароль успешно изменен!</p>
      <p>Вы будете перенаправлены на страницу входа...</p>
    </div>
  {:else}
    <div class="form-group">
      <label for="new-password">Новый пароль *</label>
      <input
        type="password"
        id="new-password"
        bind:value={newPassword}
        placeholder="Придумайте новый пароль"
        required
      />
      
      {#if newPassword.length > 0}
        <div class="password-requirements">
          <p>Требования к паролю:</p>
          <ul>
            <li class:valid={newPassword.length >= 10}>Минимум 10 символов</li>
            <li class:valid={/[A-Z]/.test(newPassword)}>Заглавная буква</li>
            <li class:valid={/[a-z]/.test(newPassword)}>Строчная буква</li>
            <li class:valid={/[0-9]/.test(newPassword)}>Цифра</li>
            <li class:valid={/[!@#$%^&*]/.test(newPassword)}>Специальный символ</li>
          </ul>
        </div>
      {/if}
    </div>
    
    <div class="form-group">
      <label for="confirm-password">Подтвердите пароль *</label>
      <input
        type="password"
        id="confirm-password"
        bind:value={confirmPassword}
        placeholder="Повторите новый пароль"
        required
      />
      {#if confirmPassword && !passwordsMatch}
        <span class="error-text">Пароли не совпадают</span>
      {/if}
    </div>
    
    {#if localError}
      <ApiErrorDisplay error={{ message: localError, code: 'RESET_PASSWORD_ERROR' }} />
    {/if}
    
    <Button 
      type="submit" 
      variant="primary" 
      disabled={isLoading || !isPasswordValid || !passwordsMatch}
    >
      {isLoading ? 'Сохранение...' : 'Сохранить новый пароль'}
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
    box-shadow: 0 0 0 3px rgba(255, 204, 0, 0.1), 0 0 15px rgba(255, 204, 0, 0.1);
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
