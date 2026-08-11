<script lang="ts">
  import { page } from "$app/stores";
  import RegisterForm from "$components/organisms/RegisterForm.svelte";
  import AuthCard from "$widgets/auth/AuthCard.svelte";
  import { initAuth } from "$shared/stores/auth.svelte.js";
  import { useAnonymousGuard } from "$shared/composables/auth";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  // Redirect if already authenticated (but allow viewing auth pages in skip-auth tests)
  $effect(() => {
    const redirectTo = $page.url.searchParams.get("redirect") || "/";
    useAnonymousGuard(redirectTo);
  });

  // Initialize auth on mount
  $effect(() => {
    initAuth();
  });
</script>

<AuthCard title={t("auth.registerTitle")} subtitle={t("auth.registerSubtitle")} showIcon={true}>
  <RegisterForm />
</AuthCard>
