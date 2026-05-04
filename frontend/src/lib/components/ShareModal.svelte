<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import { shareNote, createShareLink, revokeShare } from '$lib/api/sharing.js';
  import type { NoteShare, ShareLink } from '$lib/types.js';
  
  interface Props {
    noteId: string;
    noteTitle: string;
  }
  
  let { noteId, noteTitle }: Props = $props();
  
  const dispatch = createEventDispatcher<{
    close: void;
    shared: { share: NoteShare | ShareLink };
    revoked: { shareId: string };
  }>();
  
  let activeTab = $state<'users' | 'link'>('users');
  let isLoading = $state(false);
  let error = $state<string | null>(null);
  let success = $state<string | null>(null);
  
  // User sharing
  let userEmail = $state('');
  let permission = $state<'view' | 'edit'>('view');
  
  // Link sharing
  let linkPermission = $state<'view' | 'edit'>('view');
  let expiresIn = $state<number | undefined>(undefined);
  let maxUses = $state<number | undefined>(undefined);
  let generatedLink = $state<ShareLink | null>(null);
  
  function closeModal() {
    dispatch('close');
  }
  
  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      closeModal();
    }
  }
  
  async function handleShareWithUser() {
    if (!userEmail.trim()) {
      error = 'Введите email пользователя';
      return;
    }
    
    isLoading = true;
    error = null;
    success = null;
    
    try {
      const share = await shareNote(noteId, {
        shared_with_email: userEmail.trim(),
        permission
      });
      success = `Доступ предоставлен пользователю ${userEmail}`;
      userEmail = '';
      dispatch('shared', { share });
    } catch (err) {
      error = err instanceof Error ? err.message : 'Ошибка при предоставлении доступа';
    } finally {
      isLoading = false;
    }
  }
  
  async function handleCreateLink() {
    isLoading = true;
    error = null;
    success = null;
    
    try {
      const link = await createShareLink(noteId, {
        permission: linkPermission,
        expires_in: expiresIn,
        max_uses: maxUses
      });
      generatedLink = link;
      success = 'Ссылка для доступа создана';
      dispatch('shared', { share: link });
    } catch (err) {
      error = err instanceof Error ? err.message : 'Ошибка при создании ссылки';
    } finally {
      isLoading = false;
    }
  }
  
  function copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
    success = 'Ссылка скопирована в буфер обмена';
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
<div class="modal-backdrop" on:click={handleBackdropClick}>
  <div class="modal">
    <div class="modal-header">
      <h2>Поделиться заметкой</h2>
      <p class="note-title">«{noteTitle}»</p>
      <button class="close-button" on:click={closeModal}>×</button>
    </div>
    
    <div class="modal-tabs">
      <button 
        class="tab-button"
        class:active={activeTab === 'users'}
        on:click={() => activeTab = 'users'}
      >
        Пользователю
      </button>
      <button 
        class="tab-button"
        class:active={activeTab === 'link'}
        on:click={() => activeTab = 'link'}
      >
        По ссылке
      </button>
    </div>
    
    <div class="modal-content">
      {#if error}
        <div class="error-message">{error}</div>
      {/if}
      
      {#if success}
        <div class="success-message">{success}</div>
      {/if}
      
      {#if activeTab === 'users'}
        <div class="share-section">
          <label class="field-label">
            Email пользователя
            <input
              type="email"
              bind:value={userEmail}
              placeholder="user@example.com"
              disabled={isLoading}
            />
          </label>
          
          <label class="field-label">
            Уровень доступа
            <select bind:value={permission} disabled={isLoading}>
              <option value="view">Только просмотр</option>
              <option value="edit">Редактирование</option>
            </select>
          </label>
          
          <button 
            class="action-button"
            on:click={handleShareWithUser}
            disabled={isLoading || !userEmail.trim()}
          >
            {isLoading ? 'Создание доступа...' : 'Предоставить доступ'}
          </button>
        </div>
      {:else}
        <div class="share-section">
          <label class="field-label">
            Уровень доступа
            <select bind:value={linkPermission} disabled={isLoading}>
              <option value="view">Только просмотр</option>
              <option value="edit">Редактирование</option>
            </select>
          </label>
          
          <div class="field-row">
            <label class="field-label">
              Истекает через (часы)
              <input
                type="number"
                bind:value={expiresIn}
                placeholder="Бессрочно"
                min="1"
                disabled={isLoading}
              />
            </label>
            
            <label class="field-label">
              Максимум использований
              <input
                type="number"
                bind:value={maxUses}
                placeholder="Без ограничений"
                min="1"
                disabled={isLoading}
              />
            </label>
          </div>
          
          {#if generatedLink}
            <div class="generated-link">
              <label class="field-label">
                Ссылка для доступа
                <div class="link-row">
                  <input
                    type="text"
                    value={`${window.location.origin}/shared/${generatedLink.token}`}
                    readonly
                  />
                  <button 
                    class="copy-button"
                    on:click={() => copyToClipboard(`${window.location.origin}/shared/${generatedLink?.token}`)}
                  >
                    Копировать
                  </button>
                </div>
              </label>
            </div>
          {/if}
          
          <button 
            class="action-button"
            on:click={handleCreateLink}
            disabled={isLoading}
          >
            {isLoading ? 'Создание ссылки...' : generatedLink ? 'Создать новую ссылку' : 'Создать ссылку'}
          </button>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }
  
  .modal {
    background: var(--color-surface);
    border-radius: var(--radius-lg);
    width: 100%;
    max-width: 480px;
    max-height: 90vh;
    overflow: hidden;
    box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
  }
  
  .modal-header {
    position: relative;
    padding: 1.5rem;
    border-bottom: 1px solid var(--color-border);
  }
  
  .modal-header h2 {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--color-text-primary);
  }
  
  .note-title {
    margin: 0.25rem 0 0;
    color: var(--color-text-secondary);
    font-size: 0.875rem;
  }
  
  .close-button {
    position: absolute;
    top: 1rem;
    right: 1rem;
    width: 32px;
    height: 32px;
    background: none;
    border: none;
    font-size: 1.5rem;
    color: var(--color-text-secondary);
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: var(--radius-sm);
    transition: all 0.2s;
  }
  
  .close-button:hover {
    background: var(--color-background);
    color: var(--color-text-primary);
  }
  
  .modal-tabs {
    display: flex;
    border-bottom: 1px solid var(--color-border);
  }
  
  .tab-button {
    flex: 1;
    padding: 1rem;
    background: none;
    border: none;
    color: var(--color-text-secondary);
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    border-bottom: 2px solid transparent;
  }
  
  .tab-button.active {
    color: var(--color-primary);
    border-bottom-color: var(--color-primary);
  }
  
  .tab-button:hover:not(.active) {
    color: var(--color-text-primary);
    background: var(--color-background);
  }
  
  .modal-content {
    padding: 1.5rem;
    overflow-y: auto;
  }
  
  .share-section {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }
  
  .field-label {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-primary);
  }
  
  .field-label input,
  .field-label select {
    padding: 0.625rem 0.75rem;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    font-size: 0.875rem;
    background: var(--color-background);
    color: var(--color-text-primary);
  }
  
  .field-label input:focus,
  .field-label select:focus {
    outline: none;
    border-color: var(--color-primary);
  }
  
  .field-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 1rem;
  }
  
  .action-button {
    padding: 0.75rem 1.5rem;
    background: var(--color-primary);
    color: white;
    border: none;
    border-radius: var(--radius-md);
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .action-button:hover:not(:disabled) {
    background: var(--color-primary-dark);
  }
  
  .action-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
  
  .generated-link {
    padding: 1rem;
    background: var(--color-background);
    border-radius: var(--radius-md);
  }
  
  .link-row {
    display: flex;
    gap: 0.5rem;
  }
  
  .link-row input {
    flex: 1;
  }
  
  .copy-button {
    padding: 0.5rem 1rem;
    background: var(--color-surface);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    font-size: 0.875rem;
    cursor: pointer;
    white-space: nowrap;
  }
  
  .copy-button:hover {
    background: var(--color-background);
  }
  
  .error-message {
    padding: 0.75rem;
    background: var(--color-error-bg, #fee2e2);
    color: var(--color-error, #dc2626);
    border-radius: var(--radius-md);
    font-size: 0.875rem;
  }
  
  .success-message {
    padding: 0.75rem;
    background: var(--color-success-bg, #dcfce7);
    color: var(--color-success, #16a34a);
    border-radius: var(--radius-md);
    font-size: 0.875rem;
  }
</style>
