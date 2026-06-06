# Инструменты Frontend Агента

## 🎯 Основные задачи

1. Svelte 5 компоненты
2. State management (stores)
3. API интеграция
4. Performance optimization
5. Accessibility (WCAG 2.1)
6. Тестирование (Vitest, Testing Library, Playwright)

---

## 🛠️ Разработка компонентов

### 1. Svelte 5 Patterns

#### Компонент с props и state
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import type { Note } from '$lib/types';
  
  interface Props {
    note: Note;
    onSelect?: (note: Note) => void;
    editable?: boolean;
  }
  
  let { note, onSelect, editable = false }: Props = $props();
  let isHovered = $state(false);
  let isEditing = $state(false);
  let editedContent = $state(note.content);
  
  function handleClick() {
    onSelect?.(note);
  }
  
  function handleSave() {
    // Save logic
    isEditing = false;
  }
  
  function handleCancel() {
    editedContent = note.content;
    isEditing = false;
  }
</script>

<div 
  class="note-card" 
  class:hovered={isHovered}
  onclick={handleClick}
  onmouseenter={() => isHovered = true}
  onmouseleave={() => isHovered = false}
>
  {#if isEditing}
    <textarea bind:value={editedContent} />
    <div class="actions">
      <button onclick={handleSave}>Save</button>
      <button onclick={handleCancel}>Cancel</button>
    </div>
  {:else}
    <h3>{note.title}</h3>
    <p>{editedContent.slice(0, 100)}...</p>
    {#if editable}
      <button onclick={() => isEditing = true}>Edit</button>
    {/if}
    <span class="type">{note.type}</span>
  {/if}
</div>

<style>
  .note-card {
    padding: 1rem;
    border: 1px solid #e5e7eb;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .note-card:hovered {
    border-color: #3b82f6;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.2);
  }
  
  textarea {
    width: 100%;
    min-height: 100px;
    padding: 0.5rem;
    border: 1px solid #d1d5db;
    border-radius: 4px;
  }
  
  .actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }
</style>
```

#### Modal компонент
```svelte
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  interface Props {
    open: boolean;
    title: string;
    size?: 'sm' | 'md' | 'lg';
    onClose: () => void;
  }
  
  let { open, title, size = 'md', onClose }: Props = $props();
  const dispatch = createEventDispatcher<{
    close: {};
    submit: { data: any };
  }>();
  
  function handleClose() {
    onClose();
    dispatch('close');
  }
  
  function handleSubmit(data: any) {
    dispatch('submit', { data });
  }
</script>

{#if open}
  <div class="modal-overlay" onclick={handleClose} role="dialog" aria-modal="true">
    <div 
      class="modal-container" 
      class:modal-sm={size === 'sm'}
      class:modal-md={size === 'md'}
      class:modal-lg={size === 'lg'}
      onclick={(e) => e.stopPropagation()}
    >
      <div class="modal-header">
        <h2>{title}</h2>
        <button 
          class="close-btn" 
          onclick={handleClose} 
          aria-label="Закрыть"
        >
          ×
        </button>
      </div>
      <div class="modal-body">
        <slot />
      </div>
      <div class="modal-footer">
        <slot name="footer">
          <button class="btn btn-secondary" onclick={handleClose}>
            Отмена
          </button>
          <button class="btn btn-primary" onclick={() => handleSubmit({})}>
            Ок
          </button>
        </slot>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    animation: fadeIn 0.2s;
  }
  
  .modal-container {
    background: white;
    border-radius: 16px;
    width: 90%;
    max-height: 90vh;
    overflow: auto;
    animation: slideUp 0.3s;
  }
  
  .modal-sm { max-width: 400px; }
  .modal-md { max-width: 600px; }
  .modal-lg { max-width: 900px; }
  
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .modal-body {
    padding: 1.5rem;
  }
  
  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 1rem 1.5rem;
    border-top: 1px solid #e5e7eb;
  }
  
  .close-btn {
    background: none;
    border: none;
    font-size: 2rem;
    cursor: pointer;
    color: #6b7280;
    line-height: 1;
  }
  
  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  
  @keyframes slideUp {
    from { transform: translateY(20px); opacity: 0; }
    to { transform: translateY(0); opacity: 1; }
  }
