<script lang="ts">
  import type { Snippet } from "svelte";
  import type { HTMLAttributes } from "svelte/elements";

  type ChipSize = "sm" | "md";

  interface Props extends HTMLAttributes<HTMLSpanElement> {
    children?: Snippet;
    color?: string;
    borderColor?: string;
    background?: string;
    size?: ChipSize;
    glow?: boolean;
  }

  let {
    children,
    color,
    borderColor,
    background,
    size = "md",
    glow = false,
    class: className = "",
    ...rest
  }: Props = $props();
</script>

<span
  class="chip {size} {className}"
  class:chip--glow={glow}
  style="{color ? `--chip-color: ${color};` : ''}{borderColor
    ? ` --chip-border: ${borderColor};`
    : ''}{background ? ` --chip-bg: ${background};` : ''}"
  {...rest}
>
  {@render children?.()}
</span>

<style>
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.35rem 0.75rem;
    border-radius: 999px;
    font-size: 0.85rem;
    font-weight: 600;
    border: 1px solid var(--chip-border, var(--carbon-border, #2d2d3d));
    background: var(--chip-bg, var(--carbon-graphene, #12121a));
    color: var(--chip-color, var(--carbon-text, #f0f0f5));
    transition:
      background 0.2s ease,
      border-color 0.2s ease,
      box-shadow 0.2s ease;
  }

  .chip--glow {
    box-shadow: 0 0 12px color-mix(in srgb, var(--chip-color, var(--carbon-glow-cyan, #22d3ee)) 20%, transparent);
  }

  .chip.sm {
    padding: 0.25rem 0.6rem;
    border-radius: 6px;
    font-size: 0.8rem;
  }
</style>
