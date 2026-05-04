<script lang="ts">
  import { goto } from '$app/navigation';
  import Button from './Button.svelte';
  import ApiErrorDisplay from './ApiErrorDisplay.svelte';
  import YandexLoginButton from './YandexLoginButton.svelte';
  import { login, loginWithApiKey, isLoading, error } from '$lib/stores/auth.svelte.js';
  
  // Check if API Key mode is enabled
  const apiKeyEnabled = import.meta.env.VITE_API_KEY_ENABLED === 'true';
  const yandexEnabled = import.meta.env.VITE_YANDEX_ENABLED === 'true';
  
  let loginValue = $state('');
  let password = $state('');
  let apiKeyValue = $state('');
  let useApiKey = $state(false);
  let localError = $state<string | null>(null);
  
  async function handleSubmit(e: Event) {
    e.preventDefault();
    localError = null;
    
    if (useApiKey && apiKeyEnabled) {
      // Login with API Key
      const success = await loginWithApiKey(apiKeyValue.trim());
      if (success) {
        goto('/');
      } else {
        localError = error || 'Invalid API key';
      }
    } else {
      // Normal login
      if (!loginValue.trim() || !password) {
        localError = 'Please enter login and password';
        return;
      }
      
      const success = await login(loginValue.trim(), password);
      if (success) {
        goto('/');
      } else {
        localError = error || 'Invalid credentials';
      }
    }
  }
  
  function toggleApiKeyMode() {
    useApiKey = !useApiKey;
    localError = null;
  }
</script>

<form class="login-form" onsubmit={handleSubmit}>
  <h2>Вход в систему</h2>
  
  {#if apiKeyEnabled}
    <div class="auth-mode-toggle">
      <button 
        type="button" 
        class="mode-btn" 
        class:active={!useApiKey}
        onclick={() => useApiKey = false}
      >
        Логин / Пароль
      </button>
      <button 
        type="button" 
        class="mode-btn" 
        class:active={useApiKey}
        onclick={() => useApiKey = true}
      >
        API Key
      </button>
    </div>
  {/if}
  
  {#if useApiKey && apiKeyEnabled}
    <div class="form-group">
      <label for="api-key">API Key</label>
      <input
        type="password"
        id="api-key"
        bind:value={apiKeyValue}
        placeholder="Введите ваш API ключ"
        required
      />
    </div>
  {:else}
    <div class="form-group">
      <label for="login">Логин</label>
      <input
        type="text"
        id="login"
        bind:value={loginValue}
        placeholder="Введите логин"
        required
      />
    </div>
    
    <div class="form-group">
      <label for="password">Пароль</label>
      <input
        type="password"
        id="password"
        bind:value={password}
        placeholder="Введите пароль"
        required
      />
    </div>
  {/if}
  
  {#if localError}
    <ApiErrorDisplay error={{ message: localError, code: 'AUTH_ERROR' }} />
  {/if}
  
  <Button type="submit" variant="primary" disabled={isLoading}>
    {isLoading ? 'Вход...' : 'Войти'}
  </Button>
  
  <div class="form-links">
    <a href="/auth/register">Регистрация</a>
    <a href="/auth/forgot-password">Забыли пароль?</a>
  </div>
  
  {#if yandexEnabled && !useApiKey}
    <div class="divider">
      <span>или</span>
    </div>
    <YandexLoginButton />
  {/if}
</form>

<style>
  .login-form {
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
    margin: 0 0 1rem;
    text-align: center;
    color: var(--color-text-primary);
  }
  
  .auth-mode-toggle {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }
  
  .mode-btn {
    flex: 1;
    padding: 0.5rem;
    border: 1px solid var(--color-border);
    background: var(--color-surface);
    color: var(--color-text-secondary);
    border-radius: var(--radius-md);
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .mode-btn.active {
    background: var(--color-primary);
    color: white;
    border-color: var(--color-primary);
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
  
  .form-links {
    display: flex;
    justify-content: space-between;
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
  
  .divider {
    display: flex;
    align-items: center;
    text-align: center;
    margin: 0.5rem 0;
  }
  
  .divider::before,
  .divider::after {
    content: '';
    flex: 1;
    border-bottom: 1px solid var(--color-border);
  }
  
  .divider span {
    padding: 0 0.75rem;
    color: var(--color-text-secondary);
    font-size: 0.875rem;
  }
</style>
