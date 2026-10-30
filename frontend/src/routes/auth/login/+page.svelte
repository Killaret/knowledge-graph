<script lang="ts">
  import { page } from "$app/stores";
  import LoginForm from "$components/organisms/LoginForm.svelte";
  import AuthCard from "$widgets/auth/AuthCard.svelte";
  import PreloadIndicator from "$components/organisms/PreloadIndicator.svelte";
  import { isAuthenticated, initAuth } from "$shared/stores/auth.svelte.js";
  import { useAnonymousGuard } from "$shared/composables/auth";
  import { startPreload } from "$shared/services/PreloadService";
  import { onMount } from "svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  // Redirect if already authenticated (but allow viewing auth pages in skip-auth tests)
  $effect(() => {
    const redirectTo = $page.url.searchParams.get("redirect") || "/";
    useAnonymousGuard(redirectTo);
  });

  // Initialize auth once on mount
  onMount(() => {
    initAuth();
  });

  // Start background preload if not authenticated
  $effect(() => {
    if (!isAuthenticated()) {
      startPreload();
    }
  });
</script>

<AuthCard title={t("app.title")} subtitle={t("auth.subtitle")} showIcon={true}>
  <LoginForm />
  <PreloadIndicator />
</AuthCard>
