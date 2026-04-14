# Knowledge Graph - Полный код всех файлов

## 1️⃣ Страница 3D-графа: `src/routes/graph/3d/[id]/+page.svelte`

```svelte
<script lang="ts">
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { getGraphData, getFullGraphData } from '$lib/api/graph';
  import LazyGraph3D from '$lib/components/LazyGraph3D.svelte';
  import type { GraphData } from '$lib/api/graph';

  let graphData: GraphData | null = $state(null);
  let loading = $state(true);
  let error = $state('');
  let showFullGraph = $state(false);
  let currentNoteId: string | null = $state(null);

  $effect(() => {
    if (!browser) return;
    const id = $page.params.id;
    const mode = showFullGraph;
    if (id && id !== currentNoteId) currentNoteId = id;
    if (currentNoteId) loadGraph(currentNoteId);
  });

  async function loadGraph(noteId: string) {
    console.log('[3D Page] loadGraph:', { noteId, showFullGraph });
    loading = true;
    error = '';
    try {
      const newData = showFullGraph 
        ? await getFullGraphData() 
        : await getGraphData(noteId, 2);
      graphData = newData;
      console.log('[3D Page] Loaded:', newData.nodes.length, 'nodes');
    } catch (e) {
      error = 'Failed to load graph data';
      console.error(e);
    } finally {
      loading = false;
    }
  }
</script>

<div class="page">
  <div class="controls">
    <label class="toggle">
      <input type="checkbox" bind:checked={showFullGraph} />
      <span>Показать все заметки ({showFullGraph ? 'вкл' : 'выкл'})</span>
    </label>
  </div>
  
  {#if !loading && !error && graphData}
    <div class="stats-bar">
      <span class="stats-item"><strong>{graphData.nodes.length}</strong> nodes</span>
      <span class="stats-item"><strong>{graphData.links.length}</strong> links</span>
      <span class="stats-mode">({showFullGraph ? 'Full' : 'Local'})</span>
    </div>
  {/if}
  
  {#if error}
    <div class="center error">{error}</div>
  {:else if graphData}
    <div class="graph-wrapper" class:loading>
      <LazyGraph3D data={graphData} />
      {#if loading}
        <div class="loading-overlay">
          <div class="spinner"></div>
          <span>Загрузка...</span>
        </div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .page { position: relative; width: 100%; height: 100vh; overflow: hidden; }
  .controls {
    position: absolute; top: 80px; right: 20px; z-index: 10000;
    background: rgba(0, 0, 0, 0.85); padding: 12px 16px; border-radius: 8px;
    color: white; border: 1px solid rgba(255, 255, 255, 0.3);
    backdrop-filter: blur(10px);
  }
  .toggle { display: flex; align-items: center; gap: 8px; cursor: pointer; font-size: 14px; }
  .toggle input { cursor: pointer; width: 18px; height: 18px; }
  .center { display: flex; align-items: center; justify-content: center; height: 100%; color: white; }
  .error { color: #ff6666; }
  .stats-bar {
    position: absolute; top: 140px; right: 20px; z-index: 10000;
    display: flex; align-items: center; gap: 16px; padding: 10px 16px;
    background: rgba(0, 0, 0, 0.85); border-radius: 8px; font-size: 14px;
    color: #94a3b8; border: 1px solid rgba(255, 255, 255, 0.2); backdrop-filter: blur(10px);
  }
  .stats-item { display: flex; align-items: center; gap: 4px; }
  .stats-item strong { color: #88aaff; font-weight: 600; }
  .stats-mode { margin-left: 8px; font-style: italic; color: #64748b; font-size: 12px; }
  .graph-wrapper { position: relative; width: 100%; height: 100%; }
  .loading-overlay {
    position: absolute; top: 0; left: 0; right: 0; bottom: 0;
    background: rgba(0, 0, 0, 0.7); display: flex; flex-direction: column;
    align-items: center; justify-content: center; color: white; z-index: 1000; gap: 12px;
  }
  .spinner {
    width: 40px; height: 40px; border: 3px solid rgba(255, 255, 255, 0.3);
    border-top-color: #88aaff; border-radius: 50%; animation: spin 1s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
```

