<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { browser } from '$app/environment';
  import { initScene } from '$lib/three/core/sceneSetup';
  import { createSimulation } from '$lib/three/simulation/forceSimulation';
  import { ObjectManager } from '$lib/three/rendering/objectManager';
  import { autoZoomToFit } from '$lib/three/camera/cameraUtils';
  import type { GraphData } from '$lib/api/graph';
  import * as THREE from 'three';

  const { 
    data, 
    onNodeClick
  }: { 
    data: GraphData; 
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
  } = $props();

  let container: HTMLDivElement;
  let scene: THREE.Scene;
  let camera: THREE.PerspectiveCamera;
  let renderer: THREE.WebGLRenderer;
  let labelRenderer: any;
  let controls: any;
  let simulation: any;
  let objectManager: ObjectManager;
  let animationFrame: number;

  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let isAutoRotating = $state(true);
  let isInitialized = $state(false);
  let lastProcessedKey = $state(0);
  
  // Создаем ключ для отслеживания изменений данных
  let dataUpdateKey = $state(0);
  
  // Отслеживаем изменения в данных и обновляем ключ
  $effect(() => {
    const nodesLen = data.nodes.length;
    const linksLen = data.links.length;
    // Инкрементируем ключ только если данные реально изменились
    const newKey = nodesLen + linksLen * 1000;
    if (newKey !== dataUpdateKey) {
      dataUpdateKey = newKey;
      console.log('[Graph3D] Data changed:', { nodesLen, linksLen, key: newKey });
    }
  });

  // Reactively update graph when data changes
  $effect(() => {
    const _key = dataUpdateKey;
    
    // Пропускаем если данные уже обработаны или инициализация не завершена
    if (!isInitialized || _key === lastProcessedKey) {
      return;
    }
    
    // Ограничение для очень больших графов
    if (data.nodes.length > 500) {
      console.warn('[Graph3D] Large graph detected:', data.nodes.length, 'nodes. Limiting to 500 for performance.');
    }
    
    console.log('[Graph3D] Creating simulation:', data.nodes.length, 'nodes');
    lastProcessedKey = _key;
    createGraphSimulation();
  });

  function onResize() {
    if (!container || !renderer || !camera || !labelRenderer) return;
    const width = container.clientWidth;
    const height = container.clientHeight;
    renderer.setSize(width, height);
    labelRenderer.setSize(width, height);
    camera.aspect = width / height;
    camera.updateProjectionMatrix();
  }

  function handleClick(event: MouseEvent) {
    if (!camera || !scene || !objectManager) return;
    
    const raycaster = new THREE.Raycaster();
    const mouse = new THREE.Vector2();
    const rect = container.getBoundingClientRect();
    mouse.x = ((event.clientX - rect.left) / rect.width) * 2 - 1;
    mouse.y = -((event.clientY - rect.top) / rect.height) * 2 + 1;
    
    raycaster.setFromCamera(mouse, camera);
    const intersects = raycaster.intersectObjects(scene.children, true);
    const nodeIntersect = intersects.find((i: any) => i.object.userData?.type === 'node');
    
    if (nodeIntersect) {
      const nodeData = nodeIntersect.object.userData?.nodeData;
      if (nodeData && onNodeClick) {
        onNodeClick(nodeData);
      }
    }
  }

  onMount(async () => {
    if (!browser || !container) return;

    try {
      console.log('[Graph3D] Initializing scene...');
      const setup = initScene(container);
      scene = setup.scene;
      camera = setup.camera;
      renderer = setup.renderer;
      labelRenderer = setup.labelRenderer;
      controls = setup.controls;

      objectManager = new ObjectManager(scene);
      
      // Animation loop
      function animate() {
        animationFrame = requestAnimationFrame(animate);
        controls.update();
        renderer.render(scene, camera);
        labelRenderer.render(scene, camera);
      }
      animate();

      // Resize handler
      window.addEventListener('resize', onResize);
      
      // Click handler for nodes
      container.addEventListener('click', handleClick);
      
      isInitialized = true;
      console.log('[Graph3D] Scene initialized, isInitialized = true');
      
      // Если данные уже есть - создаем симуляцию сразу
      if (data.nodes.length > 0) {
        console.log('[Graph3D] Initial data available, creating simulation:', data.nodes.length, 'nodes');
        createGraphSimulation();
      }
      
    } catch (e) {
      console.error('Failed to initialize 3D graph:', e);
      error = 'Failed to initialize 3D visualization';
      isLoading = false;
    }
  });
  
  function createGraphSimulation() {
    console.log('[Graph3D] createGraphSimulation called:', { 
      hasObjectManager: !!objectManager, 
      hasScene: !!scene,
      nodeCount: data.nodes.length,
      linkCount: data.links.length
    });
    
    if (!objectManager || !scene) {
      console.error('[Graph3D] Cannot create simulation - missing objectManager or scene');
      return;
    }
    
    console.log('[Graph3D] Clearing existing objects...');
    // Clear existing objects
    objectManager.clear();
    if (simulation) {
      simulation.stop();
    }
    
    console.log('[Graph3D] Calling createSimulation...');
    isLoading = true;
    simulation = createSimulation(data, objectManager);
    
    console.log('[Graph3D] Simulation created, setting up event handlers...');
    simulation.on('end', () => {
      console.log('[Graph3D] Simulation ended, nodes:', simulation?.nodes()?.length || 0);
      if (simulation && camera && controls) {
        console.log('[Graph3D] Calling autoZoomToFit...');
        autoZoomToFit(simulation.nodes(), camera, controls);
      }
      isLoading = false;
    });
    
    simulation.on('tick', () => {
      // Обновляем позиции объектов при каждом тике симуляции
      if (objectManager) {
        objectManager.updatePositions();
      }
    });
  }

  onDestroy(() => {
    if (animationFrame) cancelAnimationFrame(animationFrame);
    if (simulation) simulation.stop();
    if (renderer) renderer.dispose();
    window.removeEventListener('resize', onResize);
    if (container) {
      container.removeEventListener('click', handleClick);
    }
  });

  // Public method to reset camera
  export function resetCamera() {
    if (simulation && camera && controls) {
      autoZoomToFit(simulation.nodes(), camera, controls);
    }
  }

  // Toggle auto-rotation
  export function toggleAutoRotate() {
    isAutoRotating = !isAutoRotating;
    if (controls) {
      controls.autoRotate = isAutoRotating;
    }
  }
</script>

<div bind:this={container} class="graph-3d-container">
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

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255,255,255,0.3);
    border-top-color: #88aaff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
