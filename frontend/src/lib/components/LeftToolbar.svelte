<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  
  interface Props {
    onNewNote?: () => void;
    onImport?: () => void;
    onExport?: () => void;
    onSettings?: () => void;
  }
  
  const { onNewNote, onImport, onExport, onSettings }: Props = $props();
  
  let hoveredItem = $state<string | null>(null);
  let isMobile = $state(false);
  
  // Check mobile on mount
  $effect(() => {
    const checkMobile = () => {
      isMobile = window.innerWidth < 768;
    };
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  });
  
  function navigateTo(path: string) {
    goto(path);
  }
  
  const currentPath = $derived($page.url.pathname);
  
  const isActive = (path: string) => {
    if (path === '/') return currentPath === '/';
    if (path === '/notes') return currentPath.startsWith('/notes');
    return currentPath.startsWith(path);
  };
</script>

{#snippet toolbarItem(
  id: string,
  icon: string,
  label: string,
  onclick: () => void,
  isAccent = false
)}
  <button
    class="toolbar-item"
    class:accent={isAccent}
    class:active={isActive(id)}
    {onclick}
    onmouseenter={() => hoveredItem = id}
    onmouseleave={() => hoveredItem = null}
    aria-label={label}
  >
    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
      {@html icon}
    </svg>
    
    {#if hoveredItem === id}
      <span class="tooltip" class:mobile={isMobile}>
        {label}
      </span>
    {/if}
  </button>
{/snippet}

<nav class="left-toolbar" class:mobile={isMobile} aria-label="Main navigation">
  <!-- New Note (Accent) -->
  <button
    class="toolbar-item accent"
    onclick={() => onNewNote?.() || goto('/notes/new')}
    onmouseenter={() => hoveredItem = 'new'}
    onmouseleave={() => hoveredItem = null}
    aria-label="Новая звезда"
    data-testid="toolbar-new-note"
  >
    <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
      <line x1="12" y1="5" x2="12" y2="19"/>
      <line x1="5" y1="12" x2="19" y2="12"/>
    </svg>
    
    {#if hoveredItem === 'new'}
      <span class="tooltip" class:mobile={isMobile}>
        Новая звезда
      </span>
    {/if}
  </button>
  
  <!-- Divider -->
  <div class="divider"></div>
  
  <!-- Graph (Home) -->
  {@render toolbarItem(
    '/',
    '<circle cx="12" cy="12" r="3"/><path d="M12 2v4"/><path d="M12 18v4"/><path d="m4.93 4.93 2.83 2.83"/><path d="m16.24 16.24 2.83 2.83"/><path d="M2 12h4"/><path d="M18 12h4"/><path d="m4.93 19.07 2.83-2.83"/><path d="m16.24 7.76 2.83-2.83"/>',
    'Граф',
    () => navigateTo('/')
  )}
  
  <!-- All Notes -->
  {@render toolbarItem(
    '/notes',
    '<line x1="8" y1="6" x2="21" y2="6"/><line x1="8" y1="12" x2="21" y2="12"/><line x1="8" y1="18" x2="21" y2="18"/><line x1="3" y1="6" x2="3.01" y2="6"/><line x1="3" y1="12" x2="3.01" y2="12"/><line x1="3" y1="18" x2="3.01" y2="18"/>',
    'Все заметки',
    () => navigateTo('/notes')
  )}
  
  <!-- Import -->
  {@render toolbarItem(
    'import',
    '<path d="M4 14.899A7 7 0 1 1 15.71 8h1.79a4.5 4.5 0 0 1 2.5 8.242"/><path d="M12 12v9"/><path d="m16 16-4-4-4 4"/>',
    'Импорт',
    () => onImport?.()
  )}
  
  <!-- Export -->
  {@render toolbarItem(
    'export',
    '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7,10 12,15 17,10"/><line x1="12" y1="15" x2="12" y2="3"/>',
    'Экспорт',
    () => onExport?.()
  )}
  
  <!-- Settings -->
  {@render toolbarItem(
    'settings',
    '<path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.1a2 2 0 0 1-1-1.72v-.51a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/>',
    'Настройки',
    () => onSettings?.()
  )}
</nav>

<style>
  .left-toolbar {
    position: fixed;
    left: 1rem;
    top: 50%;
    transform: translateY(-50%);
    z-index: 30;
    
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    
    background: rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 1rem;
    padding: 0.75rem;

    /* Prevent the toolbar container from intercepting pointer events
       while keeping individual buttons interactive. This avoids UI elements
       (modals/forms) being blocked by the toolbar box in narrow viewports. */
    pointer-events: none;
  }

  /* Buttons should still receive pointer events */
  .left-toolbar .toolbar-item {
    pointer-events: auto;
  }
  
  .left-toolbar.mobile {
    left: 50%;
    right: auto;
    top: auto;
    bottom: 1rem;
    transform: translateX(-50%);
    flex-direction: row;
  }
  
  .toolbar-item {
    position: relative;
    width: 2.5rem;
    height: 2.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 0.75rem;
    color: rgba(255, 255, 255, 0.7);
    
    cursor: pointer;
    transition: all 0.2s ease;
  }
  
  .toolbar-item:hover {
    background: rgba(255, 255, 255, 0.15);
    color: white;
    transform: scale(1.05);
  }
  
  .toolbar-item.accent {
    background: rgba(99, 102, 241, 0.8);
    border-color: rgba(99, 102, 241, 0.5);
    color: white;
  }
  
  .toolbar-item.accent:hover {
    background: rgba(99, 102, 241, 1);
  }
  
  .toolbar-item.active {
    background: rgba(255, 255, 255, 0.2);
    color: white;
  }
  
  .divider {
    width: 100%;
    height: 1px;
    background: rgba(255, 255, 255, 0.1);
    margin: 0.25rem 0;
  }
  
  .left-toolbar.mobile .divider {
    width: 1px;
    height: 2rem;
    margin: 0 0.25rem;
  }
  
  .tooltip {
    position: absolute;
    left: calc(100% + 0.75rem);
    top: 50%;
    transform: translateY(-50%);
    
    background: rgba(0, 0, 0, 0.8);
    color: white;
    padding: 0.375rem 0.75rem;
    border-radius: 0.5rem;
    font-size: 0.75rem;
    font-weight: 500;
    white-space: nowrap;
    
    opacity: 0;
    animation: fadeIn 0.15s ease forwards;
    pointer-events: none;
    z-index: 40;
  }
  
  .tooltip.mobile {
    left: 50%;
    top: auto;
    bottom: calc(100% + 0.5rem);
    transform: translateX(-50%);
  }
  
  @keyframes fadeIn {
    to { opacity: 1; }
  }
  
  @media (max-width: 768px) {
    .left-toolbar {
      left: 50%;
      right: auto;
      top: auto;
      bottom: 1rem;
      transform: translateX(-50%);
      flex-direction: row;
    }
    
    .divider {
      width: 1px;
      height: 2rem;
      margin: 0 0.25rem;
    }
    
    .tooltip {
      left: 50%;
      top: auto;
      bottom: calc(100% + 0.5rem);
      transform: translateX(-50%);
    }
  }
</style>
