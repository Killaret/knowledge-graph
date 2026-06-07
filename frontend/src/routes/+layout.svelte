  import '$lib/styles/global.css';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import QuickCaptureWidget from '$lib/components/QuickCaptureWidget.svelte';
  import ToastNotification from '$lib/components/ToastNotification.svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { initAuth, isAuthenticated, isInitialized, isLoading } from '$lib/stores/auth.svelte.js';
  import { initAuth, isAuthenticated, isInitialized, isLoading, skipAuthMode } from '$lib/stores/auth.svelte.js';
  import { startPreload } from '$lib/services/PreloadService';
  import { achievementsStore } from '$lib/stores/achievements';
  import { mode, getMessage } from '$lib/stores/lexicon-settings';

  const { children } = $props();

  // Public routes that don't require authentication
  const publicRoutes = [
    '/auth/login',
    '/auth/register',
    '/auth/forgot-password',
    '/auth/reset-password',
    '/auth/yandex/callback',
    '/health',
    '/test'  // Test routes for visual regression testing
  ];

  // Toast notification state
  let toastMessage = $state('');
  let toastType = $state<'success' | 'error' | 'info' | 'warning'>('info');
  let showToast = $state(false);
  let toastGalacticMode = $state(false);

  // Initialize auth on mount
  // Check if we're in SKIP_AUTH mode
  let isSkipAuth = $state(false);
  $effect(() => {
    initAuth();

    // After hydration from localStorage, preload only for guests (avoids duplicate fetch + wrong UX)
    if (isInitialized() && !isAuthenticated()) {
      startPreload();
    }
  });

  // Route protection — wait for initAuth(); isLoading() is only for login/register actions, not bootstrap
    isSkipAuth = skipAuthMode();
  });

  // Initialize auth on mount
  $effect(() => {
    const currentPath = $page.url.pathname;
    const isPublicRoute = publicRoutes.some(route => currentPath.startsWith(route));

    if (
      isInitialized() &&
      !isLoading() &&
      !isPublicRoute &&
      !isAuthenticated()
    ) {
      const returnUrl = encodeURIComponent(currentPath);
      goto(`/auth/login?redirect=${returnUrl}`);
    }
  });

  // Achievement notifications
    initAuth();

    // After hydration from localStorage, preload only for guests (avoids duplicate fetch + wrong UX)
    if (isInitialized() && !isAuthenticated()) {
      startPreload();
    }
  });

  // Route protection — wait for initAuth(); isLoading() is only for login/register actions, not bootstrap
  $effect(() => {
    if (!isAuthenticated()) return;

    // Start polling for achievements when authenticated
    achievementsStore.startPolling();

    // Subscribe to new achievements
    const unsubscribe = achievementsStore.subscribe(({ new: newAchievements }) => {
      if (newAchievements.length > 0) {
        // Show notification for each new achievement
        newAchievements.forEach(achievement => {
          showAchievementNotification(achievement);
          // Mark as seen after showing
          achievementsStore.dismiss(achievement.id);
        });
      }
    });

    return () => {
      achievementsStore.stopPolling();
      unsubscribe();
    };
  });

  // Update galactic mode for toasts
    const currentPath = $page.url.pathname;
    const isPublicRoute = publicRoutes.some(route => currentPath.startsWith(route));

    if (
      isInitialized() &&
      !isLoading() &&
      !isPublicRoute &&
      !isAuthenticated() &&
      !isSkipAuth
    ) {
      const returnUrl = encodeURIComponent(currentPath);
      goto(`/auth/login?redirect=${returnUrl}`);
    }
  });

  // Achievement notifications
  $effect(() => {
    let currentMode = 'standard';
    mode.subscribe(m => currentMode = m)();
    toastGalacticMode = currentMode === 'galactic';
    if (!isAuthenticated()) return;

    // Start polling for achievements when authenticated
    achievementsStore.startPolling();

    // Subscribe to new achievements
    const unsubscribe = achievementsStore.subscribe(({ new: newAchievements }) => {
      if (newAchievements.length > 0) {
        // Show notification for each new achievement
        newAchievements.forEach(achievement => {
          showAchievementNotification(achievement);
          // Mark as seen after showing
          achievementsStore.dismiss(achievement.id);
        });
      }
    });

    return () => {
      achievementsStore.stopPolling();
      unsubscribe();
    };
  });

  // Update galactic mode for toasts
  $effect(() => {
    let currentMode = 'standard';
    mode.subscribe(m => currentMode = m)();
    toastGalacticMode = currentMode === 'galactic';
  });

  async function showAchievementNotification(achievement: any) {
    try {
      const msg = await getMessage('achievement', 'unlocked', achievement.title || achievement.name_en || 'Achievement');
      toastMessage = msg;
      toastType = 'success';
      showToast = true;
    } catch (e) {
      console.error('Failed to get achievement message:', e);
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
  <!-- SKIP_AUTH Indicator -->
  {#if isSkipAuth}
    <div class="skip-auth-badge" title="Auth is disabled for testing (SKIP_AUTH=true)">
      🔑 SKIP_AUTH
    </div>
  {/if}
  
  <Sidebar />
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

  /* SKIP_AUTH Badge */
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
    0%, 100% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.8;
      transform: scale(1.05);
    }
  }
</style>
