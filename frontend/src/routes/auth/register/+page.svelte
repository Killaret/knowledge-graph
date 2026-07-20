<script lang="ts">
  import { goto } from "$app/navigation";
  import RegisterForm from "$components/organisms/RegisterForm.svelte";
  import AuthCard from "$components/organisms/AuthCard.svelte";
  import { isAuthenticated, initAuth } from "$shared/stores/auth.svelte.js";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

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
  title={t("auth.registerTitle")}
  subtitle={t("auth.registerSubtitle")}
  showIcon={true}
>
  <RegisterForm />
</AuthCard>
