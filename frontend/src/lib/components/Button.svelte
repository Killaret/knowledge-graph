<script lang="ts">
  type ButtonVariant = 'primary' | 'secondary' | 'danger';
  type ButtonType = 'button' | 'submit' | 'reset';

  interface Props {
    variant?: ButtonVariant;
    type?: ButtonType;
    disabled?: boolean;
    onClick?: (e: MouseEvent) => void;
    children?: import('svelte').Snippet;
  }

  const {
    variant = 'primary',
    type = 'button',
    disabled = false,
    onClick,
    children
  }: Props = $props();

  function handleClick(e: MouseEvent) {
    if (!disabled) {
      onClick?.(e);
    }
  }
</script>

<button
  {type}
  class="button {variant}"
  class:disabled
  onclick={handleClick}
  {disabled}
>
  {@render children?.()}
</button>

<style>
  .button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 0.5rem 1rem;
    border-radius: 6px;
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    border: 1px solid transparent;
    outline: none;
  }

  .button:focus {
    box-shadow: 0 0 0 3px var(--focus-ring-color);
  }

  /* Primary variant */
  .button.primary {
    background: var(--color-primary, #3b82f6);
    color: white;
    --focus-ring-color: var(--color-primary-light, rgba(59, 130, 246, 0.3));
  }

  .button.primary:hover:not(.disabled) {
    background: var(--color-primary-hover, #2563eb);
    transform: translateY(-1px);
    box-shadow: var(--shadow-md);
  }

  .button.primary:active:not(.disabled) {
    transform: translateY(0);
    box-shadow: var(--shadow-sm);
  }

  /* Secondary variant */
  .button.secondary {
    background: white;
    color: var(--color-text, #1f2937);
    border-color: var(--color-border, #e5e7eb);
    --focus-ring-color: var(--color-secondary-light, rgba(107, 114, 128, 0.2));
  }

  .button.secondary:hover:not(.disabled) {
    background: var(--color-surface-elevated, #f9fafb);
    border-color: var(--color-secondary, #6b7280);
  }

  .button.secondary:active:not(.disabled) {
    background: var(--color-background, #f3f4f6);
  }

  /* Danger variant */
  .button.danger {
    background: var(--color-danger, #ef4444);
    color: white;
    --focus-ring-color: var(--color-danger-light, rgba(239, 68, 68, 0.3));
  }

  .button.danger:hover:not(.disabled) {
    background: var(--color-danger-hover, #dc2626);
    transform: translateY(-1px);
    box-shadow: var(--shadow-md);
  }

  .button.danger:active:not(.disabled) {
    transform: translateY(0);
    box-shadow: var(--shadow-sm);
  }

  /* Disabled state */
  .button.disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none !important;
    box-shadow: none !important;
  }
</style>
