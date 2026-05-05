<script lang="ts">
  import { goto } from '$app/navigation';
  import RegisterForm from '$lib/components/RegisterForm.svelte';
  import AuthCard from '$lib/components/AuthCard.svelte';
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

<AuthCard 
  title="Создать аккаунт" 
  subtitle="Присоединитесь к вселенной знаний"
  showIcon={true}
>
  <RegisterForm />
</AuthCard>
