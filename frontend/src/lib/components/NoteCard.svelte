<script lang="ts">
  import type { Note } from '$lib/api/notes';
  import { goto } from '$app/navigation';
  import { formatDate } from '$lib/utils/date';

  const { 
    note, 
    highlightQuery = '',
    onClick
  }: { 
    note: Note; 
    highlightQuery?: string;
    onClick?: (note: Note) => void;
  } = $props();

  function highlightText(text: string, query: string): string {
    if (!query.trim()) return text;
    const escapedQuery = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const regex = new RegExp(`(${escapedQuery})`, 'gi');
    return text.replace(regex, '<mark>$1</mark>');
  }

  function handleClick() {
    if (onClick) {
      onClick(note);
    } else {
      goto(`/notes/${note.id}`);
    }
  }

  function truncateText(text: string, maxLength: number): string {
    if (text.length <= maxLength) return text;
    return text.slice(0, maxLength) + '...';
  }
</script>

<div 
    class="note-card" 
    onclick={handleClick}
    onkeydown={(e) => e.key === 'Enter' || e.key === ' ' ? handleClick() : null}
    role="button" 
    tabindex="0"
    aria-label={`Open note: ${note.title}`}
  >
  <div class="note-header">
    <h3 class="note-title">{@html highlightQuery ? highlightText(note.title, highlightQuery) : note.title}</h3>
    <div class="note-meta">
      {#if note.type}
        <span class="type-badge" data-testid="note-type">{note.type}</span>
      {/if}
      <div class="note-date">{formatDate(note.created_at)}</div>
    </div>
  </div>

  <div class="note-content">
    {@html highlightQuery ? highlightText(truncateText(note.content, 200), highlightQuery) : truncateText(note.content, 200)}
  </div>
  
  {#if note.metadata && Object.keys(note.metadata).length > 0}
    <div class="note-metadata">
      {#each Object.entries(note.metadata) as [key, value]}
        <span class="metadata-item">
          {key}: {typeof value === 'string' ? value : JSON.stringify(value)}
        </span>
      {/each}
    </div>
  {/if}
</div>

<style>
  .note-card {
    background: white;
    border: 1px solid var(--color-border, #e2e8f0);
    border-radius: 0.75rem;
    padding: 1.5rem;
    cursor: pointer;
    transition: all 0.2s ease;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  }

  .note-card:hover {
    border-color: var(--color-primary, #3b82f6);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    transform: translateY(-2px);
  }

  .note-card:focus {
    outline: none;
    border-color: var(--color-primary, #3b82f6);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .note-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 1rem;
    gap: 1rem;
  }

  .note-title {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--color-text, #1f2937);
    line-height: 1.4;
    flex: 1;
  }

  .note-date {
    font-size: 0.875rem;
    color: var(--color-text-secondary, #6b7280);
    white-space: nowrap;
  }

  .note-meta {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    gap: 0.25rem;
  }

  .type-badge {
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    background: var(--color-background-secondary, #f3f4f6);
    color: var(--color-text-secondary, #6b7280);
    text-transform: capitalize;
  }

  .note-content {
    color: var(--color-text-secondary, #6b7280);
    line-height: 1.6;
    margin-bottom: 1rem;
  }

  .note-metadata {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-top: 1rem;
  }

  .metadata-item {
    background: var(--color-background-secondary, #f3f4f6);
    color: var(--color-text-secondary, #6b7280);
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
  }

  :global(mark) {
    background: linear-gradient(120deg, #fef08a 0%, #fde047 100%);
    color: #1f2937;
    padding: 0.1em 0.2em;
    border-radius: 0.2em;
    font-weight: 600;
  }
</style>
