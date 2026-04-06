# setup_frontend.ps1
# Скрипт для создания структуры фронтенда SvelteKit
# Запускать из корня проекта: D:\knowledge-graph

$frontendPath = Join-Path (Get-Location) "frontend"

Write-Host "=== Настройка фронтенда Knowledge Graph ===" -ForegroundColor Cyan
Write-Host ""

# Проверяем, существует ли папка frontend
if (-not (Test-Path $frontendPath)) {
    Write-Host "Папка frontend не найдена. Сначала создайте проект SvelteKit:" -ForegroundColor Yellow
    Write-Host "  npx sv create frontend" -ForegroundColor Yellow
    Write-Host "  cd frontend" -ForegroundColor Yellow
    Write-Host "  npm install" -ForegroundColor Yellow
    exit 1
}

# Функция для создания файла с содержимым (перезаписывает)
function Write-FileWithContent {
    param($Path, $Content)
    $fullPath = Join-Path $frontendPath $Path
    $dir = Split-Path $fullPath -Parent
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }
    Set-Content -Path $fullPath -Value $Content -Encoding utf8 -Force
    Write-Host "  Создан/обновлён: $Path" -ForegroundColor Green
}

# ========== 1. API-клиенты ==========
Write-Host "Создание API-клиентов..." -ForegroundColor Yellow

# src/lib/api/notes.ts
$notesApi = @'
// API-клиент для работы с заметками и рекомендациями
import ky from 'ky';

// Базовый URL с прокси /api → http://localhost:8080
const api = ky.create({ prefixUrl: '/api' });

// Тип данных заметки (соответствует ответу бэкенда)
export interface Note {
  id: string;
  title: string;
  content: string;
  metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
}

// Тип рекомендации (похожая заметка)
export interface Suggestion {
  note_id: string;
  title: string;
  score: number;      // вес от 0 до 1
}

// Получить все заметки (GET /notes)
export async function getNotes(): Promise<Note[]> {
  return api.get('notes').json();
}

// Получить одну заметку по ID
export async function getNote(id: string): Promise<Note> {
  return api.get(`notes/${id}`).json();
}

// Создать новую заметку
export async function createNote(data: { title: string; content: string; metadata?: any }): Promise<Note> {
  return api.post('notes', { json: data }).json();
}

// Обновить существующую заметку
export async function updateNote(id: string, data: Partial<Note>): Promise<Note> {
  return api.put(`notes/${id}`, { json: data }).json();
}

// Удалить заметку
export async function deleteNote(id: string): Promise<void> {
  return api.delete(`notes/${id}`).json();
}

// Получить рекомендации для заметки (похожие по явным связям и эмбеддингам)
export async function getSuggestions(id: string, limit = 10): Promise<Suggestion[]> {
  return api.get(`notes/${id}/suggestions`, { searchParams: { limit } }).json();
}
'@
Write-FileWithContent -Path "src/lib/api/notes.ts" -Content $notesApi

# src/lib/api/graph.ts
$graphApi = @'
// API-клиент для получения данных графа (узлы и связи)
import ky from 'ky';

const api = ky.create({ prefixUrl: '/api' });

// Узел графа – заметка (звезда)
export interface GraphNode {
  id: string;
  title: string;
}

// Ребро графа – связь между заметками
export interface GraphLink {
  source: string;   // ID исходной заметки
  target: string;   // ID целевой заметки
  weight: number;    // вес связи (толщина линии)
}

// Данные графа: список узлов и рёбер
export interface GraphData {
  nodes: GraphNode[];
  links: GraphLink[];
}

// Запросить граф для заметки (возвращает все прямые связи и связанные заметки)
export async function getGraphData(noteId: string): Promise<GraphData> {
  return api.get(`notes/${noteId}/graph`).json();
}
'@
Write-FileWithContent -Path "src/lib/api/graph.ts" -Content $graphApi

# ========== 2. Компоненты ==========
Write-Host "Создание компонентов..." -ForegroundColor Yellow

