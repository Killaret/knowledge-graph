<script lang="ts">
  import { page } from "$app/stores";
  import ForgotPasswordForm from "$components/organisms/ForgotPasswordForm.svelte";
  import AuthCard from "$widgets/auth/AuthCard.svelte";
  import { useAnonymousGuard } from "$shared/composables/auth";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  // Redirect if already authenticated (but allow viewing auth pages in skip-auth tests)
  $effect(() => {
    const redirectTo = $page.url.searchParams.get("redirect") || "/";
    useAnonymousGuard(redirectTo);
  });
</script>

<AuthCard
  title={t("auth.forgotPasswordTitle")}
  subtitle={t("auth.forgotPasswordSubtitle")}
  showIcon={true}
>
  <ForgotPasswordForm />
</AuthCard>
