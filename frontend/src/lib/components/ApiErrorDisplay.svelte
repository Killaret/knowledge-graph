<script lang="ts">
  import type { ErrorResponse } from '$lib/types/errors';

  interface Props {
    error: ErrorResponse | null;
    onClose?: () => void;
  }

  let { error, onClose }: Props = $props();

  function handleClose() {
    onClose?.();
  }
</script>

{#if error}
  <div 
    class="error-container"
    role="alert"
    aria-live="assertive"
  >
    <button 
      class="close-button" 
      onclick={handleClose}
      aria-label="Закрыть ошибку"
      type="button"
    >
      ×
    </button>
    
    <div class="error-header">
      <span class="error-icon" aria-hidden="true">⚠️</span>
      <span class="error-code">Ошибка: {error.code}</span>
    </div>
    
    <p class="error-message">{error.message}</p>
    
    {#if error.details && error.details.length > 0}
      <div class="error-details">
        <p class="details-title">Детали:</p>
        <ul class="details-list">
          {#each error.details as detail}
            <li class="detail-item">
              <span class="detail-field">{detail.field}</span>
              <span class="detail-separator">—</span>
              <span class="detail-reason">{detail.message}</span>
              {#if detail.received !== undefined}
                <span class="detail-received">(получено: {JSON.stringify(detail.received)})</span>
              {/if}
            </li>
          {/each}
        </ul>
      </div>
    {/if}
  </div>
{/if}

<style>
  .error-container {
    position: relative;
    padding: 1rem 1.25rem;
    margin: 0.5rem 0;
    background: var(--color-surface, #1e293b);
    border-left: 4px solid var(--color-danger, #ef4444);
    border-radius: 0 8px 8px 0;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06);
    color: var(--color-text, #f1f5f9);
  }

  .close-button {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    background: none;
    border: none;
    font-size: 1.5rem;
    line-height: 1;
    color: var(--color-text-muted, #94a3b8);
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    border-radius: 4px;
    transition: color 0.2s, background-color 0.2s;
  }

  .close-button:hover {
    color: var(--color-text, #f1f5f9);
    background-color: rgba(255, 255, 255, 0.1);
  }

  .close-button:focus {
    outline: 2px solid var(--color-danger, #ef4444);
    outline-offset: 2px;
  }

  .error-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.5rem;
    padding-right: 2rem;
  }

  .error-icon {
    font-size: 1.25rem;
    flex-shrink: 0;
  }

  .error-code {
    font-weight: 600;
    font-size: 1rem;
    color: var(--color-danger, #ef4444);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .error-message {
    margin: 0 0 0.75rem 0;
    font-size: 0.95rem;
    line-height: 1.5;
    color: var(--color-text, #f1f5f9);
    padding-right: 2rem;
  }

  .error-details {
    margin-top: 0.75rem;
    padding-top: 0.75rem;
    border-top: 1px solid var(--color-border, rgba(255, 255, 255, 0.1));
  }

  .details-title {
    margin: 0 0 0.5rem 0;
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--color-text-muted, #94a3b8);
  }

  .details-list {
    margin: 0;
    padding-left: 1.25rem;
    list-style-type: disc;
  }

  .detail-item {
    margin-bottom: 0.375rem;
    font-size: 0.875rem;
    line-height: 1.4;
  }

  .detail-field {
    font-weight: 600;
    color: var(--color-danger-light, #f87171);
  }

  .detail-separator {
    margin: 0 0.25rem;
    color: var(--color-text-muted, #94a3b8);
  }

  .detail-reason {
    color: var(--color-text, #f1f5f9);
  }

  .detail-received {
    display: block;
    margin-top: 0.125rem;
    font-size: 0.8rem;
    color: var(--color-text-muted, #94a3b8);
    font-family: monospace;
  }
</style>