# src/lib/components/GraphCanvas.svelte
$graphCanvas = @'
<script lang="ts">
  // Компонент интерактивного графа на canvas с силовой раскладкой (d3-force)
  import { onMount } from 'svelte';
  import * as d3Force from 'd3-force';
  import { goto } from '$app/navigation';

  // Входные параметры: узлы и рёбра
  export let nodes: Array<{ id: string; title: string }>;
  export let links: Array<{ source: string; target: string; weight: number }>;

  let canvas: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D;
  let width = 800;
  let height = 600;
  
  // Трансформации для зумирования и панорамирования
  let transform = { x: 0, y: 0, k: 1 };
  let dragging = false;
  let dragStart = { x: 0, y: 0 };
  
  // Симуляция физики (d3-force)
  let simulation: d3Force.Simulation<d3Force.SimulationNodeDatum, undefined>;

  onMount(() => {
    ctx = canvas.getContext('2d')!;
    resize();
    window.addEventListener('resize', resize);
    startSimulation();
    return () => {
      window.removeEventListener('resize', resize);
      if (simulation) simulation.stop();
    };
  });

  // Адаптация под размер родительского контейнера
  function resize() {
    const rect = canvas.parentElement?.getBoundingClientRect();
    if (rect) {
      width = rect.width;
      height = rect.height;
      canvas.width = width;
      canvas.height = height;
    }
  }

  // Запуск симуляции (расположение узлов)
  function startSimulation() {
    // Подготовка узлов для d3
    const simulationNodes = nodes.map(n => ({ id: n.id, title: n.title, x: width/2, y: height/2 }));
    const edges = links.map(l => ({ source: l.source, target: l.target, weight: l.weight }));

    simulation = d3Force.forceSimulation(simulationNodes)
      .force('link', d3Force.forceLink(edges).id((d: any) => d.id).distance(150).strength(0.5))
      .force('charge', d3Force.forceManyBody().strength(-300))  // отталкивание
      .force('center', d3Force.forceCenter(width/2, height/2))  // центр
      .force('collision', d3Force.forceCollide().radius(40))     // чтобы узлы не наезжали друг на друга
      .alphaDecay(0.02)  // скорость затухания (чем меньше, тем дольше движутся)
      .on('tick', () => draw());

    // Быстрое начальное схождение (не дожидаясь полной стабилизации)
    simulation.tick(100);
  }

  // Отрисовка графа на canvas
  function draw() {
    if (!ctx) return;
    ctx.clearRect(0, 0, width, height);
    ctx.save();
    ctx.translate(transform.x, transform.y);
    ctx.scale(transform.k, transform.k);

    // Рисуем связи (рёбра) – чем больше вес, тем толще и ярче линия
    links.forEach(link => {
      const sourceNode = simulation.nodes().find(n => n.id === link.source);
      const targetNode = simulation.nodes().find(n => n.id === link.target);
      if (!sourceNode || !targetNode) return;
      ctx.beginPath();
      ctx.moveTo(sourceNode.x!, sourceNode.y!);
      ctx.lineTo(targetNode.x!, targetNode.y!);
      const lineWidth = Math.max(1, link.weight * 3);
      ctx.lineWidth = lineWidth;
      ctx.strokeStyle = `rgba(100, 150, 200, ${0.3 + link.weight * 0.7})`;
      ctx.stroke();
    });

    // Рисуем узлы (звёзды)
    simulation.nodes().forEach(node => {
      const radius = 20;
      ctx.beginPath();
      ctx.arc(node.x!, node.y!, radius, 0, 2 * Math.PI);
      ctx.fillStyle = '#ffd966';
      ctx.shadowBlur = 10;
      ctx.shadowColor = 'rgba(0,0,0,0.3)';
      ctx.fill();
      ctx.shadowBlur = 0;
      ctx.fillStyle = '#333';
      ctx.font = '12px sans-serif';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      ctx.fillText(node.title.slice(0, 12), node.x!, node.y!);
    });

    ctx.restore();
  }

  // Обработка зумирования колёсиком мыши
  function handleZoom(e: WheelEvent) {
    e.preventDefault();
    const delta = e.deltaY > 0 ? 0.95 : 1.05;
    const newK = transform.k * delta;
    if (newK < 0.2 || newK > 5) return;
    transform.k = newK;
    draw();
  }

  // Обработка панорамирования (перетаскивание полотна)
  function handlePanStart(e: MouseEvent) {
    dragging = true;
    dragStart = { x: e.clientX - transform.x, y: e.clientY - transform.y };
    canvas.style.cursor = 'grabbing';
  }

  function handlePanMove(e: MouseEvent) {
    if (!dragging) return;
    transform.x = e.clientX - dragStart.x;
    transform.y = e.clientY - dragStart.y;
    draw();
  }

  function handlePanEnd() {
    dragging = false;
    canvas.style.cursor = 'grab';
  }

  // Клик по узлу – переход на страницу заметки
  function handleClick(e: MouseEvent) {
    const rect = canvas.getBoundingClientRect();
    const clickX = (e.clientX - rect.left - transform.x) / transform.k;
    const clickY = (e.clientY - rect.top - transform.y) / transform.k;
    const node = simulation.nodes().find(n => {
      const dx = n.x! - clickX;
      const dy = n.y! - clickY;
      return Math.hypot(dx, dy) < 20;
    });
    if (node) {
      goto(`/notes/${node.id}`);
    }
  }
