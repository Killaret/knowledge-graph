<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { MessageFormatter } from '$lib/utils/galactic-lexicon';

  interface Props {
    message: string;
    type?: 'success' | 'error' | 'info' | 'warning';
    duration?: number;
    useGalacticMode?: boolean;
    onClose?: () => void;
  }

  let { 
    message, 
    type = 'info', 
    duration = 5000, 
    useGalacticMode = false, 
    onClose = () => {} 
  }: Props = $props();

  let formatter: MessageFormatter;
  let visible = $state(false);
  let progress = $state(100);
  let intervalId: ReturnType<typeof setInterval>;
  let timeoutId: ReturnType<typeof setTimeout>;

  // Icons for different types
  const icons = {
    success: '✅',
    error: '❌',
    info: 'ℹ️',
    warning: '⚠️',
  };

  // Galactic icons
  const galacticIcons = {
    success: '⭐',
    error: '💥',
    info: '🔭',
    warning: '🚨',
  };

  // CSS classes for different types
  const typeClasses = {
    success: 'toast-success',
    error: 'toast-error',
    info: 'toast-info',
    warning: 'toast-warning',
  };

  onMount(() => {
    formatter = new MessageFormatter(useGalacticMode);
    
    // Animate in
    setTimeout(() => {
      visible = true;
    }, 10);

    // Start progress bar
    const interval = 50;
    const decrement = 100 / (duration / interval);
    
    intervalId = setInterval(() => {
      progress -= decrement;
      if (progress <= 0) {
        progress = 0;
        clearInterval(intervalId);
      }
    }, interval);

    // Auto close
    timeoutId = setTimeout(() => {
      closeToast();
    }, duration);
  });

  onDestroy(() => {
    if (intervalId) clearInterval(intervalId);
    if (timeoutId) clearTimeout(timeoutId);
  });

  function closeToast() {
    visible = false;
    setTimeout(() => {
      onClose();
    }, 300); // Wait for exit animation
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      closeToast();
    }
  }

  let displayIcon = $derived(useGalacticMode ? galacticIcons[type] : icons[type]);
  let toastClass = $derived(`toast-notification ${typeClasses[type]} ${visible ? 'visible' : ''}`);
</script>

<svelte:window on:keydown={handleKeyDown} />

<div 
  class={toastClass}
  role="alert"
  aria-live="polite"
>
  <div class="toast-content">
    <span class="toast-icon">{displayIcon}</span>
    <span class="toast-message">{message}</span>
  </div>
  
  <button 
    class="toast-close"
    onclick={closeToast}
    aria-label="Закрыть уведомление"
  >
    ×
  </button>
  
  <div class="toast-progress">
    <div class="toast-progress-bar" style="width: {progress}%"></div>
  </div>
</div>

<style>
  .toast-notification {
    position: fixed;
    bottom: 20px;
    right: 20px;
    min-width: 300px;
    max-width: 400px;
    padding: 16px;
    border-radius: 12px;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(10px);
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    opacity: 0;
    transform: translateY(20px);
    transition: all 0.3s ease;
    z-index: 1000;
    cursor: pointer;
  }

  .toast-notification.visible {
    opacity: 1;
    transform: translateY(0);
  }

  .toast-notification:hover {
    transform: translateY(-2px);
    box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
  }

  .toast-success {
    background: linear-gradient(135deg, rgba(34, 197, 94, 0.95), rgba(21, 128, 61, 0.95));
    border: 1px solid rgba(34, 197, 94, 0.3);
  }

  .toast-error {
    background: linear-gradient(135deg, rgba(239, 68, 68, 0.95), rgba(185, 28, 28, 0.95));
    border: 1px solid rgba(239, 68, 68, 0.3);
  }

  .toast-info {
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.95), rgba(37, 99, 235, 0.95));
    border: 1px solid rgba(59, 130, 246, 0.3);
  }

  .toast-warning {
    background: linear-gradient(135deg, rgba(245, 158, 11, 0.95), rgba(217, 119, 6, 0.95));
    border: 1px solid rgba(245, 158, 11, 0.3);
  }

  .toast-content {
    display: flex;
    align-items: center;
    gap: 10px;
    flex: 1;
  }

  .toast-icon {
    font-size: 20px;
    flex-shrink: 0;
  }

  .toast-message {
    color: white;
    font-size: 14px;
    font-weight: 500;
    line-height: 1.4;
  }

  .toast-close {
    background: none;
    border: none;
    color: white;
    font-size: 20px;
    cursor: pointer;
    padding: 0;
    width: 24px;
    height: 24px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    transition: background 0.2s;
    flex-shrink: 0;
  }

  .toast-close:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .toast-progress {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 3px;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 0 0 12px 12px;
    overflow: hidden;
  }

  .toast-progress-bar {
    height: 100%;
    background: rgba(255, 255, 255, 0.8);
    transition: width 0.05s linear;
  }

  /* Galactic theme enhancements */
  .toast-success:global(.galactic) {
    background: linear-gradient(135deg, rgba(139, 92, 246, 0.95), rgba(124, 58, 237, 0.95));
    box-shadow: 0 4px 20px rgba(139, 92, 246, 0.4);
  }

  .toast-notification:global(.galactic) .toast-message {
    font-family: 'Courier New', monospace;
    letter-spacing: 0.5px;
  }
</style>
