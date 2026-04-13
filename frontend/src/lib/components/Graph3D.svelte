<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import type { GraphData } from '$lib/api/graph';
  import { goto } from '$app/navigation';

  // Types for dynamic imports
  type THREE_Module = typeof import('three');
  
  // Constructor types
  type OrbitControlsCtor = new (camera: any, domElement: HTMLElement) => any;
  type CSS2DRendererCtor = new () => any;
  type CSS2DObjectCtor = new (element: HTMLElement) => any;

  // Extended window interface for dynamic imports
  interface ExtendedWindow extends Window {
    OrbitControls?: OrbitControlsCtor;
    CSS2DRenderer?: CSS2DRendererCtor;
    CSS2DObject?: CSS2DObjectCtor;
    d3Force3d?: typeof import('d3-force-3d');
  }

  // Helper for window access
  const win = () => window as ExtendedWindow;

  const { 
    data, 
    centerNodeId,
    onNodeClick
  }: { 
    data: GraphData; 
    centerNodeId: string;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
  } = $props();

  let container: HTMLDivElement;
  let THREE: THREE_Module | null = null;
  let scene: any;
  let camera: any;
  let renderer: any;
  let labelRenderer: any;
  let controls: any;
  let animationFrame: number;
  let simulation: any;
  let raycaster: any;
  let mouse: any;

  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let isAutoRotating = $state(true);
  let tooltipVisible = $state(false);
  let tooltipTitle = $state('');
  let tooltipType = $state('');
  let tooltipPosition = $state({ x: 0, y: 0 });

  const nodeObjects = new Map<string, any>();
  const linkObjects = new Map<string, any>();
  const labelObjects = new Map<string, any>();

  const typeColors: Record<string, { color: string; label: string }> = {
    star: { color: '#ffaa44', label: 'Star' },
    planet: { color: '#44aaff', label: 'Planet' },
    comet: { color: '#ff44aa', label: 'Comet' },
    galaxy: { color: '#aa44ff', label: 'Galaxy' },
    default: { color: '#88aaff', label: 'Default' }
  };

  const initialCameraPosition = { x: 20, y: 15, z: 30 };
  const initialCameraTarget = { x: 0, y: 0, z: 0 };

  onMount(() => {
    if (!browser) return;

    let isActive = true;

    const init = async () => {
      try {
        // Dynamic imports for SSR safety
        THREE = await import('three');
        const { OrbitControls } = await import('three/examples/jsm/controls/OrbitControls.js');
        const { CSS2DRenderer, CSS2DObject } = await import('three/examples/jsm/renderers/CSS2DRenderer.js');
        const d3Force3d = await import('d3-force-3d');

        if (!isActive) return;

        // Store constructors
        win().OrbitControls = OrbitControls;
        win().CSS2DRenderer = CSS2DRenderer;
        win().CSS2DObject = CSS2DObject;
        win().d3Force3d = d3Force3d;

        // Check if graph is empty
        if (!data.nodes || data.nodes.length === 0) {
          isLoading = false;
          error = 'No connected notes found';
          return;
        }

        initThree();
        initSimulation(data);
        startAnimationLoop();
        window.addEventListener('resize', onResize);
        container.addEventListener('mousemove', onMouseMove);
        isLoading = false;
      } catch (e) {
        error = 'Error initializing 3D scene';
        console.error(e);
      }
    };

    init();

    return () => {
      isActive = false;
      cleanup();
    };
  });

  function initThree() {
    if (!container || !THREE) return;

    scene = new THREE.Scene();
    scene.background = new THREE.Color(0x050510);

    const aspect = container.clientWidth / container.clientHeight;
    camera = new THREE.PerspectiveCamera(45, aspect, 0.1, 1000);
    camera.position.set(initialCameraPosition.x, initialCameraPosition.y, initialCameraPosition.z);
    camera.lookAt(initialCameraTarget.x, initialCameraTarget.y, initialCameraTarget.z);

    renderer = new THREE.WebGLRenderer({ antialias: true, alpha: false });
    renderer.setSize(container.clientWidth, container.clientHeight);
    renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
    renderer.shadowMap.enabled = true;
    renderer.shadowMap.type = THREE.PCFSoftShadowMap;
    container.appendChild(renderer.domElement);

    const CSS2DRendererCtor = win().CSS2DRenderer;
    if (!CSS2DRendererCtor) throw new Error('CSS2DRenderer not loaded');
    labelRenderer = new CSS2DRendererCtor();
    labelRenderer.setSize(container.clientWidth, container.clientHeight);
    labelRenderer.domElement.style.position = 'absolute';
    labelRenderer.domElement.style.top = '0';
    labelRenderer.domElement.style.left = '0';
    labelRenderer.domElement.style.pointerEvents = 'none';
    container.appendChild(labelRenderer.domElement);

    const OrbitControlsCtor = win().OrbitControls;
    if (!OrbitControlsCtor) throw new Error('OrbitControls not loaded');
    controls = new OrbitControlsCtor(camera, renderer.domElement);
    controls.enableDamping = true;
    controls.dampingFactor = 0.05;
    controls.autoRotate = isAutoRotating;
    controls.autoRotateSpeed = 0.8;
    controls.enableZoom = true;
    controls.enablePan = true;
    controls.maxPolarAngle = Math.PI / 1.8;

    // Lighting
    const ambientLight = new THREE.AmbientLight(0x404060, 0.6);
    scene.add(ambientLight);

    const dirLight = new THREE.DirectionalLight(0xfff5e6, 1.2);
    dirLight.position.set(10, 30, 10);
    dirLight.castShadow = true;
    scene.add(dirLight);

    const fillLight1 = new THREE.PointLight(0x4466ff, 0.8);
    fillLight1.position.set(15, 5, 20);
    scene.add(fillLight1);

    const fillLight2 = new THREE.PointLight(0xff66aa, 0.5);
    fillLight2.position.set(-15, 10, -20);
    scene.add(fillLight2);

    addStarfield();
    scene.fog = new THREE.FogExp2(0x050510, 0.002);

    // Raycaster for hover
    raycaster = new THREE.Raycaster();
    mouse = new THREE.Vector2();

    // Click handler
    renderer.domElement.addEventListener('click', onMouseClick);
  }

  function addStarfield() {
    if (!THREE) return;
    const geometry = new THREE.BufferGeometry();
    const count = 4000;
    const positions = new Float32Array(count * 3);
    for (let i = 0; i < count * 3; i += 3) {
      const r = 80 + Math.random() * 70;
      const theta = Math.random() * Math.PI * 2;
      const phi = Math.acos(2 * Math.random() - 1);
      positions[i] = Math.sin(phi) * Math.cos(theta) * r;
      positions[i+1] = Math.sin(phi) * Math.sin(theta) * r;
      positions[i+2] = Math.cos(phi) * r;
    }
    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    const material = new THREE.PointsMaterial({
      color: 0xffffff,
      size: 0.15,
      transparent: true,
      blending: THREE.AdditiveBlending
    });
    const stars = new THREE.Points(geometry, material);
    scene.add(stars);
  }

  function onMouseMove(event: MouseEvent) {
    if (!container || !camera) return;

    const rect = container.getBoundingClientRect();
    mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;

    raycaster.setFromCamera(mouse, camera);
    const intersects = raycaster.intersectObjects(Array.from(nodeObjects.values()));

    if (intersects.length > 0) {
      const intersectedNode = intersects[0].object;
      const nodeId = intersectedNode.userData.id;
      const node = data.nodes.find(n => n.id === nodeId);

      if (node) {
        tooltipTitle = node.title;
        tooltipType = node.type || 'default';
        tooltipPosition = { x: event.clientX - rect.left + 10, y: event.clientY - rect.top - 30 };
        tooltipVisible = true;
        container.style.cursor = 'pointer';
      }
    } else {
      tooltipVisible = false;
      container.style.cursor = 'default';
    }
  }

  function onMouseClick(_event: MouseEvent) {
    if (!camera || !scene) return;

    raycaster.setFromCamera(mouse, camera);
    const intersects = raycaster.intersectObjects(Array.from(nodeObjects.values()));

    if (intersects.length > 0) {
      const clickedNode = intersects[0].object;
      const nodeId = clickedNode.userData.id;
      const node = data.nodes.find(n => n.id === nodeId);
      if (nodeId && node) {
        if (onNodeClick) {
          onNodeClick({ id: nodeId, title: node.title, type: node.type });
        } else {
          goto(`/notes/${nodeId}`);
        }
      }
    }
  }

  function initSimulation(graphData: GraphData) {
    if (!win().d3Force3d || !THREE) return;

    const nodes = graphData.nodes.map(n => ({
      ...n,
      x: (Math.random() - 0.5) * 30,
      y: (Math.random() - 0.5) * 30,
      z: (Math.random() - 0.5) * 30
    }));

    const links = graphData.links.map(l => ({
      ...l,
      source: l.source,
      target: l.target,
      value: l.weight || 1
    }));

    createVisualObjects(nodes, links);

    const d3Force = win().d3Force3d;
    if (!d3Force) throw new Error('d3-force-3d not loaded');
    const { forceSimulation, forceLink, forceManyBody, forceCenter } = d3Force;

    simulation = forceSimulation(nodes)
      .force('link', forceLink(links)
        .id((d: any) => d.id)
        .distance((l: any) => 20 / (l.value * 0.8))
        .strength(0.5)
      )
      .force('charge', forceManyBody().strength(-200).distanceMax(50))
      .force('center', forceCenter(0, 0, 0))
      .on('tick', () => {
        updatePositions(nodes);
        updateLinks(links);
      });

    simulation.alpha(1).restart();
  }

  function getColorByType(type?: string): number {
    const colors: Record<string, number> = {
      star: 0xffaa44,
      planet: 0x44aaff,
      comet: 0xff44aa,
      galaxy: 0xaa44ff,
      default: 0x88aaff
    };
    return colors[type || 'default'] || colors.default;
  }

  function createVisualObjects(nodes: any[], links: any[]) {
    if (!THREE || !win().CSS2DObject) return;

    nodeObjects.forEach(obj => scene.remove(obj));
    linkObjects.forEach(obj => scene.remove(obj));
    labelObjects.forEach(obj => scene.remove(obj));
    nodeObjects.clear();
    linkObjects.clear();
    labelObjects.clear();

    nodes.forEach(node => {
      const geometry = new THREE!.SphereGeometry(node.size || 2, 32, 32);
      const color = getColorByType(node.type);
      const material = new THREE!.MeshPhysicalMaterial({
        color: color,
        emissive: color,
        emissiveIntensity: 0.3,
        roughness: 0.4,
        metalness: 0.6,
        clearcoat: 0.8
      });
      const sphere = new THREE!.Mesh(geometry, material);
      sphere.position.set(node.x || 0, node.y || 0, node.z || 0);
      sphere.castShadow = true;
      sphere.receiveShadow = true;
      sphere.userData = { id: node.id };
      scene.add(sphere);
      nodeObjects.set(node.id, sphere);

      const div = document.createElement('div');
      div.className = 'node-label';
      div.textContent = node.title;
      div.style.color = 'white';
      div.style.fontSize = '12px';
      div.style.fontWeight = 'bold';
      div.style.textShadow = '0 0 4px rgba(0,0,0,0.8)';
      div.style.pointerEvents = 'none';
      div.style.whiteSpace = 'nowrap';
      const CSS2DObjectCtor = win().CSS2DObject;
      if (!CSS2DObjectCtor) throw new Error('CSS2DObject not loaded');
      const label = new CSS2DObjectCtor(div);
      label.position.set(0, (node.size || 2) + 1, 0);
      sphere.add(label);
      labelObjects.set(node.id, label);
    });

    links.forEach(link => {
      const geometry = new THREE!.BufferGeometry();
      const material = new THREE!.LineBasicMaterial({
        color: 0x6688cc,
        transparent: true,
        opacity: 0.4,
        blending: THREE!.AdditiveBlending
      });
      const line = new THREE!.Line(geometry, material);
      scene.add(line);
      linkObjects.set(`${link.source}-${link.target}`, line);
    });
  }

  function updatePositions(nodes: any[]) {
    nodes.forEach(node => {
      const obj = nodeObjects.get(node.id);
      if (obj) {
        obj.position.set(node.x, node.y, node.z);
      }
    });
  }

  function updateLinks(links: any[]) {
    links.forEach(link => {
      // After d3-force simulation, source/target may be objects instead of IDs
      const sourceId = typeof link.source === 'object' ? link.source.id : link.source;
      const targetId = typeof link.target === 'object' ? link.target.id : link.target;

      const key = `${sourceId}-${targetId}`;
      const line = linkObjects.get(key);
      if (!line) return;

      const sourceNode = nodeObjects.get(sourceId);
      const targetNode = nodeObjects.get(targetId);
      if (!sourceNode || !targetNode) return;

      const points = [sourceNode.position, targetNode.position];
      line.geometry.dispose();
      line.geometry = new THREE!.BufferGeometry().setFromPoints(points);
    });
  }

  function startAnimationLoop() {
    function animate() {
      animationFrame = requestAnimationFrame(animate);
      controls.update();
      renderer.render(scene, camera);
      labelRenderer.render(scene, camera);
    }
    animate();
  }

  function onResize() {
    if (!container) return;
    const width = container.clientWidth;
    const height = container.clientHeight;
    renderer.setSize(width, height);
    labelRenderer.setSize(width, height);
    camera.aspect = width / height;
    camera.updateProjectionMatrix();
  }

  function resetCamera() {
    if (!camera || !controls) return;
    camera.position.set(initialCameraPosition.x, initialCameraPosition.y, initialCameraPosition.z);
    camera.lookAt(initialCameraTarget.x, initialCameraTarget.y, initialCameraTarget.z);
    controls.target.set(initialCameraTarget.x, initialCameraTarget.y, initialCameraTarget.z);
    controls.update();
  }

  function toggleAutoRotate() {
    isAutoRotating = !isAutoRotating;
    if (controls) {
      controls.autoRotate = isAutoRotating;
    }
  }

  function cleanup() {
    if (animationFrame) cancelAnimationFrame(animationFrame);
    if (simulation) simulation.stop();
    window.removeEventListener('resize', onResize);
    if (container) {
      container.removeEventListener('mousemove', onMouseMove);
    }
    if (renderer) {
      renderer.dispose();
      renderer.domElement.remove();
    }
    if (labelRenderer) labelRenderer.domElement.remove();
    nodeObjects.forEach(obj => {
      obj.traverse((child: any) => {
        if (child.geometry) child.geometry.dispose();
        if (child.material) {
          if (Array.isArray(child.material))
            child.material.forEach((m: { dispose(): void }) => m.dispose());
          else
            child.material.dispose();
        }
      });
    });
    scene?.clear();
  }
