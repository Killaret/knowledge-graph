<script lang="ts">
  import type { Snippet } from "svelte";

  type BevelVariant = "cockpit" | "note" | "card";

  interface Props {
    children?: Snippet;
    class?: string;
    variant?: BevelVariant;
    /** Outer bevel cut (in px) — the deeper corner facing the screen edge. */
    outer?: number;
    /** Inner bevel cut (in px) — the slimmer corner facing the content. */
    inner?: number;
    /** Border color. Falls back to a subtle cyan border. */
    borderColor?: string;
    /** Glow color for the soft outer shadow. */
    glowColor?: string;
    /** Color used for the top light gradient. */
    shadeColor?: string;
    /** Whether the shade pseudo-elements are rendered. */
    shaded?: boolean;
    /** Fill the parent height. */
    fullHeight?: boolean;
    /** Inline style additions. */
    style?: string;
  }

  const {
    children,
    class: className = "",
    variant = "cockpit",
    outer = variant === "note" ? 24 : 22,
    inner = 6,
    borderColor,
    glowColor,
    shadeColor,
    shaded = true,
    fullHeight = false,
    style = "",
  }: Props = $props();

  const outerCss = $derived(`${outer}px`);
  const innerCss = $derived(`${inner}px`);
</script>

<div
  class="bevel-surface {variant} {className}"
  class:full-height={fullHeight}
  class:shaded
  style="--bevel-outer: {outerCss}; --bevel-inner: {innerCss};{borderColor
    ? ` --bevel-border: ${borderColor};`
    : ''}{glowColor ? ` --bevel-glow: ${glowColor};` : ''}{shadeColor
    ? ` --bevel-shade: ${shadeColor};`
    : ''} {style}"
>
  {@render children?.()}
</div>

<style>
  .bevel-surface {
    position: relative;
    z-index: 0;
    background:
      linear-gradient(
        135deg,
        rgba(22, 24, 36, 0.96) 0%,
        rgba(10, 10, 15, 0.96) 50%,
        rgba(12, 14, 22, 0.96) 100%
      ),
      var(--carbon-gradient-card, linear-gradient(145deg, rgba(30, 30, 42, 0.7) 0%, rgba(18, 18, 26, 0.9) 100%));
    border: 1px solid var(--bevel-border, rgba(45, 212, 191, 0.25));
    box-shadow:
      0 0 28px rgba(0, 0, 0, 0.55),
      0 0 12px var(--bevel-glow, rgba(45, 212, 191, 0.08));
    overflow: hidden;
    clip-path: polygon(
      var(--bevel-outer) 0%,
      calc(100% - var(--bevel-inner)) 0%,
      100% var(--bevel-inner),
      100% calc(100% - var(--bevel-outer)),
      calc(100% - var(--bevel-outer)) 100%,
      var(--bevel-inner) 100%,
      0% calc(100% - var(--bevel-inner)),
      0% var(--bevel-outer)
    );
    transition:
      clip-path 0.35s ease,
      box-shadow 0.35s ease,
      border-color 0.35s ease;
  }

  .bevel-surface.full-height {
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .bevel-surface.shaded::before,
  .bevel-surface.shaded::after {
    content: "";
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    pointer-events: none;
    z-index: -1;
  }

  .bevel-surface.shaded::before {
    background: linear-gradient(
      180deg,
      color-mix(in srgb, var(--bevel-shade, var(--carbon-glow-cyan, #22d3ee)) 12%, transparent) 0%,
      transparent 32%
    );
  }

  .bevel-surface.shaded::after {
    background: linear-gradient(
      0deg,
      rgba(0, 0, 0, 0.35) 0%,
      transparent 18%,
      transparent 78%,
      color-mix(in srgb, var(--bevel-shade, var(--carbon-glow-cyan, #22d3ee)) 8%, transparent) 100%
    );
  }

  @media (prefers-reduced-motion: reduce) {
    .bevel-surface {
      transition: none;
    }
  }
</style>
