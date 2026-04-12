<script lang="ts" module>
  export interface Props {
    open?: boolean;
    title?: string;
    message?: string;
    confirmText?: string;
    cancelText?: string;
    onConfirm?: () => void;
    onCancel?: () => void;
  }
</script>

<script lang="ts">
  import { Dialog } from './dialog/index.js';

  const props = $props();
  let open = props.open ?? false;
  const { title = 'Подтвердите действие', message = 'Вы уверены?', confirmText = 'Подтвердить', cancelText = 'Отмена', onConfirm, onCancel } = props;

  function handleConfirm() {
    open = false;
    onConfirm?.();
  }

  function handleCancel() {
    open = false;
    onCancel?.();
  }
</script>

<Dialog {open} onClose={handleCancel} className="confirm-dialog">
  <div class="confirm-dialog-content">
    <h2 class="dialog-title">{title}</h2>
    <p class="dialog-message">{message}</p>
    <div class="dialog-actions">
      <button class="btn-cancel" onclick={handleCancel}>
        {cancelText}
      </button>
      <button class="btn-confirm" onclick={handleConfirm}>
        {confirmText}
      </button>
    </div>
  </div>
</Dialog>

<style>
  .confirm-dialog-content {
    text-align: center;
  }

  .dialog-title {
    font-size: 1.25rem;
    font-weight: 600;
    margin-bottom: 1rem;
    color: #1a1a1a;
  }

  .dialog-message {
    color: #666;
    margin-bottom: 1.5rem;
    line-height: 1.5;
  }

  .dialog-actions {
    display: flex;
    gap: 0.75rem;
    justify-content: center;
  }

  .btn-cancel,
  .btn-confirm {
    padding: 0.5rem 1.5rem;
    border-radius: 0.5rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    border: 1px solid #e2e8f0;
  }

  .btn-cancel {
    background: white;
    color: #64748b;
  }

  .btn-cancel:hover {
    background: #f8fafc;
    border-color: #cbd5e1;
  }

  .btn-confirm {
    background: #dc2626;
    color: white;
    border-color: #dc2626;
  }

  .btn-confirm:hover {
    background: #b91c1c;
    border-color: #b91c1c;
  }

  :global(.confirm-dialog) {
    max-width: 400px;
  }
</style>
