<script lang="ts">
  import { Dialog } from './ui/dialog/index.js';
  import { createLink } from '$lib/api/links';
  import type { Note } from '$lib/api/notes';
  
  interface Props {
    open?: boolean;
    selectedNotes?: Note[];
    preselectedSource?: Note | null;
    onClose?: () => void;
    onSuccess?: () => void;
  }
  
  const { 
    open = false, 
    selectedNotes = [],
    preselectedSource = null,
    onClose,
    onSuccess
  }: Props = $props();
  
  // Form state
  let linkType = $state<'reference' | 'dependency' | 'related' | 'custom'>('reference');
  let weight = $state(0.5);
  let directionMode = $state<'full' | 'star'>('star');
  let customTypeName = $state('');
  let isSubmitting = $state(false);
  let progress = $state({ current: 0, total: 0 });
  
  // Compute source note (first selected or preselected)
  const sourceNote = $derived(preselectedSource || selectedNotes[0]);
  const targetNotes = $derived(preselectedSource 
    ? selectedNotes.filter(n => n.id !== preselectedSource.id)
    : selectedNotes.slice(1)
  );
  
  // Compute number of links to create
  const linksCount = $derived(() => {
    if (selectedNotes.length < 2) return 0;
    if (directionMode === 'star') {
      return targetNotes.length;
    } else {
      // Full graph: n * (n-1) / 2 for undirected
      const n = selectedNotes.length;
      return (n * (n - 1)) / 2;
    }
  });
  
  async function handleSubmit() {
    if (selectedNotes.length < 2) return;
    
    isSubmitting = true;
    const type = linkType === 'custom' ? customTypeName : linkType;
    const linksToCreate: { source: string; target: string; type: string; weight: number }[] = [];
    
    // Build links array
    if (directionMode === 'star') {
      // Star: source -> all others
      targetNotes.forEach(target => {
        linksToCreate.push({
          source: sourceNote.id,
          target: target.id,
          type,
          weight
        });
      });
    } else {
      // Full graph: every node with every other (undirected)
      for (let i = 0; i < selectedNotes.length; i++) {
        for (let j = i + 1; j < selectedNotes.length; j++) {
          linksToCreate.push({
            source: selectedNotes[i].id,
            target: selectedNotes[j].id,
            type,
            weight
          });
        }
      }
    }
    
    progress.total = linksToCreate.length;
    progress.current = 0;
    
    // Create links sequentially
    try {
      for (const link of linksToCreate) {
        await createLink(link.source, link.target, link.type, link.weight);
        progress.current++;
      }
      
      onSuccess?.();
      resetForm();
    } catch (e) {
      console.error('Failed to create links:', e);
      alert('Ошибка при создании связей');
    } finally {
      isSubmitting = false;
    }
  }
  
  function resetForm() {
    linkType = 'reference';
    weight = 0.5;
    directionMode = 'star';
    customTypeName = '';
    progress = { current: 0, total: 0 };
  }
  
  function handleCancel() {
    resetForm();
    onClose?.();
  }
</script>

