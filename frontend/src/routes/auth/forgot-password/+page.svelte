<script lang="ts">
  import { goto } from "$app/navigation";
  import ForgotPasswordForm from "$components/organisms/ForgotPasswordForm.svelte";
  import AuthCard from "$widgets/auth/AuthCard.svelte";
  import { isAuthenticated } from "$shared/stores/auth.svelte.js";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  // Redirect if already authenticated
  $effect(() => {
    if (isAuthenticated()) {
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
