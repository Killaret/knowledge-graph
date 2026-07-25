<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import ResetPasswordForm from "$components/organisms/ResetPasswordForm.svelte";
  import AuthCard from "$components/organisms/AuthCard.svelte";
  import ConstellationIcon from "$components/atoms/ConstellationIcon.svelte";
  import { isAuthenticated } from "$shared/stores/auth.svelte.js";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  // Get token from URL
  let token = $state("");

  $effect(() => {
    token = $page.url.searchParams.get("token") || "";
  });

  // Redirect if already authenticated
  $effect(() => {
    if (isAuthenticated()) {
      goto("/");
    }
  });
</script>

{#if token}
  <AuthCard
    title={t("auth.resetPasswordTitle")}
    subtitle={t("auth.resetPasswordSubtitle")}
    showIcon={true}
  >
    <ResetPasswordForm {token} />
  </AuthCard>
{:else}
  <AuthCard title={t("error.fallback")} subtitle={t("auth.resetTokenMissing")} showIcon={false}>
    <div class="error-content">
      <ConstellationIcon size={48} class="error-icon" />
      <p class="error-text">{t("auth.requestNewLink")}</p>
      <a href="/auth/forgot-password" class="back-link">{t("auth.requestResetLink")}</a>
    </div>
  </AuthCard>
{/if}

<style>
  .error-content {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }

  .error-content :global(.error-icon) {
    opacity: 0.6;
  }

  .error-text {
    margin: 0;
    color: var(--color-text-secondary, #94a3b8);
    font-size: 0.875rem;
  }

  .back-link {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    text-decoration: none;
    border-radius: 8px;
    font-weight: 500;
    transition: all 0.2s ease;
  }

  .back-link:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 20px rgba(59, 130, 246, 0.4);
  }
</style>
