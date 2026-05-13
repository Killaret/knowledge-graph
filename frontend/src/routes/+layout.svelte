<script lang="ts">
  import '$lib/styles/global.css';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { initAuth, isAuthenticated, isInitialized, isLoading } from '$lib/stores/auth.svelte.js';
  import { startPreload } from '$lib/services/PreloadService';

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

  // Initialize auth on mount
  $effect(() => {
    initAuth();

    // After hydration from localStorage, preload only for guests (avoids duplicate fetch + wrong UX)
    if (isInitialized() && !isAuthenticated()) {
      startPreload();
    }
  });

  // Route protection — wait for initAuth(); isLoading() is only for login/register actions, not bootstrap
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
</script>

<!--
  APP SHELL STRUCTURE
  Sidebar зарезервирован для будущего Context Control Center (v2.0)
  Пока скрыт (width: 0), но готов к активации
-->
<div class="app-shell">
  <Sidebar />
  <main class="app-main">
    {@render children()}
  </main>
</div>

<style>
  .app-shell {
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .app-main {
    flex: 1;
    overflow-y: auto;
    min-width: 0; /* Prevent flexbox overflow issues */
    background: var(--gradient-cosmic-bg);
    color: var(--color-text-dark);
  }
</style>
