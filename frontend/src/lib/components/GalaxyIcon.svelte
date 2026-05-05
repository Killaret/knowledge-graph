<script lang="ts">
  import { onMount } from 'svelte';
  
  interface Props {
    size?: number;
    class?: string;
  }
  
  let { size = 80, class: className = '' }: Props = $props();
  
  let rotation = $state(0);
  let animationId: number;
  
  onMount(() => {
    let lastTime = 0;
    const animate = (time: number) => {
      if (time - lastTime > 16) { // ~60fps
        rotation = (time * 0.01) % 360;
        lastTime = time;
      }
      animationId = requestAnimationFrame(animate);
    };
    animationId = requestAnimationFrame(animate);
    
    return () => cancelAnimationFrame(animationId);
  });
</script>

<svg
  class="galaxy-icon {className}"
  width={size}
  height={size}
  viewBox="0 0 100 100"
  xmlns="http://www.w3.org/2000/svg"
>
  <defs>
    <!-- Gradient for the galaxy spiral -->
    <linearGradient id="galaxyGradient" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#40a9ff;stop-opacity:1" />
      <stop offset="50%" style="stop-color:#ffcc00;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#ff3333;stop-opacity:1" />
    </linearGradient>
    
    <!-- Glow filter -->
    <filter id="galaxyGlow" x="-50%" y="-50%" width="200%" height="200%">
      <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
      <feMerge>
        <feMergeNode in="coloredBlur"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
  </defs>
  
  <!-- Outer rotating ring -->
  <g transform="translate(50, 50) rotate({rotation})" filter="url(#galaxyGlow)">
    <!-- Spiral arms -->
    <ellipse 
      cx="0" 
      cy="0" 
      rx="35" 
      ry="12" 
      fill="none" 
      stroke="url(#galaxyGradient)" 
      stroke-width="1.5"
      opacity="0.6"
    />
    <ellipse 
      cx="0" 
      cy="0" 
      rx="35" 
      ry="12" 
      fill="none" 
      stroke="url(#galaxyGradient)" 
      stroke-width="1.5"
      opacity="0.6"
      transform="rotate(60)"
    />
    <ellipse 
      cx="0" 
      cy="0" 
      rx="35" 
      ry="12" 
      fill="none" 
      stroke="url(#galaxyGradient)" 
      stroke-width="1.5"
      opacity="0.6"
      transform="rotate(120)"
    />
  </g>
  
  <!-- Static stars around the galaxy -->
  <circle cx="20" cy="20" r="1.5" fill="#40a9ff" opacity="0.8">
    <animate attributeName="opacity" values="0.8;1;0.8" dur="2s" repeatCount="indefinite" />
  </circle>
  <circle cx="80" cy="25" r="1" fill="#ffcc00" opacity="0.7">
    <animate attributeName="opacity" values="0.7;1;0.7" dur="1.5s" repeatCount="indefinite" />
  </circle>
  <circle cx="25" cy="75" r="1" fill="#ff3333" opacity="0.6">
    <animate attributeName="opacity" values="0.6;1;0.6" dur="2.5s" repeatCount="indefinite" />
  </circle>
  <circle cx="75" cy="80" r="1.5" fill="#a78bfa" opacity="0.7">
    <animate attributeName="opacity" values="0.7;1;0.7" dur="1.8s" repeatCount="indefinite" />
  </circle>
  
  <!-- Central core -->
  <circle cx="50" cy="50" r="8" fill="url(#galaxyGradient)" filter="url(#galaxyGlow)">
    <animate attributeName="r" values="8;9;8" dur="3s" repeatCount="indefinite" />
    <animate attributeName="opacity" values="0.9;1;0.9" dur="3s" repeatCount="indefinite" />
  </circle>
  
  <!-- Inner core highlight -->
  <circle cx="50" cy="50" r="4" fill="#ffffff" opacity="0.9">
    <animate attributeName="opacity" values="0.9;1;0.9" dur="2s" repeatCount="indefinite" />
  </circle>
</svg>

<style>
  .galaxy-icon {
    filter: drop-shadow(0 0 10px rgba(255, 204, 0, 0.4));
  }
</style>
