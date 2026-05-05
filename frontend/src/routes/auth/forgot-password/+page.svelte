<script lang="ts">
  import { goto } from '$app/navigation';
  import ForgotPasswordForm from '$lib/components/ForgotPasswordForm.svelte';
  import AuthCard from '$lib/components/AuthCard.svelte';
  import { isAuthenticated } from '$lib/stores/auth.svelte.js';
  
  // Redirect if already authenticated
  $effect(() => {
    if (isAuthenticated()) {
      goto('/');
    }
  });
</script>

<AuthCard 
  title="Восстановление пароля" 
  subtitle="Введите email для получения ссылки"
  showIcon={true}
>
  <ForgotPasswordForm />
</AuthCard>
