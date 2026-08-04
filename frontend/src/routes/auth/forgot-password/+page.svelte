<script lang="ts">
  import { goto } from "$app/navigation";
  import ForgotPasswordForm from "$components/organisms/ForgotPasswordForm.svelte";
  import AuthCard from "$widgets/auth/AuthCard.svelte";
  import { isAuthenticated, skipAuthMode } from "$shared/stores/auth.svelte.js";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  // Redirect if already authenticated (but allow viewing auth pages in skip-auth tests)
  $effect(() => {
    if (isAuthenticated() && !skipAuthMode()) {
      goto("/");
    }
  });
</script>

<AuthCard
  title={t("auth.forgotPasswordTitle")}
  subtitle={t("auth.forgotPasswordSubtitle")}
  showIcon={true}
>
  <ForgotPasswordForm />
</AuthCard>
