<script lang="ts">
  import Button from "$components/atoms/Button.svelte";
  import Modal from "$components/atoms/Modal.svelte";
  import { mode } from "$shared/stores/lexicon-settings";
  import { Theme } from "$entities";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  interface Props {
    open: boolean;
    title?: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    danger?: boolean;
    onConfirm: () => void;
    onCancel: () => void;
  }

  /* eslint-disable prefer-const -- Svelte 5 $bindable() requires let, not const, see: https://svelte.dev/docs/svelte/$bindable */
  let {
    open = $bindable(false),
    title = t("confirmModal.title"),
    message,
    confirmText = t("confirmModal.confirm"),
    cancelText = t("confirmModal.cancel"),
    danger = false,
    onConfirm,
    onCancel,
  }: Props = $props();
  /* eslint-enable prefer-const */

  let currentMode = $state("standard");
  const theme = $derived(Theme.fromString(currentMode));

  // Subscribe to mode changes
  $effect(() => {
    const unsubscribe = mode.subscribe((m) => (currentMode = m));
    return unsubscribe;
  });

  // Compute display values reactively
  const displayTitle = $derived(
    theme.transformLabel(title, {
      [t("confirmModal.title")]: t("confirmModal.titleGalactic"),
    })
  );
  const displayConfirmText = $derived(
    theme.transformLabel(confirmText, {
      [t("confirmModal.confirm")]: t("confirmModal.confirmGalactic"),
    })
  );
  const displayCancelText = $derived(
    theme.transformLabel(cancelText, {
      [t("confirmModal.cancel")]: t("confirmModal.cancelGalactic"),
    })
  );

  function handleConfirm() {
    onConfirm();
  }

  function handleCancel() {
    onCancel();
  }

  function handleClose() {
    onCancel();
  }
</script>

<Modal bind:open title={displayTitle} onClose={handleClose}>
  <p class="modal-message">{message}</p>
  <div class="modal-actions">
    <Button variant="secondary" onClick={handleCancel} data-testid="confirm-modal-cancel">
      {displayCancelText}
    </Button>
    <Button
      variant={danger ? "danger" : "primary"}
      onClick={handleConfirm}
      data-testid="confirm-modal-confirm"
    >
      {displayConfirmText}
    </Button>
  </div>
</Modal>

<style>
  .modal-message {
    margin: 0 0 24px 0;
    font-size: 1rem;
    color: var(--color-text-secondary, #64748b);
    line-height: 1.5;
  }

  .modal-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  @media (max-width: 480px) {
    .modal-actions {
      flex-direction: column-reverse;
    }
  }
</style>
