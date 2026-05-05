<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import LoginForm from '$lib/components/LoginForm.svelte';
  import AuthCard from '$lib/components/AuthCard.svelte';
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

<AuthCard 
  title="Knowledge Graph" 
  subtitle="Ваша персональная вселенная знаний"
  showIcon={true}
>
  <LoginForm />
</AuthCard>
