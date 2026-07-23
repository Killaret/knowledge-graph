<script lang="ts">
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import {
    isAuthenticated,
    currentUser,
  } from "$shared/stores/auth.svelte.js";

  let collapsed = $state(true);

  const user = $derived(currentUser());
  const authenticated = $derived(isAuthenticated());
  const currentPath = $derived($page.url.pathname);

  type NavItem = {
    href: string;
    label: string;
    short: string;
  };

  const navItems: NavItem[] = [
    { href: "/", label: "Главная", short: "Г" },
    { href: "/graph", label: "Граф", short: "Гр" },
    { href: "/notes", label: "Заметки", short: "З" },
    { href: "/search", label: "Поиск", short: "П" },
    { href: "/settings", label: "Настройки", short: "Н" },
  ];

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

<aside
  class="sidebar-widget"
  class:collapsed
  class:expanded={!collapsed}
  data-testid="sidebar-widget"
  aria-label="Боковая панель навигации"
>
  <div class="sidebar-header">
    <button
      class="toggle-button"
      type="button"
      aria-label={collapsed ? "Развернуть панель" : "Свернуть панель"}
      onclick={toggle}
    >
      {collapsed ? "☰" : "✕"}
    </button>
    {#if !collapsed}
      <span class="sidebar-title">Knowledge Graph</span>
    {/if}
  </div>

  <nav class="sidebar-nav" aria-label="Основная навигация">
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
          <span class="user-email">{user?.email ?? "Пользователь"}</span>
        {/if}
      </div>
    {:else}
      <a
        href="/auth/login"
        class="nav-link auth-link"
        onclick={(e: MouseEvent) => {
          e.preventDefault();
          navigate("/auth/login");
        }}
      >
        <span class="nav-short">В</span>
        {#if !collapsed}
          <span class="nav-label">Войти</span>
        {/if}
      </a>
    {/if}
  </div>
</aside>

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
    transition: background 0.15s ease, color 0.15s ease;
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