</script>

<!-- Подписка на события мыши и колеса на уровне окна -->
<svelte:window on:wheel={handleZoom} />

<!-- Canvas занимает всю область родителя -->
<canvas
  bind:this={canvas}
  on:mousedown={handlePanStart}
  on:mousemove={handlePanMove}
  on:mouseup={handlePanEnd}
  on:click={handleClick}
  style="width: 100%; height: 100%; cursor: grab; background: #f0f2f5;"
></canvas>
'@
Write-FileWithContent -Path "src/lib/components/GraphCanvas.svelte" -Content $graphCanvas

# ========== 3. Страницы ==========
Write-Host "Создание страниц..." -ForegroundColor Yellow

# src/routes/+page.svelte (список заметок)
$listPage = @'
<script lang="ts">
  // Главная страница: список всех заметок
  import { onMount } from 'svelte';
  import type { Note } from '$lib/api/notes';
  import { getNotes, deleteNote } from '$lib/api/notes';

  let notes: Note[] = [];
  let loading = true;
  let error = '';

  // Загружаем заметки при монтировании компонента
  onMount(async () => {
    try {
      notes = await getNotes();
    } catch (e) {
      error = 'Failed to load notes';
      console.error(e);
    } finally {
      loading = false;
    }
  });

  // Удаление заметки с подтверждением
  async function handleDelete(id: string) {
    if (!confirm('Delete this note?')) return;
    try {
      await deleteNote(id);
      notes = notes.filter(n => n.id !== id);
    } catch (e) {
      alert('Delete failed');
    }
  }
</script>

<h1>My Notes</h1>

