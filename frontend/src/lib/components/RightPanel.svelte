<script lang="ts">
  interface NodeType {
    type: string;
    label: string;
    color: string;
  }
  
  interface Props {
    nodeCount?: number;
    linkCount?: number;
    nodeTypes?: NodeType[];
    miniMapData?: { nodes: any[]; links: any[] };
  }
  
  let { 
    nodeCount = 0, 
    linkCount = 0, 
    nodeTypes = [],
    miniMapData
  }: Props = $props();
  
  let isExpanded = $state(true);
  let canvasRef = $state<HTMLCanvasElement | null>(null);
  let isMobile = $state(false);
  
  // Check mobile
  $effect(() => {
    const checkMobile = () => {
      isMobile = window.innerWidth < 768;
      if (isMobile) isExpanded = false;
    };
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  });
  
  // Draw mini-map
  $effect(() => {
    if (!canvasRef || !miniMapData || !isExpanded) return;
    
    const ctx = canvasRef.getContext('2d');
    if (!ctx) return;
    
    const width = canvasRef.width;
    const height = canvasRef.height;
    
    // Clear canvas
    ctx.fillStyle = 'rgba(0, 0, 0, 0.3)';
    ctx.fillRect(0, 0, width, height);
    
    // Find bounds
    const nodes = miniMapData.nodes;
    if (nodes.length === 0) return;
    
    let minX = Infinity, maxX = -Infinity, minY = Infinity, maxY = -Infinity;
    nodes.forEach((node: any) => {
      const x = node.x || 0;
      const y = node.y || 0;
      minX = Math.min(minX, x);
      maxX = Math.max(maxX, x);
      minY = Math.min(minY, y);
      maxY = Math.max(maxY, y);
    });
    
    // Add padding
    const padding = 20;
    const boundsWidth = Math.max(maxX - minX, 100);
    const boundsHeight = Math.max(maxY - minY, 100);
    
    // Scale to fit canvas
    const scaleX = (width - padding * 2) / boundsWidth;
    const scaleY = (height - padding * 2) / boundsHeight;
    const scale = Math.min(scaleX, scaleY);
    
    const offsetX = (width - boundsWidth * scale) / 2 - minX * scale;
    const offsetY = (height - boundsHeight * scale) / 2 - minY * scale;
    
    // Draw links
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.2)';
    ctx.lineWidth = 1;
    miniMapData.links.forEach((link: any) => {
      const source = typeof link.source === 'object' ? link.source : nodes.find((n: any) => n.id === link.source);
      const target = typeof link.target === 'object' ? link.target : nodes.find((n: any) => n.id === link.target);
      
      if (source && target) {
        ctx.beginPath();
        ctx.moveTo(source.x * scale + offsetX, source.y * scale + offsetY);
        ctx.lineTo(target.x * scale + offsetX, target.y * scale + offsetY);
        ctx.stroke();
      }
    });
    
    // Draw nodes
    nodes.forEach((node: any) => {
      const x = node.x * scale + offsetX;
      const y = node.y * scale + offsetY;
      
      ctx.beginPath();
      ctx.arc(x, y, 3, 0, Math.PI * 2);
      ctx.fillStyle = getNodeColor(node.type);
      ctx.fill();
    });
  });
  
  function getNodeColor(type: string): string {
    const colors: Record<string, string> = {
      star: '#ffaa00',
      planet: '#44aaff',
      comet: '#aaffdd',
      galaxy: '#aa44ff',
      default: '#888888'
    };
    return colors[type] || colors.default;
  }
  
  function toggleExpanded() {
    isExpanded = !isExpanded;
  }
</script>