</style>
```

### 2. State Management

#### Svelte Stores
```typescript
// lib/stores/note-store.svelte.ts
import { writable, derived } from 'svelte/store';
import type { Note, CreateNoteRequest } from '$lib/types';
import { notesApi } from '$lib/api/notes';

// Create store
function createNoteStore() {
  const { subscribe, update, set } = writable<Note[]>([]);
  const loading = writable(false);
  const error = writable<string | null>(null);
  
  return {
    subscribe,
    loading,
    error,
    
    async loadAll(limit?: number, offset?: number) {
      loading.set(true);
      error.set(null);
      
      try {
        const notes = await notesApi.getAll({ limit, offset });
        set(notes);
      } catch (e) {
        const msg = e instanceof Error ? e.message : 'Unknown error';
        error.set(msg);
      } finally {
        loading.set(false);
      }
    },
    
    async addNote(data: CreateNoteRequest) {
      try {
        const newNote = await notesApi.create(data);
        update(notes => [...notes, newNote]);
        return newNote;
      } catch (e) {
        const msg = e instanceof Error ? e.message : 'Unknown error';
        error.set(msg);
        throw e;
      }
    },
    
    async updateNote(id: string, data: Partial<Note>) {
      try {
        const updated = await notesApi.update(id, data as any);
        update(notes => notes.map(n => n.id === id ? updated : n));
        return updated;
      } catch (e) {
        const msg = e instanceof Error ? e.message : 'Unknown error';
        error.set(msg);
        throw e;
      }
    },
    
    async deleteNote(id: string) {
      try {
        await notesApi.delete(id);
        update(notes => notes.filter(n => n.id !== id));
      } catch (e) {
        const msg = e instanceof Error ? e.message : 'Unknown error';
        error.set(msg);
        throw e;
      }
    }
  };
}

export const noteStore = createNoteStore();
```

#### Derived stores
```typescript
// lib/stores/selectors.ts
import { derived } from 'svelte/store';
import { noteStore } from './note-store.svelte';

export const noteCount = derived(noteStore, $notes => $notes.length);

export const notesByType = derived(noteStore, $notes => {
  return {
    star: $notes.filter(n => n.type === 'star'),
    planet: $notes.filter(n => n.type === 'planet'),
    comet: $notes.filter(n => n.type === 'comet'),
  };
});

export const searchResults = derived(noteStore, $notes => {
  return (query: string) => {
    if (!query) return $notes;
    const lower = query.toLowerCase();
    return $notes.filter(n => 
      n.title.toLowerCase().includes(lower) ||
      n.content.toLowerCase().includes(lower)
    );
  };
});
```

### 3. API Integration

```typescript
// lib/api/client.ts
import ky from 'ky';

const apiClient = ky.create({
  prefixUrl: import.meta.env.VITE_API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  hooks: {
    beforeRequest: [
      (request) => {
        const token = localStorage.getItem('auth_token');
        if (token) {
          request.headers.set('Authorization', `Bearer ${token}`);
        }
      },
    ],
    afterResponse: [
      async (request, options, response) => {
        if (response.status === 401) {
          // Token refresh logic
          await refreshToken();
          return fetch(request);
        }
        return response;
      },
    ],
  },
});

