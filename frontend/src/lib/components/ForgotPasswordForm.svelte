<script lang="ts">
  import Button from './Button.svelte';
  import ApiErrorDisplay from './ApiErrorDisplay.svelte';
  import { forgotPassword } from '$lib/api/auth';
  
  let email = $state('');
  let isLoading = $state(false);
  let isSent = $state(false);
  let localError = $state<string | null>(null);
  
  async function handleSubmit(e: Event) {
    e.preventDefault();
    localError = null;
    
    if (!email.trim()) {
      localError = 'Введите email';
      return;
    }
    
    isLoading = true;
    
    try {
      await forgotPassword(email.trim());
      isSent = true;
    } catch (e) {
      localError = e instanceof Error ? e.message : 'Failed to send reset email';
    } finally {
      isLoading = false;
    }
  }
</script>

<form class="forgot-form" onsubmit={handleSubmit}>
  <h2>Восстановление пароля</h2>
  
  {#if isSent}
    <div class="success-message">
      <p>✅ Письмо для сброса пароля отправлено на указанный email.</p>
      <p>Проверьте вашу почту и следуйте инструкциям.</p>
    </div>
    <a href="/auth/login" class="back-link">Вернуться к входу</a>
  {:else}
    <p class="description">
      Введите ваш email, и мы отправим вам ссылку для сброса пароля.
    </p>
    
    <div class="form-group">
      <label for="email">Email</label>
      <input
        type="email"
        id="email"
        bind:value={email}
        placeholder="Введите ваш email"
        required
      />
    </div>
    
    {#if localError}
      <ApiErrorDisplay error={{ message: localError, code: 'FORGOT_PASSWORD_ERROR' }} />
    {/if}
    
    <Button type="submit" variant="primary" disabled={isLoading}>
      {isLoading ? 'Отправка...' : 'Отправить'}
    </Button>
    
    <div class="form-links">
      <a href="/auth/login">Вспомнили пароль? Войти</a>
    </div>
  {/if}
</form>

<style>
  .forgot-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-width: 400px;
    width: 100%;
    padding: 2rem;
    background: var(--color-surface);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-md);
  }
  
  h2 {
    margin: 0 0 0.5rem;
    text-align: center;
    color: var(--color-text-primary);
  }
  
  .description {
    text-align: center;
    color: var(--color-text-secondary);
    margin: 0 0 1rem;
    font-size: 0.875rem;
  }
  
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  
  label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-secondary);
  }
  
  input {
    padding: 0.75rem;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-background);
    color: var(--color-text-primary);
    font-size: 1rem;
    transition: border-color 0.2s;
  }
  
  input:focus {
    outline: none;
    border-color: var(--color-primary);
  }
  
  .success-message {
    text-align: center;
    padding: 1rem;
    background: var(--color-success-light, rgba(34, 197, 94, 0.1));
    border-radius: var(--radius-md);
    border: 1px solid var(--color-success);
  }
  
  .success-message p {
    margin: 0.5rem 0;
    color: var(--color-text-primary);
  }
  
  .back-link {
    text-align: center;
    color: var(--color-primary);
    text-decoration: none;
    font-weight: 500;
  }
  
  .back-link:hover {
    text-decoration: underline;
  }
  
  .form-links {
    text-align: center;
    margin-top: 1rem;
    font-size: 0.875rem;
  }
  
  a {
    color: var(--color-primary);
    text-decoration: none;
  }
  
  a:hover {
    text-decoration: underline;
  }
</style>
