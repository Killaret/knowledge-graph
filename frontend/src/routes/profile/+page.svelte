<script lang="ts">
  import { goto } from '$app/navigation';
  import ProfileEditor from '$lib/components/ProfileEditor.svelte';
  import { isAuthenticated, currentUser, initAuth } from '$lib/stores/auth.svelte.js';
  import { onMount } from 'svelte';

  // Initialize auth once on mount
  onMount(() => {
    initAuth();
  });
  
  // Redirect if not authenticated
  $effect(() => {
    if (!isAuthenticated()) {
      goto('/auth/login?redirect=/profile');
    }
  });
</script>

<div class="profile-page">
  <div class="profile-container" data-testid="profile-content">
    <div class="header">
      <h1>Profile</h1>
      <p>Manage your account</p>
    </div>
    
    {#if currentUser()}
      <ProfileEditor />
    {:else}
      <div class="loading">
        <div class="spinner"></div>
        <p>Loading profile...</p>
      </div>
    {/if}
  </div>
</div>

<style>
  .profile-page {
    min-height: 100vh;
    padding: 2rem;
    background: var(--color-background);
  }
  
  .profile-container {
    max-width: 800px;
    margin: 0 auto;
  }
  
  .header {
    margin-bottom: 2rem;
    color: var(--color-text-primary);
  }
  
  .header h1 {
    margin: 0 0 0.5rem;
    font-size: 2rem;
    font-weight: 700;
  }
  
  .header p {
    margin: 0;
    color: var(--color-text-secondary);
  }
  
  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    padding: 3rem;
    color: var(--color-text-secondary);
  }
  
  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid var(--color-border);
    border-top-color: var(--color-primary);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }
  
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
