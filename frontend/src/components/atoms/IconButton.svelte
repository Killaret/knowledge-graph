<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLButtonAttributes } from "svelte/elements";

  type IconButtonVariant = "default" | "danger" | "ghost";
  type IconButtonSize = "sm" | "md";

  interface Props extends HTMLButtonAttributes {
    children?: Snippet;
    variant?: IconButtonVariant;
    size?: IconButtonSize;
    title?: string;
    onClick?: (e: MouseEvent) => void;
    disabled?: boolean;
  }

  let {
    children,
    variant = "default",
    size = "md",
    onClick,
    disabled = false,
    class: className = "",
    ...rest
  }: Props = $props();

  function handleClick(e: MouseEvent) {
    if (!disabled) onClick?.(e);
  }
</script>

<button
  type="button"
  class="icon-button {variant} {size} {className}"
  class:disabled
  {disabled}
  onclick={handleClick}
  {...rest}
>
  {@render children?.()}
</button>

<style>
  .icon-button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
    border: 1px solid var(--carbon-border, #2d2d3d);
    background: rgba(255, 255, 255, 0.06);
    color: var(--carbon-text, #f0f0f5);
    cursor: pointer;
    transition:
      background 0.2s ease,
      border-color 0.2s ease,
      color 0.2s ease,
      box-shadow 0.2s ease,
      transform 0.15s ease;
    line-height: 1;
  }

  .icon-button.md {
    width: 32px;
    height: 32px;
    font-size: 16px;
  }

  .icon-button.sm {
    width: 24px;
    height: 24px;
    font-size: 12px;
  }

  .icon-button:hover:not(.disabled) {
    background: rgba(255, 255, 255, 0.12);
    border-color: var(--carbon-border-active, #4b4b5e);
  }

  .icon-button.danger:hover:not(.disabled) {
    background: rgba(248, 113, 113, 0.12);
    border-color: rgba(248, 113, 113, 0.4);
    color: var(--carbon-glow-red, #ff3a2f);
    box-shadow: 0 0 10px rgba(248, 113, 113, 0.2);
  }

  .icon-button.ghost {
    background: transparent;
    border-color: transparent;
  }

  .icon-button.ghost:hover:not(.disabled) {
    background: rgba(255, 255, 255, 0.06);
    border-color: var(--carbon-border, #2d2d3d);
  }

  .icon-button:active:not(.disabled) {
    transform: scale(0.96);
  }

  .icon-button:disabled,
  .icon-button.disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .icon-button:focus-visible {
    outline: none;
    box-shadow: var(--carbon-focus-ring, 0 0 0 3px rgba(34, 211, 238, 0.15));
  }
</style>
