<script lang="ts">
  const { 
    tags = [],
    onChange,
    maxTags = 10,
    placeholder = 'Add tag...'
  }: { 
    tags?: string[];
    onChange?: (tags: string[]) => void;
    maxTags?: number;
    placeholder?: string;
  } = $props();

  let inputValue = $state('');
  let isFocused = $state(false);
  let error = $state<string | null>(null);
  let localTags = $state<string[]>([]);

  // Initialize and sync localTags with props
  $effect.pre(() => {
    localTags = [...tags];
  });

  function validateTag(tag: string): string | null {
    const trimmed = tag.trim();
    
    if (!trimmed) {
      return 'Tag cannot be empty';
    }
    
    if (trimmed.length > 30) {
      return 'Tag must be less than 30 characters';
    }
    
    if (!/^[a-zA-Z0-9_-]+$/.test(trimmed)) {
      return 'Tag can only contain letters, numbers, underscores and hyphens';
    }
    
    if (localTags.includes(trimmed)) {
      return 'Tag already exists';
    }
    
    if (localTags.length >= maxTags) {
      return `Maximum ${maxTags} tags allowed`;
    }
    
    return null;
  }

  function addTag() {
    const trimmed = inputValue.trim();
    const validationError = validateTag(trimmed);
    
    if (validationError) {
      error = validationError;
      return;
    }
    
    const newTags = [...localTags, trimmed.toLowerCase()];
    localTags = newTags;
    onChange?.(newTags);
    inputValue = '';
    error = null;
  }

  function removeTag(tagToRemove: string) {
    const newTags = localTags.filter(tag => tag !== tagToRemove);
    localTags = newTags;
    onChange?.(newTags);
    error = null;
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === 'Enter') {
      event.preventDefault();
      addTag();
    } else if (event.key === 'Backspace' && !inputValue && localTags.length > 0) {
      // Удаляем последний тег при Backspace на пустом поле
      removeTag(localTags[localTags.length - 1]);
    } else {
      // Очищаем ошибку при вводе
      error = null;
    }
  }

  function handleBlur() {
    isFocused = false;
    // Не добавляем тег при blur, только очищаем если пустой
    if (!inputValue.trim()) {
      inputValue = '';
    }
  }
</script>

<div class="tag-selector" class:focused={isFocused}>
  <div class="tags-container">
    {#each localTags as tag (tag)}
      <span class="tag" data-testid="tag-{tag}">
        #{tag}
        <button
          type="button"
          class="remove-btn"
          onclick={() => removeTag(tag)}
          aria-label={`Remove tag ${tag}`}
          title={`Remove ${tag}`}
        >
          ×
        </button>
      </span>
    {/each}
    
    {#if localTags.length < maxTags}
      <input
        type="text"
        bind:value={inputValue}
        onkeydown={handleKeyDown}
        onfocus={() => isFocused = true}
        onblur={handleBlur}
        placeholder={localTags.length === 0 ? placeholder : ''}
        class="tag-input"
        aria-label="Add new tag"
        maxlength="30"
      />
    {/if}
  </div>
  
  {#if error}
    <div class="error-message" role="alert" aria-live="polite">
      {error}
    </div>
  {/if}
  
  <div class="hint">
    {localTags.length}/{maxTags} tags • Press Enter to add
  </div>
</div>

<style>
  .tag-selector {
    background: #f8fafc;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    padding: 8px 12px;
    transition: border-color 0.2s, box-shadow 0.2s;
  }

  .tag-selector.focused {
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  .tags-container {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    align-items: center;
    min-height: 36px;
  }

  .tag {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 10px;
    background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
    color: white;
    font-size: 13px;
    font-weight: 500;
    border-radius: 16px;
    transition: all 0.2s;
  }

  .tag:hover {
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(59, 130, 246, 0.3);
  }

  .remove-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 18px;
    height: 18px;
    padding: 0;
    margin: 0 -2px 0 2px;
    border: none;
    background: rgba(255, 255, 255, 0.2);
    color: white;
    font-size: 14px;
    font-weight: 600;
    border-radius: 50%;
    cursor: pointer;
    transition: background 0.15s;
    line-height: 1;
  }

  .remove-btn:hover {
    background: rgba(255, 255, 255, 0.3);
  }

  .tag-input {
    flex: 1;
    min-width: 100px;
    padding: 6px 0;
    border: none;
    background: transparent;
    font-size: 14px;
    color: #334155;
    outline: none;
  }

  .tag-input::placeholder {
    color: #94a3b8;
  }

  .error-message {
    margin-top: 8px;
    padding: 8px 12px;
    background: #fee2e2;
    color: #ef4444;
    font-size: 13px;
    border-radius: 6px;
  }

  .hint {
    margin-top: 8px;
    font-size: 12px;
    color: #94a3b8;
  }

  @media (max-width: 640px) {
    .tag-selector {
      padding: 6px 10px;
    }

    .tag {
      font-size: 12px;
      padding: 3px 8px;
    }

    .tag-input {
      font-size: 16px; /* Prevents zoom on iOS */
    }
  }
</style>
