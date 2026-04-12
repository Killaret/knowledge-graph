<script lang="ts">
  import { fade, fly } from 'svelte/transition';
  import { goto } from '$app/navigation';
  
  const { 
    isOpen = false,
    onToggle,
    onImportClick,
    onExportClick
  }: {
    isOpen: boolean;
    onToggle: () => void;
    onImportClick: () => void;
    onExportClick: () => void;
  } = $props();
  
  const menuItems = [
    { id: 'graph', label: 'Граф', icon: '🕸️', href: '/' },
    { id: 'import', label: 'Импорт', icon: '📥', action: 'import' },
    { id: 'export', label: 'Экспорт', icon: '📤', action: 'export' },
    { id: 'settings', label: 'Настройки', icon: '⚙️', href: '/settings' },
  ];
  
  function handleItemClick(item: any) {
    if (item.action === 'import') {
      onImportClick();
    } else if (item.action === 'export') {
      onExportClick();
    } else if (item.href) {
      goto(item.href);
    }
    onToggle();
  }
</script>

<div class="sidebar-container">
  <!-- Hamburger Button -->
  <button 
    class="menu-btn"
    onclick={onToggle}
    aria-label={isOpen ? 'Закрыть меню' : 'Открыть меню'}
    aria-expanded={isOpen}
  >
    <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2">
      {#if isOpen}
        <line x1="18" y1="6" x2="6" y2="18"></line>
        <line x1="6" y1="6" x2="18" y2="18"></line>
      {:else}
        <line x1="3" y1="12" x2="21" y2="12"></line>
        <line x1="3" y1="6" x2="21" y2="6"></line>
        <line x1="3" y1="18" x2="21" y2="18"></line>
      {/if}
    </svg>
  </button>
  
  {#if isOpen}
    <div 
      class="sidebar-overlay"
      onclick={onToggle}
      transition:fade={{ duration: 200 }}
    ></div>
    
    <nav 
      class="sidebar"
      transition:fly={{ x: -100, duration: 300 }}
      aria-label="Главное меню"
    >
      <div class="sidebar-header">
        <h3 class="sidebar-title">Меню</h3>
      </div>
      
      <ul class="menu-list">
        {#each menuItems as item}
          <li>
            <button 
              class="menu-item"
              onclick={() => handleItemClick(item)}
            >
              <span class="item-icon">{item.icon}</span>
              <span class="item-label">{item.label}</span>
            </button>
          </li>
        {/each}
      </ul>
      
      <div class="sidebar-footer">
        <p class="version">v1.0.0</p>
      </div>
    </nav>
  {/if}
</div>

<style>
  .sidebar-container {
    position: relative;
    z-index: 1500;
  }

  .menu-btn {
    position: fixed;
    top: 16px;
    left: 16px;
    width: 44px;
    height: 44px;
    background: rgba(10, 26, 58, 0.8);
    border: 1px solid rgba(255, 221, 136, 0.2);
    border-radius: 10px;
    color: #ffdd88;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
    z-index: 1501;
  }

  .menu-btn:hover {
    background: rgba(255, 221, 136, 0.1);
    border-color: rgba(255, 221, 136, 0.4);
  }

  .sidebar-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 1499;
  }

  .sidebar {
    position: fixed;
    top: 0;
    left: 0;
    width: 260px;
    height: 100vh;
    background: linear-gradient(180deg, #0a1a3a 0%, #1a2a4a 100%);
    border-right: 1px solid rgba(255, 221, 136, 0.2);
    padding: 80px 20px 20px;
    z-index: 1500;
    display: flex;
    flex-direction: column;
  }

  .sidebar-header {
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid rgba(136, 170, 204, 0.2);
  }

  .sidebar-title {
    color: #ffdd88;
    font-size: 1.25rem;
    font-weight: 600;
    margin: 0;
  }

  .menu-list {
    list-style: none;
    padding: 0;
    margin: 0;
    flex: 1;
  }

  .menu-list li {
    margin-bottom: 8px;
  }

  .menu-item {
    width: 100%;
    padding: 14px 16px;
    background: rgba(10, 26, 58, 0.4);
    border: 1px solid rgba(136, 170, 204, 0.1);
    border-radius: 10px;
    color: #e0e0e0;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 12px;
    transition: all 0.2s;
    font-size: 1rem;
  }

  .menu-item:hover {
    background: rgba(255, 221, 136, 0.1);
    border-color: rgba(255, 221, 136, 0.3);
    color: #ffdd88;
  }

  .item-icon {
    font-size: 1.25rem;
  }

  .item-label {
    font-weight: 500;
  }

  .sidebar-footer {
    padding-top: 16px;
    border-top: 1px solid rgba(136, 170, 204, 0.2);
    text-align: center;
  }

  .version {
    color: rgba(136, 170, 204, 0.6);
    font-size: 0.85rem;
    margin: 0;
  }

  @media (max-width: 768px) {
    .sidebar {
      width: 100%;
    }
  }
</style>
