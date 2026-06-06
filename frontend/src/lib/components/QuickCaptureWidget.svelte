<script lang="ts">
  import { createNote } from '$lib/api/notes';

  // State
  let isOpen = $state(false);
  let content = $state('');
  let isSubmitting = $state(false);
  let showSuccess = $state(false);

  // Toggle widget
  function toggle() {
    isOpen = !isOpen;
    if (isOpen) {
      content = '';
      showSuccess = false;
    }
  }

  // Submit quick capture
  async function submitCapture() {
    if (!content.trim()) return;

    isSubmitting = true;
    try {
      const title = content.slice(0, 50) + (content.length > 50 ? '...' : '');
      await createNote({
        title,
        content,
        type: 'star',
        metadata: {
          tags: ['#inbox']
        }
      });

      showSuccess = true;
      content = '';
      setTimeout(() => {
        showSuccess = false;
        toggle();
      }, 1000);
    } catch (error) {
      console.error('Error creating note:', error);
    } finally {
      isSubmitting = false;
    }
  }

  // Handle keyboard shortcuts
  function handleKeydown(e: KeyboardEvent) {
    // Ctrl+Shift+N to open quick capture
    if (e.key === 'n' && e.ctrlKey && e.shiftKey && !isOpen) {
      e.preventDefault();
      toggle();
    }
    if (e.key === 'Escape' && isOpen) {
      toggle();
    }
    if (e.key === 'Enter' && (e.ctrlKey || e.metaKey) && isOpen) {
      e.preventDefault();
      submitCapture();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

<!-- Floating button -->
<div class="quick-capture-container">
  {#if isOpen}
    <div class="quick-capture-modal">
      <div class="modal-header">
        <h3>✨ Quick Capture</h3>
        <button class="close-btn" onmousedown={toggle}>×</button>
      </div>
      <div class="modal-body">
        <textarea
          bind:value={content}
          placeholder="Capture your thought... (Ctrl+Enter to submit)"
          disabled={isSubmitting}
        ></textarea>
        {#if showSuccess}
          <div class="success-message">✓ Saved!</div>
        {/if}
      </div>
      <div class="modal-footer">
        <button class="cancel-btn" onmousedown={toggle} disabled={isSubmitting}>Cancel</button>
        <button class="submit-btn" onmousedown={submitCapture} disabled={isSubmitting || !content.trim()}>
          {isSubmitting ? 'Saving...' : 'Save'}
        </button>
      </div>
    </div>
  {/if}

  <button class="quick-capture-btn" onmousedown={toggle} title="Quick Capture (Ctrl+Shift+N)">
    ✨
  </button>
</div>

<style>
  .quick-capture-container {
    position: fixed;
    bottom: 30px;
    right: 30px;
    z-index: 1000;
  }

  .quick-capture-btn {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border: none;
    color: white;
    font-size: 28px;
    cursor: pointer;
    box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
    transition: all 0.3s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .quick-capture-btn:hover {
    transform: scale(1.1);
    box-shadow: 0 6px 20px rgba(102, 126, 234, 0.6);
  }

  .quick-capture-modal {
    position: absolute;
    bottom: 80px;
    right: 0;
    width: 400px;
    background: white;
    border-radius: 16px;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
    overflow: hidden;
    animation: slideUp 0.3s ease;
  }

  @keyframes slideUp {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
  }

  .modal-header h3 {
    margin: 0;
    font-size: 18px;
    font-weight: 600;
  }

  .close-btn {
    background: none;
    border: none;
    color: white;
    font-size: 24px;
    cursor: pointer;
    padding: 0;
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 50%;
    transition: background 0.2s;
  }

  .close-btn:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .modal-body {
    padding: 20px;
  }

  .modal-body textarea {
    width: 100%;
    min-height: 120px;
    padding: 12px;
    border: 2px solid #e2e8f0;
    border-radius: 8px;
    font-size: 14px;
    font-family: inherit;
    resize: vertical;
    transition: border-color 0.2s;
  }

  .modal-body textarea:focus {
    outline: none;
    border-color: #667eea;
  }

  .modal-body textarea:disabled {
    background: #f7fafc;
  }

  .success-message {
    margin-top: 10px;
    color: #48bb78;
    font-weight: 600;
    text-align: center;
    animation: fadeIn 0.3s ease;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  .modal-footer {
    display: flex;
    gap: 10px;
    padding: 16px 20px;
    background: #f7fafc;
    border-top: 1px solid #e2e8f0;
  }

  .modal-footer button {
    flex: 1;
    padding: 10px 16px;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .cancel-btn {
    background: white;
    border: 2px solid #e2e8f0;
    color: #4a5568;
  }

  .cancel-btn:hover:not(:disabled) {
    background: #f7fafc;
    border-color: #cbd5e0;
  }

  .submit-btn {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border: none;
    color: white;
  }

  .submit-btn:hover:not(:disabled) {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  }

  .modal-footer button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
