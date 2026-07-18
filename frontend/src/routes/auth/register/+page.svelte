<script lang="ts">
  import { goto } from "$app/navigation";
  import RegisterForm from "$components/organisms/RegisterForm.svelte";
  import AuthCard from "$components/organisms/AuthCard.svelte";
  import { isAuthenticated, initAuth } from "$shared/stores/auth.svelte.js";

  // Redirect if already authenticated
  $effect(() => {
    if (isAuthenticated()) {
      goto("/");
    }
  });

  // Initialize auth on mount
  $effect(() => {
    initAuth();
  });
</script>

<AuthCard
  title="Create Account"
  subtitle="Join the universe of knowledge"
  showIcon={true}
>
  <RegisterForm />
</AuthCard>