{#if loading}
  <p>Loading...</p>
{:else if error}
  <p class="error">{error}</p>
{:else}
  <div class="notes-grid">
    {#each notes as note}
      <div class="note-card">
        <!-- Клик по заголовку или содержимому ведёт на страницу заметки -->
        <a href={`/notes/${note.id}`}>
          <h3>{note.title}</h3>
          <p>{note.content.slice(0, 100)}...</p>
        </a>
        <div class="actions">
          <a href={`/notes/${note.id}/edit`}>Edit</a>
          <button on:click={() => handleDelete(note.id)}>Delete</button>
        </div>
      </div>
    {/each}
  </div>
  <a href="/notes/new" class="new-note">+ New Note</a>
{/if}

<style>
  .notes-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1.5rem;
    margin: 2rem 0;
  }
  .note-card {
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 1rem;
    background: white;
    transition: box-shadow 0.2s;
  }
  .note-card:hover {
    box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  }
  .note-card a {
    text-decoration: none;
    color: inherit;
  }
  .actions {
    margin-top: 0.5rem;
    display: flex;
    gap: 0.5rem;
  }
  .new-note {
    display: inline-block;
    background: #3b82f6;
    color: white;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    text-decoration: none;
  }
  .error {
    color: red;
  }
</style>
'@
Write-FileWithContent -Path "src/routes/+page.svelte" -Content $listPage

# src/routes/notes/[id]/+page.svelte (просмотр)
$viewPage = @'
<script lang="ts">
  // Страница отдельной заметки: содержимое, рекомендации, кнопка графа
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { getNote, getSuggestions, deleteNote } from '$lib/api/notes';
  import type { Note, Suggestion } from '$lib/api/notes';
  import { goto } from '$app/navigation';

  let note: Note | null = null;
  let suggestions: Suggestion[] = [];
  let loading = true;
  let error = '';

  const id = $page.params.id; // ID заметки из URL

  onMount(async () => {
    try {
      // Загружаем саму заметку и рекомендации параллельно (можно оптимизировать)
      note = await getNote(id);
      suggestions = await getSuggestions(id, 5);
    } catch (e) {
      error = 'Note not found';
    } finally {
      loading = false;
    }
  });

  async function handleDelete() {
    if (!confirm('Delete this note?')) return;
    await deleteNote(id);
    goto('/'); // возврат на главную
  }
</script>

{#if loading}
  <p>Loading...</p>
{:else if error}
  <p class="error">{error}</p>
{:else if note}
  <div class="note-container">
    <h1>{note.title}</h1>
    <div class="meta">
      Created: {new Date(note.created_at).toLocaleString()}
    </div>
    <div class="content">
      <!-- Содержимое заметки с сохранением пробелов и переносов -->
      {note.content}
    </div>
    <div class="actions">
      <a href={`/notes/${note.id}/edit`}>Edit</a>
      <button on:click={handleDelete}>Delete</button>
      <a href={`/graph/${note.id}`} class="graph-link">✨ Show constellation</a>
    </div>

    {#if suggestions.length}
      <h2>Similar notes</h2>
      <ul class="suggestions">
        {#each suggestions as s}
          <li>
            <a href={`/notes/${s.note_id}`}>{s.title}</a>
            <span class="score">score: {s.score.toFixed(3)}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}

<style>
  .note-container {
    max-width: 800px;
    margin: 0 auto;
  }
  .content {
    white-space: pre-wrap;   /* сохраняем переносы строк */
    margin: 1rem 0;
  }
  .actions {
    display: flex;
    gap: 1rem;
    margin: 1rem 0;
  }
  .suggestions li {
    margin-bottom: 0.5rem;
  }
  .score {
    margin-left: 1rem;
    color: #666;
    font-size: 0.9rem;
  }
  .error {
    color: red;
  }
  .graph-link {
    background: #8b5cf6;
    color: white;
    padding: 0.25rem 0.75rem;
    border-radius: 4px;
    text-decoration: none;
  }
</style>
'@
Write-FileWithContent -Path "src/routes/notes/[id]/+page.svelte" -Content $viewPage

# src/routes/notes/new/+page.svelte (создание)
$newPage = @'
<script lang="ts">
  // Страница создания новой заметки
  import { goto } from '$app/navigation';
  import { createNote } from '$lib/api/notes';

  let title = '';
  let content = '';
  let saving = false;
  let error = '';

  async function handleSubmit() {
    if (!title.trim()) {
      error = 'Title is required';
      return;
    }
    saving = true;
    error = '';
    try {
      const note = await createNote({ title, content, metadata: {} });
      // После успешного создания переходим на страницу новой заметки
      goto(`/notes/${note.id}`);
    } catch (e) {
      error = 'Failed to create note';
    } finally {
      saving = false;
    }
  }
</script>

<h1>New Note</h1>

{#if error}
  <p class="error">{error}</p>
{/if}

<form on:submit|preventDefault={handleSubmit}>
  <input type="text" placeholder="Title" bind:value={title} required />
  <textarea
    placeholder="Content (supports [[wiki links]])"
    bind:value={content}
    rows="15"
  ></textarea>
  <button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Create'}</button>
</form>

<style>
  input, textarea {
    width: 100%;
    padding: 0.5rem;
    margin-bottom: 1rem;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-family: inherit;
  }
  .error {
    color: red;
  }
</style>
'@
Write-FileWithContent -Path "src/routes/notes/new/+page.svelte" -Content $newPage

# src/routes/notes/[id]/edit/+page.svelte (редактирование)
$editPage = @'
<script lang="ts">
  // Страница редактирования существующей заметки
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { getNote, updateNote } from '$lib/api/notes';
  import type { Note } from '$lib/api/notes';

  let note: Note | null = null;
  let title = '';
  let content = '';
  let saving = false;
  let error = '';
  let loading = true;

  const id = $page.params.id;

  onMount(async () => {
    try {
      note = await getNote(id);
      title = note.title;
      content = note.content;
    } catch (e) {
      error = 'Note not found';
    } finally {
      loading = false;
    }
  });

  async function handleSubmit() {
    if (!title.trim()) {
      error = 'Title is required';
      return;
    }
    saving = true;
    error = '';
    try {
      await updateNote(id, { title, content });
      // После сохранения возвращаемся на страницу просмотра
      goto(`/notes/${id}`);
    } catch (e) {
      error = 'Failed to update note';
    } finally {
      saving = false;
    }
  }
</script>

<h1>Edit Note</h1>

{#if loading}
  <p>Loading...</p>
{:else if error}
  <p class="error">{error}</p>
{:else}
  <form on:submit|preventDefault={handleSubmit}>
    <input type="text" placeholder="Title" bind:value={title} required />
    <textarea bind:value={content} rows="15"></textarea>
    <button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Update'}</button>
  </form>
{/if}

<style>
  input, textarea {
    width: 100%;
    padding: 0.5rem;
    margin-bottom: 1rem;
    border: 1px solid #ccc;
    border-radius: 4px;
    font-family: inherit;
  }
  .error {
    color: red;
  }
</style>
'@
Write-FileWithContent -Path "src/routes/notes/[id]/edit/+page.svelte" -Content $editPage

# src/routes/graph/[id]/+page.svelte (граф)
$graphPage = @'
<script lang="ts">
  // Страница визуализации графа для заметки с заданным ID
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import GraphCanvas from '$lib/components/GraphCanvas.svelte';
  import { getGraphData } from '$lib/api/graph';

  let nodes: Array<{ id: string; title: string }> = [];
  let links: Array<{ source: string; target: string; weight: number }> = [];
  let loading = true;
  let error = '';

  const id = $page.params.id; // ID заметки из URL

  onMount(async () => {
    try {
      const data = await getGraphData(id);
      nodes = data.nodes;
      links = data.links;
    } catch (e) {
      error = 'Failed to load graph data';
      console.error(e);
    } finally {
      loading = false;
    }
  });
</script>

<h1>Созвездие заметок</h1>

{#if loading}
  <p>Загрузка графа...</p>
{:else if error}
  <p class="error">{error}</p>
{:else}
  <div class="graph-container">
    <GraphCanvas {nodes} {links} />
  </div>
{/if}

<style>
  .graph-container {
    width: 100%;
    height: 70vh;
    border: 1px solid #ccc;
    border-radius: 8px;
    overflow: hidden;
    background: #f0f2f5;
  }
  .error {
    color: red;
  }
</style>
'@
Write-FileWithContent -Path "src/routes/graph/[id]/+page.svelte" -Content $graphPage

# ========== 4. Конфигурация Vite ==========
Write-Host "Настройка vite.config.ts..." -ForegroundColor Yellow

$viteConfig = @'
// Конфигурация Vite для SvelteKit
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [sveltekit()],
  server: {
    proxy: {
      // Прокси для запросов к API бэкенда
      // Все запросы, начинающиеся с /api, перенаправляются на http://localhost:8080
      // Например, /api/notes → http://localhost:8080/notes
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // Убираем префикс /api, чтобы не было /api/notes
        rewrite: (path) => path.replace(/^\/api/, '')
      }
    }
  }
});
'@
Write-FileWithContent -Path "vite.config.ts" -Content $viteConfig

Write-Host ""
Write-Host "✅ Все файлы созданы!" -ForegroundColor Green
Write-Host ""
Write-Host "Теперь запустите фронтенд:" -ForegroundColor Cyan
Write-Host "  cd frontend" -ForegroundColor Yellow
Write-Host "  npm run dev" -ForegroundColor Yellow
Write-Host ""
Write-Host "Убедитесь, что бэкенд запущен: docker-compose up -d" -ForegroundColor Cyan