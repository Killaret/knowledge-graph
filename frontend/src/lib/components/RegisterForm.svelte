<script lang="ts">
  import { goto } from '$app/navigation';
  import Button from './Button.svelte';
  import ApiErrorDisplay from './ApiErrorDisplay.svelte';
  import { register, isLoading, error } from '$lib/stores/auth.svelte.js';
  
  let login = $state('');
  let email = $state('');
  let password = $state('');
  let confirmPassword = $state('');
  let localError = $state<string | null>(null);
  
  // Password validation
  const passwordErrors = $derived(() => {
    const errors: string[] = [];
    if (password.length < 10) {
      errors.push('Минимум 10 символов');
    }
    if (!/[A-Z]/.test(password)) {
      errors.push('Заглавная буква');
    }
    if (!/[a-z]/.test(password)) {
      errors.push('Строчная буква');
    }
    if (!/[0-9]/.test(password)) {
      errors.push('Цифра');
    }
    if (!/[!@#$%^&*]/.test(password)) {
      errors.push('Специальный символ (!@#$%^&*)');
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
      localError = 'Введите логин';
      return;
    }
    
    if (!isPasswordValid) {
      localError = 'Пароль не соответствует требованиям';
      return;
    }
    
    if (password !== confirmPassword) {
      localError = 'Пароли не совпадают';
      return;
    }
    
    const success = await register(login.trim(), password, email.trim() || undefined);
    if (success) {
      goto('/');
    } else {
      localError = error() || 'Registration failed';
    }
  }
</script>

<form class="register-form" onsubmit={handleSubmit}>
  <h2>Регистрация</h2>
  
  <div class="form-group">
    <label for="login">Логин *</label>
    <input
      type="text"
      id="login"
      bind:value={login}
      placeholder="Придумайте логин"
      required
      minlength="3"
    />
  </div>
  
  <div class="form-group">
    <label for="email">Email</label>
    <input
      type="email"
      id="email"
      bind:value={email}
      placeholder="Введите email (необязательно)"
    />
  </div>
  
  <div class="form-group">
    <label for="password">Пароль *</label>
    <input
      type="password"
      id="password"
      bind:value={password}
      placeholder="Придумайте пароль"
      required
    />
    
    {#if password.length > 0}
      <div class="password-requirements">
        <p>Требования к паролю:</p>
        <ul>
          <li class:valid={password.length >= 10}>Минимум 10 символов</li>
          <li class:valid={/[A-Z]/.test(password)}>Заглавная буква</li>
          <li class:valid={/[a-z]/.test(password)}>Строчная буква</li>
          <li class:valid={/[0-9]/.test(password)}>Цифра</li>
          <li class:valid={/[!@#$%^&*]/.test(password)}>Специальный символ</li>
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
      placeholder="Повторите пароль"
      required
    />
    {#if confirmPassword && !passwordsMatch}
      <span class="error-text">Пароли не совпадают</span>
    {/if}
  </div>
  
  {#if localError}
    <ApiErrorDisplay error={{ message: localError, code: 'REGISTER_ERROR' }} />
  {/if}
  
  <Button 
    type="submit" 
    variant="primary" 
    disabled={isLoading() || !isPasswordValid || !passwordsMatch}
  >
    {isLoading() ? 'Регистрация...' : 'Зарегистрироваться'}
  </Button>
  
  <div class="form-links">
    <a href="/auth/login">Уже есть аккаунт? Войти</a>
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
