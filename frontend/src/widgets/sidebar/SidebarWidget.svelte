<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { isAuthenticated, currentUser, logout } from "$shared/stores/auth.svelte.js";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  let collapsed = $state(true);

  const user = $derived(currentUser());
  const authenticated = $derived(isAuthenticated());
  const currentPath = $derived($page.url.pathname);

  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, getCurrentLocale(), params);

  type NavItem = {
    href: string;
    label: string;
    short: string;
  };

  const navItems: NavItem[] = [
    { href: "/", label: t("nav.home"), short: t("nav.homeShort") },
    { href: "/graph", label: t("nav.graph"), short: t("nav.graphShort") },
    { href: "/notes", label: t("nav.notes"), short: t("nav.notesShort") },
    { href: "/search", label: t("nav.search"), short: t("nav.searchShort") },
    { href: "/settings", label: t("nav.settings"), short: t("nav.settingsShort") },
  ];

  // The graph pages have their own top floating controls, so the collapsed
  // sidebar hamburger looks like an empty, useless menu button there.
  const hiddenRoutes = ["/", "/graph", "/graph/3d"];
  const isHidden = $derived(
    hiddenRoutes.some((route) => currentPath === route || currentPath.startsWith(`${route}/`))
  );

  function toggle() {
    collapsed = !collapsed;
  }

  function navigate(href: string) {
    goto(href);
  }

  function isActive(href: string): boolean {
    return currentPath === href || currentPath.startsWith(`${href}/`);
  }
</script>

{#if !isHidden}
  <aside
    class="sidebar-widget"
    class:collapsed
    class:expanded={!collapsed}
    data-testid="sidebar-widget"
    aria-label={t("nav.ariaPanel")}
  >
    <div class="sidebar-header">
      <button
        class="toggle-button"
        type="button"
        title={collapsed ? t("nav.expand") : t("nav.collapse")}
        aria-label={collapsed ? t("nav.expandPanel") : t("nav.collapsePanel")}
        onclick={toggle}
      >
        {collapsed ? "☰" : "✕"}
      </button>
      {#if !collapsed}
        <span class="sidebar-title">{t("nav.title")}</span>
      {/if}
    </div>

    <nav class="sidebar-nav" aria-label={t("nav.aria")}>
      <ul>
        {#each navItems as item}
          <li>
            <a
              href={item.href}
              class="nav-link"
              class:active={isActive(item.href)}
              aria-current={isActive(item.href) ? "page" : undefined}
              onclick={(e: MouseEvent) => {
                e.preventDefault();
                navigate(item.href);
              }}
            >
              <span class="nav-short">{item.short}</span>
              {#if !collapsed}
                <span class="nav-label">{item.label}</span>
              {/if}
            </a>
          </li>
        {/each}
      </ul>
    </nav>

    <div class="sidebar-footer">
      {#if authenticated}
        <div class="user-info">
          <span class="user-avatar">{user?.email?.[0]?.toUpperCase() ?? "?"}</span>
          {#if !collapsed}
            <span class="user-email">{user?.email ?? t("nav.guestUser")}</span>
          {/if}
        </div>
        <a
          href="/auth/login"
          class="nav-link auth-link"
          onclick={(e: MouseEvent) => {
            e.preventDefault();
            logout();
            goto("/auth/login");
          }}
        >
          <span class="nav-short">↩</span>
          {#if !collapsed}
            <span class="nav-label">{t("nav.logout")}</span>
          {/if}
        </a>
      {:else}
        <a
          href="/auth/login"
          class="nav-link auth-link"
          onclick={(e: MouseEvent) => {
            e.preventDefault();
            navigate("/auth/login");
          }}
        >
          <span class="nav-short">{t("nav.loginShort")}</span>
          {#if !collapsed}
            <span class="nav-label">{t("nav.login")}</span>
          {/if}
        </a>
        <a
          href="/auth/register"
          class="nav-link auth-link"
          onclick={(e: MouseEvent) => {
            e.preventDefault();
            navigate("/auth/register");
          }}
        >
          <span class="nav-short">{t("nav.registerShort")}</span>
          {#if !collapsed}
            <span class="nav-label">{t("nav.register")}</span>
          {/if}
        </a>
      {/if}
    </div>
  </aside>
{/if}

<style>
  .sidebar-widget {
    display: flex;
    flex-direction: column;
    flex-shrink: 0;
    height: 100vh;
    background: var(--color-surface);
    border-right: 1px solid var(--color-border);
    transition: width 0.25s ease;
    overflow: hidden;
  }

  .sidebar-widget.collapsed {
    width: 4rem;
  }

  .sidebar-widget.expanded {
    width: 16rem;
  }

  .sidebar-header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 1rem;
    border-bottom: 1px solid var(--color-border);
    min-height: 3.5rem;
  }

  .toggle-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    background: transparent;
    border: 1px solid var(--color-border);
    border-radius: 0.375rem;
    color: var(--color-text);
    cursor: pointer;
    flex-shrink: 0;
  }

  .toggle-button:hover {
    border-color: var(--color-border-focus);
    color: var(--color-primary);
  }

  .sidebar-title {
    font-weight: 600;
    color: var(--color-text);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .sidebar-nav {
    flex: 1;
    padding: 0.75rem 0.5rem;
    overflow-y: auto;
  }

  .sidebar-nav ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .nav-link {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.625rem 0.75rem;
    border-radius: 0.5rem;
    color: var(--color-text-secondary);
    text-decoration: none;
    transition:
      background 0.15s ease,
      color 0.15s ease;
    min-height: 2.5rem;
  }

  .nav-link:hover,
  .nav-link.active {
    background: var(--color-primary-light);
    color: var(--color-primary);
  }

  .nav-short {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 1.5rem;
    height: 1.5rem;
    flex-shrink: 0;
    font-weight: 600;
    font-size: 0.75rem;
    text-align: center;
  }

  .nav-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 0.875rem;
  }

  .sidebar-footer {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem 0.5rem;
    border-top: 1px solid var(--color-border);
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.625rem 0.75rem;
    min-height: 2.5rem;
  }

  .user-avatar {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    background: var(--color-primary-light);
    color: var(--color-primary);
    font-weight: 600;
    font-size: 0.75rem;
    flex-shrink: 0;
  }

  .user-email {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    font-size: 0.75rem;
    color: var(--color-text-secondary);
  }

  .auth-link {
    color: var(--color-text-secondary);
  }
</style>