---

## 2️⃣ Главная страница с 2D-графом: `src/routes/+page.svelte`

```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import FloatingControls from '$lib/components/FloatingControls.svelte';
  import NoteSidePanel from '$lib/components/NoteSidePanel.svelte';
  import CreateNoteModal from '$lib/components/CreateNoteModal.svelte';
  import ConfirmModal from '$lib/components/ConfirmModal.svelte';
  import NoteCard from '$lib/components/NoteCard.svelte';
  import { getNotes, deleteNote, searchNotes, type Note } from '$lib/api/notes';
  import { getGraphData, getFullGraphData, type GraphData } from '$lib/api/graph';
  import GraphCanvas from '$lib/components/GraphCanvas.svelte';

  let allNotes: Note[] = $state([]);
  let filteredNotes: Note[] = $state([]);
  let loading = $state(true);
  let error = $state('');
  let selectedNodeId: string | null = $state(null);
  let showCreateModal = $state(false);
  let showConfirmDelete = $state(false);
  let noteToDelete: string | null = $state(null);
  let currentView: 'graph' | 'list' = $state('graph');
  
  let graphData: GraphData = $state({ nodes: [], links: [] });
  let graphLoading = $state(false);
  let searchQuery = $state('');
  let showFullGraph = $state(false);
  let selectedType = $state<string>('all');
  let sortOption = $state<string>('newest');

  const typeFilters = [
    { id: 'all', label: 'All', emoji: '🌌' },
    { id: 'star', label: 'Stars', emoji: '⭐' },
    { id: 'planet', label: 'Planets', emoji: '🪐' },
    { id: 'comet', label: 'Comets', emoji: '☄️' },
    { id: 'galaxy', label: 'Galaxies', emoji: '🌀' }
  ];

  onMount(async () => {
    if (!browser) return;
    await loadNotes();
  });

  async function loadNotes() {
    try {
      allNotes = await getNotes();
      applyFiltersAndSort();
      await loadGraphData();
    } catch (e) {
      error = 'Failed to load notes';
      console.error(e);
    } finally {
      loading = false;
    }
  }
  
  async function loadGraphData() {
    if (allNotes.length === 0) return;
    graphLoading = true;
    try {
      graphData = showFullGraph 
        ? await getFullGraphData()
        : await getGraphData(allNotes[0].id, 2);
    } catch (e) {
      console.error('Failed to load graph:', e);
      graphData = { nodes: allNotes.map(n => ({ id: n.id, title: n.title, type: n.type || 'star' })), links: [] };
    } finally {
      graphLoading = false;
    }
  }
  
  $effect(() => { if (browser) loadGraphData(); });

  function applyFiltersAndSort() {
    let result = [...allNotes];
    if (selectedType !== 'all') result = result.filter(n => n.type === selectedType);
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      result = result.filter(n => n.title.toLowerCase().includes(query) || n.content.toLowerCase().includes(query));
    }
    filteredNotes = result;
  }

  async function handleSearch() {
    if (!searchQuery.trim()) { await loadNotes(); return; }
    try {
      const response = await searchNotes(searchQuery, 1, 20);
      allNotes = response.data;
      applyFiltersAndSort();
    } catch (e) { console.error('Search error:', e); }
  }

  async function handleDeleteConfirm() {
    if (!noteToDelete) return;
    try {
      await deleteNote(noteToDelete);
      selectedNodeId = null;
      allNotes = allNotes.filter(n => n.id !== noteToDelete);
      await loadNotes();
    } catch { alert('Failed to delete note'); }
    finally { noteToDelete = null; showConfirmDelete = false; }
  }

  function handleToggleView() { currentView = currentView === 'graph' ? 'list' : 'graph'; }
  function handleToggle3D() { goto('/graph'); }
</script>

<div class="page-container">
  <FloatingControls
    onCreate={() => showCreateModal = true}
    onSearch={(q) => { searchQuery = q; handleSearch(); }}
    onToggleView={handleToggleView}
    onToggle3D={handleToggle3D}
    noteId={selectedNodeId ?? undefined}
  />

  <div class="content-area">
    {#if loading}
      <div class="center"><div class="spinner"></div><p>Loading...</p></div>
    {:else if error}
      <p class="error">{error}</p>
    {:else}
      <div class="controls-bar">
        <div class="filter-group">
          {#each typeFilters as filter}
            <button class:active={selectedType === filter.id} onclick={() => { selectedType = filter.id; applyFiltersAndSort(); }}>
              {filter.emoji} {filter.label}
            </button>
          {/each}
        </div>
        <input bind:value={searchQuery} onkeyup={(e) => e.key === 'Enter' && handleSearch()} placeholder="Search..." />
      </div>

      {#if currentView === 'graph'}
        <label><input type="checkbox" bind:checked={showFullGraph} /> Показать все заметки</label>
        
        <div class="graph-container">
          {#if graphLoading}
            <div class="center"><div class="spinner"></div><p>Loading graph...</p></div>
          {:else if graphData.nodes.length > 0}
            <GraphCanvas nodes={graphData.nodes} links={graphData.links} onNodeClick={(n) => selectedNodeId = n.id} />
          {:else}
            <div class="empty-state">🌌 No graph data</div>
          {/if}
        </div>
      {:else}
        <div class="notes-grid">
          {#each filteredNotes as note (note.id)}
            <NoteCard {note} onClick={() => selectedNodeId = note.id} />
          {/each}
        </div>
      {/if}
    {/if}
  </div>
</div>

{#if selectedNodeId}
  <NoteSidePanel nodeId={selectedNodeId} onClose={() => selectedNodeId = null} onEdit={(id) => goto(`/notes/${id}/edit`)} onDelete={handleDeleteConfirm} />
{/if}

<CreateNoteModal bind:open={showCreateModal} onSuccess={(n) => { selectedNodeId = n.id; loadNotes(); }} />
<ConfirmModal bind:open={showConfirmDelete} title="Delete?" message="Are you sure?" confirmText="Delete" danger onConfirm={handleDeleteConfirm} />

<style>
  .page-container { min-height: 100vh; padding-top: 100px; background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%); }
  .content-area { max-width: 1400px; margin: 0 auto; padding: 0 24px 40px; }
  .center { display: flex; flex-direction: column; align-items: center; padding: 60px; color: #64748b; }
  .spinner { width: 40px; height: 40px; border: 3px solid #e2e8f0; border-top-color: #3b82f6; border-radius: 50%; animation: spin 1s infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }
  .error { padding: 20px; background: #fee2e2; color: #dc2626; border-radius: 8px; text-align: center; }
  .graph-container { height: 600px; margin-bottom: 40px; }
  .notes-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 24px; }
  .controls-bar { display: flex; justify-content: space-between; margin-bottom: 20px; flex-wrap: wrap; gap: 16px; }
  .filter-group { display: flex; gap: 8px; flex-wrap: wrap; }
  button { padding: 8px 16px; background: white; border: 1px solid #e2e8f0; border-radius: 20px; cursor: pointer; }
  button.active { background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%); color: white; border-color: transparent; }
  input { padding: 8px 12px; border: 1px solid #e2e8f0; border-radius: 8px; }
</style>
```

