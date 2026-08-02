<script lang="ts">
  import "$shared/styles/global.css";
  import SidebarWidget from "$widgets/sidebar/SidebarWidget.svelte";
  import QuickCaptureWidget from "$widgets/quick-capture/QuickCaptureWidget.svelte";
  import ToastNotification from "$widgets/notification/ToastNotification.svelte";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { onMount } from "svelte";
  import {
    initAuth,
    isAuthenticated,
    isInitialized,
    isLoading,
    skipAuthMode,
  } from "$shared/stores/auth.svelte.js";
  import { startPreload } from "$shared/services/PreloadService";
  import {
    state as achievementsState,
    startPolling as startAchievementsPolling,
    stopPolling as stopAchievementsPolling,
    dismiss as dismissAchievement,
  } from "$shared/stores/achievements.svelte.js";
  import { mode, getMessage } from "$shared/stores/lexicon-settings";
  import { Theme } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  const { children } = $props();

  // Public routes that don't require authentication
  const publicRoutes = [
    "/", // Main page - accessible for guests
    "/graph", // Graph page - accessible for guests
    "/auth/login",
    "/auth/register",
    "/auth/forgot-password",
    "/auth/reset-password",
    "/auth/yandex/callback",
    "/health",
    "/test", // Test routes for visual regression testing
  ];

  // Toast notification state
  let toastMessage = $state("");
  let toastType: "success" | "error" | "info" | "warning" = $state("info");
  let showToast = $state(false);
  let toastGalacticMode = $state(false);

  // Check if we're in SKIP_AUTH mode
  let isSkipAuth = $state(false);

  // Initialize auth once on mount
  onMount(() => {
    initAuth();
  });

  // After hydration from localStorage, preload only for guests (avoids duplicate fetch + wrong UX)
  $effect(() => {
    if (isInitialized() && !isAuthenticated()) {
      startPreload();
    }
  });

  // Update isSkipAuth when initialized
  $effect(() => {
    if (isInitialized()) {
      isSkipAuth = skipAuthMode();
    }
  });

  // Route protection — wait for initAuth(); isLoading() is only for login/register actions, not bootstrap
  $effect(() => {
    const currentPath = $page.url.pathname;
    const isPublicRoute = publicRoutes.some((route) => currentPath.startsWith(route));

    if (isInitialized() && !isLoading() && !isPublicRoute && !isAuthenticated() && !isSkipAuth) {
      const returnUrl = encodeURIComponent(currentPath);
      goto(`/auth/login?redirect=${returnUrl}`);
    }
  });

  // Achievement polling
  $effect(() => {
    if (!isAuthenticated()) return;

    startAchievementsPolling();
    return () => {
      stopAchievementsPolling();
    };
  });

  // Achievement notifications
  $effect(() => {
    const newAchievements = achievementsState.new;
    if (newAchievements.length > 0) {
      newAchievements.forEach((achievement) => {
        showAchievementNotification(achievement);
        dismissAchievement(achievement.id);
      });
    }
  });

  // Update galactic mode for toasts
  $effect(() => {
    let currentMode = "standard";
    mode.subscribe((m) => (currentMode = m))();
    toastGalacticMode = Theme.fromString(currentMode).isGalactic;
  });

  async function showAchievementNotification(achievement: { title: string }) {
    try {
      const msg = await getMessage("achievement", "unlocked", achievement.title);
      toastMessage = msg;
      toastType = "success";
      showToast = true;
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to get achievement message:", e);
      }
    }
  }

  function closeToast() {
    showToast = false;
  }
</script>

<!--
  APP SHELL STRUCTURE
  Sidebar зарезервирован для будущего Context Control Center (v2.0)
  Пока скрыт (width: 0), но готов к активации
-->
<div class="app-shell">
  {#if isSkipAuth}
    <div class="skip-auth-badge" title={t("layout.skipAuthTitle")}>🔑 SKIP_AUTH</div>
  {/if}

  <SidebarWidget />
  <main class="app-main">
    {@render children()}
  </main>
</div>

<QuickCaptureWidget />

{#if showToast}
  <ToastNotification
    message={toastMessage}
    type={toastType}
    useGalacticMode={toastGalacticMode}
    onClose={closeToast}
    duration={5000}
  />
{/if}

<style>
  .app-shell {
    display: flex;
    height: 100vh;
    overflow: hidden;
    position: relative;
  }

  .app-main {
    flex: 1;
    overflow-y: auto;
    min-width: 0; /* Prevent flexbox overflow issues */
    background: var(--gradient-cosmic-bg);
    color: var(--color-text-dark);
  }

  .skip-auth-badge {
    position: fixed;
    top: 10px;
    right: 10px;
    z-index: 9999;
    background: linear-gradient(135deg, #fbbf24 0%, #f59e0b 100%);
    color: #78350f;
    padding: 6px 12px;
    border-radius: 20px;
    font-size: 12px;
    font-weight: 600;
    box-shadow: 0 4px 12px rgba(251, 191, 36, 0.4);
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.8;
      transform: scale(1.05);
    }
  }
</style>
