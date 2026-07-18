<script lang="ts">
  interface Props {
    size?: number;
    class?: string;
  }

  const { size = 60, class: className = "" }: Props = $props();
</script>

<svg
  class="constellation-icon {className}"
  width={size}
  height={size}
  viewBox="0 0 100 100"
  xmlns="http://www.w3.org/2000/svg"
>
  <defs>
    <!-- Glow filter -->
    <filter id="starGlow" x="-50%" y="-50%" width="200%" height="200%">
      <feGaussianBlur stdDeviation="1.5" result="coloredBlur" />
      <feMerge>
        <feMergeNode in="coloredBlur" />
        <feMergeNode in="SourceGraphic" />
      </feMerge>
    </filter>

    <!-- Line gradient -->
    <linearGradient id="lineGradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#40a9ff;stop-opacity:0.8" />
      <stop offset="100%" style="stop-color:#ffcc00;stop-opacity:0.8" />
    </linearGradient>
  </defs>

  <!-- Constellation lines -->
  <g stroke="url(#lineGradient)" stroke-width="1" opacity="0.6">
    <line x1="20" y1="20" x2="50" y2="35" />
    <line x1="50" y1="35" x2="80" y2="20" />
    <line x1="50" y1="35" x2="35" y2="65" />
    <line x1="35" y1="65" x2="65" y2="65" />
    <line x1="65" y1="65" x2="50" y2="35" />
    <line x1="35" y1="65" x2="20" y2="85" />
    <line x1="65" y1="65" x2="80" y2="85" />
  </g>

  <!-- Stars -->
  <g filter="url(#starGlow)">
    <!-- Top left -->
    <circle cx="20" cy="20" r="3" fill="#40a9ff">
      <animate
        attributeName="opacity"
        values="0.7;1;0.7"
        dur="2s"
        repeatCount="indefinite"
      />
    </circle>

    <!-- Top right -->
    <circle cx="80" cy="20" r="3" fill="#40a9ff">
      <animate
        attributeName="opacity"
        values="0.8;1;0.8"
        dur="2.5s"
        repeatCount="indefinite"
      />
    </circle>

    <!-- Center (brightest) -->
    <circle cx="50" cy="35" r="4" fill="#ffcc00">
      <animate
        attributeName="r"
        values="4;5;4"
        dur="3s"
        repeatCount="indefinite"
      />
      <animate
        attributeName="opacity"
        values="0.9;1;0.9"
        dur="3s"
        repeatCount="indefinite"
      />
    </circle>

    <!-- Bottom left -->
    <circle cx="35" cy="65" r="2.5" fill="#ffcc00">
      <animate
        attributeName="opacity"
        values="0.6;1;0.6"
        dur="1.8s"
        repeatCount="indefinite"
      />
    </circle>

    <!-- Bottom right -->
    <circle cx="65" cy="65" r="2.5" fill="#ffcc00">
      <animate
        attributeName="opacity"
        values="0.7;1;0.7"
        dur="2.2s"
        repeatCount="indefinite"
      />
    </circle>

    <!-- Bottom corners -->
    <circle cx="20" cy="85" r="2" fill="#ff3333" opacity="0.8">
      <animate
        attributeName="opacity"
        values="0.6;0.9;0.6"
        dur="2s"
        repeatCount="indefinite"
      />
    </circle>
    <circle cx="80" cy="85" r="2" fill="#ff3333" opacity="0.8">
      <animate
        attributeName="opacity"
        values="0.7;1;0.7"
        dur="1.5s"
        repeatCount="indefinite"
      />
    </circle>
  </g>
</svg>

<style>
  .constellation-icon {
    filter: drop-shadow(0 0 8px rgba(64, 169, 255, 0.5));
  }
</style>
