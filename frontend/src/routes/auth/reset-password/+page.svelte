<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import ResetPasswordForm from '$lib/components/ResetPasswordForm.svelte';
  import { isAuthenticated } from '$lib/stores/auth.svelte.js';
  
  // Get token from URL
  let token = $state('');
  
  $effect(() => {
    token = $page.url.searchParams.get('token') || '';
  });
  
  // Redirect if already authenticated
  $effect(() => {
    if (isAuthenticated()) {
      goto('/');
    }
  });
</script>

<div class="reset-password-page">
  <div class="reset-password-container">
    <div class="logo">
      <h1>Сброс пароля</h1>
      <p>Создайте новый пароль для аккаунта</p>
    </div>
    
    {#if token}
      <ResetPasswordForm {token} />
    {:else}
      <div class="error-message">
        <p>❌ Токен сброса пароля не найден.</p>
        <p>Пожалуйста, запросите новую ссылку для сброса пароля.</p>
        <a href="/auth/forgot-password" class="back-link">Запросить сброс пароля</a>
      </div>
    {/if}
  </div>
</div>

<style>
  .reset-password-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-background);
    padding: 2rem;
  }
  
  .reset-password-container {
    width: 100%;
    max-width: 480px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2rem;
  }
  
  .logo {
    text-align: center;
    color: var(--color-text-primary);
  }
  
  .logo h1 {
    margin: 0 0 0.5rem;
    font-size: 1.5rem;
    font-weight: 700;
  }
  
  .logo p {
    margin: 0;
    color: var(--color-text-secondary);
    font-size: 1rem;
  }
  
  .error-message {
    text-align: center;
    padding: 2rem;
    background: var(--color-surface);
    border-radius: var(--radius-lg);
    color: var(--color-text-primary);
  }
  
  .error-message p {
    margin: 0.5rem 0;
  }
  
  .back-link {
    display: inline-block;
    margin-top: 1rem;
    padding: 0.75rem 1.5rem;
    background: var(--color-primary);
    color: white;
    text-decoration: none;
    border-radius: var(--radius-md);
    font-weight: 500;
  }
</style>