export const api = {
  get: <T>(url: string, options?: RequestInit): Promise<T> =>
    apiClient.get(url, options).json(),
  
  post: <T>(url: string, data?: any, options?: RequestInit): Promise<T> =>
    apiClient.post(url, { json: data, ...options }).json(),
  
  put: <T>(url: string, data?: any, options?: RequestInit): Promise<T> =>
    apiClient.put(url, { json: data, ...options }).json(),
  
  delete: <T>(url: string, options?: RequestInit): Promise<T> =>
    apiClient.delete(url, options).json(),
};
```

---

## 🧪 Тестирование

### Unit Tests (Vitest)
```typescript
// src/lib/components/NoteCard.spec.ts
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import NoteCard from './NoteCard.svelte';
import { mockNote } from '$lib/test-utils';

describe('NoteCard', () => {
  it('renders note title and content', () => {
    render(NoteCard, { props: { note: mockNote } });
    
    expect(screen.getByText(mockNote.title)).toBeInTheDocument();
    expect(screen.getByText(mockNote.content.slice(0, 100) + '...'))
      .toBeInTheDocument();
  });
  
  it('calls onSelect when clicked', async () => {
    const onSelect = vi.fn();
    render(NoteCard, { props: { note: mockNote, onSelect } });
    
    const card = screen.getByRole('button');
    await fireEvent.click(card);
    
    expect(onSelect).toHaveBeenCalledWith(mockNote);
    expect(onSelect).toHaveBeenCalledTimes(1);
  });
  
  it('shows edit button when editable', () => {
    render(NoteCard, { props: { note: mockNote, editable: true } });
    
    expect(screen.getByText('Edit')).toBeInTheDocument();
  });
  
  it('does not show edit button when not editable', () => {
    render(NoteCard, { props: { note: mockNote, editable: false } });
    
    expect(screen.queryByText('Edit')).not.toBeInTheDocument();
  });
});
```

### Component Integration Tests
```typescript
// src/lib/components/NoteList.spec.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/svelte';
import NoteList from './NoteList.svelte';
import { noteStore } from '$lib/stores/note-store.svelte';
import * as notesApi from '$lib/api/notes';

vi.mock('$lib/api/notes');

describe('NoteList', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });
  
  it('loads and displays notes', async () => {
    const mockNotes = [mockNote1, mockNote2];
    vi.spyOn(notesApi.notesApi, 'getAll').mockResolvedValue(mockNotes);
    
    render(NoteList);
    
    await waitFor(() => {
      expect(screen.getByText(mockNotes[0].title)).toBeInTheDocument();
      expect(screen.getByText(mockNotes[1].title)).toBeInTheDocument();
    });
  });
  
  it('shows loading state', () => {
    render(NoteList);
    
    expect(screen.getByText('Загрузка...')).toBeInTheDocument();
  });
  
  it('shows error state on failure', async () => {
    vi.spyOn(notesApi.notesApi, 'getAll').mockRejectedValue(
      new Error('Network error')
    );
    
    render(NoteList);
    
    await waitFor(() => {
      expect(screen.getByText('Ошибка загрузки заметок'))
        .toBeInTheDocument();
    });
  });
});
```

### E2E Tests (Playwright)
```typescript
// tests/note-creation.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Note Creation', () => {
  test('create new note from list page', async ({ page }) => {
    // Navigate to notes page
    await page.goto('/notes');
    
    // Click create button
    await page.click('[data-testid="create-note-btn"]');
    
    // Fill form
    await page.fill('[data-testid="note-title"]', 'New Note');
    await page.fill('[data-testid="note-content"]', 'This is my new note');
    await page.selectOption('[data-testid="note-type"]', 'star');
    
    // Submit
    await page.click('[data-testid="submit-btn"]');
    
    // Verify note appears in list
    await expect(page.locator('text=New Note')).toBeVisible();
  });
  
  test('show validation errors', async ({ page }) => {
    await page.goto('/notes');
    await page.click('[data-testid="create-note-btn"]');
    
    // Submit empty form
    await page.click('[data-testid="submit-btn"]');
    
    // Verify errors
    await expect(page.locator('text=Title is required'))
      .toBeVisible();
    await expect(page.locator('text=Content is required'))
      .toBeVisible();
  });
});
```

---

## ♿ Accessibility (WCAG 2.1)

### Keyboard Navigation
```svelte
<!-- Доступные элементы -->
<button 
  onkeydown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick();
    }
  }}
  tabindex="0"
