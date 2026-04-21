<script lang="ts">
  import { getNote, type Note } from '$lib/api/notes';
  import { goto } from '$app/navigation';
  
  const { nodeId, onClose, onEdit, onDelete }: { 
    nodeId: string; 
    onClose: () => void;
    onEdit?: (id: string) => void;
    onDelete?: (id: string) => void;
  } = $props();
  
  let note = $state<Note | null>(null);
  let loading = $state(true);
  let error = $state('');

  // Load note when nodeId changes
  $effect(() => {
    const id = nodeId; // track dependency
    loadNote(id);
  });

  async function loadNote(id: string) {
    loading = true;
    error = '';
    try {
      note = await getNote(id);
    } catch {
      error = 'Failed to load note';
      note = null;
    } finally {
      loading = false;
    }
  }
  
  function formatDate(dateString: string): string {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  }
  
  function getTypeIcon(type: string | undefined): string {
    const icons: Record<string, string> = {
      star: '⭐',
      planet: '🪐',
      comet: '☄️',
      galaxy: '🌀',
      asteroid: '☁️'
    };
    return icons[type || 'star'] || '📄';
  }
</script>

<div class="side-panel" class:open={true}>
  <div class="panel-header">
    <button class="close-btn" onclick={onClose} aria-label="Close panel">
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="18" y1="6" x2="6" y2="18"/>
        <line x1="6" y1="6" x2="18" y2="18"/>
      </svg>
    </button>
    
    {#if !loading && note}
      <div class="actions">
        <button 
          class="action-btn"
          onclick={() => onEdit?.(nodeId)}
          aria-label="Edit note"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
          </svg>
        </button>
        <button 
          class="action-btn delete"
          onclick={() => onDelete?.(nodeId)}
          aria-label="Delete note"
        >
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="3 6 5 6 21 6"/>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
          </svg>
        </button>
      </div>
    {/if}
  </div>
  
  <div class="panel-content">
    {#if loading}
      <div class="loading">
        <div class="spinner"></div>
        <p>Loading...</p>
      </div>
    {:else if error}
      <div class="error">{error}</div>
    {:else if note}
      <div class="note-header">
        <span class="type-icon">{getTypeIcon(note.type)}</span>
        <h2 class="title">{note.title}</h2>
        <span class="type-badge">{note.type || 'Note'}</span>
      </div>
      
      <div class="meta">
        <span class="date">Created: {formatDate(note.created_at)}</span>
        <span class="date">Updated: {formatDate(note.updated_at)}</span>
      </div>
      
      <div class="content">
        {note.content}
      </div>
      
      {#if note.metadata?.tags?.length > 0}
        <div class="tags">
          {#each note.metadata.tags as tag}
            <span class="tag">#{tag}</span>
          {/each}
        </div>
      {/if}
      
      <div class="panel-footer">
        <button class="view-full-btn" onclick={() => note && goto(`/notes/${note.id}`)}>
          View Full Page →
        </button>
      </div>
    {/if}
  </div>
</div>

<style>
  .side-panel {
    position: fixed;
    top: 0;
    right: -400px;
    width: 400px;
    height: 100vh;
    background: white;
    box-shadow: -4px 0 20px rgba(0, 0, 0, 0.1);
    z-index: 200;
    transition: right 0.3s ease;
    display: flex;
    flex-direction: column;
  }

  .side-panel.open {
    right: 0;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid #e2e8f0;
  }

  .close-btn {
    padding: 8px;
    border: none;
    background: #f1f5f9;
    border-radius: 8px;
    cursor: pointer;
    color: #64748b;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: #e2e8f0;
    color: #334155;
  }

  .actions {
    display: flex;
    gap: 8px;
  }

  .action-btn {
    padding: 8px 12px;
    border: none;
    background: #f1f5f9;
    border-radius: 8px;
    cursor: pointer;
    color: #64748b;
    transition: all 0.2s;
  }

  .action-btn:hover {
    background: #e2e8f0;
    color: #334155;
  }

  .action-btn.delete:hover {
    background: #fee2e2;
    color: #ef4444;
  }

  .panel-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
  }

  .loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 40px;
    color: #94a3b8;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid #e2e8f0;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 12px;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error {
    padding: 20px;
    background: #fee2e2;
    color: #ef4444;
    border-radius: 8px;
    text-align: center;
  }

  .note-header {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-bottom: 20px;
  }

  .type-icon {
    font-size: 32px;
  }

  .title {
    font-size: 24px;
    font-weight: 600;
    color: #1e293b;
    line-height: 1.3;
    margin: 0;
  }

  .type-badge {
    display: inline-block;
    padding: 4px 12px;
    background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
    color: white;
    font-size: 12px;
    font-weight: 500;
    text-transform: uppercase;
    border-radius: 20px;
    width: fit-content;
  }

  .meta {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid #e2e8f0;
  }

  .date {
    font-size: 13px;
    color: #94a3b8;
  }

  .content {
    font-size: 15px;
    line-height: 1.7;
    color: #475569;
    white-space: pre-wrap;
    margin-bottom: 24px;
  }

  .tags {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 24px;
  }

  .tag {
    padding: 4px 10px;
    background: #f1f5f9;
    color: #64748b;
    font-size: 12px;
    border-radius: 16px;
  }

  .panel-footer {
    padding-top: 20px;
    border-top: 1px solid #e2e8f0;
  }

  .view-full-btn {
    width: 100%;
    padding: 12px;
    border: 1px solid #e2e8f0;
    background: white;
    border-radius: 8px;
    color: #3b82f6;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .view-full-btn:hover {
    background: #f8fafc;
    border-color: #3b82f6;
  }

  @media (max-width: 768px) {
    .side-panel {
      width: 100%;
      right: -100%;
    }
  }
</style>