---

## 3️⃣ Компонент 3D-графа: `src/lib/components/Graph3D.svelte`

```svelte
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { browser } from '$app/environment';
  import { initScene } from '$lib/three/core/sceneSetup';
  import { createSimulation } from '$lib/three/simulation/forceSimulation';
  import { ObjectManager } from '$lib/three/rendering/objectManager';
  import { autoZoomToFit } from '$lib/three/camera/cameraUtils';
  import type { GraphData } from '$lib/api/graph';
  import * as THREE from 'three';

  const { data, onNodeClick }: { data: GraphData; onNodeClick?: (n: any) => void; } = $props();

  let container: HTMLDivElement;
  let scene: THREE.Scene, camera: THREE.PerspectiveCamera, renderer: THREE.WebGLRenderer;
  let labelRenderer: any, controls: any, simulation: any, objectManager: ObjectManager;
  let animationFrame: number, isLoading = $state(true), error = $state<string | null>(null);
  let isInitialized = $state(false), lastProcessedKey = $state(0), dataUpdateKey = $state(0);

  $effect(() => {
    const newKey = data.nodes.length + data.links.length * 1000;
    if (newKey !== dataUpdateKey) dataUpdateKey = newKey;
  });

  $effect(() => {
    const _key = dataUpdateKey;
    if (!isInitialized || _key === lastProcessedKey) return;
    if (data.nodes.length > 500) console.warn('[Graph3D] Large graph:', data.nodes.length, 'nodes');
    lastProcessedKey = _key;
    createGraphSimulation();
  });

  onMount(async () => {
    if (!browser || !container) return;
    try {
      const setup = initScene(container);
      scene = setup.scene; camera = setup.camera; renderer = setup.renderer;
      labelRenderer = setup.labelRenderer; controls = setup.controls;
      objectManager = new ObjectManager(scene);
      
      function animate() {
        animationFrame = requestAnimationFrame(animate);
        controls.update();
        renderer.render(scene, camera);
        labelRenderer.render(scene, camera);
      }
      animate();

      window.addEventListener('resize', onResize);
      isInitialized = true;
      if (data.nodes.length > 0) createGraphSimulation();
    } catch (e) {
      error = 'Failed to initialize 3D';
      isLoading = false;
    }
  });

  function onResize() {
    if (!container || !renderer || !camera) return;
    const w = container.clientWidth, h = container.clientHeight;
    renderer.setSize(w, h); labelRenderer.setSize(w, h);
    camera.aspect = w / h; camera.updateProjectionMatrix();
  }

  function createGraphSimulation() {
    if (!objectManager || !scene) return;
    objectManager.clear();
    if (simulation) simulation.stop();
    
    isLoading = true;
    simulation = createSimulation(data, objectManager);
    
    const loadingTimeout = setTimeout(() => {
      if (isLoading) { isLoading = false; autoZoomToFit(simulation.nodes(), camera, controls); }
    }, 5000);
    
    simulation.on('end', () => {
      clearTimeout(loadingTimeout);
      if (simulation && camera && controls) autoZoomToFit(simulation.nodes(), camera, controls);
      isLoading = false;
    });
    
    simulation.on('tick', () => { if (objectManager) objectManager.updatePositions(); });
  }

  onDestroy(() => {
    if (animationFrame) cancelAnimationFrame(animationFrame);
    if (simulation) simulation.stop();
    if (renderer) renderer.dispose();
    window.removeEventListener('resize', onResize);
  });
</script>

<div bind:this={container} class="graph-3d-container">
  {#if isLoading}
    <div class="loading-overlay"><div class="spinner"></div><p>Loading 3D...</p></div>
  {/if}
  {#if error}
    <div class="error-overlay"><span>🌌</span><h2>{error}</h2></div>
  {/if}
</div>

<style>
  .graph-3d-container { width: 100%; height: 100vh; position: relative; overflow: hidden; background: #050510; }
  .loading-overlay, .error-overlay { position: absolute; inset: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; background: rgba(5,5,16,0.95); color: white; z-index: 10; }
  .spinner { width: 40px; height: 40px; border: 3px solid rgba(255,255,255,0.3); border-top-color: #88aaff; border-radius: 50%; animation: spin 1s linear infinite; }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
```

