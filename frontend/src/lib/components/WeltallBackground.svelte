<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  // Parallax intensity (0 = disabled, 1 = full)
  let parallaxIntensity = 0.3;
  let mouseX = $state(0.5);
  let mouseY = $state(0.5);
  let containerRef: HTMLDivElement;

  onMount(() => {
    if (!browser) return;

    const handleMouseMove = (e: MouseEvent) => {
      if (!containerRef) return;
      const rect = containerRef.getBoundingClientRect();
      mouseX = (e.clientX - rect.left) / rect.width;
      mouseY = (e.clientY - rect.top) / rect.height;
    };

    window.addEventListener('mousemove', handleMouseMove, { passive: true });

    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
    };
  });

  // Calculate parallax offset
  let offsetX = $derived((mouseX - 0.5) * parallaxIntensity * 20);
  let offsetY = $derived((mouseY - 0.5) * parallaxIntensity * 20);
</script>

<div
  class="weltall-background"
  bind:this={containerRef}
  role="presentation"
  aria-hidden="true"
>
  <!-- Nebula layer with slow drift animation -->
  <div
    class="nebula-layer"
    style="transform: translate({offsetX * 0.5}px, {offsetY * 0.5}px)"
  >
    <div class="nebula-blob blob-1"></div>
    <div class="nebula-blob blob-2"></div>
    <div class="nebula-blob blob-3"></div>
  </div>

  <!-- Embers layer (static pattern) -->
  <div
    class="embers-layer"
    style="transform: translate({offsetX * 0.3}px, {offsetY * 0.3}px)"
  >
    <div class="embers-pattern"></div>
  </div>

  <!-- Cross watermark (very subtle) -->
  <div
    class="cross-watermark"
    style="transform: translate({offsetX * 0.1}px, {offsetY * 0.1}px)"
  >
    <svg viewBox="0 0 100 100" width="300" height="300" opacity="0.03">
      <g stroke="#ffcc00" stroke-width="2" fill="none">
        <line x1="50" y1="20" x2="50" y2="80" />
        <line x1="20" y1="50" x2="80" y2="50" />
        <circle cx="50" cy="50" r="35" />
      </g>
    </svg>
  </div>

  <!-- Subtle grid overlay -->
  <div class="grid-overlay"></div>
</div>

<style>
  .weltall-background {
    position: fixed;
    inset: 0;
    z-index: -1;
    overflow: hidden;
    pointer-events: none;
    background: linear-gradient(180deg, #0a0a1a 0%, #000 100%);
  }

  /* Nebula layers */
  .nebula-layer {
    position: absolute;
    inset: -20%;
    transition: transform 0.1s ease-out;
  }

  .nebula-blob {
    position: absolute;
    border-radius: 50%;
    filter: blur(60px);
    opacity: 0.15;
  }

  .blob-1 {
    width: 600px;
    height: 600px;
    background: radial-gradient(circle, rgba(64, 169, 255, 0.3) 0%, transparent 70%);
    top: 10%;
    left: 20%;
    animation: drift 20s ease-in-out infinite;
  }

  .blob-2 {
    width: 500px;
    height: 500px;
    background: radial-gradient(circle, rgba(255, 51, 51, 0.2) 0%, transparent 70%);
    top: 50%;
    right: 10%;
    animation: drift 25s ease-in-out infinite reverse;
  }

  .blob-3 {
    width: 400px;
    height: 400px;
    background: radial-gradient(circle, rgba(255, 204, 0, 0.15) 0%, transparent 70%);
    bottom: 20%;
    left: 30%;
    animation: drift 30s ease-in-out infinite;
  }

  @keyframes drift {
    0%, 100% {
      transform: translate(0, 0) scale(1);
    }
    33% {
      transform: translate(30px, -30px) scale(1.1);
    }
    66% {
      transform: translate(-20px, 20px) scale(0.9);
    }
  }

  /* Embers layer */
  .embers-layer {
    position: absolute;
    inset: 0;
    transition: transform 0.1s ease-out;
  }

  .embers-pattern {
    position: absolute;
    inset: 0;
    background-image:
      radial-gradient(circle at 20% 30%, rgba(255, 204, 0, 0.03) 0%, transparent 2px),
      radial-gradient(circle at 80% 70%, rgba(64, 169, 255, 0.03) 0%, transparent 2px),
      radial-gradient(circle at 50% 50%, rgba(255, 51, 51, 0.02) 0%, transparent 3px);
    background-size: 100px 100px, 150px 150px, 200px 200px;
    animation: twinkle 4s ease-in-out infinite;
  }

  @keyframes twinkle {
    0%, 100% {
      opacity: 0.5;
    }
    50% {
      opacity: 1;
    }
  }

  /* Cross watermark */
  .cross-watermark {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    transition: transform 0.1s ease-out;
    pointer-events: none;
  }

  /* Subtle grid overlay */
  .grid-overlay {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba(255, 255, 255, 0.01) 1px, transparent 1px),
      linear-gradient(90deg, rgba(255, 255, 255, 0.01) 1px, transparent 1px);
    background-size: 50px 50px;
    mask-image: radial-gradient(ellipse at 50% 50%, black 0%, transparent 70%);
    -webkit-mask-image: radial-gradient(ellipse at 50% 50%, black 0%, transparent 70%);
  }

  /* Reduced motion support */
  @media (prefers-reduced-motion: reduce) {
    .nebula-blob,
    .embers-pattern {
      animation: none;
    }
  }
</style>
