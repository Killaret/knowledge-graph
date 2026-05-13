<script lang="ts">
  import { onMount, onDestroy, untrack } from 'svelte';
  import { fade } from 'svelte/transition';
  import { browser } from '$app/environment';
  import { ObjectManager } from '$lib/three/rendering/objectManager';
  import {
    initScene,
    setFogDensity,
    resizeScene,
    disposeScene
  } from './Graph3D';
  import { animateFogDensity } from './Graph3D/fogManager';
  import {
    createSimulation,
    addNodesToSimulation
  } from './Graph3D';
  import {
    autoZoomToFit,
    centerCameraOnNode,
    toggleAutoRotate as toggleCameraAutoRotate,
    type CameraState
  } from './Graph3D';
  import { filterValidLinks } from '$lib/utils/graphUtils';
  import { graphConfig3D } from '$lib/config';
  import type { GraphData, GraphNode, GraphLink } from '$lib/api/graph';
  import * as THREE from 'three';
  import { hasPreloadedData } from '$lib/services/PreloadService';

  const { 
    data, 
    centerNodeId,
    onNodeClick,
    onNodeDoubleClick
  }: { 
    data: GraphData; 
    centerNodeId?: string | null;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
    onNodeDoubleClick?: (node: { id: string; title: string; type?: string }) => void;
  } = $props();

  // DOM и Three.js объекты
  let container: HTMLDivElement;
  let scene: THREE.Scene;
  let camera: THREE.PerspectiveCamera;
  let renderer: THREE.WebGLRenderer;
  let labelRenderer: any;
  let controls: any;
  let simulation: any;
  let objectManager: ObjectManager;
  let animationFrame: number;

  // Реактивные состояния
  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let isAutoRotating = $state(true); // eslint-disable-line @typescript-eslint/no-unused-vars -- Reserved for future auto-rotation feature
  let isInitialized = $state(false);
  let lastProcessedKey = $state(0);
  let currentData = $state<GraphData>({ nodes: [], links: [] });
  let isFullyLoaded = $state(false);
  let fogAnimationFrame: number | null = null;
  let dataUpdateKey = $state(0);
  let lastDataTimestamp = $state(0);

  const lastProcessedData: { nodes: GraphData['nodes']; links: GraphData['links'] } = { nodes: [], links: [] };
  const cameraState: CameraState = { autoRotate: true, isAnimating: false };

  // Отслеживание изменений данных
  $effect(() => {
    const currentNodes = untrack(() => data.nodes);
    const currentLinks = untrack(() => data.links);
    const nodesLen = currentNodes.length;
    const linksLen = currentLinks.length;
    const nodesChanged = currentNodes !== lastProcessedData.nodes;
    const linksChanged = currentLinks !== lastProcessedData.links;
    
    if (nodesChanged || linksChanged) {
      lastDataTimestamp = Date.now();
      lastProcessedData.nodes = currentNodes;
      lastProcessedData.links = currentLinks;
    }
    
    const newKey = nodesLen + linksLen * 1000 + lastDataTimestamp;
    if (newKey !== dataUpdateKey) {
      dataUpdateKey = newKey;
    }
  });

  // Реактивное обновление графа
  $effect(() => {
    const _key = dataUpdateKey;
    if (!isInitialized) return;
    if (_key === lastProcessedKey && lastProcessedKey !== 0) return;
    
    const currentNodes = untrack(() => data.nodes);
    const currentLinks = untrack(() => data.links);
    
    if (currentNodes.length > graphConfig3D.max_nodes) {
      console.warn(`[Graph3D] Large graph: ${currentNodes.length} nodes, limiting to ${graphConfig3D.max_nodes}`);
    }
    
    lastProcessedKey = _key;
    createGraphSimulation(currentNodes, currentLinks);
  });

  function onResize() {
    resizeScene(container, camera, renderer, labelRenderer);
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
      if (nodeData && onNodeClick) onNodeClick(nodeData);
    }
  }

  function handleDoubleClick(event: MouseEvent) {
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
      if (nodeData && onNodeDoubleClick) onNodeDoubleClick(nodeData);
    }
  }

  function createGraphSimulation(nodesData: GraphData['nodes'], linksData: GraphData['links']) {
    if (!objectManager || !scene) return;
    
    objectManager.clear();
    if (simulation) {
      simulation.stop();
      simulation.on('tick', null);
    }
    
    let lastFogUpdate = 0;
    let zoomApplied = false;
    
    const validLinks = filterValidLinks(nodesData, linksData);
    const filteredData = { nodes: nodesData, links: validLinks };
    
    currentData = filteredData;
    
    // Интеграция с PreloadService: если данные предзагружены, начинаем с меньшей плотности тумана
    const hasPreloaded = hasPreloadedData();
    const initialDensity = hasPreloaded ? 0.04 : 0.08; // Меньше тумана для предзагруженных данных
    
    if (scene) {
      setFogDensity(scene, initialDensity);
      if (import.meta.env.DEV) {
        console.log(`[Graph3D] Starting simulation with preloaded data: ${hasPreloaded}, initial fog density: ${initialDensity}`);
      }
    }
    
    isLoading = true;
    simulation = createSimulation(filteredData, objectManager);
    
    if (validLinks.length === 0 || nodesData.length <= 1) {
      simulation.stop();
      isLoading = false;
      if (scene) setFogDensity(scene, 0.005);
      isFullyLoaded = true;
      if (camera && controls) {
        if (centerNodeId) {
          centerCameraOnNode(centerNodeId, simulation.nodes(), camera, controls, true);
        } else {
          autoZoomToFit(simulation.nodes(), camera, controls, true);
        }
      }
      return;
    }
    
    const loadingTimeout = setTimeout(() => {
      if (isLoading) {
        isLoading = false;
        if (scene) setFogDensity(scene, 0.005);
        isFullyLoaded = true;
        if (simulation && camera && controls) {
          if (centerNodeId) {
            centerCameraOnNode(centerNodeId, simulation.nodes(), camera, controls, true);
          } else {
            autoZoomToFit(simulation.nodes(), camera, controls, true);
          }
        }
      }
    }, 5000);
    
    let zoomTimeout: any;
    
    simulation.on('end', () => {
      clearTimeout(loadingTimeout);
      isLoading = false;
      isFullyLoaded = true;
      
      if (scene) {
        const currentDensity = (scene.fog as THREE.FogExp2)?.density ?? 0.08;
        const fogAnimation = animateFogDensity(scene, 0.005, 800);
        // Останавливаем анимацию при уничтожении компонента
        fogAnimation.stop = () => {
          if (fogAnimationFrame) {
            cancelAnimationFrame(fogAnimationFrame);
            fogAnimationFrame = null;
          }
        };
      }
      
      clearTimeout(zoomTimeout);
      zoomTimeout = setTimeout(() => {
        if (simulation && camera && controls && !zoomApplied) {
          zoomApplied = true;
          if (centerNodeId) {
            centerCameraOnNode(centerNodeId, simulation.nodes(), camera, controls, true);
          } else {
            autoZoomToFit(simulation.nodes(), camera, controls, true);
          }
        }
      }, 300);
    });
    
    const totalNodes = filteredData.nodes.length;
    let lastLinkUpdateTime = 0;
    
    simulation.on('tick', () => {
      const now = performance.now();
      if (objectManager && simulation) {
        objectManager.updatePositions(simulation.nodes());
      }
      
      if (now - lastLinkUpdateTime >= 16) {
        lastLinkUpdateTime = now;
        if (objectManager && simulation) {
          const links = (simulation.force('link') as any)?.links() || [];
          objectManager.updateLinks(links);
        }
      }
      
      // Прогрессивное рассеивание тумана на основе прогресса позиционирования узлов
      if (scene && ++lastFogUpdate % 5 === 0) { // Обновляем чаще (каждые 5 тиков)
        const nodes = simulation.nodes();
        const nodesWithPosition = nodes.filter((n: any) => 
          n.x !== undefined && 
          !isNaN(n.x) && 
          n.y !== undefined && 
          !isNaN(n.y) && 
          n.z !== undefined && 
          !isNaN(n.z)
        ).length;
        
        const progress = Math.min(nodesWithPosition / totalNodes, 1);
        const startDensity = 0.08;
        const endDensity = 0.005;
        
        // Плавное уменьшение плотности тумана
        const currentDensity = startDensity - (startDensity - endDensity) * progress;
        setFogDensity(scene, currentDensity);
        
        // Скрываем загрузочный overlay когда прогресс > 10%
        if (isLoading && progress > 0.1) {
          isLoading = false;
        }
        
        // Логирование для отладки
        if (import.meta.env.DEV && lastFogUpdate % 50 === 0) {
          console.log(`[Graph3D] Fog progress: ${(progress * 100).toFixed(1)}%, density: ${currentDensity.toFixed(4)}`);
        }
      }
    });
  }

  
  export function addData(newData: GraphData) {
    if (!simulation || !objectManager || !scene) return;
    
    const newNodeIds = new Set(newData.nodes.map((n: GraphNode) => n.id));
    const validNewLinks = newData.links.filter((l: GraphLink) => 
      newNodeIds.has(l.source) && newNodeIds.has(l.target)
    );
    
    addNodesToSimulation(simulation, newData.nodes, validNewLinks);
    
    const existingNodeIds = new Set(currentData.nodes.map((n: GraphNode) => n.id));
    const existingLinkIds = new Set(currentData.links.map((l: GraphLink) => `${l.source}-${l.target}`));
    
    const mergedNodes = [...currentData.nodes, ...newData.nodes.filter((n: GraphNode) => !existingNodeIds.has(n.id))];
    const mergedLinks = [...currentData.links, ...validNewLinks.filter((l: GraphLink) => !existingLinkIds.has(`${l.source}-${l.target}`))];
    
    currentData = { nodes: mergedNodes, links: mergedLinks };
    if (scene) setFogDensity(scene, 0.005);
    isFullyLoaded = true;
    
    setTimeout(() => {
      if (simulation && camera && controls) {
        if (centerNodeId) {
          centerCameraOnNode(centerNodeId, simulation.nodes(), camera, controls, true);
        } else {
          autoZoomToFit(simulation.nodes(), camera, controls, true);
        }
      }
    }, 500);
  }

  onMount(async () => {
    if (!browser || !container) return;

    try {
      const setup = initScene(container);
      scene = setup.scene;
      camera = setup.camera;
      renderer = setup.renderer;
      labelRenderer = setup.labelRenderer;
      controls = setup.controls;

      objectManager = new ObjectManager(scene);
      
      if (browser) {
        (window as any).scene = scene;
        (window as any).camera = camera;
        (window as any).controls = controls;
        (window as any).simulation = simulation;
      }
      
      let frameCount = 0;
      function animate() {
        animationFrame = requestAnimationFrame(animate);
        controls.update();
        renderer.render(scene, camera);
        labelRenderer.render(scene, camera);
        if (++frameCount === 30) {
          console.log('[Graph3D] Frame 30 - Camera:', 
            `(${camera.position.x.toFixed(1)}, ${camera.position.y.toFixed(1)}, ${camera.position.z.toFixed(1)})`);
        }
      }
      animate();

      window.addEventListener('resize', onResize);
      container.addEventListener('click', handleClick);
      container.addEventListener('dblclick', handleDoubleClick);
      
      isInitialized = true;
      
      if (data.nodes.length > 0) {
        createGraphSimulation(data.nodes, data.links);
        lastProcessedKey = dataUpdateKey;
      } else {
        isLoading = false;
      }
    } catch (e) {
      console.error('Failed to initialize 3D graph:', e);
      
      // Проверяем тип ошибки для лучшей диагностики
      if (e instanceof Error) {
        if (e.message.includes('WebGL')) {
          error = 'WebGL не поддерживается. Попробуйте обновить браузер или использовать 2D режим.';
        } else if (e.message.includes('context')) {
          error = 'Ошибка создания контекста рендеринга. Попробуйте перезагрузить страницу.';
        } else {
          error = `Ошибка инициализации 3D графа: ${e.message}`;
        }
      } else {
        error = 'Неизвестная ошибка при инициализации 3D графа';
      }
      
      isLoading = false;
    }
  });

  onDestroy(() => {
    if (animationFrame) cancelAnimationFrame(animationFrame);
    if (fogAnimationFrame) cancelAnimationFrame(fogAnimationFrame);
    disposeScene(renderer, controls, objectManager);
    if (simulation) simulation.stop();
    window.removeEventListener('resize', onResize);
    if (container) {
      container.removeEventListener('click', handleClick);
      container.removeEventListener('dblclick', handleDoubleClick);
    }
  });

  export function resetCamera() {
    if (simulation && camera && controls) {
      if (centerNodeId) {
        centerCameraOnNode(centerNodeId, simulation.nodes(), camera, controls, isFullyLoaded);
      } else {
        autoZoomToFit(simulation.nodes(), camera, controls, isFullyLoaded);
      }
    }
  }

  export function toggleAutoRotate() {
    isAutoRotating = toggleCameraAutoRotate(controls, cameraState);
  }
</script>

<div bind:this={container} class="graph-3d-container" data-testid="graph-3d-container">
  {#if isLoading}
    <div class="loading-overlay" transition:fade={{ duration: 800 }} data-testid="loading-overlay">
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

  <!-- Camera Controls -->
  {#if !isLoading && !error}
    <div class="camera-controls">
      <button 
        class="camera-btn" 
        onclick={resetCamera}
        data-testid="reset-camera-button"
        title="Reset camera to fit all nodes"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M15 3h6v6M9 21H3v-6M21 3l-7 7M3 21l7-7"/>
        </svg>
        <span>Reset View</span>
      </button>
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

  .loading-overlay {
    pointer-events: none; /* Allow interaction with graph during loading */
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

  .camera-controls {
    position: absolute;
    bottom: 20px;
    right: 20px;
    z-index: 100;
  }

  .camera-btn {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: white;
    padding: 0.75rem 1rem;
    border-radius: 8px;
    cursor: pointer;
    font-size: 0.9rem;
    transition: all 0.2s ease;
  }

  .camera-btn:hover {
    background: rgba(255, 255, 255, 0.2);
    transform: translateY(-2px);
  }

  .camera-btn svg {
    width: 18px;
    height: 18px;
  }
</style>
