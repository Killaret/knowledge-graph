<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import LoginForm from '$lib/components/LoginForm.svelte';
  import AuthCard from '$lib/components/AuthCard.svelte';
  import { isAuthenticated, initAuth } from '$lib/stores/auth.svelte.js';
  import { startPreload } from '$lib/services/PreloadService';
  import { onMount } from 'svelte';
  
  // Redirect if already authenticated
  $effect(() => {
    if (isAuthenticated()) {
      const redirectTo = $page.url.searchParams.get('redirect') || '/';
      goto(redirectTo);
    }
  });
  
  // Initialize auth once on mount
  onMount(() => {
    initAuth();
  });

  // Start background preload if not authenticated
  $effect(() => {
    if (!isAuthenticated()) {
      startPreload();
    }
  });
</script>

<AuthCard 
  title="Knowledge Graph" 
  subtitle="Your personal multiverse of knowledge"
  showIcon={true}
>
  <LoginForm />
</AuthCard>
