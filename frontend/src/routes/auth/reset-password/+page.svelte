<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import ResetPasswordForm from '$lib/components/ResetPasswordForm.svelte';
  import AuthCard from '$lib/components/AuthCard.svelte';
  import ConstellationIcon from '$lib/components/ConstellationIcon.svelte';
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

{#if token}
  <AuthCard 
    title="Сброс пароля" 
    subtitle="Создайте новый пароль для аккаунта"
    showIcon={true}
  >
    <ResetPasswordForm {token} />
  </AuthCard>
{:else}
  <AuthCard 
    title="Ошибка" 
    subtitle="Токен сброса пароля не найден"
    showIcon={false}
  >
    <div class="error-content">
      <ConstellationIcon size={48} class="error-icon" />
      <p class="error-text">Пожалуйста, запросите новую ссылку для сброса пароля.</p>
      <a href="/auth/forgot-password" class="back-link">Запросить сброс пароля</a>
    </div>
  </AuthCard>
{/if}

<style>
  .error-content {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }
  
  .error-content :global(.error-icon) {
    opacity: 0.6;
  }
  
  .error-text {
    margin: 0;
    color: var(--color-text-secondary, #94a3b8);
    font-size: 0.875rem;
  }
  
  .back-link {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    text-decoration: none;
    border-radius: 8px;
    font-weight: 500;
    transition: all 0.2s ease;
  }
  
  .back-link:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 20px rgba(59, 130, 246, 0.4);
  }
</style>