---

## 4️⃣ Компонент 2D-графа: `src/lib/components/GraphCanvas.svelte`

```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';

  const { nodes, links, onNodeClick }: { 
    nodes: Array<{ id: string; title: string; type?: string }>; 
    links: Array<{ source: string; target: string; weight?: number; link_type?: string }>;
    onNodeClick?: (n: any) => void;
  } = $props();

  const linkTypeColors: Record<string, string> = { reference: '#3366ff', dependency: '#ff6600', related: '#999999', custom: '#ff66ff' };
  function getLinkColor(w: number, t?: string) {
    const c = linkTypeColors[t || 'related'];
    const o = 0.4 + (w || 0.5) * 0.4;
    return `rgba(${parseInt(c.slice(1,3),16)}, ${parseInt(c.slice(3,5),16)}, ${parseInt(c.slice(5,7),16)}, ${o})`;
  }

  let canvas: HTMLCanvasElement, ctx: CanvasRenderingContext2D;
  let w = 800, h = 600, animationId: number;
  const angles: Map<string, number> = new Map(), speeds: Map<string, number> = new Map();
  let d3Force: any = null, simulation: any = null;
  const transform = $state({ x: 0, y: 0, k: 1 });
  let dragging = $state(false), dragStart = $state({ x: 0, y: 0 });

  onMount(() => {
    if (!browser) return;
    let cleanup = () => {};
    import('d3-force').then(d3 => {
      d3Force = d3;
      ctx = canvas.getContext('2d')!;
      resize(); window.addEventListener('resize', resize);
      startSimulation(); startAnimation();
      cleanup = () => { window.removeEventListener('resize', resize); if (simulation) simulation.stop(); cancelAnimationFrame(animationId); };
    });
    return () => cleanup();
  });

  $effect(() => {
    if (d3Force && nodes.length > 0) { if (simulation) simulation.stop(); angles.clear(); speeds.clear(); startSimulation(); }
  });

  function startAnimation() {
    function animate() {
      for (const node of simulation?.nodes() || []) {
        const t = node.type || 'star', id = node.id;
        let bs = t === 'planet' ? 0.02 : t === 'comet' ? 0.03 : t === 'galaxy' ? 0.01 : 0.005;
        let s = speeds.get(id); if (s === undefined) { s = bs * (0.7 + Math.random() * 0.6); speeds.set(id, s); }
        angles.set(id, (angles.get(id) || 0) + s);
      }
      draw(); animationId = requestAnimationFrame(animate);
    }
    animate();
  }

  function resize() {
    const r = canvas.parentElement?.getBoundingClientRect();
    if (r) { w = r.width; h = r.height; canvas.width = w; canvas.height = h; }
  }

  function startSimulation() {
    if (!d3Force) return;
    const sn = nodes.map(n => ({ ...n, x: w/2, y: h/2 }));
    const e = links.map(l => ({ source: l.source, target: l.target, weight: l.weight || 1 }));
    simulation = d3Force.forceSimulation(sn)
      .force('link', d3Force.forceLink(e).id((d: any) => d.id).distance(150).strength(0.5))
      .force('charge', d3Force.forceManyBody().strength(-300))
      .force('center', d3Force.forceCenter(w/2, h/2))
      .force('collision', d3Force.forceCollide().radius(40))
      .alphaDecay(0.02)
      .on('tick', draw);
    simulation.tick(100);
  }

  function drawStar(x: number, y: number, r: number, a: number) {
    if (!ctx) return;
    ctx.beginPath();
    for (let i = 0, rot = a; i < 5; i++) {
      ctx.lineTo(x + Math.cos(rot) * r, y + Math.sin(rot) * r); rot += Math.PI / 5;
      ctx.lineTo(x + Math.cos(rot) * r * 0.4, y + Math.sin(rot) * r * 0.4); rot += Math.PI / 5;
    }
    ctx.closePath();
  }

  function draw() {
    if (!ctx) return;
    ctx.clearRect(0, 0, w, h);
    ctx.save(); ctx.translate(transform.x, transform.y); ctx.scale(transform.k, transform.k);

    links.forEach(l => {
      const s = simulation.nodes().find((n: any) => n.id === l.source);
      const t = simulation.nodes().find((n: any) => n.id === l.target);
      if (!s || !t) return;
      ctx.beginPath(); ctx.moveTo(s.x, s.y); ctx.lineTo(t.x, t.y);
      ctx.lineWidth = Math.max(1, (l.weight || 0.5) * 4);
      ctx.strokeStyle = getLinkColor(l.weight || 0.5, l.link_type);
      ctx.stroke();
    });

    const r = 24;
    simulation.nodes().forEach((n: any) => {
      const a = angles.get(n.id) || 0;
      switch (n.type || 'star') {
        case 'star': drawStar(n.x, n.y, r, a); ctx.fillStyle = '#ffdd88'; ctx.fill(); break;
        default: drawStar(n.x, n.y, r, a); ctx.fillStyle = '#cccccc'; ctx.fill();
      }
      ctx.font = `bold 14px sans-serif`; ctx.textAlign = 'center';
      ctx.fillStyle = '#ffffff'; ctx.fillText(n.title.slice(0, 14), n.x, n.y + r + 6);
    });
    ctx.restore();
  }

  function handleZoom(e: WheelEvent) {
    e.preventDefault();
    const newK = transform.k * (e.deltaY > 0 ? 0.95 : 1.05);
    if (newK >= 0.2 && newK <= 5) { transform.k = newK; draw(); }
  }

  function handlePanStart(e: MouseEvent) { dragging = true; dragStart = { x: e.clientX - transform.x, y: e.clientY - transform.y }; }
  function handlePanMove(e: MouseEvent) { if (!dragging) return; transform.x = e.clientX - dragStart.x; transform.y = e.clientY - dragStart.y; draw(); }
  function handlePanEnd() { dragging = false; }

  function handleClick(e: MouseEvent) {
    const rect = canvas.getBoundingClientRect();
    const cx = (e.clientX - rect.left - transform.x) / transform.k;
    const cy = (e.clientY - rect.top - transform.y) / transform.k;
    const node = simulation.nodes().find((n: any) => Math.hypot(n.x - cx, n.y - cy) < 24);
    if (node) onNodeClick ? onNodeClick(node) : goto(`/notes/${node.id}`);
  }
</script>

<canvas
  bind:this={canvas}
  onmousedown={handlePanStart}
  onmousemove={handlePanMove}
  onmouseup={handlePanEnd}
  onclick={handleClick}
  onwheel={handleZoom}
  style="width: 100%; height: 100%; cursor: grab; background: linear-gradient(145deg, #0a1a3a, #020617);"
/>
```