<Dialog {open} onClose={handleCancel} className="link-creation-modal">
  <div class="modal-content">
    <h2 class="modal-title">Создать связи</h2>
    
    <div class="selected-info">
      <span class="info-label">Выбрано заметок:</span>
      <span class="info-value">{selectedNotes.length}</span>
    </div>
    
    {#if sourceNote}
      <div class="source-note">
        <span class="info-label">Источник:</span>
        <span class="note-title">{sourceNote.title}</span>
      </div>
    {/if}
    
    <!-- Link Type -->
    <div class="form-group">
      <label for="link-type">Тип связи</label>
      <select id="link-type" bind:value={linkType} disabled={isSubmitting}>
        <option value="reference">Ссылка (reference)</option>
        <option value="dependency">Зависимость (dependency)</option>
        <option value="related">Связано (related)</option>
        <option value="custom">Пользовательский...</option>
      </select>
    </div>
    
    {#if linkType === 'custom'}
      <div class="form-group">
        <label for="custom-type">Название типа</label>
        <input
          id="custom-type"
          type="text"
          bind:value={customTypeName}
          placeholder="Введите название..."
          disabled={isSubmitting}
        />
      </div>
    {/if}
    
    <!-- Weight -->
    <div class="form-group">
      <label for="weight">
        Вес связи: <span class="weight-value">{weight.toFixed(1)}</span>
      </label>
      <input
        id="weight"
        type="range"
        min="0"
        max="1"
        step="0.1"
        bind:value={weight}
        disabled={isSubmitting}
        class="weight-slider"
      />
      <div class="weight-labels">
        <span>Слабая</span>
        <span>Сильная</span>
      </div>
    </div>
    
    <!-- Direction Mode -->
    <div class="form-group">
      <label>Режим связывания</label>
      <div class="radio-group">
        <label class="radio-label">
          <input
            type="radio"
            name="direction"
            value="star"
            bind:group={directionMode}
            disabled={isSubmitting}
          />
          <span class="radio-text">
            Звезда: {sourceNote?.title || 'первая'} → остальные ({targetNotes.length} связей)
          </span>
        </label>
        <label class="radio-label">
          <input
            type="radio"
            name="direction"
            value="full"
            bind:group={directionMode}
            disabled={isSubmitting}
          />
          <span class="radio-text">
            Полный граф: все со всеми ({linksCount()} связей)
          </span>
        </label>
      </div>
    </div>
    
    <!-- Progress -->
    {#if isSubmitting && progress.total > 0}
      <div class="progress-bar">
        <div class="progress-fill" style="width: {(progress.current / progress.total) * 100}%"></div>
        <span class="progress-text">{progress.current} / {progress.total}</span>
      </div>
    {/if}
    
    <!-- Actions -->
    <div class="modal-actions">
      <button class="btn-cancel" onclick={handleCancel} disabled={isSubmitting}>
        Отмена
      </button>
      <button 
        class="btn-submit" 
        onclick={handleSubmit} 
        disabled={isSubmitting || selectedNotes.length < 2 || (linkType === 'custom' && !customTypeName)}
      >
        {#if isSubmitting}
          Создание...
        {:else}
          Создать {linksCount()} связей
        {/if}
      </button>
    </div>
  </div>
</Dialog>

<style>
  .modal-content {
    padding: 1.5rem;
    min-width: 400px;
    max-width: 500px;
  }
  
  .modal-title {
    font-size: 1.25rem;
    font-weight: 600;
    color: white;
    margin: 0 0 1.25rem;
  }
  
  .selected-info,
  .source-note {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
    padding: 0.75rem;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 0.5rem;
  }
  
  .info-label {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.6);
  }
  
  .info-value {
    font-size: 1rem;
    font-weight: 600;
    color: white;
  }
  
  .note-title {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.9);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  
  .form-group {
    margin-bottom: 1.25rem;
  }
  
  .form-group label {
    display: block;
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.8);
    margin-bottom: 0.5rem;
  }
  
  select,
  input[type="text"] {
    width: 100%;
    padding: 0.625rem 0.875rem;
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 0.5rem;
    color: white;
    font-size: 0.9375rem;
  }
  
  select:focus,
  input[type="text"]:focus {
    outline: none;
    border-color: rgba(99, 102, 241, 0.8);
  }
  
  .weight-value {
    color: rgba(99, 102, 241, 1);
    font-weight: 600;
  }
  
  .weight-slider {
    width: 100%;
    height: 6px;
    margin: 0.75rem 0;
    -webkit-appearance: none;
    appearance: none;
    background: rgba(255, 255, 255, 0.2);
    border-radius: 3px;
    outline: none;
  }
  
  .weight-slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 18px;
    height: 18px;
    background: rgba(99, 102, 241, 1);
    border-radius: 50%;
    cursor: pointer;
  }
  
  .weight-slider::-moz-range-thumb {
    width: 18px;
    height: 18px;
    background: rgba(99, 102, 241, 1);
    border-radius: 50%;
    cursor: pointer;
    border: none;
  }
  
  .weight-labels {
    display: flex;
    justify-content: space-between;
    font-size: 0.75rem;
    color: rgba(255, 255, 255, 0.5);
  }
  
  .radio-group {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  
  .radio-label {
    display: flex;
    align-items: flex-start;
    gap: 0.625rem;
    cursor: pointer;
    padding: 0.625rem;
    border-radius: 0.5rem;
    transition: background 0.15s;
  }
  
  .radio-label:hover {
    background: rgba(255, 255, 255, 0.05);
  }
  
  .radio-label input[type="radio"] {
    margin-top: 2px;
    width: auto;
  }
  
  .radio-text {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.85);
    line-height: 1.4;
  }
  
  .progress-bar {
    height: 8px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
    overflow: hidden;
    margin-bottom: 1rem;
    position: relative;
  }
  
  .progress-fill {
    height: 100%;
    background: rgba(99, 102, 241, 0.8);
    transition: width 0.3s ease;
  }
  
  .progress-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 0.75rem;
    color: white;
    font-weight: 500;
  }
  
  .modal-actions {
    display: flex;
    gap: 0.75rem;
    justify-content: flex-end;
    margin-top: 1.5rem;
    padding-top: 1rem;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
  }
  
  .btn-cancel,
  .btn-submit {
    padding: 0.625rem 1.25rem;
    border-radius: 0.5rem;
    font-size: 0.9375rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }
  
  .btn-cancel {
    background: rgba(255, 255, 255, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.2);
    color: rgba(255, 255, 255, 0.8);
  }
  
  .btn-cancel:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.2);
  }
  
  .btn-submit {
    background: rgba(99, 102, 241, 0.8);
    border: none;
    color: white;
  }
  
  .btn-submit:hover:not(:disabled) {
    background: rgba(99, 102, 241, 1);
    transform: translateY(-1px);
  }
  
  .btn-submit:disabled,
  .btn-cancel:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  
  @media (max-width: 640px) {
    .modal-content {
      min-width: auto;
      max-width: 100%;
      padding: 1rem;
    }
    
    .modal-actions {
      flex-direction: column;
    }
    
    .btn-cancel,
    .btn-submit {
      width: 100%;
    }
  }
  
  :global(.link-creation-modal) {
    max-width: 500px;
  }
</style>
