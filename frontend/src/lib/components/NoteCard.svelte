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

  // Get type color for the left border indicator
  function getTypeColor(type?: string): string {
    if (!type) return '#94a3b8'; // asteroid gray default
    const colors: Record<string, string> = {
      'star': '#fbbf24',      // yellow
      'planet': '#60a5fa',    // blue
      'comet': '#f472b6',     // pink
      'galaxy': '#a78bfa',    // purple
      'asteroid': '#94a3b8'   // gray
    };
    return colors[type.toLowerCase()] || colors['asteroid'];
  }
</script>

<div 
    class="note-card" 
    style="--type-color: {getTypeColor(note.type)}"
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
    background:
      linear-gradient(135deg, rgba(255, 255, 255, 0.05) 0%, rgba(255, 255, 255, 0.02) 100%),
      rgba(10, 10, 26, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.08);
    border-left: 3px solid var(--type-color, #94a3b8);
    border-radius: 0.75rem;
    padding: 1.5rem;
    cursor: pointer;
    transition: all 0.3s ease;
    box-shadow:
      0 1px 3px rgba(0, 0, 0, 0.3),
      inset 0 0 20px rgba(255, 255, 255, 0.02);
    position: relative;
    overflow: hidden;
  }

  /* Nebula glow effect overlay */
  .note-card::before {
    content: '';
    position: absolute;
    top: -50%;
    left: -50%;
    width: 200%;
    height: 200%;
    background: radial-gradient(
      circle at var(--type-color, #94a3b8) 50% 50%,
      rgba(255, 255, 255, 0.03) 0%,
      transparent 50%
    );
    opacity: 0;
    transition: opacity 0.3s ease;
    pointer-events: none;
  }

  .note-card:hover {
    border-color: rgba(255, 204, 0, 0.3);
    border-left-color: var(--type-color, #94a3b8);
    box-shadow:
      0 8px 25px rgba(0, 0, 0, 0.4),
      0 0 30px rgba(255, 204, 0, 0.1),
      inset 0 0 30px rgba(255, 255, 255, 0.03);
    transform: translateY(-4px);
  }

  .note-card:hover::before {
    opacity: 1;
  }

  .note-card:focus {
    outline: none;
    border-color: rgba(255, 204, 0, 0.5);
    box-shadow:
      0 0 0 3px rgba(255, 204, 0, 0.1),
      0 8px 25px rgba(0, 0, 0, 0.4);
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
    color: var(--color-text-dark, #e0e0e0);
    line-height: 1.4;
    flex: 1;
    text-shadow: 0 0 10px rgba(255, 255, 255, 0.1);
  }

  .note-date {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.5);
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
    background: rgba(255, 255, 255, 0.1);
    color: var(--type-color, #94a3b8);
    text-transform: capitalize;
    border: 1px solid rgba(255, 255, 255, 0.1);
    text-shadow: 0 0 5px var(--type-color, rgba(255, 255, 255, 0.3));
  }

  .note-content {
    color: rgba(255, 255, 255, 0.6);
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
    background: rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.5);
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    border: 1px solid rgba(255, 255, 255, 0.08);
  }

  :global(mark) {
    background: linear-gradient(120deg, rgba(254, 240, 138, 0.3) 0%, rgba(253, 224, 71, 0.3) 100%);
    color: #ffcc00;
    padding: 0.1em 0.2em;
    border-radius: 0.2em;
    font-weight: 600;
    box-shadow: 0 0 5px rgba(255, 204, 0, 0.3);
  }
</style>