</script>

<div class="graph-3d-container" bind:this={container}>
  {#if isLoading}
    <div class="loading-overlay">
      <div class="spinner"></div>
      <p>Loading 3D constellation...</p>
    </div>
  {/if}

  {#if error}
    <div class="error-overlay">
      <div class="error-content">
        <span class="error-icon">🌌</span>
        <h2>{error}</h2>
        <p>This note has no connections yet.</p>
        <a href="/notes/{centerNodeId}" class="back-button">Back to Note</a>
      </div>
    </div>
  {/if}

  <!-- Controls -->
  <div class="controls-panel">
    <button class="control-btn" onclick={resetCamera} title="Reset Camera">
      <span>🏠</span>
      <span class="btn-label">Reset</span>
    </button>
    <button class="control-btn" onclick={toggleAutoRotate} title={isAutoRotating ? 'Pause Rotation' : 'Resume Rotation'}>
      <span>{isAutoRotating ? '⏸️' : '▶️'}</span>
      <span class="btn-label">{isAutoRotating ? 'Pause' : 'Play'}</span>
    </button>
  </div>

  <!-- Legend -->
  <div class="legend-panel">
    <h3 class="legend-title">Types</h3>
    {#each Object.entries(typeColors) as [type, { color, label }]}
      {#if type !== 'default'}
        <div class="legend-item">
          <span class="legend-color" style="background-color: {color}"></span>
          <span class="legend-label">{label}</span>
        </div>
      {/if}
    {/each}
  </div>

  <!-- Tooltip -->
  {#if tooltipVisible}
    <div class="tooltip" style="left: {tooltipPosition.x}px; top: {tooltipPosition.y}px;">
      <div class="tooltip-title">{tooltipTitle}</div>
      <div class="tooltip-type" style="color: {typeColors[tooltipType]?.color || typeColors.default.color}">
        {typeColors[tooltipType]?.label || typeColors.default.label}
      </div>
    </div>
  {/if}
</div>

<style>
  .graph-3d-container {
    width: 100%;
    height: 100vh;
    position: relative;
    overflow: hidden;
    background: #050510;
  }

  .loading-overlay, .error-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: rgba(5,5,16,0.95);
    color: white;
    z-index: 10;
  }

  .error-content {
    text-align: center;
    padding: 2rem;
  }

  .error-icon {
    font-size: 48px;
    margin-bottom: 1rem;
    display: block;
  }

  .back-button {
    display: inline-block;
    margin-top: 1rem;
    padding: 0.75rem 1.5rem;
    background: #3b82f6;
    color: white;
    text-decoration: none;
    border-radius: 8px;
    font-weight: 500;
  }

  .back-button:hover {
    background: #2563eb;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255,255,255,0.3);
    border-top-color: #88aaff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .controls-panel {
    position: absolute;
    top: 20px;
    left: 20px;
    display: flex;
    gap: 10px;
    z-index: 5;
  }

  .control-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 10px 16px;
    background: rgba(255,255,255,0.1);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255,255,255,0.2);
    border-radius: 8px;
    color: white;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .control-btn:hover {
    background: rgba(255,255,255,0.2);
    transform: translateY(-2px);
  }

  .control-btn span:first-child {
    font-size: 18px;
  }

  .legend-panel {
    position: absolute;
    bottom: 20px;
    right: 20px;
    background: rgba(255,255,255,0.1);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255,255,255,0.2);
    border-radius: 12px;
    padding: 16px;
    z-index: 5;
    min-width: 140px;
  }

  .legend-title {
    margin: 0 0 12px 0;
    color: white;
    font-size: 14px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  .legend-item {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 8px;
    color: white;
    font-size: 13px;
  }

  .legend-item:last-child {
    margin-bottom: 0;
  }

  .legend-color {
    width: 14px;
    height: 14px;
    border-radius: 50%;
    box-shadow: 0 0 8px currentColor;
  }

  .tooltip {
    position: absolute;
    background: rgba(0,0,0,0.85);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255,255,255,0.2);
    border-radius: 8px;
    padding: 12px 16px;
    z-index: 20;
    pointer-events: none;
    min-width: 180px;
    box-shadow: 0 4px 20px rgba(0,0,0,0.5);
  }

  .tooltip-title {
    color: white;
    font-weight: 600;
    font-size: 14px;
    margin-bottom: 4px;
    max-width: 200px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .tooltip-type {
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    font-weight: 500;
  }

  /* Mobile adjustments */
  @media (max-width: 768px) {
    .controls-panel {
      top: 10px;
      left: 10px;
    }

    .control-btn .btn-label {
      display: none;
    }

    .control-btn {
      padding: 8px 12px;
    }

    .legend-panel {
      bottom: 10px;
      right: 10px;
      padding: 12px;
      min-width: auto;
    }

    .legend-title {
      font-size: 12px;
    }

    .legend-item {
      font-size: 11px;
    }

    .legend-label {
      display: none;
    }
  }

  /* Node labels */
  :global(.node-label) {
    font-family: system-ui, -apple-system, sans-serif;
  }
</style>