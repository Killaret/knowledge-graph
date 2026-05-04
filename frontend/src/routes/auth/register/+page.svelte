<script lang="ts">
  import { goto } from '$app/navigation';
  import RegisterForm from '$lib/components/RegisterForm.svelte';
  import { isAuthenticated, initAuth } from '$lib/stores/auth.svelte.js';
  
  // Redirect if already authenticated
  $effect(() => {
    if (isAuthenticated()) {
      goto('/');
    }
  });
  
  // Initialize auth on mount
  $effect(() => {
    initAuth();
  });
</script>

<div class="register-page">
  <div class="register-container">
    <div class="logo">
      <h1>Создать аккаунт</h1>
      <p>Присоединитесь к вселенной знаний</p>
    </div>
    
    <RegisterForm />
  </div>
</div>

<style>
  .register-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-background);
    padding: 2rem;
  }
  
  .register-container {
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
    font-size: 1.75rem;
    font-weight: 700;
  }
  
  .logo p {
    margin: 0;
    color: var(--color-text-secondary);
    font-size: 1rem;
  }
</style>
