<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import ResetPasswordForm from '$components/organisms/ResetPasswordForm.svelte';
  import AuthCard from '$components/organisms/AuthCard.svelte';
  import ConstellationIcon from '$components/atoms/ConstellationIcon.svelte';
  import { isAuthenticated } from '$shared/stores/auth.svelte.js';
  
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
    title="Reset Password" 
    subtitle="Create a new password for your account"
    showIcon={true}
  >
    <ResetPasswordForm {token} />
  </AuthCard>
{:else}
  <AuthCard 
    title="Error" 
    subtitle="Reset password token not found"
    showIcon={false}
  >
    <div class="error-content">
      <ConstellationIcon size={48} class="error-icon" />
      <p class="error-text">Please request a new password reset link.</p>
      <a href="/auth/forgot-password" class="back-link">Request password reset</a>
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
