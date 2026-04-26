<script lang="ts">
  import { Modal, Button } from './index';

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

  /* eslint-disable prefer-const */
  let {
    open = $bindable(false),
    title = 'Confirm',
    message,
    confirmText = 'Confirm',
    cancelText = 'Cancel',
    danger = false,
    onConfirm,
    onCancel
  }: Props = $props();
  /* eslint-enable prefer-const */

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

<Modal bind:open {title} onClose={handleClose}>
  <p class="modal-message">{message}</p>
  <div class="modal-actions">
    <Button variant="secondary" onClick={handleCancel}>
      {cancelText}
    </Button>
    <Button variant={danger ? 'danger' : 'primary'} onClick={handleConfirm}>
      {confirmText}
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
