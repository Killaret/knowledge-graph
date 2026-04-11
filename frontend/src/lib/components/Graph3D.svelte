<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';
  import { detectDeviceCapabilities, shouldUse3D, type DeviceCapabilities } from '$lib/utils/deviceCapabilities';

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

  let { 
    nodes = [] as GraphNode[],
    links = [] as GraphLink[]
  } = $props<{
    nodes: GraphNode[];
    links: GraphLink[];
  }>();

  let containerRef: HTMLDivElement;
  let renderer: any;
  let scene: any;
  let camera: any;
  let controls: any;
  let graphInstance: any;
  let animationId: number;
  let resizeObserver: ResizeObserver;
  let THREE: typeof import('three');
  let ThreeForceGraph: typeof import('three-forcegraph').default;
  let OrbitControls: any;
  
  // Device capabilities for optimization
  let deviceCaps: DeviceCapabilities;
  let isWebGLSupported = true;

  const typeColors: Record<string, string> = {
    star: '#ffaa00',
    planet: '#44aaff',
    comet: '#aaffdd',
    galaxy: '#aa44ff',
    default: '#888888'
  };

  function getColorByType(type: string): string {
    return typeColors[type] || typeColors.default;
  }

  function generateGlowTexture(color: string): THREE.CanvasTexture {
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
    return texture;
  }

  onMount(async () => {
    if (!containerRef || !browser) return;

    // Detect device capabilities for optimization
    deviceCaps = detectDeviceCapabilities();
    
    // Check WebGL support
    try {
      const canvas = document.createElement('canvas');
      const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
      if (!gl) {
        isWebGLSupported = false;
        return;
      }
    } catch (e) {
      isWebGLSupported = false;
      return;
    }

    // Dynamic imports for SSR compatibility
    THREE = await import('three');
    const forceGraphModule = await import('three-forcegraph');
    ThreeForceGraph = forceGraphModule.default;
    const controlsModule = await import('three/examples/jsm/controls/OrbitControls.js');
    OrbitControls = controlsModule.OrbitControls;

    // Get container dimensions
    const rect = containerRef.getBoundingClientRect();
    const width = rect.width;
    const height = rect.height;

    // Initialize renderer with optimizations
    renderer = new THREE.WebGLRenderer({ 
      antialias: deviceCaps.gpuTier !== 'low', // Disable antialiasing on low-end
      alpha: true,
      powerPreference: deviceCaps.isLowPower ? 'low-power' : 'high-performance'
    });
    renderer.setSize(width, height);
    renderer.setPixelRatio(deviceCaps.pixelRatio); // Use optimized pixel ratio
    containerRef.appendChild(renderer.domElement);

    // Initialize scene
    scene = new THREE.Scene();
    scene.background = new THREE.Color(0x0a1a3a);

    // Add starfield background (optimized count based on device)
    const starGeometry = new THREE.BufferGeometry();
    const starCount = deviceCaps.starCount; // Use optimized star count
    const starPositions = new Float32Array(starCount * 3);
    for (let i = 0; i < starCount * 3; i += 3) {
      starPositions[i] = (Math.random() - 0.5) * 2000;
      starPositions[i + 1] = (Math.random() - 0.5) * 2000;
      starPositions[i + 2] = (Math.random() - 0.5) * 2000;
    }
    starGeometry.setAttribute('position', new THREE.BufferAttribute(starPositions, 3));
    const starMaterial = new THREE.PointsMaterial({ 
      color: 0xffffff, 
      size: deviceCaps.isLowPower ? 1 : 2, // Smaller stars on low-end
      transparent: true, 
      opacity: deviceCaps.isLowPower ? 0.4 : 0.6 
    });
    const stars = new THREE.Points(starGeometry, starMaterial);
    scene.add(stars);

    // Initialize camera
    camera = new THREE.PerspectiveCamera(60, width / height, 0.1, 2000);
    camera.position.z = 200;

    // Initialize controls
    controls = new OrbitControls(camera, renderer.domElement);
    controls.enablePan = true;
    controls.enableZoom = true;
    controls.enableRotate = true;
    controls.minDistance = 50;
    controls.maxDistance = 500;

    // Add lights (simplified for low-end devices)
    const ambientLight = new THREE.AmbientLight(0x404040, deviceCaps.isLowPower ? 2.0 : 1.5);
    scene.add(ambientLight);

    if (!deviceCaps.isLowPower) {
      const directionalLight = new THREE.DirectionalLight(0xffffff, 0.8);
      directionalLight.position.set(100, 100, 100);
      scene.add(directionalLight);

      const pointLight = new THREE.PointLight(0x4488ff, 1, 500);
      pointLight.position.set(0, 0, 100);
      scene.add(pointLight);
    }

    // Create force graph
    graphInstance = new ThreeForceGraph<GraphNode, GraphLink>()
      .nodeRelSize(6)
      .nodeColor((node) => getColorByType((node as GraphNode).type))
      .nodeThreeObject((node) => {
        const n = node as GraphNode;
        const color = getColorByType(node.type);
        const group = new THREE.Group();

        // Main node sprite
        const spriteMaterial = new THREE.SpriteMaterial({
          map: generateGlowTexture(color),
          blending: THREE.AdditiveBlending,
          depthWrite: false,
          transparent: true
        });
        const sprite = new THREE.Sprite(spriteMaterial);
        sprite.scale.set(20, 20, 1);
        group.add(sprite);

        // Label
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d')!;
        canvas.width = 256;
        canvas.height = 64;
        ctx.font = 'bold 20px sans-serif';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillStyle = '#ffffff';
        ctx.shadowColor = 'rgba(0,0,0,0.8)';
        ctx.shadowBlur = 4;
        let title = n.title;
        if (title.length > 20) title = title.slice(0, 18) + '...';
        ctx.fillText(title, 128, 32);

        const labelTexture = new THREE.CanvasTexture(canvas);
        const labelMaterial = new THREE.SpriteMaterial({ 
          map: labelTexture, 
          transparent: true,
          depthWrite: false
        });
        const labelSprite = new THREE.Sprite(labelMaterial);
        labelSprite.scale.set(40, 10, 1);
        labelSprite.position.y = -15;
        group.add(labelSprite);

        return group;
      })
      .linkColor(() => 'rgba(100, 150, 200, 0.4)')
      .linkWidth((link) => Math.max(0.5, (link as GraphLink).weight * 2))
      .linkDirectionalParticles(deviceCaps.enableParticles ? 2 : 0) // Disable particles on low-end
      .linkDirectionalParticleWidth(deviceCaps.enableParticles ? 1.5 : 0)
      .linkDirectionalParticleSpeed(0.01)
      .linkDirectionalParticleColor(() => 'rgba(150, 200, 255, 0.8)');
    
    // Add click handler via DOM events
    const raycaster = new THREE.Raycaster();
    const mouse = new THREE.Vector2();
    
    const handleClick = (event: MouseEvent) => {
      const rect = renderer.domElement.getBoundingClientRect();
      mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
      mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
      
      raycaster.setFromCamera(mouse, camera);
      const intersects = raycaster.intersectObjects(graphInstance.children, true);
      
      if (intersects.length > 0) {
        // Find the node data from the intersected object
        let obj: THREE.Object3D | null = intersects[0].object;
        while (obj) {
          const nodeData = (obj as any).__data;
          if (nodeData && nodeData.id) {
            goto(`/notes/${nodeData.id}`);
            break;
          }
          obj = obj.parent;
        }
      }
    };
    
    const handleMouseMove = (event: MouseEvent) => {
      const rect = renderer.domElement.getBoundingClientRect();
      mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
      mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
      
      raycaster.setFromCamera(mouse, camera);
      const intersects = raycaster.intersectObjects(graphInstance.children, true);
      
      containerRef.style.cursor = intersects.length > 0 ? 'pointer' : 'default';
    };
    
    renderer.domElement.addEventListener('click', handleClick);
    renderer.domElement.addEventListener('mousemove', handleMouseMove);

    // Set data
    const graphData = {
      nodes: nodes.map(n => ({ ...n })),
      links: links.map(l => ({ ...l }))
    };
    graphInstance.graphData(graphData);

    scene.add(graphInstance);

    // Handle resize
    resizeObserver = new ResizeObserver((entries) => {
      const entry = entries[0];
      const { width: newWidth, height: newHeight } = entry.contentRect;
      camera.aspect = newWidth / newHeight;
      camera.updateProjectionMatrix();
      renderer.setSize(newWidth, newHeight);
    });
    resizeObserver.observe(containerRef);

    // Animation loop
    const animate = () => {
      animationId = requestAnimationFrame(animate);
      graphInstance.tickFrame();
      controls.update();
      renderer.render(scene, camera);
    };
    animate();

    return () => {
      renderer.domElement.removeEventListener('click', handleClick);
      renderer.domElement.removeEventListener('mousemove', handleMouseMove);
      cancelAnimationFrame(animationId);
      resizeObserver.disconnect();
      controls.dispose();
      renderer.dispose();
      if (containerRef && renderer.domElement.parentNode === containerRef) {
        containerRef.removeChild(renderer.domElement);
      }
    };
  });
</script>

<div class="graph-3d-container" bind:this={containerRef}>
  {#if nodes.length === 0}
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
