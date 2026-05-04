<script lang="ts">
  import { goto } from '$app/navigation';
  import ForgotPasswordForm from '$lib/components/ForgotPasswordForm.svelte';
  import { isAuthenticated } from '$lib/stores/auth.svelte.js';
  
  // Redirect if already authenticated
  $effect(() => {
    if (isAuthenticated()) {
      goto('/');
    }
  });
</script>

<div class="forgot-password-page">
  <div class="forgot-password-container">
    <div class="logo">
      <h1>Восстановление пароля</h1>
      <p>Введите email для получения ссылки</p>
    </div>
    
    <ForgotPasswordForm />
  </div>
</div>

<style>
  .forgot-password-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-background);
    padding: 2rem;
  }
  
  .forgot-password-container {
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
</style>
