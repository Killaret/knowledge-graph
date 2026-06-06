# knowledge-graph-frontend-svelte

**Version:** 1.0  
**Purpose:** Frontend development with Svelte 5, UI/UX, components  
**Status:** Active  
**Priority:** 🟢 High

---

## Overview

`knowledge-graph-frontend-svelte` specializes in frontend development using Svelte 5, TypeScript, and modern web technologies.

**Key Areas:**
- Svelte 5 components
- State management (stores)
- API integration
- Performance optimization
- Accessibility (WCAG 2.1)
- Component testing
- E2E testing

---

## Component Patterns

### 1. Svelte 5 Components

#### Basic Component
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import type { Note } from '$lib/types';
  
  interface Props {
    note: Note;
    onSelect?: (note: Note) => void;
  }
  
  let { note, onSelect }: Props = $props();
  let isHovered = $state(false);
  
  function handleClick() {
    onSelect?.(note);
  }
</script>

<div 
  class="note-card" 
  class:hovered={isHovered}
  onclick={handleClick}
  onmouseenter={() => isHovered = true}
  onmouseleave={() => isHovered = false}
>
  <h3>{note.title}</h3>
  <p>{note.content.slice(0, 100)}...</p>
  <span class="type">{note.type}</span>
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
</style>
```

#### Modal Component
```svelte
<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  
  interface Props {
    open: boolean;
    title: string;
    onClose: () => void;
  }
  
  let { open, title, onClose }: Props = $props();
  const dispatch = createEventDispatcher();
  
  function handleClose() {
    onClose();
    dispatch('close');
  }
</script>

{#if open}
  <div class="modal-overlay" onclick={handleClose}>
    <div class="modal-container" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2>{title}</h2>
        <button class="close-btn" onclick={handleClose} aria-label="Close">×</button>
      </div>
      <div class="modal-body">
        <slot />
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
  }
  
  .modal-container {
    background: white;
    border-radius: 16px;
    max-width: 600px;
    width: 90%;
    max-height: 90vh;
    overflow: auto;
  }
  
  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }
  
  .close-btn {
    background: none;
    border: none;
    font-size: 2rem;
    cursor: pointer;
    color: #6b7280;
  }
</style>
```

### 2. State Management

#### Stores
```typescript
// lib/stores/note-store.svelte.ts
import { writable } from 'svelte/store';
import type { Note } from '$lib/types';
import { notesApi } from '$lib/api/notes';

export const notes = $state<Note[]>([]);
export const loading = $state(false);
export const error = $state<string | null>(null);

export async function loadNotes() {
  loading = true;
  error = null;
  
  try {
    notes = await notesApi.getAll();
  } catch (e) {
    error = e instanceof Error ? e.message : 'Unknown error';
  } finally {
    loading = false;
  }
}

export async function addNote(noteData: CreateNoteRequest) {
  const newNote = await notesApi.create(noteData);
  notes = [...notes, newNote];
  return newNote;
}
```

### 3. API Integration

```typescript
// lib/api/notes.ts
import { api } from './client';
import type { Note, CreateNoteRequest } from '$lib/types';

export const notesApi = {
  getAll: async (params?: { limit?: number }) => {
    const response = await api.get<Note[]>('/api/v1/notes', { 
      searchParams: params 
    });
    return response;
  },
  
  create: async (data: CreateNoteRequest) => {
    const response = await api.post<Note>('/api/v1/notes', data);
    return response;
  },
};
```

---

## Testing

### Component Tests
```typescript
// src/lib/components/NoteCard.spec.ts
import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import NoteCard from './NoteCard.svelte';
import { mockNote } from '$lib/test-utils';

describe('NoteCard', () => {
  it('renders note title and content', () => {
    render(NoteCard, { props: { note: mockNote } });
    
    expect(screen.getByText(mockNote.title)).toBeInTheDocument();
    expect(screen.getByText(mockNote.content.slice(0, 100) + '...')).toBeInTheDocument();
  });
  
  it('calls onSelect when clicked', async () => {
    const onSelect = vi.fn();
    render(NoteCard, { props: { note: mockNote, onSelect } });
    
    const card = screen.getByRole('button');
    await fireEvent.click(card);
    
    expect(onSelect).toHaveBeenCalledWith(mockNote);
  });
});
```

### E2E Tests
```typescript
// tests/note-creation.spec.ts
import { test, expect } from '@playwright/test';

test('create new note', async ({ page }) => {
  await page.goto('/notes');
  
  await page.click('[data-testid="create-note-btn"]');
  await page.fill('[data-testid="note-title"]', 'New Note');
  await page.fill('[data-testid="note-content"]', 'Note content');
  await page.click('[data-testid="submit-btn"]');
  
  await expect(page.locator('text=New Note')).toBeVisible();
});
```

---

## Performance Optimization

### Lazy Loading
```svelte
<script>
  import { lazy } from 'svelte';
  
  const HeavyComponent = lazy(() => import('./HeavyComponent.svelte'));
</script>

{#if showHeavy}
  <HeavyComponent />
{/if}
```

### Memoization
```svelte
<script>
  let items = $state([]);
  
  $effect(() => {
    const sorted = [...items].sort((a, b) => a.title.localeCompare(b.title));
    // Use sorted for rendering
  });
</script>
```

### Virtual Scrolling
```svelte
<!-- For large lists -->
<script>
  let visibleItems = $state([]);
  let startIndex = $state(0);
  
  $effect(() => {
    visibleItems = items.slice(startIndex, startIndex + 50);
  });
</script>
```

---

## Accessibility

### WCAG 2.1 Compliance
```svelte
<!-- Semantic HTML -->
<nav aria-label="Main navigation">
  <ul role="menubar">
    <li role="none">
      <a role="menuitem" href="/notes">Notes</a>
    </li>
  </ul>
</nav>

<!-- Keyboard navigation -->
<button 
  onkeydown={(e) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleClick();
    }
  }}
  tabindex="0"
>
  Submit
</button>

<!-- Focus management -->
<div 
  bind:this={modalRef}
  onfocusout={(e) => {
    if (!modalRef.contains(e.relatedTarget)) {
      onClose();
    }
  }}
>
  <slot />
</div>
```

---

## Best Practices

### Component Structure
```
components/
├── common/
│   ├── Button/
│   │   ├── Button.svelte
│   │   ├── Button.spec.ts
│   │   └── index.ts
│   ├── Modal/
│   └── Input/
├── features/
│   ├── notes/
│   │   ├── NoteCard.svelte
│   │   ├── NoteList.svelte
│   │   └── NoteEditor.svelte
│   └── graph/
└── layouts/
    ├── Header.svelte
    └── Sidebar.svelte
```

### Type Safety
```typescript
// Always use TypeScript strict mode
// tsconfig.json
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true
  }
}
```

### Error Boundaries
```svelte
<script>
  let { fallback } = $props();
  let error = $state(null);
  
  function handleError(err: Error) {
    error = err;
  }
</script>

{#if error}
  <slot name="fallback" {error} />
{:else}
  <slot on:error={handleError} />
{/if}
```

---

## Commands

### Development
```bash
npm run dev
npm run dev -- --host  # For remote access
```

### Testing
```bash
# Unit tests
npm run test:unit

# E2E tests
npm run test:e2e

# Coverage
npm run test:coverage
```

### Build
```bash
npm run build
npm run preview  # Preview production build
```

### Linting
```bash
npm run lint
npm run format
```

---

**Tools:** Component testing tools from `performance-tools.md`  
**Coverage Target:** > 60%  
**Lighthouse Score:** > 90