---

## 5️⃣ API клиент для заметок: `src/lib/api/notes.ts`

```typescript
import ky from 'ky';

const api = ky.create({ prefixUrl: '/api', timeout: 30000, retry: { limit: 2, methods: ['get', 'post', 'put', 'delete'], statusCodes: [408, 413, 429, 500, 502, 503, 504] }});

export interface Note { id: string; title: string; content: string; metadata: Record<string, any>; type?: string; created_at: string; updated_at: string; }
export interface Suggestion { note_id: string; title: string; score: number; }

export async function getNotes(): Promise<Note[]> { return (await api.get('notes').json<{ notes: Note[] }>()).notes; }
export async function getNote(id: string): Promise<Note> { return api.get(`notes/${id}`).json(); }
export async function createNote(data: { title: string; content: string; type?: string; metadata?: any }): Promise<Note> { return api.post('notes', { json: data }).json(); }
export async function updateNote(id: string, data: Partial<Note>): Promise<Note> { return api.put(`notes/${id}`, { json: data }).json(); }
export async function deleteNote(id: string): Promise<void> { return api.delete(`notes/${id}`).json(); }
export async function getSuggestions(id: string, limit = 10): Promise<Suggestion[]> { return api.get(`notes/${id}/suggestions`, { searchParams: { limit } }).json(); }

export interface SearchResponse { data: Note[]; total: number; page: number; size: number; totalPages: number; }
export async function searchNotes(query: string, page = 1, size = 20): Promise<SearchResponse> {
  return api.get(`notes/search?q=${query}&page=${page}&size=${size}`).json();
}
```

