<script lang="ts">
  type ButtonVariant = "primary" | "secondary" | "danger" | "ghost";
  type ButtonType = "button" | "submit" | "reset";
  type ButtonSize = "sm" | "md" | "lg";

  interface Props {
    variant?: ButtonVariant;
    type?: ButtonType;
    size?: ButtonSize;
    disabled?: boolean;
    onClick?: (e: MouseEvent) => void;
    children?: import("svelte").Snippet;
    "data-testid"?: string;
    "aria-label"?: string;
    class?: string;
    style?: string;
  }

  const {
    variant = "primary",
    type = "button",
    size = "md",
    disabled = false,
    onClick,
    children,
    class: className = "",
    style = "",
    ...restProps
  }: Props = $props();

  function handleClick(e: MouseEvent) {
    if (!disabled) {
      onClick?.(e);
    }
  }
</script>

<button
  {type}
  class="button {variant} {size} {className}"
  class:disabled
  onclick={handleClick}
  {disabled}
  {style}
  {...restProps}
>
  {@render children?.()}
</button>

<style>
  .button {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 0.55rem 1.1rem;
    border-radius: 10px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all var(--carbon-transition, 0.25s ease);
    border: 1px solid transparent;
    outline: none;
    text-shadow: 0 1px 2px rgba(0, 0, 0, 0.25);
  }

  .button:focus-visible {
    box-shadow: var(--carbon-focus-ring, 0 0 0 3px rgba(34, 211, 238, 0.15));
  }

  /* Primary — diamond/cyan + graphene violet */
  .button.primary {
    background: var(--carbon-gradient-primary, linear-gradient(135deg, #22d3ee 0%, #8b5cf6 100%));
    color: white;
    box-shadow: var(--carbon-glow-primary, 0 4px 15px rgba(34, 211, 238, 0.3));
  }

  .button.primary:hover:not(.disabled) {
    transform: translateY(-2px);
    filter: brightness(1.08);
    box-shadow:
      0 8px 25px rgba(34, 211, 238, 0.35),
      0 0 30px rgba(139, 92, 246, 0.25);
  }

  .button.primary:active:not(.disabled) {
    transform: translateY(0);
    filter: brightness(0.95);
  }

  /* Secondary — graphite */
  .button.secondary {
    background: var(--carbon-graphene, #12121a);
    color: var(--carbon-text, #f0f0f5);
    border-color: var(--carbon-border, #2d2d3d);
  }

  .button.secondary:hover:not(.disabled) {
    background: var(--carbon-c70, #1a1a24);
    border-color: var(--carbon-border-active, #4b4b5e);
    box-shadow: 0 0 12px rgba(139, 92, 246, 0.12);
  }

  .button.secondary:active:not(.disabled) {
    background: var(--carbon-black, #050508);
  }

  /* Danger — fullerene red */
  .button.danger {
    background: var(--carbon-gradient-danger, linear-gradient(135deg, #ff3a2f 0%, #c2410c 100%));
    color: white;
    box-shadow: var(--carbon-glow-red, 0 4px 15px rgba(255, 58, 47, 0.3));
  }

  .button.danger:hover:not(.disabled) {
    transform: translateY(-2px);
    filter: brightness(1.1);
    box-shadow: 0 8px 25px rgba(255, 58, 47, 0.4);
  }

  .button.danger:active:not(.disabled) {
    transform: translateY(0);
    filter: brightness(0.95);
  }

  /* Ghost — transparent carbon with amber shimmer */
  .button.ghost {
    background: rgba(255, 255, 255, 0.04);
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    color: var(--carbon-text, #f0f0f5);
    border: 1px solid var(--carbon-border, #2d2d3d);
  }

  .button.ghost:hover:not(.disabled) {
    background: rgba(245, 158, 11, 0.08);
    border-color: rgba(245, 158, 11, 0.4);
    box-shadow: var(--carbon-glow-amber, 0 0 16px rgba(245, 158, 11, 0.35));
    transform: translateY(-1px);
  }

  .button.ghost:active:not(.disabled) {
    background: rgba(245, 158, 11, 0.12);
    transform: translateY(0);
  }

  /* Sizes */
  .button.sm {
    padding: 0.35rem 0.7rem;
    font-size: 12px;
    border-radius: 8px;
  }

  .button.lg {
    padding: 0.75rem 1.5rem;
    font-size: 16px;
    border-radius: 12px;
  }

  /* Disabled state */
  .button.disabled {
    opacity: 0.5;
    cursor: not-allowed;
    transform: none !important;
    box-shadow: none !important;
    filter: grayscale(0.5);
  }

  @media (prefers-reduced-motion: reduce) {
    .button,
    .button:hover:not(.disabled),
    .button:active:not(.disabled) {
      transition: none;
      transform: none;
    }
  }
</style>
