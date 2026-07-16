<script lang="ts">
  import { goto } from '$app/navigation';
  import ForgotPasswordForm from '$components/organisms/ForgotPasswordForm.svelte';
  import AuthCard from '$components/organisms/AuthCard.svelte';
  import { isAuthenticated } from '$shared/stores/auth.svelte.js';
  
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