---

## 6️⃣ Симуляция 3D-графа: `src/lib/three/simulation/forceSimulation.ts`

```typescript
import { forceSimulation, forceLink, forceManyBody, forceCenter } from 'd3-force-3d';
import type { GraphData } from '$lib/api/graph';
import type { ObjectManager } from '$lib/three/rendering/objectManager';

export function createSimulation(data: GraphData, objectManager: ObjectManager) {
  const nodes = data.nodes.map(n => ({ ...n, x: (n as any).x ?? (Math.random() - 0.5) * 30, y: (n as any).y ?? (Math.random() - 0.5) * 30, z: (n as any).z ?? (Math.random() - 0.5) * 30 }));
  const links = data.links.map(l => ({ ...l, source: l.source, target: l.target, value: l.weight ?? 1 }));
  objectManager.createAll(nodes, links);

  return forceSimulation(nodes)
    .force('link', forceLink(links).id((d: any) => d.id).distance((l: any) => 20 / (l.value * 0.8)).strength(0.5))
    .force('charge', forceManyBody().strength(-200).distanceMax(50))
    .force('center', forceCenter(0, 0, 0))
    .alphaDecay(0.02)
    .on('tick', () => { objectManager.updatePositions(nodes); objectManager.updateLinks(links); });
}
```

