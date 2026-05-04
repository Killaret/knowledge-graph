<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import { handleYandexCallback, isLoading, error } from '$lib/stores/auth.svelte.js';
  
  let isProcessing = $state(true);
  let localError = $state<string | null>(null);
  
  onMount(async () => {
    const code = $page.url.searchParams.get('code');
    const state = $page.url.searchParams.get('state');
    
    if (!code || !state) {
      localError = 'Отсутствуют необходимые параметры авторизации';
      isProcessing = false;
      return;
    }
    
    const success = await handleYandexCallback(code, state);
    
    if (success) {
      goto('/');
    } else {
      localError = error || 'Ошибка авторизации через Яндекс';
      isProcessing = false;
    }
  });
</script>

<div class="yandex-callback-page">
  <div class="callback-container">
    {#if isProcessing}
      <div class="loading">
        <div class="spinner"></div>
        <p>Выполняется вход через Яндекс...</p>
      </div>
    {:else if localError}
      <div class="error">
        <p>❌ {localError}</p>
        <a href="/auth/login" class="back-link">Вернуться к входу</a>
      </div>
    {/if}
  </div>
</div>

<style>
  .yandex-callback-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--color-background);
    padding: 2rem;
  }
  
  .callback-container {
    text-align: center;
    color: var(--color-text-primary);
  }
  
  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }
  
  .spinner {
    width: 48px;
    height: 48px;
    border: 4px solid var(--color-border);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  
  .error {
    padding: 2rem;
    background: var(--color-surface);
    border-radius: var(--radius-lg);
  }
  
  .error p {
    margin: 0 0 1rem;
    color: var(--color-error);
  }
  
  .back-link {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    background: var(--color-primary);
    color: white;
    text-decoration: none;
    border-radius: var(--radius-md);
    font-weight: 500;
  }
</style>
