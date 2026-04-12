<script lang="ts">
  import { fade } from 'svelte/transition';
  
  const { 
    title = 'Нет данных',
    message = 'Здесь пока ничего нет',
    actionText = '',
    onAction,
    icon = '🔭'
  }: {
    title?: string;
    message?: string;
    actionText?: string;
    onAction?: () => void;
    icon?: string;
  } = $props();
</script>

<div class="empty-state" transition:fade={{ duration: 300 }}>
  <div class="icon">{icon}</div>
  <h3 class="title">{title}</h3>
  <p class="message">{message}</p>
  {#if actionText && onAction}
    <button class="action-btn" onclick={onAction}>
      {actionText}
    </button>
  {/if}
</div>

<style>
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 3rem 2rem;
    color: #e0e0e0;
  }

  .icon {
    font-size: 4rem;
    margin-bottom: 1rem;
    opacity: 0.8;
    animation: float 3s ease-in-out infinite;
  }

  @keyframes float {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-10px); }
  }

  .title {
    font-size: 1.5rem;
    font-weight: 600;
    color: #ffdd88;
    margin: 0 0 0.75rem 0;
    text-shadow: 0 0 20px rgba(255, 221, 136, 0.3);
  }

  .message {
    font-size: 1rem;
    color: #88aacc;
    margin: 0 0 1.5rem 0;
    max-width: 300px;
    line-height: 1.5;
  }

  .action-btn {
    padding: 12px 24px;
    background: linear-gradient(135deg, #ffdd88 0%, #ffaa44 100%);
    color: #0a1a3a;
    border: none;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 4px 12px rgba(255, 170, 68, 0.4);
  }

  .action-btn:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 20px rgba(255, 170, 68, 0.5);
  }
</style>