---

## 7️⃣ Управление объектами 3D: `src/lib/three/rendering/objectManager.ts`

```typescript
import * as THREE from 'three';
import { CSS2DObject } from 'three/examples/jsm/renderers/CSS2DRenderer.js';
import { createNodeMesh } from './nodeFactory';
import { createLinkLine } from './linkFactory';
import { createLabel } from './labelFactory';
import type { GraphNode, GraphLink } from '$lib/api/graph';

export class ObjectManager {
  private nodeMap = new Map<string, THREE.Group>();
  private linkMap = new Map<string, THREE.Line>();
  private labelMap = new Map<string, CSS2DObject>();

  constructor(private scene: THREE.Scene) {}

  createAll(nodes: GraphNode[], links: GraphLink[]) {
    this.clear();
    nodes.forEach(node => {
      const mesh = createNodeMesh(node);
      mesh.position.set((node as any).x ?? 0, (node as any).y ?? 0, (node as any).z ?? 0);
      mesh.userData = { type: 'node', id: node.id, nodeData: node };
      this.scene.add(mesh); this.nodeMap.set(node.id, mesh);
      const label = createLabel(node.title || node.id.slice(0, 6), () => { window.location.href = `/notes/${node.id}`; });
      label.position.copy(mesh.position); this.scene.add(label); this.labelMap.set(node.id, label);
    });
    links.forEach(link => {
      const s = nodes.find(n => n.id === link.source), t = nodes.find(n => n.id === link.target);
      if (!s || !t) return;
      const line = createLinkLine(new THREE.Vector3((s as any).x ?? 0, (s as any).y ?? 0, (s as any).z ?? 0), new THREE.Vector3((t as any).x ?? 0, (t as any).y ?? 0, (t as any).z ?? 0), link.weight ?? 1, link.link_type);
      this.scene.add(line); this.linkMap.set(`${link.source}-${link.target}`, line);
    });
  }

  updatePositions(nodes: GraphNode[]) {
    nodes.forEach(node => {
      const obj = this.nodeMap.get(node.id), label = this.labelMap.get(node.id);
      if (obj) { obj.position.set((node as any).x ?? 0, (node as any).y ?? 0, (node as any).z ?? 0); if (label) label.position.copy(obj.position); }
    });
  }

  updateLinks(links: GraphLink[]) {
    links.forEach(link => {
      const line = this.linkMap.get(`${link.source}-${link.target}`);
      if (!line) return;
      const s = this.nodeMap.get(link.source), t = this.nodeMap.get(link.target);
      if (!s || !t) return;
      line.geometry.dispose();
      line.geometry = new THREE.BufferGeometry().setFromPoints([s.position.clone(), t.position.clone()]);
    });
  }

  clear() {
    this.nodeMap.forEach(o => this.scene.remove(o));
    this.linkMap.forEach(o => this.scene.remove(o));
    this.labelMap.forEach(o => this.scene.remove(o));
    this.nodeMap.clear(); this.linkMap.clear(); this.labelMap.clear();
  }

  getNode(id: string) { return this.nodeMap.get(id); }
}
```

