<script lang="ts">
  interface Props {
    hotkeyLines: string[];
    helpContent?: string;
    onClose: () => void;
  }

  const { hotkeyLines, helpContent, onClose }: Props = $props();
</script>

<div class="modal-backdrop" data-testid="help-modal-backdrop" onclick={onClose} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onClose(); } }} role="button" tabindex="0">
  <div class="modal-content" data-testid="help-modal" onclick={(e) => e.stopPropagation()} onkeydown={(e) => { e.stopPropagation(); if (e.key === 'Escape' || e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onClose(); } }} role="dialog" aria-modal="true" tabindex="-1">
    <div class="modal-header">
      <h2>Keyboard Shortcuts</h2>
      <button class="close-btn" onclick={onClose} aria-label="Close">×</button>
    </div>
    <div class="modal-body">
      {#if helpContent}
        <p class="help-content">{helpContent}</p>
      {/if}
      <ul class="hotkey-list">
        {#each hotkeyLines as line}
          <li>{line}</li>
        {/each}
      </ul>
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .modal-content {
    background: #1a1a2e;
    border: 1px solid #3d3d5c;
    border-radius: 12px;
    padding: 1.5rem;
    min-width: 320px;
    max-width: 90vw;
    color: #e0e0ff;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: #e0e0ff;
    font-size: 1.5rem;
    cursor: pointer;
  }

  .help-content {
    margin: 0 0 1rem;
    color: #a0a0cc;
  }

  .hotkey-list {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .hotkey-list li {
    padding: 0.5rem 0;
    border-bottom: 1px solid #2a2a40;
  }

  .hotkey-list li:last-child {
    border-bottom: none;
  }
</style>
