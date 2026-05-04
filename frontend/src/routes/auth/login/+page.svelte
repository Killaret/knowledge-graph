<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import LoginForm from '$lib/components/LoginForm.svelte';
  import { isAuthenticated, initAuth } from '$lib/stores/auth.svelte.js';
  
  // Redirect if already authenticated
  $effect(() => {
    if (isAuthenticated()) {
      const redirectTo = $page.url.searchParams.get('redirect') || '/';
      goto(redirectTo);
    }
  });
  
  // Initialize auth on mount
  $effect(() => {
    initAuth();
  });
</script>

<div class="login-page">
  <div class="login-container">
    <div class="logo">
      <h1>Knowledge Graph</h1>
      <p>Ваша персональная вселенная знаний</p>
    </div>
    
    <LoginForm />
  </div>
</div>

<style>
  .login-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-background);
    padding: 2rem;
  }
  
  .login-container {
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
    font-size: 2rem;
    font-weight: 700;
  }
  
  .logo p {
    margin: 0;
    color: var(--color-text-secondary);
    font-size: 1rem;
  }
</style>