---

## 8️⃣ Плавающие кнопки: `src/lib/components/FloatingControls.svelte`

```svelte
<script lang="ts">
  import { goto } from '$app/navigation';
  
  const { onCreate, onSearch, onToggleView, onToggle3D, noteId = '' }: {
    onCreate?: () => void; onSearch?: (q: string) => void; onToggleView?: () => void; onToggle3D?: () => void; noteId?: string;
  } = $props();
  
  let searchQuery = $state(''), currentView = $state<'graph' | 'list'>('graph'), showMenu = $state(false);
  
  function handleSearch() { onSearch?.(searchQuery); }
  function handleToggle3D() { goto(noteId ? `/graph/3d/${noteId}` : '/graph/3d'); }
  function toggleView() { currentView = currentView === 'graph' ? 'list' : 'graph'; onToggleView?.(); }
</script>

<div class="floating-controls">
  <div class="view-toggle">
    <button class:active={currentView === 'graph'} onclick={toggleView} title="2D">📊 2D</button>
    <button onclick={() => onToggle3D?.()} title="3D">🧊 3D</button>
    <button class:active={currentView === 'list'} onclick={toggleView} title="List">📋 List</button>
  </div>
  <div class="search-container">
    <input bind:value={searchQuery} onkeyup={(e) => e.key === 'Enter' && handleSearch()} placeholder="Search..." />
    <button onclick={handleSearch}>🔍</button>
  </div>
  <button class="create-btn" onclick={() => onCreate?.()}>➕</button>
</div>

<style>
  .floating-controls { position: fixed; top: 20px; left: 50%; transform: translateX(-50%); display: flex; gap: 12px; padding: 12px 20px; background: rgba(255,255,255,0.95); backdrop-filter: blur(10px); border-radius: 50px; box-shadow: 0 4px 20px rgba(0,0,0,0.1); z-index: 100; }
  .view-toggle { display: flex; gap: 4px; padding: 4px; background: #f1f5f9; border-radius: 25px; }
  button { padding: 8px 12px; border: none; background: transparent; border-radius: 20px; cursor: pointer; transition: all 0.2s; }
  button.active { background: white; color: #3b82f6; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
  .search-container { display: flex; align-items: center; gap: 8px; padding: 4px 4px 4px 16px; background: #f8fafc; border-radius: 25px; border: 1px solid #e2e8f0; }
  input { border: none; background: transparent; outline: none; font-size: 14px; width: 180px; }
  .create-btn { padding: 12px; background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%); color: white; border-radius: 50%; box-shadow: 0 4px 12px rgba(59,130,246,0.4); }
</style>
```

---

## 📋 Всего 8 файлов

| # | Файл | Назначение |
|---|------|-----------|
| 1 | `src/routes/graph/3d/[id]/+page.svelte` | Страница 3D-графа с переключателем |
| 2 | `src/routes/+page.svelte` | Главная страница с 2D-графом |
| 3 | `src/lib/components/Graph3D.svelte` | Компонент 3D-графа (Three.js) |
| 4 | `src/lib/components/GraphCanvas.svelte` | Компонент 2D-графа (Canvas) |
| 5 | `src/lib/api/notes.ts` | API клиент для заметок |
| 6 | `src/lib/three/simulation/forceSimulation.ts` | d3-force-3d симуляция |
| 7 | `src/lib/three/rendering/objectManager.ts` | Управление 3D объектами |
| 8 | `src/lib/components/FloatingControls.svelte` | Плавающие кнопки UI |
