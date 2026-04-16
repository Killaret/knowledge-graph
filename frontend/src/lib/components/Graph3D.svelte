<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { browser } from '$app/environment';
  import { initScene, setFogDensity } from '$lib/three/core/sceneSetup';
  import { createSimulation, addNodesToSimulation } from '$lib/three/simulation/forceSimulation';
  import { ObjectManager } from '$lib/three/rendering/objectManager';
  import { autoZoomToFit, centerCameraOnNode } from '$lib/three/camera/cameraUtils';
  import type { GraphData } from '$lib/api/graph';
  import * as THREE from 'three';

  const { 
    data, 
    centerNodeId,
    onNodeClick
  }: { 
    data: GraphData; 
    centerNodeId?: string | null;
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

  let isLoading = $state(true); // Show loading while graph initializes
  let error = $state<string | null>(null);
  let isAutoRotating = $state(true);
  let isInitialized = $state(false);
  let lastProcessedKey = $state(0);
  
  // Progressive loading state
  let currentData = $state<GraphData>({ nodes: [], links: [] });
  let isFullyLoaded = $state(false);
  let fogAnimationFrame: number | null = null;
  
  // Создаем ключ для отслеживания изменений данных (включает timestamp для гарантированного обновления)
  let dataUpdateKey = $state(0);
  let lastDataTimestamp = $state(0);

  // Отслеживаем изменения в данных и обновляем ключ
  $effect(() => {
    const nodesLen = data.nodes.length;
    const linksLen = data.links.length;
    // Проверяем, действительно ли данные изменились (сравниваем ссылку на массив)
    const dataChanged = data.nodes !== lastProcessedData.nodes || data.links !== lastProcessedData.links;
    if (dataChanged) {
      lastDataTimestamp = Date.now();
      lastProcessedData = { nodes: data.nodes, links: data.links };
    }
    // Ключ включает timestamp для гарантированного обновления при переключении режима
    const newKey = nodesLen + linksLen * 1000 + lastDataTimestamp;
    if (newKey !== dataUpdateKey) {
      dataUpdateKey = newKey;
      console.log('[Graph3D] Data changed:', { nodesLen, linksLen, key: newKey, timestamp: lastDataTimestamp });
    }
  });

  // Храним ссылку на последние обработанные данные
  let lastProcessedData = $state<{ nodes: any[], links: any[] }>({ nodes: [], links: [] });

  // Reactively update graph when data changes
  $effect(() => {
    const _key = dataUpdateKey;
    
    // Wait for initialization to complete
    if (!isInitialized) {
      return;
    }
    
    // Skip only if exact same data state was already processed
    if (_key === lastProcessedKey && lastProcessedKey !== 0) {
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
      
      // Export objects for debugging
      if (browser) {
        (window as any).scene = scene;
        (window as any).camera = camera;
        (window as any).controls = controls;
        (window as any).simulation = simulation;
      }
      
      // Animation loop
      let frameCount = 0;
      function animate() {
        animationFrame = requestAnimationFrame(animate);
        controls.update();
        renderer.render(scene, camera);
        labelRenderer.render(scene, camera);
        
        // Log debug info once after 30 frames (~0.5 sec at 60fps)
        if (++frameCount === 30) {
          console.log('[Graph3D] Render frame 30 - Camera:', 
            `(${camera.position.x.toFixed(1)}, ${camera.position.y.toFixed(1)}, ${camera.position.z.toFixed(1)})`,
            'Scene children:', scene.children.length,
            'Fog density:', (scene.fog as any)?.density ?? 'no fog'
          );
        }
      }
      animate();
      console.log('[Graph3D] Animation loop started');

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
        // Mark this state as processed so effect doesn't re-trigger for same data
        lastProcessedKey = dataUpdateKey;
      } else {
        // No nodes - show empty state immediately
        console.log('[Graph3D] No nodes in data, showing empty state');
        isLoading = false;
        // Don't set lastProcessedKey here - we want the effect to trigger when data arrives
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
    
    // Filter links to only include those where both source and target nodes exist
    // (for local graph, API may return links to nodes outside the graph)
    const nodeIds = new Set(data.nodes.map(n => n.id));
    const validLinks = data.links.filter(l => nodeIds.has(l.source) && nodeIds.has(l.target));
    if (validLinks.length !== data.links.length) {
      console.warn(`[Graph3D] Filtered out ${data.links.length - validLinks.length} orphan links`);
    }
    const filteredData = {
      nodes: data.nodes,
      links: validLinks
    };
    
    console.log('[Graph3D] Calling createSimulation...');
    // Track current data for progressive loading
    currentData = filteredData;
    // Set dense fog for "veil of creation" effect - will clear when nodes appear
    if (scene) {
      setFogDensity(scene, 0.06); // Dense but visible veil
    }
    isLoading = true;
    simulation = createSimulation(filteredData, objectManager);
    
    console.log('[Graph3D] Simulation created, setting up event handlers...');
    
    // If no links or single node, simulation won't emit 'end' (already at equilibrium)
    // Stop immediately and show the graph
    if (filteredData.links.length === 0 || filteredData.nodes.length <= 1) {
      console.log('[Graph3D] No links or single node, stopping simulation immediately');
      simulation.stop();
      isLoading = false;
      // Animate fog dissipation for this edge case too
      console.log('[Graph3D] Starting fog animation for single node/no links');
      animateFog(0.02, 0.005, 1500);
      isFullyLoaded = true;
      if (camera && controls) {
        // Use centerCameraOnNode for local graph, autoZoomToFit for full graph
        if (centerNodeId) {
          centerCameraOnNode(centerNodeId, simulation.nodes(), camera, controls, true);
        } else {
          autoZoomToFit(simulation.nodes(), camera, controls, true);
        }
      }
      return;
    }
    
    // Резервный таймаут - принудительно выключаем загрузку через 5 секунд
    const loadingTimeout = setTimeout(() => {
      if (isLoading) {
        console.log('[Graph3D] Loading timeout reached, forcing isLoading = false');
        isLoading = false;
        // Also animate fog on timeout
        animateFog(0.02, 0.005, 2000);
        isFullyLoaded = true;
        if (simulation && camera && controls) {
          // Use centerCameraOnNode for local graph, autoZoomToFit for full graph
          if (centerNodeId) {
            centerCameraOnNode(centerNodeId, simulation.nodes(), camera, controls, true);
          } else {
            autoZoomToFit(simulation.nodes(), camera, controls, true);
          }
        }
      }
    }, 5000);
    
    // Zoom timeout for delayed auto-zoom
    let zoomTimeout: any;
    
    simulation.on('end', () => {
      console.log('[Graph3D] Simulation ended, nodes:', simulation?.nodes()?.length || 0);
      clearTimeout(loadingTimeout);
      isLoading = false;

      // Animate fog dissipation for initial load (same as addData)
      console.log('[Graph3D] Starting fog animation for initial load (0.02 -> 0.005)');
      animateFog(0.02, 0.005, 2500);
      isFullyLoaded = true;

      clearTimeout(zoomTimeout);
      zoomTimeout = setTimeout(() => {
        if (simulation && camera && controls) {
          // Use centerCameraOnNode for local graph, autoZoomToFit for full graph
          if (centerNodeId) {
            console.log('[Graph3D] Calling centerCameraOnNode for local graph...');
            centerCameraOnNode(centerNodeId, simulation.nodes(), camera, controls, true);
          } else {
            console.log('[Graph3D] Calling autoZoomToFit for full graph...');
            autoZoomToFit(simulation.nodes(), camera, controls, true);
          }
        }
      }, 300);
    });
    
    // Track if first node is visible to clear fog
    let firstNodeVisible = false;
    
    simulation.on('tick', () => {
      // Обновляем позиции объектов при каждом тике симуляции
      if (objectManager && simulation) {
        objectManager.updatePositions(simulation.nodes());
      }
      
      // Check if first node is visible and clear fog
      if (!firstNodeVisible && scene && camera) {
        const nodes = simulation.nodes();
        if (nodes.length > 0 && nodes[0].x !== undefined) {
          // First node has position - clear the fog veil
          firstNodeVisible = true;
          console.log('[Graph3D] First node visible, clearing fog veil');
          animateFog(0.06, 0.005, 1500); // Clear from dense veil to light fog
          // Also hide loading overlay when first node appears
          if (isLoading) {
            isLoading = false;
          }
        }
      }
    });
  }

  // Animate fog density from start to end over duration
  function animateFog(startDensity: number, endDensity: number, duration: number = 2000) {
    if (!scene) return;
    
    const startTime = performance.now();
    
    function updateFog() {
      const elapsed = performance.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);
      
      // Ease out cubic
      const ease = 1 - Math.pow(1 - progress, 3);
      const currentDensity = startDensity + (endDensity - startDensity) * ease;
      
      setFogDensity(scene, currentDensity);
      
      if (progress < 1) {
        fogAnimationFrame = requestAnimationFrame(updateFog);
      } else {
        fogAnimationFrame = null;
      }
    }
    
    // Cancel any existing fog animation
    if (fogAnimationFrame) {
      cancelAnimationFrame(fogAnimationFrame);
    }
    
    updateFog();
  }

  // Public method to add more data (for progressive loading)
  export function addData(newData: GraphData) {
    if (!simulation || !objectManager || !scene) {
      console.warn('[Graph3D] addData called but simulation not ready');
      return;
    }
    
    console.log('[Graph3D] addData called:', { 
      currentNodes: currentData.nodes.length, 
      newNodes: newData.nodes.length,
      currentLinks: currentData.links.length,
      newLinks: newData.links.length
    });
    
    // Filter new data to only include valid links
    const newNodeIds = new Set(newData.nodes.map(n => n.id));
    const validNewLinks = newData.links.filter(l => 
      newNodeIds.has(l.source) && newNodeIds.has(l.target)
    );
    
    const filteredNewData = {
      nodes: newData.nodes,
      links: validNewLinks
    };
    
    // Add new data to simulation
    addNodesToSimulation(simulation, filteredNewData, currentData, objectManager);
    
    // Update current data tracking
    const existingNodeIds = new Set(currentData.nodes.map(n => n.id));
    const existingLinkIds = new Set(currentData.links.map(l => `${l.source}-${l.target}`));
    
    const mergedNodes = [
      ...currentData.nodes,
      ...newData.nodes.filter(n => !existingNodeIds.has(n.id))
    ];
    const mergedLinks = [
      ...currentData.links,
      ...validNewLinks.filter(l => !existingLinkIds.has(`${l.source}-${l.target}`))
    ];
    
    currentData = { nodes: mergedNodes, links: mergedLinks };
    
    // Animate fog dissipation (dense to clear)
    console.log('[Graph3D] Starting fog animation (0.08 -> 0.005)');
    animateFog(0.08, 0.005, 2500);
    
    // Mark as fully loaded
    isFullyLoaded = true;
    
    // Call camera positioning with animation after a delay to let simulation settle
    setTimeout(() => {
      if (simulation && camera && controls) {
        // Use centerCameraOnNode for local graph, autoZoomToFit for full graph
        if (centerNodeId) {
          console.log('[Graph3D] Centering on central node after addData');
          centerCameraOnNode(centerNodeId, simulation.nodes(), camera, controls, true);
        } else {
          console.log('[Graph3D] Auto zooming to fit full graph with animation');
          autoZoomToFit(simulation.nodes(), camera, controls, true);
        }
      }
    }, 500);
  }

  onDestroy(() => {
    if (animationFrame) cancelAnimationFrame(animationFrame);
    if (fogAnimationFrame) cancelAnimationFrame(fogAnimationFrame);
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
      // Use centerCameraOnNode for local graph, autoZoomToFit for full graph
      if (centerNodeId) {
        centerCameraOnNode(centerNodeId, simulation.nodes(), camera, controls, isFullyLoaded);
      } else {
        autoZoomToFit(simulation.nodes(), camera, controls, isFullyLoaded);
      }
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
      <p class="loading-text">The universe isn't still created, but soon</p>
      <p class="loading-subtext">First celestial body appearing...</p>
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

  {#if !isLoading && !error && data.nodes.length === 0}
    <div class="no-data-message">
      <div class="empty-content">
        <span class="empty-icon">🔭</span>
        <h2>No Connections Yet</h2>
        <p>This note has no links to other notes.</p>
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

  .loading-overlay, .error-overlay, .no-data-message {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: rgba(5,5,16,0.85);
    color: white;
    z-index: 10;
    backdrop-filter: blur(5px);
  }

  .loading-text {
    font-size: 1.8rem;
    font-weight: 500;
    text-align: center;
    color: rgba(136, 170, 255, 0.9);
    text-shadow: 0 0 30px rgba(136, 170, 255, 0.6);
    background: rgba(5, 5, 16, 0.7);
    padding: 1.5rem 2rem;
    border-radius: 16px;
    backdrop-filter: blur(8px);
    border: 1px solid rgba(136, 170, 255, 0.2);
    animation: pulse-dim 3s ease-in-out infinite;
  }

  @keyframes pulse-dim {
    0%, 100% { 
      opacity: 0.6;
      text-shadow: 0 0 20px rgba(136, 170, 255, 0.4);
    }
    50% { 
      opacity: 1;
      text-shadow: 0 0 40px rgba(136, 170, 255, 0.8);
    }
  }

  .loading-subtext {
    font-size: 1rem;
    margin-top: 1rem;
    color: rgba(100, 116, 139, 0.8);
    font-style: italic;
    background: rgba(5, 5, 16, 0.5);
    padding: 0.5rem 1rem;
    border-radius: 8px;
    backdrop-filter: blur(4px);
  }

  .error-content, .empty-content {
    text-align: center;
    padding: 2rem;
  }

  .error-icon, .empty-icon {
    font-size: 48px;
    margin-bottom: 1rem;
    display: block;
  }
</style>
