# Cursor Rule: knowledge-graph-frontend-svelte

SvelteKit + Svelte 5 + TypeScript strict. Adapter: `@sveltejs/adapter-node`.
Source root: `frontend/src/`.

## Manual Found → Automated Covered

If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue. Choose the test level by severity and scope:

- **unit** — pure logic, validators, or utilities (e.g. `errorMessage.ts`, email validation).
- **integration** — handlers, repositories, or routes (e.g. `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
- **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g. public graph, achievements, SSE fallback).

The test should fail before the fix (where safe) and pass after the fix. If the defect depends on manual data or config setup, fix the seed or config script — not only the instructions.

---

## Svelte 5 Runes — Only Allowed Reactive Primitives

| Rune | Purpose | Example |
|------|---------|---------|
| `$state` | Mutable reactive variable | `let count = $state(0)` |
| `$derived` | Computed from other state | `let doubled = $derived(count * 2)` |
| `$effect` | Side effects on state change | `$effect(() => { console.log(count) })` |
| `$props` | Typed component props | `const { title, onClose } = $props()` |
| `$bindable` | Two-way bound prop | `let value = $bindable('')` |

```svelte
<!-- ✅ Svelte 5 — correct -->
<script lang="ts">
  let count = $state(0);
  let doubled = $derived(count * 2);

  $effect(() => {
    document.title = `Count: ${count}`;
  });
</script>

<!-- ❌ Svelte 4 store — NEVER use in new code -->
<!-- import { writable } from 'svelte/store'; -->
<!-- const count = writable(0); -->
```

---

## Cross-Component State (Module-Level $state)

For shared auth/graph state, use a `.svelte.ts` module file that exports
getter functions over a module-level `$state` object.

```typescript
// frontend/src/shared/stores/auth.svelte.ts  (real file)
const authState = $state({
  currentUser: null as User | null,
  accessToken: null as string | null,
  isLoading: false,
  error: null as string | null,
});

export function currentUser(): User | null { return authState.currentUser; }
export function isLoading(): boolean { return authState.isLoading; }
export async function login(login: string, password: string): Promise<boolean> {
  authState.isLoading = true;
  // ...
}
```

Components import and call these functions directly — no store subscriptions.

---

## Atomic Design Component Hierarchy

```
frontend/src/components/
  atoms/        Button.svelte, Input.svelte, Badge.svelte, Icon.svelte
  molecules/    SearchBar.svelte, NoteCard.svelte, TagSelector.svelte, Modal.svelte, ToastNotification.svelte
  organisms/    GraphCanvas.svelte, NoteSidePanel.svelte, Sidebar.svelte, NoteEditor.svelte
frontend/src/routes/
  +page.svelte  (route pages / templates)
```

**Rule:** Atoms must have zero business logic. Organisms may call API functions
but must not contain domain computation.

---

## Component File Structure

```svelte
<!-- frontend/src/components/molecules/NoteCard.svelte -->
<script lang="ts">
  import type { Note } from '$shared/types';

  // Props — always typed, never untyped object
  const {
    note,
    onDelete,
    compact = false
  }: {
    note: Note;
    onDelete: (id: string) => void;
    compact?: boolean;
  } = $props();

  // Local state
  let isDeleting = $state(false);

  // Derived
  let truncatedContent = $derived(
    note.content.length > 120 ? note.content.slice(0, 120) + '…' : note.content
  );

  async function handleDelete() {
    isDeleting = true;
    await onDelete(note.id);
  }
</script>

<article class="note-card" class:compact>
  <h3>{note.title}</h3>
  <p>{truncatedContent}</p>
  <button onclick={handleDelete} disabled={isDeleting}>Delete</button>
</article>
```

---

## API Client Patterns (ky library)

```typescript
// frontend/src/shared/api/notes.ts
import { api } from './client';
import { getAccessToken } from '$shared/stores/auth.svelte';
import type { Note, CreateNoteRequest } from '$shared/types';

export async function createNote(data: CreateNoteRequest): Promise<Note> {
  return api.post('notes', { json: data }).json<Note>();
}

export async function getNotes(limit = 20, offset = 0): Promise<Note[]> {
  const response = await api.get('notes', { searchParams: { limit, offset } }).json<{ notes: Note[]; total: number; limit: number; offset: number }>();
  return response.notes;
}
```

---

## D3-Force Graph Patterns

The graph canvas (`frontend/src/components/organisms/GraphCanvas.svelte`) delegates simulation logic to `frontend/src/features/graph-interaction/` and graph rendering internals under `frontend/src/components/organisms/GraphCanvas/` (renderer, simulation modules).

```typescript
// Key pattern: separate simulation state from Svelte state
import { startSimulation, draw, handleZoom } from './GraphCanvas';

let canvas: HTMLCanvasElement;
let simState: SimulationState;

$effect(() => {
  // Re-run simulation when nodes/links props change
  if (nodes.length > 0) {
    simState = startSimulation(canvas, nodes, links);
  }
  return () => clearSimulation(simState);  // cleanup
});
```

Canvas is always preferred over SVG for 200+ nodes (performance).
D3-force lives entirely in `.ts` modules — never inline in `.svelte`.

---

## Routes

```
frontend/src/routes/
  +page.svelte              ← splash / landing
  +layout.svelte            ← global layout, auth guard
  auth/login/+page.svelte
  auth/register/+page.svelte
  auth/yandex/callback/+page.svelte
  graph/+page.svelte        ← 2D D3-force graph
  graph/3d/+page.svelte     ← Three.js 3D graph
  graph/[id]/+page.svelte   ← single note graph
  notes/[id]/+page.svelte
  notes/new/+page.svelte
  profile/+page.svelte
  search/+page.svelte
```

SvelteKit `+page.ts` loaders call the backend. SSR hooks in `hooks.server.ts`
use `VITE_API_TARGET` (direct to `backend:8080`) not the public nginx URL.

---

## Testing with @testing-library/svelte

```typescript
// frontend/src/components/molecules/NoteCard.spec.ts
import { render, screen, fireEvent } from '@testing-library/svelte';
import { vi, describe, it, expect } from 'vitest';
import NoteCard from './NoteCard.svelte';

describe('NoteCard', () => {
  it('renders note title', () => {
    render(NoteCard, {
      props: {
        note: { id: '1', title: 'Test Note', content: 'Body' },
        onDelete: vi.fn()
      }
    });
    expect(screen.getByText('Test Note')).toBeInTheDocument();
  });

  it('calls onDelete when delete button clicked', async () => {
    const onDelete = vi.fn();
    render(NoteCard, { props: { note: { id: '42', title: 'X', content: '' }, onDelete } });
    await fireEvent.click(screen.getByRole('button', { name: /delete/i }));
    expect(onDelete).toHaveBeenCalledWith('42');
  });
});
```

---

## TypeScript Rules

- `strict: true` in `tsconfig.json` — no implicit `any`.
- All shared API response types live in `frontend/src/shared/types/`.
- Import paths use aliases (`$shared/*`, `$components/*`, `$features/*`), avoid deep relative `../../`.
- Never use `as unknown as T` to silence type errors — fix the type.

---

## Anti-Patterns

```typescript
// ❌ Svelte 4 store in new code
import { writable } from 'svelte/store';
const notes = writable<Note[]>([]);

// ❌ Business logic in component script
// (e.g., computing cosine similarity, parsing JWT)

// ❌ Direct fetch() instead of ky API client
fetch('/api/v1/notes').then(r => r.json())  // no auth header, no error handling

// ❌ $effect without cleanup for event listeners
$effect(() => {
  window.addEventListener('resize', handler);
  // missing: return () => window.removeEventListener('resize', handler);
});

// ❌ Mutating props directly
const { note } = $props();
note.title = 'changed';  // props are read-only

// ❌ Non-$props access pattern (Svelte 4 style)
export let title: string;  // old syntax — use $props() instead
```
