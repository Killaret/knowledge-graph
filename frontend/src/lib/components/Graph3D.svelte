<script lang="ts">
  import { onMount, untrack } from 'svelte';
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';
  import type * as ThreeType from 'three';
  let THREE: any = null;
  let OrbitControls: any = null;
  let ThreeForceGraph: any = null;
  import type ThreeForceGraphType from 'three-forcegraph';

  // Types
  interface GraphNode {
    id: string;
    title: string;
    type: string;
  }

  interface GraphLink {
    source: string;
    target: string;
    weight: number;
  }

  interface GraphData {
    nodes: GraphNode[];
    links: GraphLink[];
  }

  interface ForceGraphNode extends GraphNode {
    __threeObj?: ThreeType.Group;
  }

  // Props
  const { 
    data,
    onNodeClick,
    onNodeRightClick
  }: { 
    data: GraphData;
    onNodeClick?: (node: GraphNode, event: MouseEvent) => void;
    onNodeRightClick?: (node: GraphNode, event: MouseEvent) => void;
  } = $props();

  // DOM refs
  let containerRef: HTMLDivElement;

  // Three.js refs (initialized in onMount)
  let graphInstance: ThreeForceGraphType | null = $state(null);
  let animationId: number = $state(0);
  let glowTextures: ThreeType.CanvasTexture[] = [];
  let labelTextures: ThreeType.Texture[] = [];
  let createdMaterials: ThreeType.Material[] = [];

  // WebGL support check
  function isWebGLSupported(): boolean {
    try {
      const canvas = document.createElement('canvas');
      return !!(
        window.WebGLRenderingContext &&
        (canvas.getContext('webgl') || canvas.getContext('experimental-webgl'))
      );
    } catch { /* ignore */
      return false;
    }
  }

  // Constants
  const CONFIG = {
    STAR_COUNT: 1000,
    CAMERA_Z: 200,
    NODE_REL_SIZE: 2,
    LABEL_MAX_LENGTH: 20,
    RESIZE_THROTTLE_MS: 100
  } as const;

  const TYPE_COLORS: Record<string, string> = {
    star: '#ffaa00',
    planet: '#44aaff',
    comet: '#aaffdd',
    galaxy: '#aa44ff',
    default: '#888888'
  };

  function getColorByType(type: string): string {
    return TYPE_COLORS[type] || TYPE_COLORS.default;
  }

  function generateGlowTexture(color: string): ThreeType.CanvasTexture {
    const canvas = document.createElement('canvas');
    canvas.width = 64;
    canvas.height = 64;
    const ctx = canvas.getContext('2d')!;

    const gradient = ctx.createRadialGradient(32, 32, 0, 32, 32, 32);
    gradient.addColorStop(0, color);
    gradient.addColorStop(0.4, color + '80');
    gradient.addColorStop(1, 'rgba(0,0,0,0)');

    ctx.fillStyle = gradient;
    ctx.fillRect(0, 0, 64, 64);

    const texture = new THREE.CanvasTexture(canvas);
    glowTextures.push(texture);
    return texture;
  }

  function disposeTextures(): void {
    glowTextures.forEach(texture => texture.dispose());
    glowTextures = [];
    labelTextures.forEach(t => t.dispose());
    labelTextures = [];
    createdMaterials.forEach(m => {
      try { (m as any).dispose?.(); } catch { /* ignore */ }
    });
    createdMaterials = [];
  }

  // Throttled resize handler
  function createThrottledResizeHandler(
    camera: ThreeType.PerspectiveCamera,
    renderer: ThreeType.WebGLRenderer
  ) {
    let timeoutId: ReturnType<typeof setTimeout> | null = null;
    
    return (entries: ResizeObserverEntry[]) => {
      if (timeoutId) return;
      
      timeoutId = setTimeout(() => {
        timeoutId = null;
        const entry = entries[0];
        const { width, height } = entry.contentRect;
        camera.aspect = width / height;
        camera.updateProjectionMatrix();
        renderer.setSize(width, height);
      }, CONFIG.RESIZE_THROTTLE_MS);
    };
  }

  // Reactive data update
  $effect(() => {
    const currentData = data;
    const currentGraph = graphInstance;
    
    if (currentGraph && currentData?.nodes && currentData?.links) {
      untrack(() => {
        currentGraph.graphData({
          nodes: currentData.nodes.map(n => ({ ...n })),
          links: currentData.links.map(l => ({ ...l }))
        });
      });
    }
  });

  onMount(() => { let _cleanup: (() => void) | undefined; (async () => {
    if (!containerRef || !browser) {
      return;
    }

    // If WebGL is not supported (headless browsers), render a testable placeholder
    if (!isWebGLSupported()) {
      const placeholder = document.createElement('div');
      placeholder.className = 'no-webgl-placeholder';
      placeholder.setAttribute('data-testid', 'main-graph-canvas');
      placeholder.textContent = 'WebGL not available';
      containerRef.appendChild(placeholder);

      // cleanup for this path
      _cleanup = () => {
        try { if (placeholder.parentNode === containerRef) containerRef.removeChild(placeholder); } catch { /* ignore */ }
      };
      return;
    }

    // Dynamic imports at runtime to avoid SSR issues
    THREE = await import('three');
    const orbitModule = await import('three/examples/jsm/controls/OrbitControls.js');
    OrbitControls = (orbitModule as any).OrbitControls ?? (orbitModule as any).default ?? orbitModule;
    const tfModule = await import('three-forcegraph');
    ThreeForceGraph = (tfModule as any).default ?? tfModule;

    const rect = containerRef.getBoundingClientRect();
    const width = rect.width;
    const height = rect.height;

    // Renderer
    const newRenderer = new THREE.WebGLRenderer({ antialias: true, alpha: true });
    newRenderer.setSize(width, height);
    newRenderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    containerRef.appendChild(newRenderer.domElement);
    // Mark the actual canvas with test id so tests can reliably wait for it
    try { newRenderer.domElement.setAttribute('data-testid', 'main-graph-canvas'); } catch { /* ignore in non-DOM env */ }


    // Scene
    const newScene = new THREE.Scene();
    newScene.background = new THREE.Color(0x0a1a3a);

    // Starfield
    const starGeometry = new THREE.BufferGeometry();
    const starPositions = new Float32Array(CONFIG.STAR_COUNT * 3);
    for (let i = 0; i < CONFIG.STAR_COUNT * 3; i += 3) {
      starPositions[i] = (Math.random() - 0.5) * 2000;
      starPositions[i + 1] = (Math.random() - 0.5) * 2000;
      starPositions[i + 2] = (Math.random() - 0.5) * 2000;
    }
    starGeometry.setAttribute('position', new THREE.BufferAttribute(starPositions, 3));
    const starMaterial = new THREE.PointsMaterial({
      color: 0xffffff,
      size: 1,
      transparent: true,
      opacity: 0.6
    });
    const stars = new THREE.Points(starGeometry, starMaterial);
    newScene.add(stars);

    // Camera
    const newCamera = new THREE.PerspectiveCamera(60, width / height, 0.1, 2000);
    newCamera.position.z = CONFIG.CAMERA_Z;

    // Controls
    const newControls = new OrbitControls(newCamera, newRenderer.domElement);
    newControls.enablePan = true;
    newControls.enableZoom = true;
    newControls.enableRotate = true;

    // Lights
    newScene.add(new THREE.AmbientLight(0x404040, 1.5));
    const dirLight = new THREE.DirectionalLight(0xffffff, 0.8);
    dirLight.position.set(100, 100, 100);
    newScene.add(dirLight);

    // Graph
    const newGraph = new ThreeForceGraph()
      .nodeRelSize(CONFIG.NODE_REL_SIZE)
      .nodeColor((node: ForceGraphNode) => getColorByType(node.type))
      .nodeThreeObject((node: ForceGraphNode) => {
        const color = getColorByType(node.type);
        const group = new THREE.Group();

        // Glow sprite
        const spriteMaterial = new THREE.SpriteMaterial({
          map: generateGlowTexture(color),
          blending: THREE.AdditiveBlending,
          depthWrite: false,
          transparent: true
        });
        const sprite = new THREE.Sprite(spriteMaterial);
        sprite.scale.set(6, 6, 1);
        group.add(sprite);

        // Label
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d')!;
        canvas.width = 256;
        canvas.height = 64;
        ctx.font = 'bold 20px sans-serif';
        ctx.textAlign = 'center';
        ctx.fillStyle = '#ffffff';
        ctx.shadowColor = 'rgba(0,0,0,0.8)';
        ctx.shadowBlur = 4;
        
        let title = node.title;
        if (title.length > CONFIG.LABEL_MAX_LENGTH) {
          title = title.slice(0, CONFIG.LABEL_MAX_LENGTH - 2) + '...';
        }
        ctx.fillText(title, 128, 32);

        const labelTexture = new THREE.CanvasTexture(canvas);
        const labelMaterial = new THREE.SpriteMaterial({
          map: labelTexture,
          transparent: true
        });
        const labelSprite = new THREE.Sprite(labelMaterial);
        labelSprite.scale.set(16, 4, 1);
        labelSprite.position.y = -15;
        group.add(labelSprite);

        return group;
      })
      .linkColor(() => 'rgba(100, 150, 200, 0.4)')
      .linkWidth((link: GraphLink) => Math.max(0.5, link.weight * 2))
      .linkDirectionalParticles(2)
      .linkDirectionalParticleWidth(1.5)
      .linkDirectionalParticleSpeed(0.01)
      .linkDirectionalParticleColor(() => 'rgba(150, 200, 255, 0.8)');

    graphInstance = newGraph;

    // Set initial data
    if (data?.nodes?.length && data?.links) {
      newGraph.graphData({
        nodes: data.nodes.map(n => ({ ...n })),
        links: data.links.map(l => ({ ...l }))
      });
      newScene.add(newGraph);
    }

    // Click handler
    const raycaster = new THREE.Raycaster();
    const mouse = new THREE.Vector2();

    const handleClick = (event: MouseEvent) => {
      const rect = newRenderer.domElement.getBoundingClientRect();
      mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
      mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

      raycaster.setFromCamera(mouse, newCamera);
      const intersects = raycaster.intersectObjects(newGraph.children, true);

      if (intersects.length > 0) {
        let obj: ThreeType.Object3D | null = intersects[0].object;
        while (obj) {
          const nodeData = (obj as any).__data;
          if (nodeData?.id) {
            if (onNodeClick) {
              onNodeClick(nodeData as GraphNode, event);
            } else {
              goto(`/notes/${nodeData.id}`);
            }
            break;
          }
          obj = obj.parent;
        }
      }
    };

    newRenderer.domElement.addEventListener('click', handleClick);
    
    // Right-click handler for context menu
    const handleRightClick = (event: MouseEvent) => {
      event.preventDefault();
      const rect = newRenderer.domElement.getBoundingClientRect();
      mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
      mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
      
      raycaster.setFromCamera(mouse, newCamera);
      const intersects = raycaster.intersectObjects(newGraph.children, true);
      
      if (intersects.length > 0) {
        let obj: ThreeType.Object3D | null = intersects[0].object;
        while (obj) {
          const nodeData = (obj as any).__data;
          if (nodeData?.id) {
            onNodeRightClick?.(nodeData as GraphNode, event);
            break;
          }
          obj = obj.parent;
        }
      }
    };
    
    newRenderer.domElement.addEventListener('contextmenu', handleRightClick);

    // Resize with throttle
    const resizeHandler = createThrottledResizeHandler(newCamera, newRenderer);
    const newResizeObserver = new ResizeObserver(resizeHandler);
    newResizeObserver.observe(containerRef);

    // Animation
    const animate = () => {
      animationId = requestAnimationFrame(animate);
      newGraph.tickFrame();
      newControls.update();
      newRenderer.render(newScene, newCamera);
    };
    animate();

    // Cleanup
    _cleanup = () => {
      try {
        newRenderer.domElement.removeEventListener('click', handleClick);
        newRenderer.domElement.removeEventListener('contextmenu', handleRightClick);
      } catch { /* ignore */ }
      cancelAnimationFrame(animationId);
      try { newResizeObserver.disconnect(); } catch { /* ignore */ }
      try { newControls.dispose(); } catch { /* ignore */ }

      // Pause/stop graph engine if available
      try {
        const maybePause = (newGraph as any)?.pause ?? (newGraph as any)?.stop ?? (newGraph as any)?.dispose;
        if (typeof maybePause === 'function') maybePause.call(newGraph);
      } catch { /* ignore */ }

      // Remove graph from scene
      try { if (newScene && newGraph) newScene.remove(newGraph as any); } catch { /* ignore */ }

      // Dispose tracked textures & materials
      disposeTextures();

      // Traverse scene to dispose leftover materials/textures
      try {
        newScene?.traverse((obj: any) => {
          if (obj?.material) {
            try { obj.material.dispose(); } catch { /* ignore */ }
          }
          if (obj?.geometry) {
            try { obj.geometry.dispose(); } catch { /* ignore */ }
          }
          if (obj?.texture) {
            try { obj.texture.dispose(); } catch { /* ignore */ }
          }
        });
      } catch { /* ignore */ }

      try { newRenderer.dispose(); } catch { /* ignore */ }
      try {
        if (newRenderer.domElement.parentNode === containerRef) {
          containerRef.removeChild(newRenderer.domElement);
        }
      } catch { /* ignore */ }

      // clear refs
      graphInstance = null;
      animationId = 0;
    };
      })();

    return () => { _cleanup?.(); };
  });
</script>

<div class="graph-3d-container" bind:this={containerRef} data-legacy-testid="graph-canvas">
  {#if !data?.nodes?.length}
    <div class="empty-state">No nodes to display</div>
  {/if}
</div>

<style>
  .graph-3d-container {
    width: 100%;
    height: 100%;
    position: relative;
    overflow: hidden;
  }

  .graph-3d-container :global(canvas) {
    display: block;
    width: 100%;
    height: 100%;
  }

  .empty-state {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    color: #888;
    font-size: 1.2rem;
  }
</style>
