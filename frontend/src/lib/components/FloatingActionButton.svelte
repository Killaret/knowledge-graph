<script lang="ts">
  import { goto } from '$app/navigation';
  import { fade } from 'svelte/transition';
  
  let isHovered = $state(false);
  
  function handleClick() {
    goto('/notes/new');
  }
</script>

<button
  class="fab"
  data-testid="fab-new-note"
  onclick={handleClick}
  onmouseenter={() => isHovered = true}
  onmouseleave={() => isHovered = false}
  aria-label="Создать заметку"
  title="Создать заметку"
>
  <svg class="icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
    <line x1="12" y1="5" x2="12" y2="19"></line>
    <line x1="5" y1="12" x2="19" y2="12"></line>
  </svg>
  
  {#if isHovered}
    <span class="tooltip" transition:fade={{ duration: 150 }}>
      Создать заметку
    </span>
  {/if}
</button>

<style>
  .fab {
    position: fixed;
    bottom: 24px;
    right: 24px;
    width: 56px;
    height: 56px;
    border-radius: 50%;
    background: linear-gradient(135deg, #ffdd88 0%, #ffaa44 100%);
    border: none;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    box-shadow: 0 4px 12px rgba(255, 170, 68, 0.4), 0 2px 4px rgba(0, 0, 0, 0.2);
    transition: all 0.3s ease;
    z-index: 1000;
  }

  .fab:hover {
    transform: scale(1.1);
    box-shadow: 0 6px 20px rgba(255, 170, 68, 0.5), 0 4px 8px rgba(0, 0, 0, 0.3);
  }

  .fab:active {
    transform: scale(0.95);
  }

  .icon {
    width: 24px;
    height: 24px;
    color: #0a1a3a;
  }

  .tooltip {
    position: absolute;
    right: 70px;
    bottom: 50%;
    transform: translateY(50%);
    background: rgba(10, 26, 58, 0.95);
    color: #ffdd88;
    padding: 8px 12px;
    border-radius: 6px;
    font-size: 14px;
    white-space: nowrap;
    pointer-events: none;
    border: 1px solid rgba(255, 221, 136, 0.3);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  }

  .tooltip::after {
    content: '';
    position: absolute;
    right: -6px;
    top: 50%;
    transform: translateY(-50%);
    border-width: 6px 0 6px 6px;
    border-style: solid;
    border-color: transparent transparent transparent rgba(10, 26, 58, 0.95);
  }

  @media (max-width: 768px) {
    .fab {
      bottom: 16px;
      right: 16px;
      width: 48px;
      height: 48px;
    }

    .icon {
      width: 20px;
      height: 20px;
    }
  }
</style>
