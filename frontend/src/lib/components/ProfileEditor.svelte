<script lang="ts">
  import { goto } from '$app/navigation';
  import Button from './Button.svelte';
  import ApiErrorDisplay from './ApiErrorDisplay.svelte';
  import Modal from './Modal.svelte';
  import { currentUser, updateUserInfo, logout } from '$lib/stores/auth.svelte.js';
  import * as usersApi from '$lib/api/users';
  
  let name = $state('');
  let email = $state('');
  let isLoading = $state(false);
  let isSaving = $state(false);
  let isDeleting = $state(false);
  let showDeleteConfirm = $state(false);
  let deletePassword = $state('');
  let localError = $state<string | null>(null);
  let successMessage = $state<string | null>(null);
  
  // Load current user data
  $effect(() => {
    if (currentUser) {
      name = currentUser.login || '';
      email = currentUser.email || '';
    }
  });
  
  async function handleSave() {
    localError = null;
    successMessage = null;
    isSaving = true;
    
    try {
      await usersApi.updateMe({ email: email || undefined });
      await updateUserInfo();
      successMessage = 'Профиль обновлен';
    } catch (e) {
      localError = e instanceof Error ? e.message : 'Failed to update profile';
    } finally {
      isSaving = false;
    }
  }
  
  async function handleDelete() {
    if (!deletePassword) {
      localError = 'Введите пароль для подтверждения';
      return;
    }
    
    isDeleting = true;
    localError = null;
    
    try {
      await usersApi.deleteMe(deletePassword);
      await logout();
      goto('/auth/login');
    } catch (e) {
      localError = e instanceof Error ? e.message : 'Failed to delete account';
      isDeleting = false;
    }
  }
  
  function openDeleteConfirm() {
    showDeleteConfirm = true;
    deletePassword = '';
    localError = null;
  }
  
  function closeDeleteConfirm() {
    showDeleteConfirm = false;
    deletePassword = '';
    localError = null;
  }
</script>

<div class="profile-editor">
  <h2>Редактирование профиля</h2>
  
  {#if successMessage}
    <div class="success-message">{successMessage}</div>
  {/if}
  
  <div class="form-group">
    <label for="name">Логин (только чтение)</label>
    <input
      type="text"
      id="name"
      value={name}
      readonly
      disabled
    />
    <span class="hint">Логин нельзя изменить</span>
  </div>
  
  <div class="form-group">
    <label for="email">Email</label>
    <input
      type="email"
      id="email"
      bind:value={email}
      placeholder="Введите ваш email"
    />
  </div>
  
  {#if localError}
    <ApiErrorDisplay error={{ message: localError, code: 'PROFILE_ERROR' }} />
  {/if}
  
  <div class="actions">
    <Button 
      variant="primary" 
      disabled={isSaving}
      on:click={handleSave}
    >
      {isSaving ? 'Сохранение...' : 'Сохранить'}
    </Button>
    
    <Button 
      variant="danger" 
      on:click={openDeleteConfirm}
    >
      Удалить аккаунт
    </Button>
  </div>
</div>

{#if showDeleteConfirm}
  <Modal title="Подтверждение удаления" open={showDeleteConfirm} onClose={closeDeleteConfirm}>
    <div class="delete-confirm">
      <p class="warning">
        ⚠️ Внимание! Это действие необратимо. Все ваши данные будут удалены.
      </p>
      
      <div class="form-group">
        <label for="delete-password">Введите пароль для подтверждения</label>
        <input
          type="password"
          id="delete-password"
          bind:value={deletePassword}
          placeholder="Ваш пароль"
        />
      </div>
      
      {#if localError}
        <ApiErrorDisplay error={{ message: localError, code: 'DELETE_ERROR' }} />
      {/if}
      
      <div class="modal-actions">
        <Button variant="secondary" on:click={closeDeleteConfirm}>
          Отмена
        </Button>
        <Button 
          variant="danger" 
          disabled={isDeleting || !deletePassword}
          on:click={handleDelete}
        >
          {isDeleting ? 'Удаление...' : 'Подтвердить удаление'}
        </Button>
      </div>
    </div>
  </Modal>
{/if}

<style>
  .profile-editor {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-width: 500px;
    padding: 2rem;
    background: var(--color-surface);
    border-radius: var(--radius-lg);
    box-shadow: var(--shadow-md);
  }
  
  h2 {
    margin: 0 0 1rem;
    color: var(--color-text-primary);
  }
  
  .form-group {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  
  label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--color-text-secondary);
  }
  
  input {
    padding: 0.75rem;
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    background: var(--color-background);
    color: var(--color-text-primary);
    font-size: 1rem;
    transition: border-color 0.2s;
  }
  
  input:focus {
    outline: none;
    border-color: var(--color-primary);
  }
  
  input:disabled {
    background: var(--color-surface-elevated);
    color: var(--color-text-muted);
    cursor: not-allowed;
  }
  
  .hint {
    font-size: 0.75rem;
    color: var(--color-text-muted);
  }
  
  .success-message {
    padding: 0.75rem;
    background: var(--color-success-light, rgba(34, 197, 94, 0.1));
    border: 1px solid var(--color-success);
    border-radius: var(--radius-md);
    color: var(--color-success);
  }
  
  .actions {
    display: flex;
    gap: 1rem;
    margin-top: 1rem;
  }
  
  .delete-confirm {
    padding: 1rem;
  }
  
  .warning {
    color: var(--color-warning);
    font-weight: 500;
    margin-bottom: 1rem;
  }
  
  .modal-actions {
    display: flex;
    gap: 1rem;
    justify-content: flex-end;
    margin-top: 1rem;
  }
</style>
