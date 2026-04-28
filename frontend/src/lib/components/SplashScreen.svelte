<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { fade } from 'svelte/transition';

  interface Props {
    duration?: number;
  }

  const { duration = 2500 }: Props = $props();

  let visible = $state(true);
  let hasShown = $state(false);

  onMount(() => {
    // Check if already shown in this session
    if (browser) {
      hasShown = sessionStorage.getItem('splash-shown') === 'true';
      if (hasShown) {
        visible = false;
        return;
      }
      sessionStorage.setItem('splash-shown', 'true');
    }

    // Auto-hide after duration
    const timer = setTimeout(() => {
      visible = false;
    }, duration);

    return () => clearTimeout(timer);
  });
</script>

{#if visible}
  <div
    class="splash-screen"
    out:fade={{ duration: 800 }}
    role="img"
    aria-label="Weltall Protocol - Knowledge Graph"
  >
    <!-- Dark radial gradient background -->
    <div class="splash-background"></div>

    <!-- SVG Cross (Xenogears symbol) -->
    <div class="cross-container">
      <svg
        class="cross-svg"
        viewBox="0 0 100 100"
        width="120"
        height="120"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <!-- Gradient matching cosmic theme -->
          <linearGradient id="grad-x" x1="0%" y1="0%" x2="100%" y2="100%">
            <stop offset="0%" style="stop-color:#40a9ff;stop-opacity:1" />
            <stop offset="50%" style="stop-color:#ffcc00;stop-opacity:1" />
            <stop offset="100%" style="stop-color:#ff3333;stop-opacity:1" />
          </linearGradient>
          <!-- Roughness filter for organic feel -->
          <filter id="roughness">
            <feTurbulence type="fractalNoise" baseFrequency="0.04" numOctaves="3" result="noise" />
            <feDisplacementMap in="SourceGraphic" in2="noise" scale="2" />
          </filter>
          <!-- Glow filter -->
          <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur stdDeviation="3" result="coloredBlur" />
            <feMerge>
              <feMergeNode in="coloredBlur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        <!-- Pulsing circle background -->
        <circle
          class="pulse-circle"
          cx="50"
          cy="50"
          r="35"
          fill="none"
          stroke="url(#grad-x)"
          stroke-width="1"
          opacity="0.5"
          filter="url(#glow)"
        />

        <!-- The cross symbol -->
        <g filter="url(#roughness)">
          <!-- Vertical bar -->
          <rect x="46" y="20" width="8" height="60" fill="url(#grad-x)" rx="1" />
          <!-- Horizontal bar -->
          <rect x="20" y="46" width="60" height="8" fill="url(#grad-x)" rx="1" />
        </g>

        <!-- Center gem -->
        <circle cx="50" cy="50" r="8" fill="#ffcc00" filter="url(#glow)">
          <animate
            attributeName="opacity"
            values="0.8;1;0.8"
            dur="2s"
            repeatCount="indefinite"
          />
        </circle>
      </svg>
    </div>

    <!-- App name -->
    <div class="app-title">
      <h1>Knowledge Graph</h1>
      <p class="subtitle">Explore the cosmos of ideas</p>
    </div>

    <!-- Loading indicator -->
    <div class="loading-bar">
      <div class="loading-progress"></div>
    </div>
  </div>
{/if}

<style>
  .splash-screen {
    position: fixed;
    inset: 0;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    overflow: hidden;
  }

  .splash-background {
    position: absolute;
    inset: 0;
    background: radial-gradient(ellipse at 50% 100%, #0a0a1a 0%, #000 80%);
    z-index: -1;
  }

  .cross-container {
    position: relative;
    animation: float 3s ease-in-out infinite;
  }

  .cross-svg {
    filter: drop-shadow(0 0 20px rgba(255, 204, 0, 0.4));
  }

  .pulse-circle {
    animation: pulse 2s ease-in-out infinite;
  }

  .app-title {
    margin-top: 2rem;
    text-align: center;
    animation: fadeInUp 1s ease 0.5s both;
  }

  .app-title h1 {
    font-size: 2rem;
    font-weight: 700;
    color: #e0e0e0;
    margin: 0;
    text-shadow: 0 0 20px rgba(255, 204, 0, 0.5);
    letter-spacing: 0.1em;
  }

  .subtitle {
    font-size: 1rem;
    color: rgba(255, 255, 255, 0.6);
    margin: 0.5rem 0 0;
    font-style: italic;
  }

  .loading-bar {
    position: absolute;
    bottom: 100px;
    width: 200px;
    height: 2px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 1px;
    overflow: hidden;
  }

  .loading-progress {
    height: 100%;
    background: linear-gradient(90deg, #40a9ff, #ffcc00);
    animation: load 2.5s ease-out forwards;
    box-shadow: 0 0 10px rgba(255, 204, 0, 0.5);
  }

  @keyframes float {
    0%, 100% {
      transform: translateY(0);
    }
    50% {
      transform: translateY(-10px);
    }
  }

  @keyframes pulse {
    0%, 100% {
      transform: scale(1);
      opacity: 0.5;
    }
    50% {
      transform: scale(1.1);
      opacity: 0.8;
    }
  }

  @keyframes fadeInUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @keyframes load {
    from {
      width: 0%;
    }
    to {
      width: 100%;
    }
  }

  /* Reduced motion support */
  @media (prefers-reduced-motion: reduce) {
    .cross-container,
    .pulse-circle,
    .loading-progress {
      animation: none;
    }

    .app-title {
      animation: none;
      opacity: 1;
    }
  }
</style>
