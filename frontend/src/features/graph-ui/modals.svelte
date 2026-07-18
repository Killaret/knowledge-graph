<script lang="ts">
  import type { NoteFormState } from '$features/graph-forms/note-form';
  import type { LinkFormState } from '$features/graph-forms/link-form';

  const {
    activeForm,
    noteFormState,
    linkFormState,
    onSave,
    onCancel
  }: {
    activeForm: 'note' | 'link' | null;
    noteFormState: NoteFormState;
    linkFormState: LinkFormState;
    onSave: (form: 'note' | 'link') => void;
    onCancel: (form: 'note' | 'link') => void;
  } = $props();
</script>

{#if activeForm === 'note'}
  <div
    class="note-form"
    data-testid="ghost-note-form"
    style="position: absolute; left: {noteFormState.noteFormPosition.x}px; top: {noteFormState.noteFormPosition.y}px; background: rgba(10, 26, 58, 0.98); border: 1px solid rgba(138, 43, 226, 0.6); border-radius: 12px; padding: 20px; min-width: 320px; box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6); z-index: 100; backdrop-filter: blur(12px);"
  >
    <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;">
      <h3 style="margin: 0; color: #a78bfa; font-size: 16px; font-weight: 600;">Create New Note</h3>
      <button
        data-testid="ghost-note-close"
        onclick={() => onCancel('note')}
        style="background: none; border: none; color: rgba(255,255,255,0.6); font-size: 20px; cursor: pointer; padding: 4px 8px; border-radius: 4px; transition: all 0.2s;"
        aria-label="Close"
      >
        ×
      </button>
    </div>
    <input
      data-testid="ghost-note-title"
      type="text"
      placeholder="Title"
      bind:value={noteFormState.newNoteTitle}
      style="width: 100%; padding: 12px; margin-bottom: 12px; border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; background: rgba(0,0,0,0.4); color: white; box-sizing: border-box; font-size: 14px; transition: border-color 0.2s;"
      onkeydown={(e) => e.key === 'Enter' && onSave('note')}
    />
    <textarea
      data-testid="ghost-note-content"
      placeholder="Content (optional)"
      bind:value={noteFormState.newNoteContent}
      style="width: 100%; padding: 12px; margin-bottom: 12px; border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; background: rgba(0,0,0,0.4); color: white; min-height: 100px; box-sizing: border-box; font-size: 14px; resize: vertical; transition: border-color 0.2s;"
    ></textarea>
    <select
      data-testid="ghost-note-type"
      bind:value={noteFormState.newNoteType}
      style="width: 100%; padding: 12px; margin-bottom: 16px; border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; background: rgba(0,0,0,0.4); color: white; font-size: 14px; cursor: pointer; transition: border-color 0.2s;"
    >
      <option value="star">⭐ Star</option>
      <option value="planet">🪐 Planet</option>
      <option value="comet">☄️ Comet</option>
      <option value="galaxy">🌀 Galaxy</option>
      <option value="asteroid">🌑 Asteroid</option>
    </select>
    <div style="display: flex; gap: 12px; justify-content: flex-end;">
      <button
        data-testid="ghost-note-cancel"
        onclick={() => onCancel('note')}
        style="padding: 10px 20px; border: 1px solid rgba(255,255,255,0.3); border-radius: 8px; background: transparent; color: white; cursor: pointer; font-size: 14px; transition: all 0.2s;"
      >
        Cancel
      </button>
      <button
        data-testid="ghost-note-create"
        onclick={() => onSave('note')}
        style="padding: 10px 20px; border: none; border-radius: 8px; background: linear-gradient(135deg, #8b5cf6, #6366f1); color: white; cursor: pointer; font-size: 14px; font-weight: 500; transition: all 0.2s; box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);"
      >
        Create
      </button>
    </div>
  </div>
{/if}

{#if activeForm === 'link'}
  <div
    class="link-form"
    data-testid="link-form"
    style="position: absolute; left: {linkFormState.linkFormPosition.x}px; top: {linkFormState.linkFormPosition.y}px; background: rgba(10, 26, 58, 0.98); border: 1px solid rgba(255, 204, 0, 0.6); border-radius: 12px; padding: 20px; min-width: 300px; box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6); z-index: 100; backdrop-filter: blur(12px);"
  >
    <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;">
      <h3 style="margin: 0; color: #fbbf24; font-size: 16px; font-weight: 600;">Create Link</h3>
      <button
        onclick={() => onCancel('link')}
        style="background: none; border: none; color: rgba(255,255,255,0.6); font-size: 20px; cursor: pointer; padding: 4px 8px; border-radius: 4px; transition: all 0.2s;"
        aria-label="Close"
      >
        ×
      </button>
    </div>
    <label for="link-type" style="display: block; color: rgba(255,255,255,0.8); font-size: 13px; margin-bottom: 8px; font-weight: 500;">Link Type</label>
    <select
      id="link-type"
      bind:value={linkFormState.newLinkType}
      style="width: 100%; padding: 12px; margin-bottom: 16px; border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; background: rgba(0,0,0,0.4); color: white; font-size: 14px; cursor: pointer; transition: border-color 0.2s;"
    >
      <option value="reference">📖 Reference</option>
      <option value="dependency">🔗 Dependency</option>
      <option value="related">🔀 Related</option>
      <option value="custom">✨ Custom</option>
    </select>
    <label for="link-strength" style="display: block; color: rgba(255,255,255,0.8); font-size: 13px; margin-bottom: 8px; font-weight: 500;">Link Strength: {linkFormState.newLinkWeight.toFixed(1)}</label>
    <input
      id="link-strength"
      type="range"
      min="0.1"
      max="1.0"
      step="0.1"
      bind:value={linkFormState.newLinkWeight}
      style="width: 100%; margin-bottom: 16px; accent-color: #fbbf24;"
    />
    <div style="display: flex; gap: 12px; justify-content: flex-end;">
      <button
        data-testid="link-form-cancel"
        onclick={() => onCancel('link')}
        style="padding: 10px 20px; border: 1px solid rgba(255,255,255,0.3); border-radius: 8px; background: transparent; color: white; cursor: pointer; font-size: 14px; transition: all 0.2s;"
      >
        Cancel
      </button>
      <button
        data-testid="link-form-create"
        onclick={() => onSave('link')}
        style="padding: 10px 20px; border: none; border-radius: 8px; background: linear-gradient(135deg, #fbbf24, #f59e0b); color: #000; cursor: pointer; font-size: 14px; font-weight: 500; transition: all 0.2s; box-shadow: 0 4px 12px rgba(251, 191, 36, 0.3);"
      >
        Create Link
      </button>
    </div>
  </div>
{/if}