<div class="right-panel" class:mobile={isMobile} class:collapsed={!isExpanded}>
  <button class="panel-header" onclick={toggleExpanded}>
    <span class="panel-title">Обзор</span>
    <svg 
      class="chevron" 
      class:rotated={!isExpanded}
      viewBox="0 0 24 24" 
      width="16" 
      height="16" 
      fill="none" 
      stroke="currentColor" 
      stroke-width="2"
    >
      <polyline points="6,9 12,15 18,9"/>
    </svg>
  </button>
  
  {#if isExpanded}
    <div class="panel-content">
      <!-- Stats -->
      <div class="stats">
        <div class="stat-item">
          <span class="stat-value">{nodeCount}</span>
          <span class="stat-label">узлов</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="stat-value">{linkCount}</span>
          <span class="stat-label">связей</span>
        </div>
      </div>
      
      <!-- Mini Map -->
      <div class="mini-map">
        <canvas 
          bind:this={canvasRef} 
          width="192" 
          height="128"
          class="map-canvas"
        ></canvas>
        {#if !miniMapData}
          <div class="map-placeholder">
            <svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="12" r="3"/>
              <path d="M12 2v4"/><path d="M12 18v4"/>
              <path d="m4.93 4.93 2.83 2.83"/><path d="m16.24 16.24 2.83 2.83"/>
              <path d="M2 12h4"/><path d="M18 12h4"/>
              <path d="m4.93 19.07 2.83-2.83"/><path d="m16.24 7.76 2.83-2.83"/>
            </svg>
            <span>Мини-карта</span>
          </div>
        {/if}
      </div>
      
      <!-- Legend -->
      {#if nodeTypes.length > 0}
        <div class="legend">
          <div class="legend-title">Типы узлов</div>
          <div class="legend-items">
            {#each nodeTypes as type}
              <div class="legend-item">
                <span class="legend-dot" style="background-color: {type.color}"></span>
                <span class="legend-label">{type.label}</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .right-panel {
    position: fixed;
    top: 1rem;
    right: 1rem;
    width: 14rem;
    z-index: 30;
    
    background: rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 1rem;
    overflow: hidden;
  }
  
  .right-panel.collapsed {
    width: auto;
  }
  
  .right-panel.mobile {
    top: auto;
    bottom: 5rem;
    right: 0.5rem;
    width: auto;
    max-width: 10rem;
  }
  
  .panel-header {
    width: 100%;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem 1rem;
    background: transparent;
    border: none;
    color: rgba(255, 255, 255, 0.8);
    cursor: pointer;
    transition: background 0.2s ease;
  }
  
  .panel-header:hover {
    background: rgba(255, 255, 255, 0.05);
  }
  
  .panel-title {
    font-weight: 500;
    font-size: 0.875rem;
  }
  
  .chevron {
    transition: transform 0.2s ease;
  }
  
  .chevron.rotated {
    transform: rotate(180deg);
  }
  
  .panel-content {
    padding: 0 1rem 1rem;
  }
  
  .stats {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    padding: 0.75rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }
  
  .stat-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.125rem;
  }
  
  .stat-value {
    font-size: 1.25rem;
    font-weight: 600;
    color: white;
  }
  
  .stat-label {
    font-size: 0.6875rem;
    color: rgba(255, 255, 255, 0.5);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  
  .stat-divider {
    width: 1px;
    height: 2rem;
    background: rgba(255, 255, 255, 0.1);
  }
  
  .mini-map {
    position: relative;
    margin: 0.75rem 0;
    border-radius: 0.5rem;
    overflow: hidden;
    background: rgba(0, 0, 0, 0.2);
  }
  
  .map-canvas {
    display: block;
    width: 100%;
    height: auto;
  }
  
  .map-placeholder {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    color: rgba(255, 255, 255, 0.3);
    font-size: 0.75rem;
  }
  
  .legend {
    margin-top: 0.5rem;
  }
  
  .legend-title {
    font-size: 0.6875rem;
    color: rgba(255, 255, 255, 0.5);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 0.5rem;
  }
  
  .legend-items {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }
  
  .legend-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.8125rem;
    color: rgba(255, 255, 255, 0.7);
  }
  
  .legend-dot {
    width: 0.625rem;
    height: 0.625rem;
    border-radius: 50%;
    flex-shrink: 0;
  }
  
  .legend-label {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  @media (max-width: 768px) {
    .right-panel {
      top: auto;
      bottom: 5rem;
      right: 0.5rem;
      width: auto;
      max-width: 10rem;
    }
  }
</style>