>
  Click me
</button>

<!-- Focus management -->
<div 
  bind:this={modalRef}
  onfocusout={(e) => {
    if (!modalRef?.contains(e.relatedTarget as Node)) {
      onClose();
    }
  }}
>
  <slot />
</div>
```

### ARIA Labels
```svelte
<nav aria-label="Основная навигация">
  <ul role="menubar">
    <li role="none">
      <a role="menuitem" href="/notes">Заметки</a>
    </li>
  </ul>
</nav>

<button 
  aria-expanded={isOpen}
  aria-controls="dropdown-menu"
  onclick={toggleDropdown}
>
  Меню
</button>

<div 
  id="dropdown-menu"
  role="menu"
  aria-hidden={!isOpen}
>
  <slot />
</div>
```

---

## ⚡ Performance Optimization

### Lazy Loading
```svelte
<script>
  import { onMount } from 'svelte';
  
  let HeavyComponent;
  let loaded = $state(false);
  
  onMount(async () => {
    HeavyComponent = (await import('./HeavyComponent.svelte')).default;
    loaded = true;
  });
</script>

{#if loaded}
  <svelte:component this={HeavyComponent} />
{:else}
  <div class="loading">Загрузка...</div>
{/if}
```

### Virtual Scrolling
```svelte
<script>
  let items = $state([]);
  let startIndex = $state(0);
  const pageSize = 50;
  
  $effect(() => {
    // Only render visible items
    visibleItems = items.slice(startIndex, startIndex + pageSize);
  });
  
  function handleScroll(e) {
    const { scrollTop, scrollHeight, clientHeight } = e.target;
    if (scrollTop + clientHeight > scrollHeight - 100) {
      startIndex += pageSize;
    }
  }
</script>
```

### Memoization
```svelte
<script>
  let items = $state([]);
  
  // Computed value (automatically cached)
  const sortedItems = $derived(
    [...items].sort((a, b) => a.title.localeCompare(b.title))
  );
  
  const groupedByType = $derived(
    items.reduce((acc, item) => {
      acc[item.type] = [...(acc[item.type] || []), item];
      return acc;
    }, {})
  );
</script>
```

---

## 🔧 Команды

### Development
```bash
npm run dev              # Запуск dev сервера
npm run dev -- --host    # Для доступа извне
```

### Testing
```bash
npm run test:unit        # Unit тесты (Vitest)
npm run test:e2e         # E2E тесты (Playwright)
npm run test:coverage    # С покрытием
npm run test:unit -- --run  # Запустить один раз
```

### Build
```bash
npm run build            # Production build
npm run preview          # Preview production build
```

### Linting
```bash
npm run lint             # ESLint + Svelte checker
npm run format           # Prettier formatting
```

---

## 📚 Best Practices

### Компонентная структура
```
src/
├── lib/
│   ├── components/
│   │   ├── common/
│   │   │   ├── Button/
│   │   │   ├── Modal/
│   │   │   └── Input/
│   │   ├── features/
│   │   │   ├── notes/
│   │   │   └── graph/
│   │   └── layouts/
│   ├── stores/
│   ├── api/
│   └── utils/
├── routes/
└── tests/
```

### Типизация
```typescript
// Всегда использовать TypeScript strict mode
// tsconfig.json
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "noImplicitReturns": true
  }
}
```

### Обработка ошибок
```typescript
// Error boundaries
try {
  const data = await api.get('/api/data');
} catch (error) {
  if (error instanceof APIError) {
    errorStore.set(error.message);
  } else {
    errorStore.set('Неизвестная ошибка');
  }
}
```

---

**Tools:** Этот файл + `integration-tools.md`  
**Coverage Target:** > 60%  
**Lighthouse Score:** > 90