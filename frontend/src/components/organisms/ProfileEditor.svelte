<script lang="ts">
  import { goto } from "$app/navigation";
  import Button from "$components/atoms/Button.svelte";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import Modal from "$components/atoms/Modal.svelte";
  import { currentUser, updateUserInfo, logout } from "$shared/stores/auth.svelte.js";
  import * as usersApi from "$shared/api/users";
  import { getCurrentLocale, setLocale, formatMessage, type Locale } from "$shared/utils/i18n";

  let name = $state("");
  let email = $state("");
  let selectedLocale = $state<Locale>("en");
  let isSaving = $state(false);
  let isDeleting = $state(false);
  let showDeleteConfirm = $state(false);
  let deletePassword = $state("");
  let localError = $state<string | null>(null);
  let successMessage = $state<string | null>(null);

  // Load current user data and locale
  $effect(() => {
    const user = currentUser();
    if (user) {
      name = user.login || "";
      email = user.email || "";
    }
    selectedLocale = getCurrentLocale();
  });

  async function handleSave() {
    localError = null;
    successMessage = null;
    isSaving = true;

    try {
      await usersApi.updateMe({ email: email || undefined });
      await updateUserInfo();
      successMessage = formatMessage("settings.saved", selectedLocale);
    } catch (e) {
      localError = e instanceof Error ? e.message : formatMessage("server.error", selectedLocale);
    } finally {
      isSaving = false;
    }
  }

  async function handleDelete() {
    if (!deletePassword) {
      localError = formatMessage("password.confirm", selectedLocale);
      return;
    }

    isDeleting = true;
    localError = null;

    try {
      await usersApi.deleteMe(deletePassword);
      await logout();
      goto("/auth/login");
    } catch (e) {
      localError = e instanceof Error ? e.message : formatMessage("server.error", selectedLocale);
      isDeleting = false;
    }
  }

  function openDeleteConfirm() {
    showDeleteConfirm = true;
    deletePassword = "";
    localError = null;
  }

  function closeDeleteConfirm() {
    showDeleteConfirm = false;
    deletePassword = "";
    localError = null;
  }

  async function handleLocaleChange(locale: Locale) {
    selectedLocale = locale;
    setLocale(locale);

    // Persist locale to backend user settings
    try {
      await usersApi.updateSetting("preferred_language", locale);
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to save locale setting:", e);
      }
    }

    // Reload page to apply new locale
    window.location.reload();
  }
</script>

<div class="profile-editor">
  <h2>{formatMessage("profile.editTitle", selectedLocale)}</h2>

  {#if successMessage}
    <div class="success-message">{successMessage}</div>
  {/if}

  <div class="form-group">
    <label for="name"
      >{formatMessage("title", selectedLocale)} ({formatMessage("readonly", selectedLocale)})</label
    >
    <input type="text" id="name" value={name} readonly disabled />
    <span class="hint">{formatMessage("login.readonly", selectedLocale)}</span>
  </div>

  <div class="form-group">
    <label for="email">{formatMessage("profile.emailLabel", selectedLocale)}</label>
    <input
      type="email"
      id="email"
      bind:value={email}
      placeholder={formatMessage("email.placeholder", selectedLocale)}
    />
  </div>

  <div class="form-group">
    <label for="locale">{formatMessage("profile.languageLabel", selectedLocale)}</label>
    <select
      id="locale"
      bind:value={selectedLocale}
      onchange={(e) => handleLocaleChange(e.currentTarget.value as Locale)}
    >
      <option value="en">{formatMessage("locale.en", selectedLocale)}</option>
      <option value="ru">{formatMessage("locale.ru", selectedLocale)}</option>
    </select>
    <span class="hint">{formatMessage("profile.languageHint", selectedLocale)}</span>
  </div>

  {#if localError}
    <ApiErrorDisplay error={{ message: localError, code: "PROFILE_ERROR" }} />
  {/if}

  <div class="actions">
    <Button variant="primary" disabled={isSaving} onClick={handleSave}>
      {isSaving ? formatMessage("saving", selectedLocale) : formatMessage("save", selectedLocale)}
    </Button>

    <Button variant="danger" onClick={openDeleteConfirm}>
      {formatMessage("delete.account", selectedLocale)}
    </Button>
  </div>
</div>

{#if showDeleteConfirm}
  <Modal
    title={formatMessage("delete.confirm", selectedLocale, { item: "account" })}
    open={showDeleteConfirm}
    onClose={closeDeleteConfirm}
  >
    <div class="delete-confirm">
      <p class="warning">
        ⚠️ {formatMessage("profile.deleteWarning", selectedLocale)}
      </p>

      <div class="form-group">
        <label for="delete-password">{formatMessage("password.confirm", selectedLocale)}</label>
        <input
          type="password"
          id="delete-password"
          bind:value={deletePassword}
          placeholder={formatMessage("profile.passwordPlaceholder", selectedLocale)}
        />
      </div>

      {#if localError}
        <ApiErrorDisplay error={{ message: localError, code: "DELETE_ERROR" }} />
      {/if}

      <div class="modal-actions">
        <Button variant="secondary" onClick={closeDeleteConfirm}>
          {formatMessage("cancel", selectedLocale)}
        </Button>
        <Button variant="danger" disabled={isDeleting || !deletePassword} onClick={handleDelete}>
          {isDeleting
            ? formatMessage("deleting", selectedLocale)
            : formatMessage("confirm.delete", selectedLocale)}
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
