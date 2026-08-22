---
name: Frontend Svelte 5 Rules
alwaysApply: false
globs: ["frontend/**/*.{svelte,ts,js}"]
description: Svelte 5 runes, TypeScript strict, atomic design, SvelteKit patterns, D3/Three.js
---

# Frontend Svelte 5 Rules

## Manual Found → Automated Covered

If a defect or questionable behavior is discovered during manual testing, create at least one regression test before closing the issue. Choose the test level by severity and scope:

- **unit** — pure logic, validators, or utilities (e.g. `errorMessage.ts`, email validation).
- **integration** — handlers, repositories, or routes (e.g. `PUT /users/me`, `DELETE /users/me` in `router_test.go`).
- **E2E / Playwright** — user-facing scenarios spanning frontend, backend, and data (e.g. public graph, achievements, SSE fallback).

The test should fail before the fix (where safe) and pass after the fix. If the defect depends on manual data or config setup, fix the seed or config script — not only the instructions.

## Svelte 5 Runes — MANDATORY

This project uses Svelte 5 runes syntax exclusively. Never use Svelte 4 syntax.

### $props() — Component Props

```svelte
<script lang="ts">
  interface Props {
    title: string;
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
    disabled?: boolean;
    onClick?: (e: MouseEvent) => void;
    children?: import('svelte').Snippet;
    'data-testid'?: string;
  }

  const {
    title,
    variant = 'primary',
    disabled = false,
    onClick,
    children,
    ...restProps
  }: Props = $props();
</script>
```

### $state() — Reactive State

```svelte
<script lang="ts">
  // Simple state
  let count = $state(0);
  let notes = $state<Note[]>([]);

  // Object state (reactive at property level)
  const authState = $state({
    currentUser: null as User | null,
    accessToken: null as string | null,
    isLoading: false,
    error: null as string | null
  });
</script>
```

### $derived() — Computed Values

```svelte
<script lang="ts">
  let notes = $state<Note[]>([]);
  const noteCount = $derived(notes.length);
  const hasNotes = $derived(notes.length > 0);
  const sortedNotes = $derived(
    [...notes].sort((a, b) => b.updatedAt.getTime() - a.updatedAt.getTime())
  );
</script>
```

### $effect() — Side Effects

```svelte
<script lang="ts">
  let searchQuery = $state('');

  $effect(() => {
    // Runs when searchQuery changes
    const timeout = setTimeout(() => {
      fetchResults(searchQuery);
    }, 300);
    return () => clearTimeout(timeout); // cleanup
  });
</script>
```

## Atomic Design Pattern

```
frontend/src/components/
├── atoms/        # Button, Input, Badge, Icon (single-purpose)
├── molecules/    # SearchBar, NoteCard, TagSelector (composed atoms)
├── organisms/    # Sidebar, GraphCanvas, NoteEditor (complex sections)
frontend/src/features/
├── graph-interaction/  # Drag-and-drop, hotkeys, zoom-pan
└── graph-forms/        # Note form, link form
frontend/src/shared/
├── api/          # ky-based API clients
├── stores/       # runes-based state modules
├── services/     # business logic / side effects
├── utils/        # helpers
├── types/        # TypeScript types
├── mocks/        # SvelteKit mocks for Vitest
└── lib/graph/    # graph rendering helpers (legacy FSD remnant)
```

### Component Example (Atom)

```svelte
<!-- Button.svelte -->
<script lang="ts">
  interface Props {
    variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
    disabled?: boolean;
    onClick?: (e: MouseEvent) => void;
    children?: import('svelte').Snippet;
  }

  const { variant = 'primary', disabled = false, onClick, children }: Props = $props();
</script>

<button
  class="button {variant}"
  class:disabled
  onclick={(e) => !disabled && onClick?.(e)}
  {disabled}
>
  {@render children?.()}
</button>
```

## Store Pattern (Svelte 5 Runes)

```typescript
// auth.svelte.ts — runes-based store
import { browser } from '$app/environment';

const authState = $state({
  currentUser: null as User | null,
  accessToken: null as string | null,
  isInitialized: false,
});

// Export reactive getters
export function currentUser(): User | null { return authState.currentUser; }
export function isAuthenticated(): boolean { return authState.currentUser !== null; }

// Export actions
export async function login(email: string, password: string): Promise<void> {
  const tokens = await authApi.login({ email, password });
  authState.accessToken = tokens.access_token;
  authState.currentUser = await usersApi.getProfile();
}
```

## TypeScript Strict Mode

- All files use `lang="ts"` in `<script>` blocks
- Explicit types for props interfaces
- No `any` — use `unknown` and narrow
- API response types defined in `$shared/types/`

```typescript
// frontend/src/shared/types/note.ts
export interface Note {
  id: string;
  title: string;
  content: string;
  type: string;
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

// frontend/src/shared/types/user.ts
export interface User {
  id: string;
  email: string;
  username: string;
}
```

## Anti-Patterns

```svelte
<!-- ❌ Bad — Svelte 4 syntax -->
<script>
  export let title;
  let count = 0;
  $: doubled = count * 2;
</script>

<!-- ✅ Good — Svelte 5 runes -->
<script lang="ts">
  const { title }: Props = $props();
  let count = $state(0);
  const doubled = $derived(count * 2);
</script>
```

```svelte
<!-- ❌ Bad — on:click directive (Svelte 4) -->
<button on:click={handler}>Click</button>

<!-- ✅ Good — onclick attribute (Svelte 5) -->
<button onclick={handler}>Click</button>
```

```svelte
<!-- ❌ Bad — slot (Svelte 4) -->
<slot />

<!-- ✅ Good — Snippet (Svelte 5) -->
{@render children?.()}
```

```typescript
// ❌ Bad — writable store (Svelte 4)
import { writable } from 'svelte/store';
const count = writable(0);

// ✅ Good — $state rune (Svelte 5)
let count = $state(0);
```

## API Client Pattern

```typescript
// frontend/src/shared/api/notes.ts
import { api } from './client';

export async function getNotes(limit = 100, offset = 0): Promise<NotesResponse> {
  const response = await api.get('notes', { searchParams: { limit, offset } }).json<{ notes: Note[]; total: number; limit: number; offset: number }>();
  return response.notes;
}

export async function createNote(data: CreateNoteRequest): Promise<Note> {
  return api.post('notes', { json: data }).json<Note>();
}
```

## Vitest Test Pattern

```typescript
import { render, screen, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, vi } from 'vitest';
import Button from './Button.svelte';

describe('Button', () => {
  it('renders with text', () => {
    render(Button, { props: { children: /* snippet */ } });
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  it('calls onClick when clicked', async () => {
    const onClick = vi.fn();
    render(Button, { props: { onClick } });
    await fireEvent.click(screen.getByRole('button'));
    expect(onClick).toHaveBeenCalledOnce();
  });

  it('does not fire when disabled', async () => {
    const onClick = vi.fn();
    render(Button, { props: { onClick, disabled: true } });
    await fireEvent.click(screen.getByRole('button'));
    expect(onClick).not.toHaveBeenCalled();
  });
});
```

## SvelteKit Routing

```
frontend/src/routes/
├── +layout.svelte       # Root layout (auth check, sidebar)
├── +page.svelte         # Home / dashboard
├── graph/+page.svelte   # 2D graph view
├── graph/3d/+page.svelte # 3D graph view (Three.js)
├── notes/[id]/+page.svelte  # Note detail
├── auth/login/+page.svelte  # Login page
└── search/+page.svelte  # Search results
```
